# 上游同步记录

本文档记录 USA0 二开版本与官方 Sub2API 的对应关系。发布 tag 按 USA0 自己的版本线命名，官方基线通过本文件、tag message 和 Git 提交记录追踪。

## USA0 v1.0.12

- 发布版本：v1.0.12
- 官方基线版本：v0.1.169（`26d894ef4f50645a4bf1030e378ac892f17d0223`）
- 上一 USA0 版本：v1.0.11
- 上游集成提交：`d3633d0eccb3fcb58e51322c9ffe91ba68c45635`
- 发布范围：同步 Passkey/WebAuthn、公开 Model Plaza、Optional JWT、用户/API Key 字段级更新、模型协议兼容、价格数据、SMTP、代理熔断和部署安全更新，并保留 USA0 App JWT/ZeroBox、工单、分销、支付、模型市场、模型监控及软件中心
- 发布状态：已合入并推送 `usa0/main`，已创建并推送 v1.0.12 annotated tag；Release workflow 已触发，未在本次操作中持续等待其最终结果
- 记录日期：2026-07-31

## 官方 v0.1.169 同步（本地集成完成）

- 集成分支：`codex/merge-upstream-v0.1.169`
- 官方目标版本：v0.1.169
- 官方目标提交：`26d894ef4f50645a4bf1030e378ac892f17d0223`
- 上一官方基线：v0.1.166（`dc893dd0b8eab41df5be595ae9fcd1aa74a062b8`）
- 集成前 USA0 提交：`0e6a554deb32f2859e7406b9fe81bf63506c9262`
- USA0 发布目标：v1.0.12
- 同步状态：已完成上游合并、冲突解决、USA0 适配和完整代码验证，已合入并推送 `usa0/main`，并创建/推送 USA0 v1.0.12 tag
- 记录日期：2026-07-31

### 同步范围

- 从官方 v0.1.166 同步至 v0.1.169，共纳入 74 个上游提交、228 个官方变更文件。
- 新增 Passkey/WebAuthn 注册、登录、凭据管理、管理员开关和部署配置，并为反向代理来源、RP ID 与 Origin 增加校验和测试。
- 新增独立公开 Model Plaza（`/model-plaza`），按平台、分组、倍率和模型名筛选并展示用户实付价、官方参考价、阶梯价与缓存价格。
- 新增 Optional JWT 中间件，公开接口可在匿名访问与已登录个性化数据之间安全切换；用户与 API Key 更新改为字段级更新，避免并发请求覆盖未修改字段。
- 加固上游 URL 路径校验、OAuth/会话撤销、审计日志和 Prompt Audit 配置边界；代理流熔断在状态异常时 fail-open，减少错误阻断正常流量。
- 修复 SMTP 邮件消息构造、Gemini 上游 URL、Claude/OpenAI 兼容与用量归因，并更新模型价格、容器资源限制及发布构建配置。

### 保留与适配的 USA0 内容

- 31 个文本冲突均按业务语义合并；保留 USA0 App JWT/ZeroBox、用户授权与设备会话、工单、分销、支付、模型监控、软件中心和对应路由/门控。
- 保留用户模型市场 `/model-marketplace`，官方新增 Model Plaza 独立使用 `/model-plaza`；两者的数据口径、访问路径和功能开关不互相覆盖。
- Wire 保留 USA0 `ProvideAuthService` 的分销依赖并接入 Passkey、Model Plaza、Optional JWT；移除重复 `NewAuthService` provider 后重新生成，App JWT、工单、分销、模型监控和软件中心依赖均在生成结果中保留。
- Passkey、Model Plaza 和新增设置界面全部迁入 USA0 语义设计令牌、响应式及无障碍约束；首页的软件中心开关由已有公开设置显式传入，避免展示组件脱离 Pinia 上下文时失败。
- 保留 USA0 `VERSION` 为已发布的 `1.0.11`；未提前推断下一发布版本。
- 官方 Passkey 迁移原编号 `191` 与 USA0 已发布迁移冲突，SQL 内容保持不变并改为 `207_passkey_credentials.sql`；最新迁移区间 `176` 至 `207` 编号唯一。
- Ent 与 Wire 从合并后的 schema/provider 重新生成，未发现未暂存生成漂移。

