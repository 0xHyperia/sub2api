# 上游同步记录

本文档记录 USA0 二开版本与官方 Sub2API 的对应关系。发布 tag 按 USA0 自己的版本线命名，官方基线通过本文件、tag message 和 Git 提交记录追踪。

## 官方 v0.2.1 同步（集成分支）

- 集成分支：`codex/merge-upstream-v0.2.1`
- USA0 合并基线：`usa0/main` / `c2f16ccae9640c3dce07b0a1fc57809a12de11`
- 官方目标版本：v0.2.1
- 官方 annotated tag 对象：`adc26f68f687685e847bfb997559f48e79cac475`
- 官方目标提交：`578785ee7fb35030b094b69624efe25670a36f5f`
- 上一官方基线：v0.1.183（`e8cb019fabf8b55199436229044cbf9aa7a82564`）
- 当前 USA0 版本：`1.0.19`
- 同步范围：338 个上游提交、590 个变更文件，39,911 行新增、3,269 行删除；与 USA0 历史变更存在 111 个重叠路径
- 同步状态：已完成无提交合并、冲突解决和迁移编号适配；待生成代码、完整测试和代码审计后再提交
- 迁移适配：USA0 已发布 `176–247` 保持不变，上游 11 个迁移顺延为 `248–258`，非事务 upstream request ID 索引改用 `255`
- 兼容决策：保留 USA0 模型市场、Model Monitor、ZeroAgent 授权、支付/分销/经营分析和响应式管理界面；接入上游 Astra、Fast/Ultrafast、reasoning effort、价格热加载、请求 ID、原生 compaction 和图片 URL 安全下载能力
- 验证状态：当前环境缺少 Go 工具链，后端生成、Go 测试和 lint 尚未执行；前端与静态检查待完成
- 发布方式：本轮不合回 `usa0/main`、不推送、不创建 tag、不发布

## USA0 v1.0.19

- 发布版本：v1.0.19
- 官方基线版本：v0.1.183（`e8cb019fabf8b55199436229044cbf9aa7a82564`）
- 上一 USA0 版本：v1.0.18
- 上游集成提交：`2b7ccd440a38915870a309176f1ac0678298c815`
- 发布范围：同步官方 v0.1.183，包含 OpenAI/Codex 会话与额度保护、Kimi/Antigravity/Grok 稳定性修复、Model Plaza 上下文与分时定价、OAuth 出站插件系统、媒体/工具/Responses 兼容性和 Channel Monitor V2 修复
- USA0 兼容：保留 ZeroAgent 授权、模型市场与模型监控、支付分销及经营分析；插件管理入口继续由 `plugin_management_enabled` 门控
- 发布方式：创建并推送 v1.0.19 annotated tag 后触发 Release workflow；Actions 运行结果由后续单独确认
- 记录日期：2026-08-27

## 官方 v0.1.183 同步（集成分支）

- 集成分支：`codex/merge-upstream-v0.1.183`
- USA0 合并基线：`usa0/main` / `4c0725fbf4358318628c52af01df8241e40cc3dc`
- 官方目标版本：v0.1.183
- 官方 annotated tag 对象：`c21fd3382a1c39fe491a96ac6780bac927327ae4`
- 官方目标提交：`e8cb019fabf8b55199436229044cbf9aa7a82564`
- 上一官方基线：v0.1.179（`75f88be5f75c27771836b586f7de1503afa0e3bc`）
- 当前 USA0 版本：`1.0.19`
- 同步状态：已完成合并、冲突解决、USA0 适配、代码审计和自动化验证，并已合回本地 `usa0/main`；已完成 v1.0.19 发布准备，尚未推送、创建 tag 或发布
- 记录日期：2026-08-27

### 同步范围

- 从官方 v0.1.179 到 v0.1.183 共纳入 225 个上游提交、516 个官方变更文件，官方差异为 42,026 行新增、3,281 行删除；基于三方合并结果识别 23 个文本冲突文件。
- OpenAI OAuth 新增 5 小时/7 天额度耗尽 429 的重置时间识别与账号暂停；普通瞬时 429 继续重试。Codex 支持 `session-id` 粘性会话，容量溢出时保留持久绑定；Responses Lite、parallel tool calls、custom tool/tool search 恢复和工具 ID 前缀兼容性同步修复。
- Kimi 并发限制 403 改为可恢复临时冷却并保留故障转移；Antigravity 兼容模式 token 上限收敛到 64,000；Grok CLI 使用官方 UA；邮箱换绑增加别名去重和事务级并发保护。
- Model Plaza 增加上下文阶梯定价、渠道分时定价及工作日规则；Fast/service tier 计费、响应模型和模型列表读取上限贯穿路由、用量、审计、经营分析和公开模型价格。
- OAuth 出站传输插件系统落地，增加插件清单、签名/兼容性检查、运行时、管理端上传/配置页面和 OpenAI 出站插件路由；插件迁移在 USA0 中顺延为 `246_plugins.sql`、`247_plugin_artifacts.sql`。
- Gemini schema、Grok 媒体/工具调用、Responses Lite、OpenAI 文件输入、Realtime、图片和国产供应商协议兼容性获得修复；补充配额自动重置、请求体诊断和多种 failover/会话边界测试。
- Channel Monitor V2 修复 Composite 平台聚合 SQL 条件；同步渠道缓存失效、监控吞吐、usage 守卫、OpenAI 配额暂停以及备份/插件运行稳定性修复。

