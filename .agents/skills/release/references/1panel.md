# Casdoor 单实例 1Panel 部署

本参考仅供 release 的部署阶段使用。Docker 发布成功且目标镜像核实后再读取凭据和查询部署配置；编辑技能时不发送部署请求。

## 本地配置

读取仓库根目录 `.1panel.local.json` 的顶层字段，仅支持单实例。文件不存在、为空、JSON 无效或必需字段缺失时停止部署，说明需补齐的字段，不修改原文件、不复制其他项目凭据、不请求用户在聊天中粘贴 Key。

配置结构示例（全部为占位值，不代表当前部署）：

```json
{
  "host": "192.0.2.10",
  "baseUrl": "http://{host}:8080/api/v2",
  "apiKey": "<在本地填写 1Panel API Key>",
  "node": "local",
  "statusUrl": "http://{host}:8000/api/health",
  "accessUrl": "https://casdoor.example.com/",
  "composeName": "casdoor",
  "serviceName": "casdoor",
  "containerName": "casdoor",
  "imageRepository": "jahnli/casdoor"
}
```

- 必需：非空字符串 `baseUrl`、`apiKey`、`statusUrl`。`baseUrl` 必须包含 `/api/v2`，`statusUrl` 是 Casdoor 健康接口的完整地址，不能填 new-api 的 `/api/status`。
- `node` 默认 `local`；`composeName`、`serviceName`、`containerName` 默认 `casdoor`；`imageRepository` 默认 `jahnli/casdoor`。提供时均须为非空字符串；镜像仓库不能包含 tag 或 digest。用户明确指定值优先，但不能跳过实际目标唯一性核验。
- `accessUrl` 为应用访问入口的完整 HTTP(S) 地址。缺失或非法时不阻断部署，仅报告“访问地址未核实”，不从面板或健康地址推算。
- 可选 `versionUrl` 为 Casdoor `/api/get-version-info` 的完整地址；不配置时不查询应用版本，不从其他 URL 拼接。应用版本核验不需要另行获取登录凭据或修改接口权限。
- `baseUrl`、`statusUrl`、`versionUrl` 支持在主机部分使用 `{host}`。仅在使用占位符时要求 `host` 为非空主机名、IPv4 或带方括号的 IPv6，不含协议、端口、路径、用户信息、空白或花括号。用字面替换（PowerShell：`$url.Replace('{host}', $config.host)`），不执行表达式或环境变量展开。替换后校验 HTTP(S) URL、主机及残余占位符；失败时停止请求。`accessUrl` 不替换占位符，只接受完整地址。
- 不读取 `targets`，不自动把其他项目的对内／对外配置选作本实例。配置只有 `targets` 时提示转换为单实例顶层结构。
- 凭据文件不得加入提交。只在请求进程内存中使用 Key、签名、原 Compose、env 和请求体，不输出文件内容或完整 API 响应，不写临时文件。

## 鉴权与请求

所有面板请求、任务查询和核验固定使用解析后的同一 `baseUrl`、Key 和节点。接口路径追加到去掉末尾斜杠的 `baseUrl`，不重复追加 `/api/v2`。

每次请求生成新的 Unix 秒级时间戳：

```text
1Panel-Timestamp: <timestamp>
1Panel-Token: <小写十六进制签名>
CurrentNode: <node>
Content-Type: application/json; charset=utf-8
```

