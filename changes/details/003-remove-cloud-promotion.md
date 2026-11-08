# 移除 Casdoor Cloud 推广及调整主题配色

**日期**：2026-11-08

## 涉及文件

- `web/src/components/layout/Header.tsx`、`web/src/components/layout/Sidebar.tsx`、`web/src/pages/Dashboard.tsx` — 移除顶部、账户菜单、侧边栏和数据看板的推广入口。
- `web/src/components/auth/AuthLayout.tsx`、`web/src/components/common/CloudPromo.tsx` — 移除登录页推广链接及专用推广组件。
- `web/src/pages/SettingEditPage.tsx`、`web/src/Conf.ts` — 移除推广显示设置及前端配置。
- `object/setting.go`、`conf/web_config.go` — 移除推广设置字段和下发逻辑；已有数据库的旧设置列保留但不再使用。
- `web/src/locales/*/data.json` — 清理推广专用多语言文案。
- `README.md`、`.github/ISSUE_TEMPLATE/bug_report.yml` — 清理项目文档和问题模板中的 Cloud 推广及选项。

## 2026-11-02：调整主题配色

默认主色由黑白配色改为参考 ai-gateway 蓝色预设的配色：浅色模式为 `#1249e5`，深色模式为 `#2e70f4`。同步调整背景、卡片、弹层、文字、边框、状态色及前五组图表色；侧栏采用由主色和背景混合生成的淡蓝选中底色，属于对参考预设的适配，并非完整照搬实色侧栏背景。组织和应用仍可配置自定义品牌色。

### 涉及文件

- `web/src/Conf.ts` — 更新默认品牌主色。
- `web/src/index.css` — 调整深浅色主题及登录页浅色面板的配色；侧栏选中背景按浅色 12%、深色 20% 的主色比例混合生成，保留额外三组图表色。
- `web/src/hooks/use-application-chrome.ts` — 默认品牌色使用主题提供的深浅色主色和焦点环，自定义品牌色继续沿用原有覆盖及深色适配逻辑。
- `web/src/components/ui/sidebar.tsx` — 一级菜单选中文字使用侧栏主色，子菜单使用主题主色，一级菜单悬停时保持选中文字颜色。

## 2026-11-04：登录极光背景与静态资源配置

登录页在非嵌入模式且未设置对应设备的自定义背景图片时显示蓝色极光背景，适配深浅色主题并响应鼠标移动。减少动态效果偏好或不支持所需图形能力时保留静态渐变；页面隐藏时暂停动画。组织和应用配置的登录背景图片继续优先使用。

前后端默认静态资源基址调整为 `http://10.16.7.144:9002/casdoor`，仍可通过 `staticBaseUrl` 覆盖。页面图标、触屏图标及清单使用同一资源来源；后台未配置组织图标时也从该基址读取。后端替换静态资源地址时同时兼容旧公共地址与内网默认地址，组织自定义图标继续优先；自定义样式资源允许使用配置的静态资源基址。前端启动时先读取配置 Cookie，静态资源地址使用动态导出，避免启动阶段固定为旧值。

### 涉及文件

- `web/src/components/auth/LoginAurora.tsx` — 实现极光背景、鼠标响应、深浅色适配、静态回退及动画资源清理。
- `web/src/components/auth/AuthLayout.tsx` — 根据嵌入状态和背景图片配置决定是否显示极光背景。
- `web/src/pages/auth/LoginPage.tsx` — 为登录页启用极光背景。
- `conf/conf.go` — 定义内网默认静态资源基址，保留配置覆盖。
- `object/application_util.go` — 将配置的静态资源基址加入样式资源允许前缀。
- `routers/branding.go`、`routers/static_filter.go` — 兼容两种默认图标地址及静态资源替换，保留组织品牌覆盖。
- `web/index.html`、`web/src/Conf.ts` — 更新默认资源入口与前端静态资源基址。
- `web/src/hooks/use-application-chrome.ts` — 后台默认图标跟随静态资源配置。
- `web/src/lib/setting.tsx` — 提前加载配置 Cookie，并动态导出静态资源基址。
- `web/README.md` — 更新默认图标与静态资源地址的说明。

## 2026-11-08：应用用户名示例配置

应用的用户名登录项新增“显示用户名示例”开关和可编辑的示例用户名，默认关闭，已有应用不自动开启。开启且原示例为空时默认填入 `lijiaheng`，关闭后保留已有非空示例。配置仅作用于当前应用，不增加组织级设置。

后台在用户名项下方显示附属配置行，第一列沿用登录项的下拉框组件且仅提供“显示用户名示例”选项，第二列放置开关，开启后第三列显示示例输入框，不增加表格列。登录项可见性开关及示例开关与输入框、下拉框垂直居中对齐。

登录页在用户名标签右侧以灰色小字显示“示例：用户名”，应用编辑页登录预览使用同一配置。开关关闭或示例为空白时不显示；验证码及魔法链接登录不显示用户名示例。提示不自动填充登录账号，也不参与认证或校验。

### 涉及文件

- `object/application.go` — 在应用登录项中增加示例开关及示例文本字段，随现有登录项配置保存，不新增独立数据库列。
- `web/src/components/application/ApplicationUiCustomizationTab.tsx` — 增加用户名示例附属配置行、默认示例值与编辑交互，统一下拉框样式和开关垂直对齐。
- `web/src/components/crud/EditableTable.tsx` — 支持在数据行之后渲染附属配置行，复用原行的更新逻辑。
- `web/src/pages/auth/LoginPage.tsx` — 根据开关、示例内容和登录方式渲染用户名示例，登录预览同步生效。
- `web/src/locales/en/data.json`、`web/src/locales/zh/data.json` — 增加开关、示例用户名及登录页示例提示的中英文文案。

### 验证与限制

- `go build ./...` 通过；前端多轮布局调整期间的 TypeScript 类型检查通过，最后一次开关垂直居中调整的 ESLint 检查通过。
- Vite 生产打包通过，构建后处理重命名 `web/build` 为 `web/build-old` 时遇到 Windows `EPERM` 错误，完整构建命令未成功。
- 未执行自动化浏览器交互及配置保存后的重新加载验证。运行中的后端需要重启才能识别和保存新增字段。

## 2026-11-08：登录页当前账号展示优化

“使用以下账号继续”区域参考用户列表的用户列，将原先的“用户名（姓名）”单行文字改为姓名在上、用户名在下的两行布局。姓名使用常规文字颜色，用户名使用 13 像素字号和 `#4b5672` 灰蓝色，均采用常规字重；姓名缺失时仅显示用户名，长文本保留省略处理。

头像从 32 像素放大至 36 像素，并禁止在空间不足时被压缩，与两行文字配合展示。实际登录页及应用编辑页登录预览同步生效，继续使用当前账号登录的行为保持不变。

### 涉及文件

- `web/src/pages/auth/LoginPage.tsx` — 调整当前账号姓名与用户名的排列、字号、颜色和字重，放大头像并防止压缩。

### 验证与限制

- 姓名与用户名两行布局调整后，修改文件的 ESLint 检查通过。
- 后续头像尺寸调整未重新执行检查或构建，未执行自动化浏览器验证；不涉及接口或数据库变更。
