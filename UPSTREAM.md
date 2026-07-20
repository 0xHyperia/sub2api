# 上游同步记录

本文档记录 USA0 二开版本与官方 Sub2API 的对应关系。发布 tag 按 USA0 自己的版本线命名，官方基线通过本文件、tag message 和 Git 提交记录追踪。

## 官方 v0.1.162 同步（待发布）

- 集成分支：`codex/merge-upstream-v0.1.162`
- 官方基线版本：v0.1.162
- 官方基线提交：`27f094e0960ebd8e52de7ff7e763c6fec2ff4057`
- 上一官方基线：v0.1.161（`19149ca196eeae4a4482e5299dc6fa4ba0b06c8c`）
- 上一 USA0 版本：v1.0.7
- USA0 发布版本：待确认
- 集成提交：`1ae4c01bda2e496fd7fd6c0004e061ee521e320e`
- 同步状态：已完成冲突解决、USA0 适配、完整代码验证和本地 Docker 健康检查并生成集成合并提交；尚未合入或推送 `usa0/main`，尚未创建 USA0 发布 tag
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
- 后端 `go test ./...`、`go test -tags=unit ./...` 通过，`golangci-lint run ./...` 为 0 issues。
- 前端全量 Vitest、`pnpm run lint:check`、`pnpm run typecheck` 和 `pnpm run build` 通过；Lint 仅保留 1 条既有 warning。
- 语义设计令牌、Toggle 无障碍守卫、四份 Compose 配置渲染和 `deploy/tests/install-github-token-test.sh` 均通过。
- 官方 v0.1.162 未新增迁移或 Ent schema 变更；USA0 已发布迁移保持不变且编号唯一。
- Docker 镜像构建成功，并以 `--no-deps` 仅替换 `sub2api-dev`；PostgreSQL 与 Redis 未重启，`http://localhost:8081/health` 返回 `{"status":"ok"}`。
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
