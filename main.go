package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mylxsw/go-utils/diff"
	"github.com/mylxsw/go-utils/file"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoURI, diffName string
var dataDir string
var contextLine, keepVersion uint
var noDiff bool

func main() {
	flag.StringVar(&mongoURI, "mongo-uri", "mongodb://localhost:27017", "MongoDB URI，参考文档 https://docs.mongodb.com/manual/reference/connection-string/")
	flag.StringVar(&dataDir, "data-dir", "./tmp", "diff 状态数据存储目录")
	flag.UintVar(&contextLine, "context-line", 2, "diff 上下文信息数量")
	flag.UintVar(&keepVersion, "keep-version", 100, "保留多少个版本的历史记录")
	flag.BoolVar(&noDiff, "no-diff", false, "只输出基本信息，不执行 diff")
	flag.StringVar(&diffName, "name", "mongodb", "Diff 名称")

	flag.Parse()

	if noDiff {
		if err := mongoInfo(mongoURI, os.Stdout); err != nil {
			panic(err)
		}

		return
	}

	buffer := bytes.NewBuffer(nil)
	if err := mongoInfo(mongoURI, buffer); err != nil {
		panic(err)
	}

	fs := file.LocalFS{}
	if err := fs.MkDir(dataDir); err != nil {
		panic(err)
	}

	differ := diff.NewDiffer(fs, dataDir, int(contextLine))
	latest := differ.DiffLatest(diffName, buffer.String())
	if err := latest.PrintAndSave(os.Stdout); err != nil {
		panic(err)
	}

	_ = latest.Clean(keepVersion)
}

func mongoInfo(mongoURI string, out io.Writer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOption := options.Client().ApplyURI(mongoURI)
	connect, err := mongo.Connect(ctx, clientOption)
	if err != nil {
		return err
	}
	defer connect.Disconnect(context.TODO())

	mm := NewMongoManager(connect)
	databaseNames, err := mm.AllDatabaseNames(ctx)
	if err != nil {
		return err
	}
	for _, name := range databaseNames {
		_, _ = fmt.Fprintf(out, "DB: %s\n", name)
	}

	users, err := mm.AllUsers(ctx)
	if err != nil {
		return err
	}
	for _, user := range users {
		_, _ = fmt.Fprintf(out, "USER: db=%s, user=%s\n", user.DB, user.User)
		for _, role := range user.Roles {
			_, _ = fmt.Fprintf(out, "USER_ROLE: db=%s, user=%s, role=%s/%s\n", user.DB, user.User, role.DB, role.Role)
		}
	}

	conf, err := mm.Config(ctx)
	if err != nil {
		return err
	}
	for _, setting := range conf.Members {
		_, _ = fmt.Fprintf(out, "SETTING: id=%d, host=%s, vote=%d, arbiterOnly=%v, buildIndexes=%v, hidden=%v, priority=%d\n", setting.ID, setting.Host, setting.Votes, setting.ArbiterOnly, setting.BuildIndexes, setting.Hidden, setting.Priority)
	}

	replStat, err := mm.ReplStatus(ctx)
	if err != nil {
		return err
	}
	for _, stat := range replStat.Members {
		_, _ = fmt.Fprintf(out, "REPL_STAT: id=%d, name=%s, state=%s, health=%d, syncSourceHost=%s, syncingTo=%s\n", stat.ID, stat.Name, stat.StateStr, stat.Health, stat.SyncSourceHost, stat.SyncingTo)
	}

	return nil
}

type MongoManager struct {
	conn *mongo.Client
}

func NewMongoManager(conn *mongo.Client) *MongoManager {
	return &MongoManager{conn: conn}
}

func (mm *MongoManager) AllDatabaseNames(ctx context.Context) ([]string, error) {
	return mm.conn.ListDatabaseNames(ctx, bson.M{})
}

func (mm *MongoManager) AllUsers(ctx context.Context) ([]User, error) {
	var users UsersResp
	if err := mm.conn.Database("admin").RunCommand(ctx, bson.M{"usersInfo": bson.M{"forAllDBs": true}}).Decode(&users); err != nil {
		return nil, err
	}

	return users.Users, nil
}

