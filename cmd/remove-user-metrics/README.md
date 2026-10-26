# 清理积分、信誉分和排名

当前代码不再创建、读写或输出用户的 `score`、`karma`、`ranking` 和组织的 `init_score`。常规服务启动不会删除旧数据库中的列。此工具用于清理已有数据库，默认只读预览。

在仓库根目录执行：

```powershell
go run ./cmd/remove-user-metrics -config conf/app.conf
```

确认预览中的目标表、组织和同步器后，停服、备份数据库，再执行：

```powershell
go run ./cmd/remove-user-metrics -config conf/app.conf -apply
```

使用服务自身的配置加载、环境变量、`.env`、表名前缀和 PostgreSQL schema 规则。连接信息不在预览中输出。支持 PostgreSQL、MySQL 和 SQLite 3.35+；其他数据库需由管理员执行对应迁移。

清理范围：

- 用户表：删除 `score`、`karma`、`ranking`。
- 组织表：删除 `init_score`，移除 `account_items` 和 `default_token_fields` 中的相关项目。
- 应用表：移除 `token_fields` 中的相关字段。
- 同步器：移除相关 `table_columns` 映射，并停用受影响的同步器。调整映射、核对主键和哈希字段后再启用。如果此前借用 `Score` 保存单位 ID，应将外部单位 ID 列映射到 `Affiliation`，并保留所属单位表配置；同步时会在 ID 与名称之间转换。

其他用户数据、所属单位名称、自定义属性（包括 `Properties.score`）、无关表和未受影响的同步器保持原样。人脸识别置信度不属于本次移除范围。历史审计记录与已签发的 JWT 不会被改写；新签发的内置令牌不再含这些用户字段。

迁移可重复执行，已清理的列和配置会跳过。工具先检查全部 JSON，再执行计划；非法 JSON 会阻止迁移。PostgreSQL 和 SQLite 在事务中执行；MySQL 的 DDL 隐式提交，失败后可能需要修复依赖索引等问题并重新运行。工具不会级联删除自定义索引、视图或其他依赖。

隔离验证：

```powershell
go test ./cmd/remove-user-metrics
go test ./object -run 'TestRetiredUserMetrics|TestAffiliationSyncWithoutUserMetrics'
```

测试使用 SQLite 内存数据库，不连接业务库；覆盖预览、实际清理、重复执行、表名前缀、非法 JSON、失败回滚、无关数据保留及同步映射。MySQL/PostgreSQL 仍需在对应部署的备份库验证。
