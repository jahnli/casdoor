# 移除商业模块、用户信誉分、积分和排名及旧前端

**日期**：2026-11-12

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

商业模块移除时，`web-old` 按项目约定保持只读，其中的历史商业代码不属于当前前端的构建入口；该目录已于 2026-11-01 删除，详见下方补充记录。后续功能迁移和上游同步不得重新引入已移除的商业功能。

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

此前借用 `Score` 保存单位 ID 的同步器，应将外部单位 ID 列映射到 `Affiliation` 并保留所属单位表配置。其他用户数据、所属单位名称、无关表、自定义属性及人脸识别置信度保留；历史审计记录和已签发 JWT 不改写。此次裁剪时，`web-old` 仍作为已停用的参考代码保留，后于 2026-11-01 删除。

清理工具支持 PostgreSQL、MySQL 和 SQLite 3.35+，遵循服务配置、环境变量、表名前缀和 PostgreSQL schema 规则。迁移可重复执行，非法 JSON 会阻止执行；PostgreSQL 和 SQLite 使用事务，MySQL 的 DDL 隐式提交，失败后可能需要修复依赖再重试。

### 本次验证

- `go build ./...` 通过；变更 Go 文件的 `gofumpt` 检查通过。
- 前端 `npm run typecheck`、`npm run lint`、`npm run build`（包括构建后处理）通过。
- `go test ./cmd/remove-user-metrics ./object -run 'TestCleanup|TestRetiredUserMetrics|TestAffiliationSyncWithoutUserMetrics' -count=1` 通过，使用 SQLite 内存数据库。
- 验证了新表和接口/JWT 用户数据不含旧字段、旧配置过滤、更新拦截、所属单位双向映射及哈希一致性；迁移覆盖只读预览、实际删除、重复执行、表名前缀、非法 JSON、失败回滚及无关数据保留。
- 24 个变更 JSON 文件解析通过，`git diff --check` 通过。
- 未对实际业务数据库执行删列；MySQL/PostgreSQL 迁移仍需在对应部署的备份库验证。

## 2026-11-01：移除已停用的旧前端

删除 Ant Design 旧前端 `web-old`，仅保留基于 shadcn/Tailwind 的 `web` 作为前端。原先保留旧前端作为只读参考的约定终止；现有前端已迁移的校验、字段格式和条件逻辑继续保留。

### 涉及文件

- `web-old/` — 删除全部 356 个跟踪文件，包括页面、组件、接口调用、多语言资源、静态回调脚本、测试、构建配置、依赖清单和锁文件。
- `.gitignore` — 删除旧前端专用的 Cypress 截图和视频忽略规则。
- `README.md`、`web/README.md` — 移除旧前端保留说明及失效路径，将组件映射标明为历史迁移说明。
- `web/cypress/support/commands.js`、`web/src/components/auth/mfa/MfaItemsTable.tsx` — 清理指向旧前端文件或组件的注释，运行逻辑不变。
- `CLAUDE.md`、`.agents/skills/upstream-merge/SKILL.md` — 明确只维护 `web`，禁止恢复旧前端，继续保留不迁移 AI 助手和 Web3 钱包登录的约束。

### 行为与验证

- 旧前端此前已不参与构建或提供页面服务，本次删除不新增用户可见的界面变化，也不涉及数据库迁移。
- 当前 `web` 的构建后处理脚本及后端仍在引用的静态登录回调脚本保留。
- 未发现构建或部署配置依赖 `web-old`；目录残留检查无输出，`git diff --check HEAD` 通过。
- 本次未运行完整构建或业务回归；此前记录的构建和业务验证结果不代表本次已重新执行。

## 2026-11-04：用户独立属性与资料字段排序

用户编辑页增加 Open ID、部门名称、工号、工作日期、职位类型、职级的独立输入项，继续存储在用户属性中。组织个人页面设置可分别配置这些字段的可见性、查看权限、修改权限及正则规则；独立属性项的权限优先于整体属性配置，属性表隐藏已配置独立入口的键，编辑其他属性时保留这些键。

用户编辑页与组织个人页面设置共用字段顺序：头像、组织、显示名称、账号名称、姓氏、名字、邮箱、手机号；随后为群组、工作单位、部门名称、工号、职位类型、职级、工作日期及补充个人资料。账号类型、标签、权限与密码设置位于 UID 号前；UID 号后依次展示 IP 白名单、最近登录时间、最近登录 IP、被删除、删除时间、属性、国家/地区、城市、地址、地址列表与授权同意，后接认证凭据及实名认证。双列表单连续排列，不为字段配对强制留空；原有字段显示条件、编辑权限、正则规则和标签页配置保留。组织配置中的旧 Company 项不再展示或保存，工作单位继续使用 Affiliation 字段。