### 验证结果

- `go generate ./ent` 与 `go generate ./cmd/server` 通过；后端 `go test ./...`、`go test -tags=unit ./...` 通过，`golangci-lint run ./...` 为 0 issues。
- 前端 `pnpm run test:run` 共 289 个测试文件、1809 项测试通过；Passkey、Model Plaza、首页、语义令牌和 Toggle 无障碍定向测试通过。
- 前端 `pnpm run lint:check` 为 0 errors、1 条既有 warning，`pnpm run typecheck` 与 `pnpm run build` 通过；构建仅保留既有动态导入与大 chunk 警告。
- `git diff --check` 通过，无未解决 Git 冲突；源码冲突标记扫描通过，迁移 `176` 至 `207` 前缀唯一。
- 本次仅准备本地集成提交，未运行 Docker 服务替换或登录态浏览器复核；不得将运行时健康检查、桌面/移动视觉或生产发布表述为已验证。

---

## USA0 v1.0.11

- 发布版本：v1.0.11
- 官方基线版本：v0.1.166（`dc893dd0b8eab41df5be595ae9fcd1aa74a062b8`）
- 上一 USA0 版本：v1.0.10
- 发布范围：修复 EasyPay 主动查询协议和错误响应校验，增加待支付订单后台对账，并在支付结果页主动恢复回调丢失或刚过期的已付款订单
- 记录日期：2026-07-30

## USA0 v1.0.10

- 发布版本：v1.0.10
- 官方基线版本：v0.1.166（`dc893dd0b8eab41df5be595ae9fcd1aa74a062b8`）
- 上一 USA0 版本：v1.0.9
- 发布范围：将应用授权拆分为用户授权与设备会话，支持查看、重命名和单独撤销 ZeroAgent 设备，并新增迁移 `206_rebuild_app_oauth_grants.sql`
- 记录日期：2026-07-30

## 官方 v0.1.166 同步（已合并）

- 集成分支：`codex/merge-upstream-v0.1.166`
- 官方基线版本：v0.1.166
- 官方基线提交：`dc893dd0b8eab41df5be595ae9fcd1aa74a062b8`
- 上一官方基线：v0.1.165（`e9a58c1cb8b5ef626a75c93b4d953fde5e67aa29`）
- 集成前 USA0 提交：`3f63fbd69de8c536a7af5efc3b9e9357af465318`
- 集成提交：`651f68136b8a8dc8438e6ba4eb0e102e2abb69e2`
- USA0 发布版本：v1.0.9
- 同步状态：已完成上游合并、冲突解决、USA0 适配、完整代码验证、本地 Docker 健康检查和登录态浏览器复核，并纳入 USA0 v1.0.9 发布
- 记录日期：2026-07-29

### 同步范围

- 从官方 v0.1.165 同步至 v0.1.166，共纳入 62 个上游提交；官方范围变更 142 个文件，本次集成相对 USA0 基线变更 141 个文件。
- 新增面板 API 限流，登录接口按账号、重查询按独立阈值、公开设置按真实客户端 IP 计数；默认启用并允许管理员豁免，Redis 异常时按 fail-open 处理。
- 设置全量 PUT 增加字段存在性判断，局部请求不再把未发送的值类型设置重置为零值；USA0 可空设置继续保持省略即保留语义。
- OpenAI WebSocket 按回合记录模型与计费信息，模型映射后的实际上游模型归因更准确，并补充 reasoning failover、Responses 工具配对及使用记录类型支持。
- 支付管理统计升级为多币种聚合，日收入、支付方式和用户排行均按币种分组；统计继续使用用户实付 `pay_amount`，不使用服务商结算额。
- 完善 Antigravity/Gemini 兼容、账号账单探测、Prompt Audit 配置、Caddy SSE 缓存规则和依赖安全更新；注册页新增可选推广邀请码字段。
- 使用记录增加请求类型筛选与聚合维度，通用 Select、分组描述、模型监控时间线和可用渠道移动布局补齐窄屏边界处理。

### 保留与适配的 USA0 内容