### 冲突与 USA0 适配

- 23 个冲突均按业务语义解决，没有整文件选择 ours/theirs；保留 USA0 首页、登录注册、语义设计令牌、移动布局、`/model-marketplace`、模型监控、支付、分销、经营分析、工单、ZeroAgent/App 授权和自有模型市场。
- 上游 OAuth 出站插件系统与 USA0 自定义 OAuth 服务依赖并存；插件仅扩展出站传输能力，不改变 ZeroAgent 授权契约、账号采用决策或分销归因流程。插件管理入口受现有管理员能力门控，secret 仍只返回 configured 状态。
- 上游 Model Plaza、标准 `/models` 与 USA0 `/model-marketplace` 保持独立。上游供应商/模型能力可进入模型市场数据源，但不替代 USA0 的模型级图标、分组图标、价格倍率、性能展示和模型监控目录。
- Channel Monitor V1/V2 与 USA0 Model Monitor 继续使用独立服务、数据表、游标、路由和开关；V2 被动聚合真实渠道流量，Model Monitor 继续执行模型×分组主动探测、持久游标、多实例认领、补录和公平补偿，两套监控可以并行运行。
- 退款、余额、佣金和经营事实仍由 USA0 财务事务统一维护；上游账号状态抢占、额度暂停、重试和可恢复错误语义仅作为请求生命周期输入，不绕过 USA0 的 affiliate/distribution 幂等事务。
- 价格解析固定为 Group → Channel → 内置基础价，再应用 USA0 用户/分组/峰值倍率及利润控制；上下文阶梯、分时、Fast/service tier 和响应模型计费全部消费同一解析结果，模型市场和经营成本不得继续使用孤立全局价。
- 用量和审计保留 requested、routed/upstream、response 三种模型维度；渠道开关决定最终计费模型，经营分析读取最终计费模型，模型监控目录仍使用请求/路由模型，避免上游降级结果污染模型可用性。
- 设置 DTO、局部 PATCH、审计键、公开脱敏响应、前端类型和严格 fixture 同步接入上游新增字段；默认值继续 fail-closed，任何 OAuth、插件和供应商凭证不进入公开响应。
- 迁移保持已发布 `176–245` 不变；上游插件迁移原编号 `229–230` 在 USA0 顺延为 `246–247`。`176–247` 每个新增区间前缀唯一连续，Ent/Wire 均从合并后的 schema/provider 重新生成。
- 前端页面采用 USA0 semantic token、共享组件和既有响应式布局；不覆盖 USA0 首页、登录注册设计、模型市场及 ZeroAgent 品牌。此前模型广场入口、账号状态颜色、插件页面和运维错误详情交互修复与本次同步合并为同一提交。

### 验证结果

- `PATH=/tmp/sub2api-go-n7aFsf/go/bin:$PATH go test -tags=integration ./... -timeout=30m` 通过，退出码 0；此前普通 `go test ./...` 与 `go test -tags=unit ./...` 也已通过。
- `mise exec golangci-lint@2.13.0 -- golangci-lint run ./... --timeout=30m` 通过，报告 `0 issues`。
- `go generate ./ent` 与 `go generate ./cmd/server` 使用仓库固定 Go 工具链通过，重复生成后无 Ent/Wire 漂移。
- 前端 `mise exec -- pnpm run lint:check` 通过（0 errors，1 条既有未使用测试常量 warning）；`typecheck` 通过；Vitest 共 351 个测试文件、2,286 项测试通过；生产 `build` 通过，仅有既有 Browserslist、动态导入和大 chunk 警告。
- `git diff --check`、暂存差异检查、未解决索引检查和源码冲突标记检查通过；迁移 `176–247` 新增区间唯一连续，合并验证时 `backend/cmd/server/VERSION` 为 `1.0.18`，发布提交更新为 `1.0.19`。
- 本次尚未执行推送、tag、发布或 GitHub Actions；不得将这些状态表述为已完成。登录态桌面/移动浏览器复核也未在本轮新增声明。

## USA0 v1.0.18

