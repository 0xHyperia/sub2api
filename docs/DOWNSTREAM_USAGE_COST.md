# 下游对接指南：按请求查询消耗金额

本文说明如何在每次调用模型之后，查询这一次请求实际扣费的金额。适用于需要逐次对账、二次计费或向终端用户展示单次费用的下游系统。

## 1. 概述

模型调用接口（`/v1/responses`、`/v1/messages`、`/v1/chat/completions` 等）的响应体保持上游原生格式，**响应里不包含费用**。费用在响应结束后异步结算，结算完成后通过独立的查询接口获取：

```
GET /v1/usage/requests/{request_id}
```

对接只需要三步：

1. 调用模型接口，从响应头读取 `X-Client-Request-ID`。
2. 用这个 ID 调用查询接口，带上 `wait` 参数让服务端等待结算完成。
3. 读取返回的 `actual_cost`。

## 2. 认证

与模型调用使用同一个 API Key：

```
Authorization: Bearer sk-xxxx
```

也支持 `x-api-key: sk-xxxx`。只能查询**该 Key 自己**发起的请求，查询其他 Key 的请求 ID 一律返回 404。

额度已用完或已过期的 Key 仍然可以调用本接口，用来查询最后一笔扣费。

## 3. 获取请求 ID

每个模型调用的响应头都会带：

```
X-Client-Request-ID: f40afdac-4f12-453c-9899-ddb62038b021
```

流式和非流式响应都有这个头，它在响应开始时就已发出，不必等流结束。

> **注意：请求 ID 由服务端生成。** 在请求里自行设置 `X-Client-Request-ID` 不会被采用，请始终以**响应头**里的值为准。响应里还可能出现 `X-Request-ID`，它不是查询接口使用的 ID，请不要用它查询。

下游应在收到响应头时就把这个 ID 与自己的业务订单号关联保存，后续用它查费用。

## 4. 查询接口

```
GET /v1/usage/requests/{request_id}?wait=<秒>
```

| 参数 | 位置 | 说明 |
|---|---|---|
| `request_id` | 路径 | 响应头 `X-Client-Request-ID` 的值，长度不超过 128 字节 |
| `wait` | 查询串 | 可选，整数 0-10，默认 3。记录尚未结算时，服务端最多等待这么多秒，查到立即返回 |

### 成功响应（200）

```json
{
  "object": "usage_request",
  "request_id": "f40afdac-4f12-453c-9899-ddb62038b021",
  "status": "billed",
  "model": "gpt-5.6-terra",
  "actual_cost": 0.0000694,
  "unit": "USD",
  "usage": {
    "input_tokens": 23,
    "output_tokens": 54,
    "cache_creation_tokens": 0,
    "cache_read_tokens": 0
  },
  "created_at": "2026-10-09T05:12:38.44854Z"
}
```

| 字段 | 说明 |
|---|---|
| `status` | 固定为 `billed`，表示该请求已结算 |
| `model` | 计费所用模型 |
| `actual_cost` | 本次请求**实际扣费金额**，单位见 `unit`。已包含分组倍率、高峰倍率、缓存、长上下文等全部计费规则 |
| `unit` | 固定为 `USD` |
| `usage` | 本次计费的 token 明细 |
| `created_at` | 结算记录的创建时间（UTC） |

`actual_cost` 与该 Key 的实际扣费完全一致。`input_tokens` 是扣除缓存读写之后的**净输入**，所以可能小于上游响应里的 `usage.input_tokens`；缓存部分在 `cache_read_tokens` 和 `cache_creation_tokens` 中单独给出。

### 错误响应

| HTTP 状态 | 含义 | 处理建议 |
|---|---|---|
| 400 | `request_id` 为空或过长，或 `wait` 不在 0-10 | 修正参数 |
| 401 | API Key 无效 | 检查密钥 |
| 404 | 在等待时间内没有查到该请求的结算记录 | 见下文“404 怎么处理” |
| 429 | 触发认证频率限制 | 退避后重试 |
| 500 | 服务端异常 | 退避后重试 |

错误体格式：

```json
{"error": {"message": "No billed usage found for this request", "type": "not_found_error"}, "type": "error"}
```

## 5. 时序与重试

结算是异步的：响应结束后，服务端在后台计算费用、扣款并写入记录，之后才能查到。开发环境实测，模型响应结束后约 **100 毫秒**即可查到。

