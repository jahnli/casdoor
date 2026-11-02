# 本地变更日志

| 编号 | 日期 | 说明 | 详情 |
| --- | --- | --- | --- |
| 001 | 2026-11-01 | 移除商业模块及注册登录中的付费订阅限制；移除信誉分、积分、排名和初始积分的前后端及数据库逻辑，提供已有数据库清理工具；删除已停用的 Ant Design 旧前端 web-old，仅保留 web | [详情](details/001-remove-commerce.md) |
| 002 | 2026-10-27 | 移除后台及认证页面默认的 Powered by Casdoor 标识及其页脚占位；优先显示应用页脚 HTML，其次使用自定义页脚配置，均未配置时不渲染页脚 | `web/src/components/layout/AppLayout.tsx`、`web/src/components/auth/AuthLayout.tsx` |
| 003 | 2026-11-02 | 移除 Casdoor Cloud 推广入口、专用设置和文案，包括后台、登录页及项目文档，已有数据库的旧设置列保留但不再使用；主题配色参考 ai-gateway 的蓝色预设，调整深浅色背景、状态色、图表及侧栏选中态，保留组织和应用自定义品牌色 | [详情](details/003-remove-cloud-promotion.md) |