- 16 个文本冲突均按业务语义合并，保留 USA0 App JWT/ZeroBox、公开模型展示、`/model-marketplace`、工单、分销、支付门控和模型监控路由，并为用户、管理员、支付、认证及公开设置组接入对应面板限流器。
- 支付仪表盘沿用 USA0 语义令牌、错误重试和响应式布局，同时移植多币种 DTO；不同币种不混算柱状比例，金额通过 ISO 货币格式化，用户排行按币种独立排序。
- 保留 USA0 可用渠道移动卡片，没有引入上游重复移动实现；补充桌面/移动稳定测试标识、加载和空态测试，长描述与模型列表不会撑破窄屏。
- Select 保留 USA0 ARIA、键盘导航和 typeahead，同时移植视口安全边距与动态最小宽度；分组描述支持换行和任意长词断行。
- 面板限流、注册推广字段和 Composite 分组提示均改用 USA0 语义设计令牌，并为新增 Toggle 补充可访问名称；语义令牌与无障碍守卫通过。
- 支付手续费模式、服务商级费率、充值/订单页可见开关、Contact Us、模型广场与监控设置均保留；局部设置更新未发现需要新增的 USA0 setting-key 别名。
- 官方 v0.1.166 未新增迁移、Ent schema 或 Wire provider；已发布迁移 `176` 至 `202` 保持不变且编号唯一。Ent/Wire 重新生成无漂移，`go.sum` 补齐 Wire 工具链校验项。

### 验证结果

- `go generate ./ent` 与 `go generate ./cmd/server` 通过，生成文件无漂移。
- 后端 `go test ./...`、`go test -tags=unit ./...`、`go test -tags=integration ./...` 全部通过，`golangci-lint run ./...` 为 0 issues。
- 前端 `pnpm run lint:check` 为 0 errors、1 条既有 warning；`pnpm run typecheck`、全量 Vitest（283 个文件、1755 项测试）和 `pnpm run build` 通过。
- 面板限流、设置局部 PUT、支付多币种统计、支付设置、Select、GroupOptionItem、可用渠道、注册推广字段、路由、语义令牌和 Toggle 无障碍定向测试通过。
- `git diff --cached --check` 通过，源码无未解决冲突标记；最新迁移区间 `176` 至 `202` 无重复编号。
- Docker 镜像构建成功，并以 `--no-deps` 仅替换 `sub2api-dev`；PostgreSQL 与 Redis 未重启，`http://localhost:8081/health` 返回 `{"status":"ok"}`。
- 登录态浏览器确认支付概览真实同时展示 CNY/USD 统计，桌面及 390x844 视口无横向溢出；安全设置限流卡片、保存控件和可见 Toggle 名称正常；可用渠道移动空态无溢出。当前数据库无可用渠道卡片数据，因此未做展开卡片的真实数据视觉复核。

---

## 官方 v0.1.165 同步（已合并）

- 集成分支：`codex/merge-upstream-v0.1.165`
- 官方基线版本：v0.1.165
- 官方基线提交：`e9a58c1cb8b5ef626a75c93b4d953fde5e67aa29`
- 上一官方基线：v0.1.162（`27f094e0960ebd8e52de7ff7e763c6fec2ff4057`）
- 集成前 USA0 提交：`a9e0071e46723c250bbfad5f589131c38ea9291e`
- 集成提交：`85474bce70f1782b5b25ade4f5e2a94179687b07`
- 上一 USA0 版本：v1.0.7
- USA0 发布版本：待确认
- 同步状态：已完成冲突解决、USA0 适配、完整代码验证和本地 Docker 健康检查，集成提交已合入本地 `usa0/main`；尚未推送，尚未创建 USA0 发布 tag
- 记录日期：2026-07-27

### 同步范围

