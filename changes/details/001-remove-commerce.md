# 移除商业模块及用户信誉分、积分和排名

**日期**：2026-10-27

本项目不再提供商店、商品、优惠券、购物车、订单、支付、计划、定价、订阅、交易和账户余额功能。注册与登录不再依赖套餐或有效订阅；账号禁用、应用访问限制、角色权限和 MFA 等认证授权规则继续生效。

用户信誉分、积分、排名和组织初始积分也已移除，涵盖当前前端、接口数据、JWT、注册与同步流程、默认配置和新数据库结构。已有数据库通过独立清理工具迁移，具体说明见下方 2026-10-27 补充记录。

## 涉及文件

- `web/src/App.tsx`、`web/src/lib/nav.ts`、`web/src/components/common/NavItemTree.tsx` — 移除商业菜单、管理页面和公开购买、定价、支付结果路由。
- `web/src/pages/`、`web/src/backend/`、`web/src/components/`、`web/src/lib/` — 删除专用商业页面、接口调用和组件，清理用户组织余额、支付提供商配置、首页统计及认证购买跳转。
- `web/public/AuthCallbackHandler.js`、`web/README.md`、`web/cypress/e2e/` — 清除静态登录回调中的购买分支、更新前端功能说明，并删除专用商业页面的旧端到端测试。
- `controllers/`、`routers/`、`form/`、`authz/authz.go` — 删除商业 API、支付回调、商业访问策略、注册套餐字段和登录订阅检查。
- `object/` — 删除商业对象及建表、初始化、导出、统计逻辑，清理用户组织和提供商中的专用商业字段。
- `pp/` — 删除支付提供商实现；身份提供商中的支付宝登录保留。
- `go.mod`、`go.sum` — 清理不再使用的支付依赖。
- `i18n/`、`web/src/locales/` — 清理专用商业文案，保留其他功能仍在使用的共享文案。
- `init_data.json.template`、`swagger/swagger.json`、`swagger/swagger.yml` — 清理初始化商业数据、已删除接口及数据模型文档。
- `README.md` — 标明本地裁剪范围及旧数据库迁移入口。
- `cmd/remove-commerce/` — 为已有数据库提供显式执行的商业数据清理工具和说明。

## 数据与兼容性

新数据库不再创建商业表和商业字段。已有数据库使用清理工具预览并执行迁移，移除商业数据并将历史 `paid-user` 转为 `normal-user`。删除数据不可逆，需停服并备份后执行；常规服务启动不会自动执行破坏性迁移。

`web-old` 按项目约定保持只读，其中的历史商业代码不属于当前前端的构建入口。后续功能迁移和上游同步不得重新引入已移除的商业功能。

## 商业模块移除时的验证（2026-10-14）

- `go mod tidy` 和 `go build ./...` 通过。
- 前端 `tsc --noEmit`、`yarn lint`、`yarn build`（包括 postbuild）与静态回调脚本的 `node --check` 通过。
- 独立 SQLite 内存数据库初始化及数据导出通过，使用独立表名前缀，未连接业务数据库。
- 隔离服务实测：管理员、普通组织的普通用户及历史 `paid-user` 均可完成密码登录和会话查询；被禁用账号仍拒绝登录；账号响应不包含余额、币种和购物车字段。
- 已认证管理员访问商品、支付、订阅、交易查询及购买、支付回调、订单支付、优惠券校验接口，均返回 HTTP 404。
- 现存 `object` 测试编译通过，未运行依赖业务数据库的测试；未添加 Go 测试。
- 清理工具通过独立 SQLite 数据库验证：只读预览、实际删除、重复执行零操作、非法 JSON 拒绝写入、失败事务回滚、表名前缀和共享数据保留。MySQL/PostgreSQL 尚未完成连接实测。
- 迁移配置按运行模式读取 INI 分区，并通过四个临时 SQLite 数据库的十种配置组合验证，包括环境变量覆盖、空值回落和前缀选择；未选中的默认区数据库保持不变。
- 使用当前 `conf/app.conf` 尝试只读迁移预览，数据库连接失败；未执行实际数据库删表或删字段。
- 未对实际业务数据库执行破坏性迁移；外部身份提供商的 SSO 和 MFA 完整流程仍需在对应环境回归。

## 2026-10-27：移除信誉分、积分和排名

### 涉及文件