### 涉及文件

- `object/user.go` — 按独立属性项的查看规则过滤用户属性，避免整体属性规则覆盖独立字段权限。
- `object/user_util.go` — 为独立属性键匹配修改权限与正则规则，逐键恢复无权修改的值。
- `web/src/components/crud/MapTable.tsx` — 支持隐藏独立字段对应的属性键，保留隐藏值。
- `web/src/lib/account-items.ts` — 定义独立属性字段、存储键、输入类型及两个页面共用的字段顺序。
- `web/src/lib/setting.tsx` — 将独立属性字段加入可选账号项。
- `web/src/pages/UserEditPage.tsx` — 展示独立属性字段，按共用顺序连续排列资料、账号状态和认证信息。
- `web/src/pages/OrganizationEditPage.tsx` — 支持独立属性字段的配置与正则规则，个人页面设置列表与用户编辑页排序一致，过滤旧 Company 项。
- `web/src/locales/en/data.json`、`web/src/locales/zh/data.json` — 增加独立属性字段的中英文文案。

### 个人封面与账号字段文案补充

用户编辑页及组织个人页面设置增加“个人封面（Profile Cover）”独立属性项，位于个人主页之后、自我介绍之前，以 URL 输入框编辑并保存到 `properties.coverUrl`，不新增数据库列。沿用独立属性字段的查看权限、修改权限与正则校验；新建组织默认包含该项，查看权限为公开、修改权限为本人，已有组织需在个人页面设置中添加并启用。初始实现仅提供地址配置与保存；后续用户列表悬停资料卡片使用该地址渲染封面，见下方补充。

共享中文文案将 UID 号、IP 白名单、WebAuthn 凭据、MFA 账户、MFA 设置项、API 密钥的英文与中文之间补上空格，组织设置与用户界面同步显示。

组织个人页面设置在新增配置行选定字段或更换字段名称后，立即按共享字段顺序重新排列，避免个人封面停留在列表末尾；个人封面紧跟个人主页，其余配置值保留。

- `web/src/lib/account-items.ts` — 新增个人封面的属性映射、URL 输入类型和共享排序。
- `object/user_util.go` — 将个人封面映射到 `coverUrl`，接入现有属性权限及正则校验。
- `web/src/pages/organization-defaults.ts` — 新建组织默认在个人主页后配置个人封面。
- `web/src/pages/OrganizationEditPage.tsx` — 选定或更换账号字段时即时排序，使个人封面自动归位到个人主页之后。
- `web/src/locales/en/data.json`、`web/src/locales/zh/data.json` — 增加封面名称及 URL 提示，规范六项账号字段的中英文间距。

## 2026-11-04：用户列表与悬停资料卡片

用户列表将头像、显示名称和账号名称合并为左侧冻结的“用户”列，头像缩小为 28 像素，与姓名保持 12 像素间距；账号名称在显示名称下方以 13 像素、`#4b5672` 颜色显示，采用普通文本，取消点击跳转和悬停下划线。“组织”和“应用”移到创建时间之后，不再冻结。

部门、职级和工作日期合并为“任职概况”：首行展示部门，过长时省略并提供完整文本提示；次行以标签展示职级和带日历图标的日期。增加最近登录时间、最后登录 IP 和删除时间，只有已删除用户显示实际删除时间。普通字段缺失时显示浅灰色 `-`，任职概况全部为空时显示一个 `-`，部分有值时仅展示已有资料；状态开关沿用原有显示。

精简独立表格列：移除账号名称、头像、电子邮箱、手机号、真实姓名、已验证、国家/地区、用户类型和标签等独立列；工号、性别、生日不作为独立表格列。用户资料仍可在编辑页或悬停卡片中查看，未删除对应数据字段。用户表格数据列宽统一收紧约 15%，操作列同步收窄；默认配置及已有配置中的用户列名称、任职概况名称和列顺序同步适配。

头像悬停资料卡片参考现有项目的资料卡布局，展示个人封面、头像、显示名称、账号名称、性别标识、管理员盾牌、工作单位、自我介绍及个人主页。资料区依次展示部门、岗位职级、入职日期、邮箱、工号、职务、职位描述、手机号、出生日期、民族和探亲地，仅显示有值的资料行；值缺失、为空或仅含空白时隐藏整行，全部为空时同时隐藏分隔线与资料区。封面未配置时使用蓝色渐变；悬停头像约 100 毫秒后打开，紧贴头像右侧并向上偏移，保留屏幕边缘避让和高度溢出滚动。

