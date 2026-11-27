# 项目技能配置

本文件只保存当前仓库差异，六个通用技能的 SKILL.md 可独立复制到其他项目。复制后按目标仓库更新本文件，不能沿用原项目远端或部署适配。用户本次明确要求优先；本文件不授权执行提交、推送或部署，也不存凭据。下列路径均相对目标仓库根目录。

```yaml
project_name: Casdoor
git:
  push_after_commit: true
  target_branch: current
  push_targets:
  - remote: github
    expected_url: https://github.com/jahnli/casdoor
  - remote: gitlab
    expected_url: https://gitlab.semi-tech.com/stc/casdoor
  read_only_remotes:
  - upstream
changelog:
  mode: local-customizations
  path: changes/CHANGELOG.md
  details_dir: changes/details
  language: zh-CN
  stage_after_update: true
announcement:
  title: Casdoor 功能更新公告
  audience: 平台使用者
  featured_topics: []
main_upstream_deploy:
  workflow: null
```

## 提交与日志约定

保留本项目原有约定：“提交”默认含推送，明确要求仅本地或限定远端时遵循用户要求。上表 github/gitlab 接收同一次提交；操作前须核对实际 push URL。upstream 即使存在也仅作只读来源。日志是有效二开清单，不是已发布版本历史；显式更新日志后沿用原有暂存约定，但不得混入用户无关或部分暂存内容。

## 公告业务约定

`update-announce` 输出通用分类公告，默认面向平台使用者。明确要求面向全体员工的公司公告、功能清单及其专用格式时，优先读取独立的 [announce](skills/announce/SKILL.md)，本次不修改该技能。任何公告均须核实角色权限、组织与应用配置，支持某认证协议不等于公司已经启用。

## main 发布目标

尚未配置 main 上游合并部署适配。不能复用其他项目服务器，也不能把本项目现有 release 或 1Panel 配置自动当作 main 的 SSH 目标。执行 main 全流程前需明确 fork/main、上游、验证、镜像发布方式及部署实例。