- 从官方 v0.1.162 同步至 v0.1.165，共纳入 166 个上游提交；当前集成相对 USA0 基线变更 426 个文件。
- 新增 Composite 组合分组和模型路由，支持公开模型别名按端点、平台、优先级及推理策略路由到具体上游。
- 新增 OpenAI ChatGPT Live 转发、平台能力探测、并发缓存隔离、使用记录类型和会话标识，并补齐跨平台 attestation 降级行为。
- 新增分组级 reasoning effort 策略与 Claude Opus 5 模型支持，完善 OpenAI Responses、WebSocket、Grok 工具协议、图片用量和上游错误兼容。
- 新增 Ollama Cloud 用量抓取、账号级与全局配置及后台账号页展示；账号模型目录和调度缓存同步适配新平台能力。
- 支付新增支付宝手机网站唤起开关及支付结果有效期处理；公告支持后台预览和统一 Markdown/HTML 样式。
- 注册与 OAuth 增加邮箱别名归一化和重复保护，使用日志新增 `session_id`，并完善认证缓存版本、调度快照和计费回退。

### 保留与适配的 USA0 内容

- 49 个文本冲突均按业务语义合并，保留 USA0 首页、登录页、语义设计令牌、响应式布局和无障碍约束；上游新增 Composite、Live、Ollama 和推理策略界面已迁入现有组件体系。
- 保留 Contact Us、工单、分销与推广追踪、模型广场、模型监控、支付门控、App JWT/ZeroBox 和 OAuth 授权，并与上游新增设置、路由及 Wire 依赖共同接入。
- Composite 分组在模型广场与首页展示中展开为具体平台账号；模型监控跳过 Composite 别名，继续按真实平台进行探测，避免别名污染监控目录。
- 管理端分组页面对旧后端或局部 API mock 缺少 Live 能力接口时安全降级；公告预览不依赖用户公告 Pinia store；通用 Select 正确转发 `id` 和 `aria-label`。
- 上游原迁移与 USA0 已发布编号冲突，因此保持 SQL 内容不变并顺延为 `193` 至 `200`；分销迁移守卫继续只禁止 `192` 之后出现新的 distribution/promotion 补丁。
- Ent 与 Wire 已从合并后的 schema/provider 重新生成，生成结果同时保留 USA0 AppAuthorization、工单、分销、模型监控与上游 Composite、Ollama、Live 依赖。
- 保持 USA0 `VERSION` 为已发布的 `1.0.7`，在发布版本确认前不预先推断下一版本号。

### 验证结果

- `go generate ./ent` 与 `go generate ./cmd/server` 通过，重新生成后无未暂存漂移。
- 后端 `go test ./...`、`go test -tags=unit ./...`、`go test -tags=integration ./...` 全部通过，`golangci-lint run ./...` 为 0 issues。
- 前端全量 Vitest、`pnpm run typecheck` 和 `pnpm run build` 通过；`pnpm run lint:check` 为 0 errors、1 条既有 warning。
- 前端语义设计令牌守卫、Select/Toggle 无障碍、分组 Live 降级、公告预览、Composite/Ollama/推理策略及支付流程测试通过。
- 迁移 `193` 至 `200` 编号连续且在新增区间内唯一；开发 PostgreSQL 已确认 8 个迁移全部应用。
- `git diff --cached --check` 通过，源码无未解决冲突标记；Ent/Wire 再生成无漂移。
- Docker 镜像构建成功，并以 `--no-deps` 仅替换 `sub2api-dev`；PostgreSQL 与 Redis 未重启，容器健康，`http://localhost:8081/health` 返回 `{"status":"ok"}`。
- 浏览器在 `http://localhost:3000/dashboard` 被重定向至登录页；登录页 1280px 桌面与 390x844 移动视口均无横向溢出、文字裁切或控制台错误。当前浏览器会话无登录状态，因此本轮未完成管理员分组、账号、设置、公告、使用记录、支付及分销页面的登录态视觉复核，不得将这些页面表述为已验证。

---

## 官方 v0.1.162 同步（待发布）

- 集成分支：`codex/merge-upstream-v0.1.162`
- 官方基线版本：v0.1.162
- 官方基线提交：`27f094e0960ebd8e52de7ff7e763c6fec2ff4057`
- 上一官方基线：v0.1.161（`19149ca196eeae4a4482e5299dc6fa4ba0b06c8c`）
- 上一 USA0 版本：v1.0.7
- USA0 发布版本：待确认
- 集成提交：`1ae4c01bda2e496fd7fd6c0004e061ee521e320e`
- 主分支修复提交：`86fb718e38120e7b49680e753083bc0ec97bd19a`
- 同步状态：已完成冲突解决、USA0 适配、完整代码验证和本地 Docker 健康检查，集成提交已快进合入并推送至 `origin/usa0/main`；尚未创建 USA0 发布 tag
- 记录日期：2026-07-20