推荐做法：**直接使用 `wait=5`**，一次调用即可拿到结果，不需要自己写轮询。

```
GET /v1/usage/requests/{request_id}?wait=5
```

### 404 怎么处理

404 表示在等待时间内没有结算记录，可能的原因：

- 结算还没完成（高负载下可能比平时慢）；
- 这次请求没有产生计费，例如请求在上游就失败了、被拒绝，或在计费前被丢弃；
- 请求 ID 写错，或这个 ID 不属于当前 Key。

建议策略：

1. 先用 `wait=5` 查询一次。
2. 仍是 404 时，退避后再重试，例如间隔 2 秒、5 秒、10 秒各一次（每次仍带 `wait`）。
3. 请求结束后 **30 秒**仍是 404，按“该请求未扣费”处理。

对于返回了 HTTP 错误状态的模型调用，通常不会产生费用，不需要查询。

## 6. 示例

### curl

```bash
BASE=https://your-gateway.example.com
KEY=sk-xxxx

# 1. 调用模型，同时保存响应头
curl -s -D headers.txt "$BASE/v1/responses" \
  -H "Authorization: Bearer $KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-5.6-terra","input":"用一句话介绍你自己","max_output_tokens":64}'

# 2. 读取请求 ID
RID=$(grep -i '^x-client-request-id' headers.txt | awk '{print $2}' | tr -d '\r')

# 3. 查询费用（最多等待 5 秒）
curl -s "$BASE/v1/usage/requests/$RID?wait=5" \
  -H "Authorization: Bearer $KEY"
```

### Python

```python
import time
import requests

BASE = "https://your-gateway.example.com"
HEADERS = {"Authorization": "Bearer sk-xxxx"}

# 非流式调用
resp = requests.post(
    f"{BASE}/v1/responses",
    headers={**HEADERS, "Content-Type": "application/json"},
    json={"model": "gpt-5.6-terra", "input": "你好", "max_output_tokens": 64},
    timeout=120,
)
resp.raise_for_status()
request_id = resp.headers["X-Client-Request-ID"]


def get_cost(request_id: str, deadline_seconds: int = 30):
    """返回 actual_cost；超过 deadline 仍无记录返回 None（按未扣费处理）。"""
    end = time.time() + deadline_seconds
    delay = 2
    while True:
        r = requests.get(
            f"{BASE}/v1/usage/requests/{request_id}",
            params={"wait": 5},
            headers=HEADERS,
            timeout=15,
        )
        if r.status_code == 200:
            return r.json()["actual_cost"]
        if r.status_code not in (404, 429, 500, 502, 503):
            r.raise_for_status()
        if time.time() >= end:
            return None
        time.sleep(delay)
        delay = min(delay * 2, 10)


cost = get_cost(request_id)
print("本次扣费(USD):", cost)
```

### 流式调用

流式响应的响应头同样带 `X-Client-Request-ID`，在流开始时就能读到：

```python
with requests.post(
    f"{BASE}/v1/responses",
    headers={**HEADERS, "Content-Type": "application/json"},
    json={"model": "gpt-5.6-terra", "input": "你好", "stream": True},
    stream=True,
    timeout=120,
) as resp:
    request_id = resp.headers["X-Client-Request-ID"]
    for line in resp.iter_lines():
        ...  # 处理 SSE 事件

cost = get_cost(request_id)  # 流结束后再查询
```

如果客户端中途断开，上游已经产生的消耗依然会结算，仍可凭 `request_id` 查询。

## 7. 常见问题

**为什么响应里不直接返回金额？**
费用在响应结束后才会最终确定并扣款。为保证“你查到的金额”与“实际扣款金额”严格一致，金额统一通过结算记录返回。

**返回的是标准价还是实际扣费？**
是实际扣费 `actual_cost`，已经应用了你的分组倍率及各项计费规则。接口不返回标准价和倍率明细。

**同一个请求 ID 可以重复查询吗？**
可以，结果固定不变，适合幂等对账。

**一次请求在网关内部发生账号切换（failover）会怎样？**
一次对外请求只对应一条结算记录和一个 `X-Client-Request-ID`，按最终成功的那次请求计费。

**WebSocket 模式（Responses WebSocket）支持吗？**
当前不支持，请使用 HTTP 调用。

**需要批量对账怎么办？**
可以把每次调用的 `X-Client-Request-ID` 与自己的订单号一起落库，定期逐条查询。