- 发布版本：v1.0.18
- 官方基线版本：v0.1.179（`75f88be5f75c27771836b586f7de1503afa0e3bc`）
- 上一 USA0 版本：v1.0.17
- 上游集成提交：`bf62fa595a16700a37cb5c67f9f499f9e693d436`
- 发布范围：同步官方 v0.1.179，新增 Kimi、智谱、DeepSeek adaptive 多协议及 Composite 路由、渠道 Fast/Flex/上下文区间倍率、可配置代理探测、Responses input token 计数和用量聚合优化，并合入 OpenAI、Codex、Grok、Channel Monitor 等稳定性修复
- 附加修复：模型广场与模型监控成功率仅统计确认成功和明确上游模型故障；余额、参数、鉴权、模型不存在、内容策略、限流、网关、网络、TLS、超时及断流等不可归因结果不再降低成功率，故障转移最终成功仅计一次成功
- 发布方式：创建并推送 v1.0.18 annotated tag 后触发 Release workflow；Actions 运行结果由后续单独确认
- 记录日期：2026-08-20

## 官方 v0.1.179 同步（已合并）

- 集成分支：`codex/merge-upstream-v0.1.179`
- 官方目标版本：v0.1.179
- 官方 annotated tag：`3c28fad50472b409e18666df617f4237d8ba7007`
- 官方目标提交：`75f88be5f75c27771836b586f7de1503afa0e3bc`
- 上一官方基线：v0.1.178（`e0c48a19ed794a565e3858662520afe0a1f9f0ba`）
- 集成前 USA0 提交：`606ceec7b53e341c717a6c1dd0b6a03d2a3b3458`
- USA0 发布版本：v1.0.18
- 集成提交：`bf62fa595a16700a37cb5c67f9f499f9e693d436`
- 同步状态：已完成合并、冲突解决、USA0 适配、代码审计和自动化验证，并已合入 `usa0/main`，作为 USA0 v1.0.18 的发布基线
- 记录日期：2026-08-20

### 同步范围

- 从官方 v0.1.178 同步至 v0.1.179，共纳入 78 个上游提交、212 个官方变更文件，官方差异为 9,567 行新增、1,446 行删除；当前 USA0 集成差异为 213 个文件。
- Kimi、智谱 GLM、DeepSeek API Key 账号新增 `adaptive` 协议，同一账号可承接 Chat Completions、Anthropic Messages 和 OpenAI Responses，并可分别配置协议 Base URL；账号连接测试、请求头覆写、配额与余额探测同步适配。
- Composite 新增 Kimi/智谱/DeepSeek 路由目标，并支持 Codex 模型列表、故障转移、Alpha Search 与 Live；管理端平台选择改用共享平台目录，补齐订阅、账号、运维与错误透传筛选。
- 渠道定价新增 `fast_multiplier`、`flex_multiplier` 和上下文区间输入/输出/缓存倍率；Anthropic `speed: fast` 纳入 Fast 计费与使用日志，渠道标准价覆盖不再抹掉模型价卡的 Fast 比例。
- 新增可配置代理探测目标和解析器，以及 `POST /v1/responses/input_tokens`：官方端点透传真实计数，自定义中转、Grok 和国产供应商使用本地估算。
- 用量统计改用 `GROUPING SETS` 单次扫描并统一筛选口径，新增 requested/upstream effective model 并发表达式索引；Channel Monitor 修复 quota 凭据判定、模式组合校验、账号解绑、模式徽标和配额标签布局。
- 同步 OpenAI 非流式读取失败 failover、容量文本错误识别、Responses WS 后续轮换号、reasoning bridge、Grok 图片/工具搜索、Grok xhigh 以及国产供应商连接测试路由修复。
- 长上下文计费采用上游破坏性变更：分组或账号任一开关启用即生效。存量 OpenAI 分组若需维持旧口径，升级后必须显式关闭分组 `long_context_pricing_enabled`。

### 冲突与 USA0 适配

- 初始 21 个文本冲突均按业务语义解决，没有整文件选择 ours/theirs；保留 USA0 首页、合并登录注册、语义设计令牌、移动布局、分组列设置菜单修复、ZeroAgent 授权、支付、分销、经营分析、模型市场和模型监控。
- 账号测试事件同时保留 USA0 TTFT 收集和上游 adaptive 多协议测试的 completion 抑制；Composite 默认候选新增 Kimi/智谱/DeepSeek，同时继续通过 USA0 导出的 `DefaultModelsListCandidateIDs` 为模型市场及管理端复用。
- `/model-marketplace`、官方 `/model-plaza` 与标准 `/models` 继续保持独立；新增供应商与 Composite 扩展的是上游接入能力，不替代 USA0 模型级图标、分组图标、发现、价格和性能展示。
- Channel Monitor V1/V2/quota 与 USA0 Model Monitor 的服务、数据表、游标、路由和开关保持独立；共享平台目录只统一平台选项，不让 `channel_monitor_mode` 控制 `model_monitor_enabled`。
- 定价基础优先级继续为 Group → Channel → 内置价卡；渠道标准价、Fast/Flex 和区间倍率在同一解析结果上组合，USA0 用户/分组/峰值倍率及利润控制继续消费最终解析价格。长上下文关闭时选择最低渠道区间，开启时按实际上下文选择。
- 设置严格公开响应仅删除已下线 Sora 的旧例外；Sora 配置、文档和示例同步清理，当前非测试源码没有可运行 Sora 路径。代理探测新增配置已接入后端配置、环境可达性测试和部署示例。
- 官方迁移 `226–228` 与 USA0 已发布历史冲突，保持 SQL 语义并顺延为 `242–244`：effective model 并发索引、Composite 国产供应商约束、渠道定价倍率。`176–244` 每个迁移前缀唯一连续，`242` 已接入非事务索引恢复逻辑。
- Ent/Wire 从合并后的 schema/provider 连续生成两次且无漂移；生成结果继续包含 ZeroAgent 应用授权、分销、支付、工单、模型监控、汇率同步与上游新增服务依赖。USA0 `VERSION` 在发布提交中更新为 `1.0.18`。