### 同步范围

- 从官方 v0.1.161 同步至 v0.1.162，共纳入 114 个上游提交、190 个变更文件。
- 客户端真实 IP 解析改为可配置可信代理、转发链和自定义请求头，并为安全设置、审计记录和反向代理部署补齐完整契约。
- 异步生图对象存储迁移至后台备份页面配置，保存后即时生效，同时修复环境变量凭证加载与 S3 临时加密密钥重启失效问题。
- Grok 新增客户端工具缓存、动态 Free 配额、连接测试与代理质量改进，并完善视频代理、Claude prompt cache 和本地 token 估算。
- OpenAI/Codex 完善标准模型列表兼容、配额错误、HTTP/SSE/WebSocket failover、Responses 事件解析和生图意图性能优化。
- 更新检查支持 GitHub Token，订阅到期精确到分钟、剩余天数向上取整，并修复自定义货币符号、API Key IP 列表及提示词审计关闭等问题。
- 批量生图指引接入中英文多语言，统一暗色主题细节；dev/local Compose 的 Redis 与 PostgreSQL 参数恢复实际生效。

### 保留与适配的 USA0 内容

- 31 个文本冲突均按业务语义合并，保留 USA0 首页、登录页、语义设计令牌、移动卡片、响应式布局和无障碍约束。
- 保留 ZeroBox/App JWT、OAuth 用户授权、工单、分销结算、模型广场、模型监控及支付功能门控，并将其与上游新增路由、中间件和设置契约共同接入。
- Wire 同时保留 USA0 工单存储提供器和上游图片存储工厂；重新生成 `wire_gen.go` 后无额外漂移。
- 为 Viper 环境变量配置补齐默认结构，新增工单与图片存储凭证测试，避免仅通过环境变量配置时字段不可达。
- dev Compose 保留 USA0 服务配置并加入 `UPDATE_GITHUB_TOKEN`；批量生图移动端操作补齐中英文翻译及完整性测试。
- 保持 USA0 `VERSION` 在已发布的 `1.0.7`，在发布版本确认前不预先推断或写入下一版本号。

### 验证结果