资料卡片左侧标签采用 13 像素字号和 `#4b5672` 颜色，标签区域从 48 像素加宽至 64 像素并禁止换行，避免岗位职级、入职日期和出生日期拆行。

### 涉及文件

- `web/src/pages/UserListPage.tsx` — 合并与精简用户表格列，实现任职概况、列冻结、顺序调整和空值占位，接入头像悬停资料卡片；账号名称改为 13 像素和指定颜色的普通文本，取消跳转与悬停下划线。
- `web/src/components/user/UserProfileHoverCard.tsx` — 新增用户悬停资料卡片，读取现有用户字段与扩展属性，展示个人封面和资料并调整浮层定位；补充职位描述（`properties.job_description`）、民族（`properties.ethnicity`）和探亲地（`user.location`），过滤空值资料行；标签改为 13 像素和指定颜色，标签区域加宽至 64 像素且不换行。
- `web/src/components/crud/CrudListPage.tsx` — 为用户表格适配已有配置中的合并列名称、顺序和列宽；其他类型表格沿用原有逻辑。
- `web/src/lib/setting.tsx` — 同步用户表格默认列定义、名称、顺序与宽度，移除不再提供的独立列配置。
- `web/src/locales/en/data.json`、`web/src/locales/zh/data.json` — 增加任职概况的中英文文案。
- `web/package.json`、`web/pnpm-lock.yaml` — 增加悬停资料卡片所需依赖及锁定信息。

### 验证与限制

- 本轮最后一次 `pnpm run typecheck` 通过，覆盖用户表格空值占位调整。
- 后续字号、颜色、标签宽度及账号名称取消跳转的调整未重新执行类型检查或构建。
- 未执行完整前端构建或自动化浏览器验证；卡片最终定位与不同屏幕尺寸下的显示仍需实际页面核对。
- 不涉及数据库迁移；资料卡片使用列表接口已返回的数据，未增加额外资料查询。

### 姓名与用户名模糊搜索补充（2026-11-04）

“用户”列的搜索框同时对姓名（`displayName`）和用户名（`name`）进行包含匹配，任意一个字段命中即可返回，中文提示改为“搜索姓名或用户名”。全局、组织和群组用户列表及分页总数共用匹配条件，保留组织范围和群组成员限制。

#### 涉及文件

- `object/ormer_session.go` — 提取用户搜索条件，为姓名查询增加用户名的或条件，保留参数化查询和字段校验。
- `object/user.go` — 全局及组织用户总数使用与列表一致的用户查询条件。
- `object/group.go` — 群组用户查询和总数接入相同搜索条件。
- `web/src/components/crud/types.ts`、`web/src/components/crud/DataTable.tsx` — 支持按列配置搜索输入提示，未配置时沿用默认提示。
- `web/src/pages/UserListPage.tsx` — 为合并后的用户列设置姓名或用户名搜索提示。
- `web/src/locales/zh/data.json` — 增加“搜索姓名或用户名”中文文案。

#### 验证与限制

- `go build ./...`、`bun run typecheck` 通过。
- 本次涉及的三个前端代码文件执行 ESLint 检查未通过，报告用户列表已有的 21 处缩进问题。
- 未执行页面实际查询验证；后端需重启后生效，不涉及数据库迁移。

### 任职概况提示与系统 Tooltip 样式补充（2026-11-04）

用户列表任职概况的部门文字改用公共 Tooltip，鼠标悬停或键盘聚焦时显示完整内容，长文本自动换行，列表内仍保持省略显示；部门为空时不展示提示。系统公共 Tooltip 统一使用 `#0a0a0a` 背景和白色文字，覆盖表单字段说明、侧边栏及其他使用该组件的提示。

#### 涉及文件

- `web/src/pages/UserListPage.tsx` — 将部门文字的原生标题提示替换为公共 Tooltip，支持键盘聚焦与完整内容换行显示，颜色沿用公共组件。
- `web/src/components/ui/tooltip.tsx` — 将默认主题色背景和文字色改为固定的 `#0a0a0a` 背景与白色文字。

#### 验证与限制