### 验证结果

- 后端 `go test ./...`、`go test -tags=unit ./...` 和 `go test -tags=integration ./... -timeout=30m` 通过；unit 套件发现并修复了 Composite 候选测试对旧私有符号的引用。integration 首轮有一项既有 Server-Timing 调度计时断言抖动，单项连续 5 次及全量复跑均通过。
- `golangci-lint v2.9 run ./... --timeout=30m` 为 0 issues。真实 PostgreSQL 已验证空库应用全部迁移，以及从 `241` 升级到 `242–244`；两个表达式索引、Composite 平台约束、倍率列、正值约束与重启幂等均通过。
- 前端 `pnpm run lint:check` 为 0 errors、1 条既有支付测试未使用变量 warning；`pnpm run typecheck`、全量 Vitest（344 个测试文件、2,207 项测试）和生产构建通过。构建仅有既有 Browserslist、动态导入和大 chunk 警告。
- `git diff --cached --check`、未解决索引检查和源码冲突标记扫描通过；合并验证时 `VERSION` 为 `1.0.17`，迁移 `176–244` 唯一连续。
- 本次未替换本地服务，也未进行登录态桌面/390px 管理页面浏览器复核；不得将运行时健康、视觉验证、远端推送或发布表述为已完成。

## USA0 v1.0.17

- 发布版本：v1.0.17
- 官方基线版本：v0.1.178（`e0c48a19ed794a565e3858662520afe0a1f9f0ba`）
- 上一 USA0 版本：v1.0.16
- 上游集成提交：`8fdf8134dfc8bc3ca27127d2700a453e2e8cacc3`
- 发布范围：同步官方 v0.1.177 至 v0.1.178，新增 Kimi、智谱、DeepSeek 一等供应商、Channel Monitor 配额模式、渠道模型分时倍率、分组用量日汇总、Codex remote compaction v2 与 turn state、OpenAI Team 联动熔断及批量账号设置；保留并适配 USA0 模型广场、模型监控、ZeroAgent 授权、支付、分销、经营分析和响应式管理界面
- 附加修复：修正分组管理列设置菜单的视口定位、滚动、末项裁切及桌面/移动端对齐
- 发布方式：创建并推送 v1.0.17 annotated tag 后触发 Release workflow；Actions 运行结果由后续单独确认
- 记录日期：2026-08-18

## 官方 v0.1.178 同步（已合并）

- 集成分支：`codex/merge-upstream-v0.1.178`
- 官方目标版本：v0.1.178
- 官方 annotated tag：`15290e66c66801a7ce435a6d24b178ee9486f284`
- 官方目标提交：`e0c48a19ed794a565e3858662520afe0a1f9f0ba`
- 上一官方基线：v0.1.176（`e803e3851c0a7e222cfadeafad7b8636ab959d11`）
- 集成前 USA0 提交：`3331cbd6b879d61b196a5d17a5d61640eb0f7f47`
- USA0 发布版本：v1.0.17
- 同步状态：已完成合并、冲突解决、USA0 适配、代码审计和自动化验证，并已快进合入 `usa0/main`，作为 USA0 v1.0.17 的发布基线
- 记录日期：2026-08-18

### 同步范围

- 从官方 v0.1.176 同步至 v0.1.178，共纳入 124 个上游提交、347 个官方变更文件，包含 v0.1.177 与 v0.1.178。
- v0.1.177 新增分组用量日汇总和自动汇聚；Codex 接入 remote compaction v2、会话 beta 头和 `x-codex-turn-state` 回传，并隔离跨账号回显、原生 v2 与旧压缩路由。
- v0.1.177 修复 Grok 长上下文及带版本媒体模型计费、账号自动刷新偏好；Codex OAuth 指纹收敛改为显式选择、默认关闭，并覆盖透传路径。
- v0.1.178 新增 Kimi、智谱、DeepSeek 一等供应商支持，覆盖多协议接入、分组、渠道定价、配额/余额、调度、限流、模型发现和计费；新增 Channel Monitor quota 模式及 8 平台配额快照。
- v0.1.178 新增渠道模型分时倍率、OpenAI Team 联动熔断和账号批量设置，并补充 Grok 本站 24h/7d/30d 用量、Ollama 用量查询及 Ops 自定义时间区间。
- 同步 OpenAI WS/HTTP 自定义工具、二进制帧策略链、Codex 身份与探针、Gemini 工具配置与 4xx、Anthropic SSE 过载、认证快照分组定价、SMTP 到期提醒、Ops 批量落库等修复。