- `go generate ./ent` 与 `go generate ./cmd/server` 通过，生成后无漂移。
- 后端 `go test ./...`、`go test -tags=unit ./...`、`go test -tags=integration ./...` 通过，`golangci-lint run ./...` 为 0 issues。
- 前端全量 Vitest、`pnpm run lint:check`、`pnpm run typecheck` 和 `pnpm run build` 通过；Lint 仅保留 1 条既有 warning。
- 语义设计令牌、Toggle 无障碍守卫、四份 Compose 配置渲染和 `deploy/tests/install-github-token-test.sh` 均通过。
- 官方 v0.1.162 未新增迁移或 Ent schema 变更；USA0 已发布迁移保持不变且编号唯一。
- Docker 镜像构建成功，并以 `--no-deps` 仅替换 `sub2api-dev`；PostgreSQL 与 Redis 未重启，`http://localhost:8081/health` 返回 `{"status":"ok"}`。
- 首次远程 integration job 暴露 USA0 分销查询未及时关闭结果集及测试清理吞掉外键错误；修复后 [CI 29744196424](https://github.com/0xHyperia/sub2api/actions/runs/29744196424) 和 [Security Scan 29744196499](https://github.com/0xHyperia/sub2api/actions/runs/29744196499) 均通过。
- 当前 Codex 会话没有可连接的浏览器实例，因此尚未补做本次发布候选的桌面端、390px 移动端及登录态页面复核；不得将此项表述为已验证。

---

## USA0 v1.0.7

- 发布版本：v1.0.7
- 发布分支：`usa0/main`
- 官方基线版本：v0.1.161
- 官方基线提交：`19149ca196eeae4a4482e5299dc6fa4ba0b06c8c`
- 上一 USA0 版本：v1.0.6
- 功能提交：`8f1aed0cef47726342a142cee6303bf7781008b9`
- 发布状态：已创建并推送 v1.0.7 tag，GitHub Release 与 x86_64 GHCR 镜像发布成功
- Release：https://github.com/0xHyperia/sub2api/releases/tag/v1.0.7
- Release 工作流：https://github.com/0xHyperia/sub2api/actions/runs/29674998261
- 记录日期：2026-07-19

### 版本变更

- 统一浅色、深色和系统主题的首屏预绘制及语义色令牌，避免页面加载闪烁，并修正主题切换图标的点击区域与视觉中心。
- 全面优化用户端手机布局，将标题与主要操作、搜索与筛选、状态与刷新按业务层级组合；复杂筛选改为展开区域或移动抽屉。
- 为管理端账号、用户、渠道、分组、订阅、兑换码、优惠码、工单、代理、模型监控、风控、运维日志和订单等页面补充移动卡片或紧凑工具栏。
- 优化模型广场、使用记录、批量图片、订单、订阅、邀请返利和可用渠道的窄屏布局，消除 390px 视口下的横向溢出。
- 模型监控与模型广场继续按账号和分组模型限制生成目录，保持两处展示口径一致。

### 验证结果

- 提交 `8f1aed0ce` 的 GitHub Actions CI 和 Security Scan 均通过。
- 前端全量 Vitest、`pnpm lint:check`、`pnpm typecheck`、`pnpm build` 和 `git diff --check` 通过；Lint 仅保留 1 条既有 warning。
- 在 `http://localhost:3000/dashboard` 登录后以 390px 视口检查 14 个主要用户路由，并抽查深浅主题；所有页面无横向溢出。
- 批量任务筛选弹窗、使用记录筛选抽屉和紧凑工具栏交互正常；主题切换按钮与图标中心偏差为 0px。

---

## USA0 v1.0.6

- 发布版本：v1.0.6
- 集成分支：`codex/merge-upstream-v0.1.161`
- 发布分支：`usa0/main`
- 官方基线版本：v0.1.161
- 官方基线提交：`19149ca196eeae4a4482e5299dc6fa4ba0b06c8c`
- 上一官方基线：v0.1.158（`26abd19a2812edba02bbef93c3e2a620141cc257`）
- 集成提交：`b1d9516700ccbd7cabe09d8ae239dea81ad88c7f`
- 同步状态：已完成冲突解决、USA0 适配和完整验证，生成合并提交并快进合入 `usa0/main`，已推送至 `origin/usa0/main`，并创建、推送 v1.0.6 tag
- 记录日期：2026-07-19

### 同步范围

- 从官方 v0.1.158 同步至 v0.1.161，共纳入 98 个上游提交、377 个变更文件。
- 合入独立的提示词安全审计引擎、审计节点池、策略配置、运行状态、事件复查及 `/admin/prompt-audit` 管理页面。
- 合入敏感操作 step-up 2FA 总开关和会话 IP/UA 绑定默认关闭策略；设置更新字段使用可空语义，避免旧客户端静默重置安全开关。
- 合入入口拒绝聚合日志、无效认证防护、认证缓存失效 Outbox、统一可信客户端 IP 识别及对应运维接口。
- 合入上游计费倍率排序与探测设置、APIKey 上游地址跳转、套餐有效期动态单位和 Stripe SDK 延迟加载。
- 合入 Grok 受保护视频同源代理、媒体资格隔离与模型映射修复，以及按模型隔离临时冷却和池模式临时规则。
- 合入 OpenAI APIKey 独立搜索调度修复、Responses 流事件补全、WebSocket 回合生命周期、瞬时耗尽 503 语义及 Anthropic 监控文本提取修复。

### 保留的 USA0 二开内容

- 保留新版首页、登录页和后台组件视觉体系及响应式布局。
- 增加按用户实际倍率排序、无限滚动的模型广场，并提供独立于渠道监控的模型可用性监控、后台配置和模型卡片状态时间线。
- 模型广场页面迁移至 `/model-marketplace`，为上游标准 OpenAI/Codex `/models` API 保留独立路由。
- 保留 App JWT/ZeroBox 授权，并与上游审计日志和 step-up 中间件共同接入认证、用户、管理员及支付路由。
- 新增审计、账号编辑、批量限额和复制能力继续使用 USA0 的表格选择、无障碍交互、响应式容器和语义设计令牌。
- 模型广场接口加入 Server Timing 用户接口白名单，前后端白名单保持一致。
- 保留 ZeroBox 应用授权、OAuth 路由和用户授权管理能力。
- 保留支付订单功能开关、路由门控和支付状态失败恢复界面。
- 保留自定义首页内容的 URL/HTML 净化、iframe 沙箱与加载失败恢复。
- 模型监控调用适配上游低倍率优先调度参数，并显式保持中性探测策略。
- 后端依赖注入同时保留 USA0 App JWT/ZeroBox、工单、模型监控及上游 Prompt Audit、Ingress Reject、认证缓存清理生命周期。
- 管理侧栏将内容审核与提示词审计组合为安全审计分组，同时保留 USA0 工单、支付、模型广场和监控入口。
- Prompt Audit、账号、设置、运维日志、DataTable、套餐和法律文档页面统一使用 USA0 语义设计令牌、响应式与无障碍约束；本次上游新增的原始色板类已全部清理。
- 账号计费倍率设置迁移至系统设置页，保留账号表格排序、批量选择、实时状态和安全上游地址跳转。
- 保留 USA0 订阅/充值重构，并移植上游动态套餐有效期与 Stripe 纯模块延迟加载。

### 验证结果

- 19 个文本冲突均完成语义合并；精确冲突标记扫描为 0，设置/API 严格契约、App 授权、模型广场、支付、工单和路由守卫测试均通过。
- `go generate ./ent` 与 `go generate ./cmd/server` 成功且生成后无漂移；USA0 `AppAuthorization`、工单、模型监控与上游新增服务同时保留。
- 后端 `go test ./...`、`go test -tags=unit ./...` 通过，`golangci-lint run ./...` 为 0 issues。
- 前端 `pnpm run test:run` 共 240 个测试文件、1509 项测试通过；`pnpm run lint:check` 为 0 errors、1 条既有 warning，`pnpm run typecheck` 和 `pnpm run build` 通过。
- 前端语义设计令牌和 Toggle 无障碍守卫通过；Prompt Audit 已生成独立生产 chunk，DataTable、设置、账号、Stripe 和法律文档回归测试通过。
- USA0 已发布迁移 `177` 至 `186` 保持不变；上游 `181` 至 `184` 原 SQL 按依赖顺延为 `187` 至 `190`，最新区间编号唯一连续。
- Docker 启动迁移后，开发 PostgreSQL 已确认 `prompt_audit_jobs`、`prompt_audit_events`、`ops_ingress_reject_aggregates`、`auth_cache_invalidation_outbox` 及 `full_prompt` 列存在。
- 实际源码 `git diff --check` 与冲突标记扫描通过；上游自带 `openspec/source-freeze/aicodex-prompt-audit-tracked.patch` 是原始补丁快照，内部保留的空白格式不作为源码格式错误修改。
- `docker compose -f deploy/docker-compose.dev.yml build sub2api` 成功，并以 `--no-deps` 仅替换应用容器；PostgreSQL 与 Redis 未重启，Docker 健康检查通过，`http://localhost:8081/health` 返回 `{"status":"ok"}`。
- 浏览器检查覆盖 1440x900 桌面登录页、390x844 移动登录页和移动首页；无横向溢出或控制台错误，管理员路由正确重定向到带原目标的登录页。

---

## USA0 v1.0.0

- 发布版本：v1.0.0
- 发布分支：usa0/main
- 官方基线版本：v0.1.146
- 官方基线提交：6631cbad67a0c15edf1006d2d0d60dc98169adf7
- 当时已知后续官方版本：v0.1.147
- 同步状态：尚未合入 v0.1.147
- 记录日期：2026-07-09

### 主要二开内容

- 重构支付设置，区分即时支付与发卡支付。
- 重构首页为 USA0 新设计。
- 保留官方 main 作为上游同步基线，二开功能在 `usa0/main` 维护。