func (mm *MongoManager) Config(ctx context.Context) (ReplSetConfig, error) {
	var replConf ReplSetConfigResp
	if err := mm.conn.Database("admin").RunCommand(ctx, bson.M{"replSetGetConfig": 1}).Decode(&replConf); err != nil {
		return ReplSetConfig{}, err
	}

	return replConf.Config, nil
}

func (mm *MongoManager) ReplStatus(ctx context.Context) (ReplSetStatus, error) {
	var replSetStatus ReplSetStatus
	if err := mm.conn.Database("admin").RunCommand(ctx, bson.M{"replSetGetStatus": 1}).Decode(&replSetStatus); err != nil {
		return ReplSetStatus{}, err
	}

	return replSetStatus, nil
}

type UsersResp struct {
	Users []User `bson:"users" json:"users"`
}

type User struct {
	ID         string   `bson:"_id" json:"id"`
	DB         string   `bson:"db" json:"db"`
	Mechanisms []string `bson:"mechanisms" json:"mechanisms"`
	Roles      []Role   `bson:"roles" json:"roles"`
	User       string   `bson:"user" json:"user"`
}

type Role struct {
	DB   string `bson:"db" json:"db"`
	Role string `bson:"role" json:"role"`
}

type ReplSetConfig struct {
	ID              string                `bson:"_id" json:"id"`
	Members         []ReplSetMemberConfig `bson:"members" json:"members"`
	ProtocolVersion int                   `bson:"protocolVersion" json:"protocol_version"`
}

type ReplSetMemberConfig struct {
	ID           int    `bson:"_id" json:"id"`
	ArbiterOnly  bool   `bson:"arbiterOnly" json:"arbiter_only"`
	BuildIndexes bool   `bson:"buildIndexes" json:"build_indexes"`
	Hidden       bool   `bson:"hidden" json:"hidden"`
	Host         string `bson:"host" json:"host"`
	Priority     int    `bson:"priority" json:"priority"`
	SlaveDelay   int    `bson:"slaveDelay" json:"slave_delay"`
	Votes        int    `bson:"votes" json:"votes"`
}

type ReplSetConfigResp struct {
	Config ReplSetConfig `json:"config" bson:"config"`
}

type ReplSetStatus struct {
	Members                 []ReplMember `bson:"members" json:"members"`
	MyState                 int          `bson:"myState" json:"my_state"`
	OK                      int          `bson:"ok" json:"ok"`
	Set                     string       `bson:"set" json:"set"`
	Term                    int          `bson:"term" json:"term"`
	SyncSourceHost          string       `bson:"syncSourceHost" json:"sync_source_host"`
	SyncSourceID            int          `bson:"syncSourceId" json:"sync_source_id"`
	SyncingTo               string       `bson:"syncingTo" json:"syncing_to"`
	HeartbeatIntervalMillis int          `bson:"heartbeatIntervalMillis" json:"heartbeat_interval_millis"`
	Date                    time.Time    `bson:"date" json:"date"`
}

type ReplMember struct {
	ID                   int       `bson:"_id" json:"id"`
	ConfigVersion        int       `bson:"configVersion" json:"config_version"`
	InfoMessage          string    `bson:"infoMessage" json:"info_message"`
	LastHeartbeat        time.Time `bson:"lastHeartbeat" json:"last_heartbeat"`
	LastHeartbeatMessage string    `bson:"lastHeartbeatMessage" json:"last_heartbeat_message"`
	LastHeartbeatRecv    time.Time `bson:"lastHeartbeatRecv" json:"last_heartbeat_recv"`
	Name                 string    `bson:"name" json:"name"`
	State                int       `bson:"state" json:"state"`
	StateStr             string    `bson:"stateStr" json:"state_str"`
	SyncSourceHost       string    `bson:"syncSourceHost" json:"sync_source_host"`
	SyncSourceID         int       `bson:"syncSourceId" json:"sync_source_id"`
	SyncingTo            string    `bson:"syncingTo" json:"syncing_to"`
	Uptime               int       `bson:"uptime" json:"uptime"`
	ElectionDate         time.Time `bson:"electionDate" json:"election_date"`
	Health               int       `bson:"health" json:"health"`
	PingMS               int       `bson:"pingMs" json:"ping_ms"`
}

func NoError(err error) {
	if err != nil {
		panic(err)
	}
}