- 添加部门 Tooltip 后执行 `bun run typecheck` 通过。
- 后续公共 Tooltip 颜色调整未重新执行类型检查或构建，未执行自动化浏览器验证。
- 不涉及接口、配置或数据库迁移。

### 资料卡片字段与空值隐藏补充（2026-11-04）

- 职位描述位于职务之后，民族与探亲地位于出生日期之后，直接使用现有数据，不新增接口或数据库字段。探亲地按确认映射读取用户 `location` 字段。
- 全部资料行统一将值转为字符串并去除首尾空白，再过滤空值；隐藏空字段的标签与占位符，保留有值字段的顺序和长文本换行。
- 新增三个字段后执行 `bun run typecheck` 通过；后续空值过滤调整未重新执行类型检查或构建，未执行自动化浏览器验证。

## 2026-11-06：应用默认登录与有效期设置

通过后台新建应用时，默认关闭用户注册，启用登录会话和自动登录；访问令牌有效期由 7 天改为 1 小时，Cookie 有效期由 30 天改为 7 天，刷新令牌有效期维持 7 天。这些默认值不批量改写已有应用的设置。

后端为登录项列表为空的应用生成默认登录项时，将“忘记密码”入口默认设为隐藏；已有显式配置的登录项保持原值。该调整仅改变默认入口可见性，不删除密码找回接口。

### 涉及文件

- `web/src/pages/defaults.ts` — 调整新建应用的注册、登录会话、自动登录及访问令牌和 Cookie 有效期默认值。
- `object/application_util.go` — 在生成默认登录项时，将“忘记密码”项的可见性设为关闭。

### 验证与限制

- 已核对提交中的实际代码差异及默认登录项生成条件。
- 本次仅补充变更日志，未重新运行构建或登录、注册、令牌有效期的业务验证；不涉及数据库迁移。

## 2026-11-07：审计列表用户资料加载优化

会话、日志、令牌和验证页面在分页请求中启用 `includeUsers`，后端按当前页去重后的组织与用户名批量查询用户展示字段，通过响应的 `data3` 返回用户资料映射，原有列表数据和分页总数仍分别保存在 `data` 与 `data2`。全局管理员可获取当前页涉及的用户资料，组织管理员仅获取本组织资料；资料应用组织账号项的查看规则，属性输出限定为悬停卡片所用字段，不返回密码、令牌或 MFA 密钥。

前端优先复用随列表返回的资料，避免为这些用户额外调用会补齐角色、权限与 MFA 的完整用户详情接口；列表和用户列准备完成后一起显示。旧后端或未附带相应资料时，仍回退到用户详情查询，同一页相同用户合并请求。已明确返回但用户不存在时，使用用户名和原有字母头像布局，不重复查询。

共用用户列在需要自行查询详情时，加载期间只保留行高，详情请求完成后再显示完整头像与名称布局，避免先显示临时字母头像再切换资料。日志详情抽屉复用列表已取得的用户资料；应用凭据类型的日志主体继续显示原有应用链接。

### 涉及文件

- `object/audit_user.go` — 按组织与用户名批量查询用户展示字段，不扩展角色、权限或 MFA。
- `controllers/audit_user.go` — 汇总当前页用户、限制管理员组织范围、复用组织配置并过滤资料，通过 `data3` 返回用户映射。
- `controllers/record.go` — 日志分页响应接入可选用户资料输出。
- `controllers/session.go` — 会话分页响应接入可选用户资料输出，保留会话标识脱敏。
- `controllers/token.go` — 令牌分页响应接入可选用户资料输出。
- `controllers/verification.go` — 验证分页响应接入可选用户资料输出。
- `web/src/backend/RecordBackend.ts` — 日志列表请求增加默认关闭的用户资料参数。
- `web/src/backend/SessionBackend.ts` — 会话列表请求增加默认关闭的用户资料参数。
- `web/src/backend/TokenBackend.ts` — 令牌列表请求增加默认关闭的用户资料参数。
- `web/src/backend/VerificationBackend.ts` — 验证列表请求增加默认关闭的用户资料参数。
- `web/src/lib/audit-users.ts` — 将用户资料映射关联到列表行，支持旧响应回退、并行查询及同页请求去重。
- `web/src/components/user/UserCell.tsx` — 区分详情尚未加载与查询完成但用户不存在，加载完成后整块渲染用户信息。
- `web/src/pages/RecordListPage.tsx` — 日志列表启用批量用户资料，列表与详情抽屉共用查询结果。
- `web/src/pages/SessionListPage.tsx` — 会话列表启用批量用户资料，直接传入完整用户列。
- `web/src/pages/TokenListPage.tsx` — 令牌列表启用批量用户资料，直接传入完整用户列。
- `web/src/pages/VerificationListPage.tsx` — 验证列表启用批量用户资料，直接传入完整用户列。

