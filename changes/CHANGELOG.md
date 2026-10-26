# 本地变更日志

| 编号 | 日期 | 说明 | 详情 |
| --- | --- | --- | --- |
| 001 | 2026-10-27 | 移除商业模块及注册登录中的付费订阅限制；移除信誉分、积分、排名和初始积分的前后端及数据库逻辑，提供已有数据库清理工具 | [详情](details/001-remove-commerce.md) |
| 002 | 2026-10-27 | 移除后台及认证页面默认的 Powered by Casdoor 标识及其页脚占位；优先显示应用页脚 HTML，其次使用自定义页脚配置，均未配置时不渲染页脚 | `web/src/components/layout/AppLayout.tsx`、`web/src/components/auth/AuthLayout.tsx` |
