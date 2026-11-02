# 移除 Casdoor Cloud 推广及调整主题配色

**日期**：2026-11-02

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