### 验证与限制

- 本轮代码修改后，`go build ./...`、`bun run typecheck` 和修改前端文件的 ESLint 检查通过。
- 尚未完成后端重启后的接口耗时对比、浏览器显示及不同权限账号的端到端验证，不记录未经实测的提速比例；日志列表自身的查询耗时尚未消除。
- 后端需重启后生效；不涉及数据库结构、索引或数据迁移。不启用用户资料参数的调用沿用原有列表响应行为。

## 2026-11-11：个人资料仅限超级管理员修改

“我的账户”及共用用户编辑页面中的头像、姓名、邮箱、手机号、部门、自定义属性、密码等字段，仅允许 `built-in` 组织的超级管理员修改。普通用户和组织管理员仍可查看原有权限范围内的资料，只能管理自己的面容 ID、MFA 账户、多因素认证和 WebAuthn 凭据；组织配置中的不可修改规则继续生效，MFA 项目策略不开放给普通用户修改。

非超级管理员仅在面容 ID 或认证账户、凭据发生变化后显示保存按钮，保存请求只包含实际修改的安全字段。多因素认证的设置、禁用和首选操作仍通过原有独立接口即时保存。未配置个人页面字段时的回退布局也应用相同限制。

后端对非超级管理员的用户更新请求校验目标账户及请求字段，拒绝混入个人资料、管理员标记、密码或任意属性；保留已有 MFA 密钥和 WebAuthn 凭据的保护逻辑，并使用实际数据库列名保存认证数据。头像、证件图片上传、实名验证、邮箱和手机号重置、密码设置及用户批量导入接口也收紧为仅超级管理员可调用。邮箱或手机号为空时，普通用户不能借 MFA 绑定流程写入个人资料，需由超级管理员先维护联系方式。

### 涉及文件

- `web/src/pages/UserEditPage.tsx` — 区分超级管理员与其他用户，只开放本人认证信息编辑，按实际修改字段保存；不显示额外的权限提示条。
- `web/src/components/user/AccountItemRow.tsx`、`web/src/lib/account-items.ts` — 为字段配置和回退布局增加统一编辑白名单，区分管理员查看权限与修改权限。
- `web/src/backend/UserBackend.ts` — 支持传递部分更新的字段列表。
- `web/src/pages/UserListPage.tsx` — 用户导入入口仅对超级管理员显示。
- `controllers/user.go`、`controllers/user_security.go` — 限制非超级管理员仅更新自己的面容 ID、MFA 账户和 WebAuthn 凭据，校验字段白名单、组织修改规则及脱敏密钥；密码设置和实名验证仅允许超级管理员。
- `controllers/resource.go`、`controllers/verification.go`、`controllers/user_upload.go` — 限制头像和证件图片上传、联系方式重置及用户批量导入。
- `controllers/mfa.go` — 阻止通过 MFA 设置流程补写未维护的邮箱和手机号。

### 验证与限制

- `go build ./...`、前端类型检查、修改文件的 ESLint 检查和 Vite 生产构建通过。
- 对真实字段组件进行静态渲染检查，覆盖普通用户、组织管理员、超级管理员及未配置组织字段的场景。
- 对后端白名单解析函数运行独立校验，5 组合法请求通过，11 组夹带资料、权限、密码、错误字段或列名的请求被拒绝。
- 未连接业务数据库执行写入验证，未进行实际账号的浏览器端到端验证；未部署或重启现有服务，不涉及数据库迁移。
- 非超级管理员无法使用密码设置接口，包括原有忘记密码流程；需要修改密码或清除强制改密状态时，由超级管理员处理。

## 2026-11-12：账户内部资料仅限超级管理员查看

用户编辑页的组织、ID、注册应用、用户类型、群组、Open ID 和部门原始 JSON，仅向超级管理员显示原值。非超级管理员有数据时统一显示六位 `******`，无数据时显示 `-`；可读部门名称保留。

账户、用户详情及用户列表响应统一清除这些原值，只附带非空字段标记。当前组织和当前应用使用固定别名，由服务端按已认证身份在鉴权前解析，继续检查实际目标权限。响应内的角色和权限来源也清理群组成员信息；账户响应不再向非超级管理员附带可能含有原值的 JWT 或上游 OAuth 令牌，对应验证器同步二维码和链接入口关闭。