### 冲突与 USA0 适配

- 初始 25 个文本冲突均按业务语义解决，没有采用整文件 ours/theirs；保留 USA0 首页、登录注册、语义设计令牌、移动布局、ZeroAgent 授权、支付、分销、affiliate、经营分析和模型监控。
- Kimi、智谱、DeepSeek 作为新的账号与渠道平台接入多协议网关、调度、计费、余额/配额和管理页面；它们解决的是供应商接入能力，不能替代用户发现与模型维度展示，因此继续保留 `/model-marketplace`、模型级 `ModelIcon`、`modelVendor.ts` 以及模型感知的 `GroupBadge`。
- 官方 `/model-plaza` 与 USA0 `/model-marketplace` 保持独立，标准 `/models` 仍作为网关模型接口；混合渠道与 Composite 场景继续使用模型图标和供应商识别，分组平台图标只表达渠道归属。
- Channel Monitor quota 与 V1/V2、USA0 Model Monitor 的服务、表、路由和开关保持独立；`channel_monitor_show_quota` 默认关闭并在用户公开响应中 fail-closed，不受 `model_monitor_enabled` 控制。
- 分时倍率只应用于 Channel 的 token 定价，并按 turn 开始时刻取值；Group → Channel → 内置基础价优先级及 USA0 用户/分组/峰值倍率和利润控制保持不变。
- 上游邀请码消费与用户创建原子化后，USA0 进一步将邮箱注册/OAuth 首次建号、邀请码占用和 durable distribution claim 放入同一 Ent 事务；失败时保留 pending retry，`distribution_code` 与 `aff_code` 继续互斥。
- 管理设置完整接入 `channel_monitor_mode`、`channel_monitor_hide_throughput` 和 `channel_monitor_show_quota` 的 DTO、PATCH、公开脱敏响应、前端类型、严格 fixture 与审计键；新增设置和 CN 供应商页面已迁入 USA0 semantic token 与无障碍契约。
- Release workflow 保留 USA0 构建与发布流程，并接入上游 QEMU 初始化；本次仅准备本地集成提交，不触发构建或发布。
- 保持已发布迁移 `176–235` 不变；上游新增迁移顺延为 `236–241`：分组用量日汇总、汇总时区、CN 平台配额、Codex 指纹种子回填、渠道模型分时定价和 Channel Monitor quota 模式。`176–241` 前缀唯一连续。
- Ent 与 Wire 已从合并后的 schema/provider 重新生成，生成结果同时保留 CN provider、Channel Monitor quota、Model Monitor、ZeroAgent、汇率、分销、affiliate 与经营分析依赖。

### 验证结果

- `go generate ./ent ./cmd/server` 通过且再生成无漂移；`go test ./...`、`go test -tags=unit ./...`、`go test -tags=integration ./... -timeout=30m` 全部通过，CI 固定版 `golangci-lint v2.9 run ./... --timeout=30m` 为 0 issues。
- 新增 PostgreSQL 集成回归覆盖注册事务中的 distribution claim；设置审计测试覆盖 Channel Monitor 三个契约字段；CN provider、quota 模式、分时定价、日汇总、Codex 指纹回填及迁移测试均通过。
- 前端全量 Vitest 共 335 个测试文件、2160 项测试通过；`vue-tsc --noEmit` 和生产构建通过；ESLint 为 0 errors、1 条既有未使用测试常量 warning。
- `git diff --check`、暂存差异检查、源码冲突标记和未解决索引扫描通过；官方 tag、目标提交、124 个提交、347 个文件及迁移 `176–241` 唯一连续均已复核；v1.0.17 发布提交中的 `backend/cmd/server/VERSION` 为 `1.0.17`。
- 本地后端健康接口返回 `{"status":"ok"}`。Playwright 在 1440×900 与 390×844 下验证登录页无横向溢出和控制台错误；现有浏览器会话无登录态，受保护的 Settings、Groups、Accounts、Channel Monitor、Model Monitor 和模型市场仅验证了正确跳转，未声明登录态页面视觉通过。`/model-plaza` 因当前功能门控回到自定义首页，首页无溢出但有一项既有资源 404。

---

## 官方 v0.1.176 同步（已合并）

