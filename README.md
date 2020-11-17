# mongo-diff

Mongo Diff 是一个命令行工具，用于记录 MongoDB 数据库系统变量、用户、数据库的变更，生成差异报告

```bash
Usage:
  -context-line uint
        diff 上下文信息数量 (default 2)
  -data-dir string
        diff 状态数据存储目录 (default "./tmp")
  -keep-version uint
        保留多少个版本的历史记录 (default 100)
  -mongo-uri string
        MongoDB URI，参考文档 https://docs.mongodb.com/manual/reference/connection-string/ (default "mongodb://localhost:27017")
  -name string
        Diff 名称 (default "mongodb")
  -no-diff
        只输出基本信息，不执行 diff
```