按目标版本及已确认配置使用鉴权算法：MD5 为 `hex(md5('1panel' + API-Key + timestamp))`；目标版本支持且已确认时优先 HMAC-SHA256，为 `hex(hmac_sha256(API-Key, '1panel:' + timestamp))`，以 [1Panel 官方 API 文档](https://1panel.cn/docs/v2/dev_manual/api_manual/) 为准。鉴权失败时检查 API 开关、IP 白名单、时钟和算法，不循环猜测、不降低面板安全设置。

使用内联终端命令，JSON 序列化器生成请求体；PowerShell 以 UTF-8 字节发送，从响应 `RawContentStream` 按 UTF-8 解码。每次网络请求设有限超时（例如 30 秒），异常仅输出经过脱敏的状态和原因。面板操作要求 HTTP 成功且响应 `code == 200`；请求接受不等于任务成功。

## 1. 取得原配置和容器

以下查询用于获取更新必需信息，不另设远端预检：

```text
POST /containers/compose/search
{"page":1,"pageSize":100}

POST /containers/inspect
{"id":"<composeName>","type":"compose","detail":"<搜索结果的实际 path>"}

POST /containers/inspect
{"id":"<containerName>","type":"container"}
```

- 分页精确匹配唯一编排；在内存中保留原 `name`、`path`、`createdBy`、完整 `env` 和 Compose。
- 容器 `data` 为 JSON 字符串时先解析；确认容器的 Compose 标签指向目标编排和服务，不仅凭名称相似判定。没有唯一目标时停止。
- 记录容器 `Id`、镜像 ID `Image`、完整引用 `Config.Image`、运行状态、健康状态和 `RestartCount`。需要核对 digest 时使用该面板版本的只读镜像检查接口取得 RepoDigests，不能将镜像 ID 当成 registry digest。
- 仅修改 `serviceName` 对应服务的镜像引用，保留其他服务、数据库、卷、挂载、端口、环境变量、启动命令及 Unicode 内容。不要全局替换所有 `image:`，不要修改仓库内的示例 `docker-compose.yml` 代替线上更新。
- 实际 Compose 含 `build`、变量、锚点、digest 或多文件合并时，先确认可直接替换的目标镜像字段、实际文件路径以及 1Panel 会执行的更新行为。无法保证只部署已发布镜像或目标不明确时停止，不擅自删除 `build`、修改 `.env`、替换其他服务或改写编排结构。
- 更新前原镜像仓库与本次目标不同且无明确切换授权时停止。预期仓库配置仅用于比较，不能证明本次构建发布到了该仓库。

## 2. 提交一次更新

说明面板 API 地址、编排、完整目标镜像及重建可能造成的短暂中断。release 已授权强制拉取和重建，即使当前标签相同也继续。

生成唯一 UUID，提交：

```text
POST /containers/compose/update
```

```json
{
  "taskID": "<UUID>",
  "name": "<原 name>",
  "path": "<原 path>",
  "detailPath": "<实际修改的 Compose 文件路径>",
  "content": "<仅替换目标服务镜像的完整 Compose>",
  "createdBy": "<原值>",
  "env": "<完整原值，保留实际类型>",
  "forcePull": true
}
```

读取后直接提交；期间有中断或并发修改迹象时重新读取，配置变化则停止，避免覆盖其他操作。多文件路径不能作为逗号拼接的单文件路径提交。超时或结果不明先查同一 `taskID` 和容器，不能重新生成 UUID 再提交。

## 3. 等待同一任务

```text
POST /logs/tasks/read
```

```json
{
  "id": 0,
  "type": "task",
  "name": "",
  "page": 1,
  "pageSize": 500,
  "latest": false,
  "taskID": "<同一 UUID>",
  "taskType": "",
  "taskOperate": "",
  "resourceID": 0
}
```

按 `data.totalLines` 分页读取必要日志，每 5–10 秒查询，最多 10 分钟。要求 `taskStatus == "Success"`，并确认目标镜像拉取成功、目标容器 `Recreated` 和 `Started`（或目标版本等价的明确事件）。日志结束、HTTP 200 或更新接口返回成功均不足以证明部署完成。输出仅摘取脱敏后的状态和必要错误，不展示完整日志。

## 4. 重新核验

- 重新查询实际 Compose 和容器，确认两者完整镜像引用等于本次目标，新容器 ID 已变化，记录新旧镜像 ID；同标签镜像 ID 可以不变。
- 容器须为 `running`；存在健康检查时等待 `healthy`，`unhealthy`、反复退出或重启增长时报告实际失败。健康状态仍为 `starting` 或服务刚启动时，在任务完成后最多再等待 3 分钟，每 5–10 秒查询；无健康检查时明确注明。
- 对解析后的 `statusUrl` 执行 GET，要求 HTTP 200、有效 JSON 且 `status == "ok"`。Casdoor 的实现是 `controllers/system_info.go` 的 `Health()` 调用 `ResponseOk()`，后者在 `controllers/util.go` 写入 `status: "ok"`；它仅检查存活，不查询数据库，不证明登录流程正常。
- 配置 `versionUrl` 时执行 GET，要求 HTTP 200、`status == "ok"`，记录 `data.version`、`data.commitId` 和 `data.commitOffset`。有效的完整 `commitId` 应等于本次 Actions 的 `headSha`；缩写仅在能唯一核实对应提交时接受。提交不一致时报告版本不一致，不宣称成功。
- Docker 镜像标签来自 `git describe --tags --always`，应用版本来自 `util/system.go` 的 Git 历史计算或内置变量；两者的标签、提交偏移表示可能不同。按本次远端构建规则和提交核对，不直接要求 `data.version` 等于镜像标签。接口缺失、要求认证、返回 `dev` / `unknown` 或无法对应本次提交时，注明“部署完成，应用版本核验未完成”，不索取新的应用凭据、不改权限。未配置 `versionUrl` 时仅说明应用版本未核实，使用已确认的镜像和健康证据判定部署结果。

失败、超时或查询受阻时停止后续更新，返回当前服务状态及 `taskID`。不自动重跑或回滚；回滚需另有授权，并核实原镜像和数据库迁移兼容性。