- `web/src/pages/UserEditPage.tsx`、`web/src/pages/OrganizationEditPage.tsx` — 删除信誉分、积分、排名和初始积分编辑入口。
- `web/src/pages/defaults.ts`、`web/src/pages/organization-defaults.ts`、`web/src/lib/setting.tsx` — 清理新建用户、组织的默认值、账号配置项、用户字段及同步和令牌字段选项。
- `web/src/components/user/AffiliationSelect.tsx`、`web/src/pages/auth/PromptPage.tsx` — 所属单位选择和地址变更不再借用积分字段传递或清空单位 ID；用户编辑页同步调整调用。
- `web/src/lib/webhook-preview.ts`、`web/src/locales/*/data.json`、`i18n/locales/*/data.json` — 清理示例用户中的积分和排名，以及相关标签、提示和初始积分错误文案。
- `object/user.go`、`object/user_util.go`、`object/organization.go` — 删除用户和组织模型字段、初始积分读取、注册排名赋值及修改权限逻辑；拒绝旧字段的直接写入和指定列更新。
- `object/user_metrics_compat.go` — 过滤旧组织配置中的信誉分、积分、排名账号项及默认令牌字段，保留其他配置和自定义属性字段。
- `controllers/account.go`、`controllers/auth.go`、`controllers/magic_link.go`、`object/token_oauth_util.go`、`object/token_verification_code.go`、`object/ldap_conn.go` — 注册、第三方首次登录、魔术链接、访客、验证码和 LDAP 创建用户时不再初始化积分。
- `object/token_jwt.go`、`object/wellknown_oidc_discovery.go`、`object/application_util.go` — 清理 JWT 用户字段、OIDC 声明列表及应用返回组织信息中的初始积分处理。
- `object/syncer_util.go`、`object/syncer_user.go` — 删除旧字段的同步映射和输出，阻止仍引用旧字段的同步器连接数据库，避免向外部系统写入空值。
- `object/syncer_affiliation.go`、`object/syncer_public_api.go`、`object/syncer_sync.go` — 所属单位通过 `Affiliation` 映射 ID 与名称，同步哈希按名称比较；反向同步遇到未知或重名单位时返回错误。
- `object/init.go`、`init_data.json.template`、`conf/app.conf`、`manifests/casdoor/values.yaml` — 删除内置用户和组织的相关字段、账号项及全局初始积分配置。
- `swagger/swagger.json`、`swagger/swagger.yml` — 删除接口模型中的 `score`、`karma`、`ranking` 和 `initScore`。
- `cmd/remove-user-metrics/main.go`、`cmd/remove-user-metrics/README.md`、`README.md` — 新增已有数据库的只读预览和显式清理工具，说明移除范围及迁移入口。
- `cmd/remove-user-metrics/main_test.go`、`object/user_metrics_compat_test.go` — 新增隔离数据库迁移、模型与响应字段、旧配置过滤、旧字段更新拦截及所属单位同步测试。

### 数据与兼容性

新数据库不再创建用户表的 `score`、`karma`、`ranking` 和组织表的 `init_score`。已有数据库的物理列不会在服务启动时自动删除；使用 `cmd/remove-user-metrics` 先预览，再通过 `-apply` 显式执行清理。迁移同时移除组织账号项、默认令牌字段、应用令牌字段和同步列映射中的相关配置，并停用受影响的同步器，待核对主键、哈希及单位映射后重新启用。

此前借用 `Score` 保存单位 ID 的同步器，应将外部单位 ID 列映射到 `Affiliation` 并保留所属单位表配置。其他用户数据、所属单位名称、无关表、自定义属性及人脸识别置信度保留；历史审计记录和已签发 JWT 不改写。`web-old` 继续作为已停用的参考代码保留。

清理工具支持 PostgreSQL、MySQL 和 SQLite 3.35+，遵循服务配置、环境变量、表名前缀和 PostgreSQL schema 规则。迁移可重复执行，非法 JSON 会阻止执行；PostgreSQL 和 SQLite 使用事务，MySQL 的 DDL 隐式提交，失败后可能需要修复依赖再重试。

### 本次验证

- `go build ./...` 通过；变更 Go 文件的 `gofumpt` 检查通过。
- 前端 `npm run typecheck`、`npm run lint`、`npm run build`（包括构建后处理）通过。
- `go test ./cmd/remove-user-metrics ./object -run 'TestCleanup|TestRetiredUserMetrics|TestAffiliationSyncWithoutUserMetrics' -count=1` 通过，使用 SQLite 内存数据库。
- 验证了新表和接口/JWT 用户数据不含旧字段、旧配置过滤、更新拦截、所属单位双向映射及哈希一致性；迁移覆盖只读预览、实际删除、重复执行、表名前缀、非法 JSON、失败回滚及无关数据保留。
- 24 个变更 JSON 文件解析通过，`git diff --check` 通过。
- 未对实际业务数据库执行删列；MySQL/PostgreSQL 迁移仍需在对应部署的备份库验证。
