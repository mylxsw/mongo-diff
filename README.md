# mongo-diff

Mongo Diff 是一个命令行工具，用于记录 MongoDB 数据库系统变量、用户、数据库的变更，生成差异报告

```bash
Usage:
  -context-line int
    	diff 上下文信息数量 (default 2)
  -data-dir string
    	diff 状态数据存储目录 (default "./tmp")
  -db-host string
    	MySQL Host (default "127.0.0.1")
  -db-password string
    	MySQL Password
  -db-port int
    	MySQL Port (default 3306)
  -db-user string
    	MySQL User (default "root")
  -diff-databases
    	是否对比数据库差异 (default true)
  -diff-users
    	是否对比用户差异 (default true)
  -diff-vars
    	是否对比系统变量差异 (default true)
  -exclude-dbs string
    	需要排除的系统变量 (default "performance_schema,information_schema,mysql,sys")
  -exclude-vars string
    	需要排除的系统变量 (default "gtid_binlog_pos,gtid_binlog_state,gtid_current_pos")
  -with-tables
    	对比数据库差异时，是否启用表名差异对比 (default true)
```