- 集成分支：`codex/merge-upstream-v0.1.176`
- 官方目标版本：v0.1.176
- 官方 annotated tag：`14e6d7ee7bdb1e4cb6bc59129a7ee1dd1110c52a`
- 官方目标提交：`e803e3851c0a7e222cfadeafad7b8636ab959d11`
- 上一官方基线：v0.1.170（`c043c24774228ba891ddf90d783aa6dc7d0855b5`）
- 集成前 USA0 提交：`3d60b4768b9cafadee7ba5af697e8165e6363971`
- 集成提交：`065239909a2d363bad06c5a5148d4de7bf495748`
- USA0 版本：保持 v1.0.16，未推断下一发布版本
- 同步状态：已完成冲突解决、USA0 适配和自动化验证，集成分支已合入 `usa0/main`；未创建 tag，未发布
- 记录日期：2026-08-15

### 同步范围

- 从官方 v0.1.170 同步至 v0.1.176，共纳入 330 个上游提交、682 个官方变更文件；官方版本链为 v0.1.171、v0.1.172、v0.1.173、v0.1.175、v0.1.176，不存在 v0.1.174。
- 新增腾讯天御与阿里云验证码 2.0、OAuth/Passkey 动作验证、Codex 客户端身份与版本同步、Composite 推理策略，以及额度缓存、连接超时、WebSocket、Responses 和计费完整性修复。
- 新增上游响应模型审计、模型错配筛选和按响应模型计费；新增 Channel Monitor V2 被动聚合、V1/V2 互斥模式、健康阈值、隐私默认值与固定时间桶。
- 扩展 Grok SSO/refresh-token 重新授权、跨实例 OAuth、Free/付费档位识别、媒体与 Voice、Realtime、TTS/STT、`/v1/web_search`、Grok-only `/x_search`、Grok 4.6/200k 阶梯价及 Chat/Responses 搜索工具兼容。
- 新增默认文本模型和跨客户端映射、团队/模型冷却和软门禁、分组视频/音频/搜索/逐模型定价、长上下文定价开关、Codex OAuth 指纹策略，以及大文件备份分卷上传和恢复。
- 合入退款强制确认、Stripe 幂等退款、余额/订阅状态抢占、定时备份 leader 锁、Realtime 漏费、未知 Grok 模型零成本、渠道缓存失效、价格模型归一化和响应探测等修复。

### 冲突与 USA0 适配

- 初始 48 个文本冲突均按业务语义解决，没有采用整文件 ours/theirs。保留 USA0 首页、登录注册、语义设计令牌、移动布局、模型市场 `/model-marketplace`、ZeroAgent 授权、工单、支付、分销、经营分析和模型监控。
- Model Monitor 与 Channel Monitor V2 保持服务、数据表、游标、路由及开关完全独立；`channel_monitor_mode` 只控制 Channel Monitor V1/V2，不能关闭或切换 `model_monitor_enabled`。V2 使用被动真实流量聚合，USA0 Model Monitor 继续按模型与分组执行主动探测、补录、多实例认领和公平补偿。
- 退款入口采用上游 `require_force`、pending 状态抢占、可用余额与幂等最终化语义；实际财务事务保留 USA0 支付金额快照、affiliate/distribution 佣金冲正和经营事实更新，使订单、余额/订阅、返利、分销及审计在同一事务内完成。
- OAuth pending 采用上游账号接管防护和动作验证码状态机，并将 USA0 `distribution_code` 贯穿 pending 创建、LinuxDo、OIDC、微信和 DingTalk 完成路径；继续与 `aff_code` 互斥，保留账户采用决策和 ZeroAgent 授权契约。
- 分组定价统一由 resolver 按 Group → Channel → 内置基础价解析，再应用 USA0 用户/分组/峰值倍率和利润控制；模型市场、经营成本、响应模型计费、长上下文以及 Grok 视频/音频/搜索均读取相同结果。
- 使用记录保留 requested、routed/upstream、response 三种模型维度；渠道开关决定最终计费模型，经营分析读取最终计费模型，模型监控目录仍使用请求/路由模型，避免上游降级结果污染监控目录。
- `/x_search`、Voice、Realtime、TTS/STT、视频和 Web Search 已接入路由、feature guard、Prompt Audit、错误归属、使用日志和计费；Composite 路由到 Grok 时保留媒体能力与对应价格。
- 管理设置同步合并 CAPTCHA、Codex 版本、Grok 映射、Channel Monitor 和 USA0 自有字段，覆盖 DTO、局部更新、审计键、公开脱敏响应、前端类型和严格 JSON fixture；secret 仅暴露 configured 状态。
- 上游新增页面和组件已适配 USA0 semantic token；Channel Monitor V1/V2 使用独立 wrapper，Grok 用量同时展示 USA0 本地请求/token/账号成本/用户成本和上游 Free/Paid 窗口，二维码仍保留固定白底以确保暗色主题可扫描。
- Ent 与 Wire 从合并后的 schema/provider 重新生成。USA0 已发布迁移 `212–215` 保持不变，上游迁移顺延为 `216–235`：`216` 响应模型、`217` V2 基表、`218` 非事务 mismatch 索引、`219–230` V2 配置/汇总/权限/隐私、`231–235` 视频/音频/搜索/清理/逐模型定价；每个编号唯一且连续。

### 验证结果

