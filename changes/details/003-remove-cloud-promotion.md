# 移除 Casdoor Cloud 推广

**日期**：2026-11-01

## 涉及文件

- `web/src/components/layout/Header.tsx`、`web/src/components/layout/Sidebar.tsx`、`web/src/pages/Dashboard.tsx` — 移除顶部、账户菜单、侧边栏和数据看板的推广入口。
- `web/src/components/auth/AuthLayout.tsx`、`web/src/components/common/CloudPromo.tsx` — 移除登录页推广链接及专用推广组件。
- `web/src/pages/SettingEditPage.tsx`、`web/src/Conf.ts` — 移除推广显示设置及前端配置。
- `object/setting.go`、`conf/web_config.go` — 移除推广设置字段和下发逻辑；已有数据库的旧设置列保留但不再使用。
- `web/src/locales/*/data.json` — 清理推广专用多语言文案。
- `README.md`、`.github/ISSUE_TEMPLATE/bug_report.yml` — 清理项目文档和问题模板中的 Cloud 推广及选项。