### 涉及文件

- `object/user.go`、`object/user_profile.go` — 增加不落库的字段存在标记，复制并脱敏用户资料、属性及角色权限来源，不修改原始对象。
- `object/profile_alias.go` — 定义当前组织和应用别名，清理账户使用的组织及应用配置中的内部标识。
- `controllers/user_profile.go`、`controllers/user.go` — 统一处理用户详情、分页与非分页列表、排序查询响应，并为同组织请求提供固定别名。
- `controllers/account.go` — 脱敏账户响应及顶层用户 ID，不向非超级管理员附带 JWT；限制无 OAuth 授权上下文的普通会话访问 UserInfo，退出响应不返回内部组织用户名。
- `controllers/application.go`、`controllers/auth.go` — 适配账户应用配置、登录组织比较及第三方绑定回调的应用别名，保留回调原始状态值。
- `controllers/mfa.go` — 非超级管理员设置 TOTP 时，验证器标签仅使用用户名和固定发行者，避免带出组织名。
- `routers/authz_filter.go`、`routers/profile_alias.go` — 在鉴权前解析查询、JSON、普通表单及文件表单中的路由别名，不改写密码、属性或回调状态。
- `web/src/pages/UserEditPage.tsx`、`web/src/hooks/use-options.ts` — 七项资料统一掩码，限制普通用户的选项和组织配置请求，按令牌可用性展示验证器同步入口。
- `web/src/pages/auth/MfaSetupPage.tsx` — 通过用户应用接口获取 MFA 配置，不依赖已隐藏的注册应用字段。
- `web/src/components/user/OAuthWidget.tsx` — ID 被隐藏时按组织别名与用户名识别本人，避免空 ID 相等造成误判。
- `web/src/components/layout/Header.tsx`、`web/src/hooks/use-logout.ts` — 普通用户菜单仅显示用户名，退出后避免跳转至别名组织登录地址。
- `object/user_profile_test.go`、`object/profile_alias_test.go`、`controllers/user_profile_test.go`、`routers/profile_alias_test.go` — 覆盖原值清理、空值标记、超级管理员保留原值、源对象不变、跨组织目标鉴权及应用回调别名。

### 验证与限制

- 相关 Go 测试、前端 TypeScript 检查、修改文件的 ESLint 检查及 `git diff --check` 通过。
- 公开登录配置、组织管理配置及 OAuth/OIDC 按授权范围输出的协议声明保留原规则，本次限制针对账户和用户资料响应，不表示所有认证协议均不再使用身份标识。
- 未运行真实数据库与浏览器端到端验证；未部署或重启服务，不涉及数据库迁移。

## 2026-11-12：新增 User ID 独立属性字段

用户编辑页及组织个人页面设置新增“User ID”字段，共享字段顺序中紧接在 Open ID 前，保存到 `properties.user_id`，与用户顶层 `id` 区分，不新增数据库列。该字段接入现有独立属性的查看、修改权限及正则校验；组织配置独立入口后，通用属性表隐藏对应的 `user_id` 键，保留原值。

User ID 沿用 Open ID 的内部资料限制：仅超级管理员可查看原值，非超级管理员有值时显示六位 `******`，无值时显示 `-`；账户和用户资料响应清除原值，仅保留非空标记。已有组织需在个人页面设置中添加并启用 User ID，本次不自动修改组织配置。

### 涉及文件

- `web/src/lib/account-items.ts` — 增加 User ID 属性映射和文本输入类型，共享顺序中排在 Open ID 前。
- `web/src/pages/UserEditPage.tsx` — 将 User ID 接入内部资料掩码及编辑限制。
- `web/src/locales/en/data.json`、`web/src/locales/zh/data.json` — 增加 User ID 显示文案。
- `object/user_util.go` — 将 User ID 映射到 `user_id`，接入独立属性权限和正则校验。
- `object/user_profile.go` — 非超级管理员的资料响应移除 `user_id` 原值并附带非空标记，保留源对象和超级管理员原值。
- `object/user_profile_test.go` — 扩展脱敏测试，覆盖 User ID 原值清理、空值标记和源对象保留。

### 验证与限制

- 本轮功能修改后，`npm run typecheck`、修改前端代码文件的 ESLint 检查及 `go test ./object -run '^TestGetMaskedUserProfile' -count=1` 通过。
- 本次日志更新未重复运行上述检查；未进行实际数据库写入或浏览器端到端验证，未部署。