- `go generate ./ent` 与 `go generate ./cmd/server` 通过，生成前后无漂移；`go test ./...`、`go test -tags=unit ./...`、`go test -tags=integration ./...` 全部通过，`golangci-lint run ./... --timeout=30m` 为 0 issues。
- 新增 PostgreSQL 集成回归测试，独立数据库先执行至迁移 `215`，再由正式 runner 执行 `216–235` 并复跑；20 个新迁移均只记录一次，响应模型字段、mismatch 索引和分组定价字段存在。全新数据库、并发 leader/advisory lock 与迁移幂等也由 integration suite 验证。
- 前端 `pnpm run test:run` 共 326 个测试文件、2065 项测试通过；`pnpm run typecheck`、`pnpm run lint:check` 和生产构建通过。Lint 仅保留 1 条既有 warning，构建仅有动态导入与大 chunk 警告。
- `git diff --check`、源码冲突标记扫描和迁移 `212–235` 唯一连续检查通过；本地前后端健康接口可访问，`backend/cmd/server/VERSION` 保持 `1.0.16`。
- 最终提交 `052a4c842b7b311bcbdb5219b32987dc2f544cf9` 的 GitHub Actions CI [31829129083](https://github.com/0xHyperia/sub2api/actions/runs/31829129083) 和 Security Scan [31829129200](https://github.com/0xHyperia/sub2api/actions/runs/31829129200) 均通过；Go 已升级至 1.26.6，`nanoid` 已锁定至 3.3.18，并移除 lint action 对远程 JSON Schema 校验的运行时依赖。
- 使用缓存的 Playwright 1.55/Chromium 对最新临时后端与前端执行登录态浏览器复核：Settings、Backup、Groups、Channel Monitor、Model Monitor、模型市场、Usage、Orders、经营分析和分销页面在 1440×900 与 390×844 下无横向溢出、页面异常、控制台或接口错误；Channel Monitor 另行在 V1/V2 两种模式下各复核两种视口，测试后设置恢复为 `enabled=false, mode=v1`。

---

## USA0 v1.0.16（发布准备）

- 发布版本：v1.0.16
- 官方基线版本：v0.1.170（`c043c24774228ba891ddf90d783aa6dc7d0855b5`）
- 上一 USA0 版本：v1.0.15
- 修复提交：`0dbb47bf65f675a427366c9dc0923b12629d9d07`
- 发布范围：修复模型监控分组配置写入 PostgreSQL 时 `interval_seconds` 参数类型推断不一致的问题；对时间片对齐表达式显式转换为 `integer` 与 `double precision`，并增加回归测试，避免主动检测配置保存失败
- 发布状态：已合入并推送 `usa0/main`，已创建并推送 v1.0.16 annotated tag；Release workflow 已自动触发 [31190985507](https://github.com/0xHyperia/sub2api/actions/runs/31190985507)，按要求不等待其完成
- 记录日期：2026-08-07

## USA0 v1.0.15

- 发布版本：v1.0.15
- 官方基线版本：v0.1.170（`c043c24774228ba891ddf90d783aa6dc7d0855b5`）
- 上一 USA0 版本：v1.0.14
- 上游集成提交：`5ec957fa171e7f4ab05b8e975429f270ae5bb12d`
- 发布范围：模型广场与模型监控整体 TPS、TTFT、平均耗时和成功率改为有效分组等权汇总；主动检测改为自然时间片调度，新增流量窗口跳过、失败补偿、多实例条件认领、下架模型筛选和对应管理界面，并增加迁移 `211_model_monitor_aligned_probe_compensation.sql`
- 发布状态：已合入并推送 `usa0/main`，已创建并推送 v1.0.15 annotated tag；通过 `workflow_dispatch` 执行 [Release workflow 31176639187](https://github.com/0xHyperia/sub2api/actions/runs/31176639187) 成功，GitHub Release 已发布并上传 darwin amd64/arm64、linux amd64/arm64、Windows amd64 资产及 checksums
- 记录日期：2026-08-07

## USA0 v1.0.14

- 发布版本：v1.0.14
- 官方基线版本：v0.1.170（`c043c24774228ba891ddf90d783aa6dc7d0855b5`）
- 上一 USA0 版本：v1.0.13
- 上游集成提交：`5ec957fa171e7f4ab05b8e975429f270ae5bb12d`
- 发布范围：优化模型广场成功率状态与分组性能展示，后台模型监控新增模型及分组成功/失败统计并修正表格布局；订阅模型用量估算统一复用系统实时汇率
- 发布状态：已合入并推送 `usa0/main`，已创建并推送 v1.0.14 annotated tag；按要求未持续查询 Release workflow 状态
- 记录日期：2026-08-06

## USA0 v1.0.13

- 发布版本：v1.0.13
- 官方基线版本：v0.1.170（`c043c24774228ba891ddf90d783aa6dc7d0855b5`）
- 上一 USA0 版本：v1.0.12
- 上游集成提交：`5ec957fa171e7f4ab05b8e975429f270ae5bb12d`
- 发布范围：同步分组利润控制、多平台上游倍率探测、批量账号管理和网关稳定性修复；重构模型监控与模型广场性能体验，适配生图模型按次价格，并新增可复用的 USD/CNY 汇率配置与自动同步
- 发布状态：已合入并推送 `usa0/main`，已创建并推送 v1.0.13 annotated tag；按要求未持续查询 Release workflow 状态
- 记录日期：2026-08-05

## 官方 v0.1.170 同步（已合并）

- 集成分支：`codex/merge-upstream-v0.1.170`
- 官方目标版本：v0.1.170
- 官方目标提交：`c043c24774228ba891ddf90d783aa6dc7d0855b5`
- 上一官方基线：v0.1.169（`26d894ef4f50645a4bf1030e378ac892f17d0223`）
- 集成前 USA0 提交：`bac85b42929f1b5e655639f4764f3c98bf1e69ae`
- 集成提交：`5ec957fa171e7f4ab05b8e975429f270ae5bb12d`
- USA0 发布版本：v1.0.13
- 同步状态：已完成上游合并、冲突解决、USA0 适配和代码验证，已合入并推送 `usa0/main`，并创建、推送 USA0 v1.0.13 tag
- 记录日期：2026-08-04

### 同步范围

- 从官方 v0.1.169 同步至 v0.1.170，共纳入 62 个上游提交、242 个官方变更文件。
- 新增分组级利润控制，按最低利润率与安全缓冲过滤 Token 调度候选，并在槽位获取后复核；提供 `profit-preview` 离线预演工具，组合分组和非 Token 计费路径保持不受门控影响。
- 上游计费倍率探测扩展至全部 API Key 平台，并可按账号自动写回倍率；同步托管账号禁止手工或批量修改倍率，写回具有值域校验、审计日志和探测快照。
- 管理端账号支持按筛选结果全选，批量删除增加并发限制；首页增加精简展示预设，内容审核支持代理，提示词安全审计增加“仅审计最新输入”范围。
- 修复 Anthropic 中断流部分用量漏计费、OpenAI WebSocket 关闭帧竞态、流内限流与容量错误重试、Codex 命名空间工具和加密压缩恢复，并完善 Grok SSE、订阅配额窗口、图片任务及 SMTP 行为。
- 修复保存设置时清空可见支付方式和支付方式选择器溢出，优化 Model Plaza 筛选与价格表布局，并更新 Codex Auto-review 价格数据。

### 保留与适配的 USA0 内容

- 22 个文本冲突均按业务语义合并；完整保留 ZeroAgent 应用授权、App JWT、OAuth grant 与设备会话，不恢复已经移除的 ZeroBox 客户端兼容。
- 保留用户模型市场 `/model-marketplace`，官方 Model Plaza 继续独立使用 `/model-plaza`；首页保留 USA0 `HomeExperiment`、HTML 净化、iframe sandbox 和失败恢复，仅接入上游精简首页模式。
- 分组利润字段接入 Ent、调度、缓存和管理界面，但从公开分组响应隐藏，避免结合用户倍率反推成本阈值；上游迁移 `192`、`193` 因已发布编号冲突顺延为 `209`、`210`，认证缓存触发器保留 USA0 图片权限等既有失效条件。
- 倍率探测覆盖全部 API Key 平台，同时与 USA0 账号编辑、批量编辑和倍率同步状态兼容；上游新增界面使用 USA0 语义设计令牌、响应式布局和共享交互组件。
- 账号批量删除保留 USA0 共享确认框，执行改用上游 `batchDelete` API；支付设置采用上游字段级 PATCH 语义，同时保留 USA0 页面可见性、手续费模式、渠道费率和支付失败恢复。
- 设置严格响应契约补齐购买订阅、汇率同步和模型市场性能图字段；Wire 重新生成后继续包含 ZeroAgent、工单、分销、模型监控、汇率同步与上游利润控制依赖。
- USA0 `VERSION` 已在发布准备提交中更新为 `1.0.13`。

### 验证结果

- `go generate ./ent` 与 `go generate ./cmd/server` 通过且生成哈希无漂移；后端 `go test ./...`、`go test -tags=unit ./...` 通过，`golangci-lint v2.9 run ./... --timeout=30m` 为 0 issues。
- 前端 `pnpm run lint:check`、`pnpm run typecheck` 和生产构建通过；全量 Vitest 共 298 个测试文件、1880 项测试通过，构建仅有既有动态导入与大 chunk 警告。
- `git diff --check` 通过，无未解决 Git 或源码冲突标记；迁移 `176` 至 `210` 前缀唯一，Ent/Wire 生成结果未发现额外漂移。
- 本次未替换本地 Docker 服务，也未执行登录态浏览器与真实数据库迁移复核；不得将运行时健康、桌面/移动视觉或生产发布表述为已验证。

---

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
