# Contract Review: channel-pricing-model-catalog-s122

## Decision

PASS

## Findings

- chatgpt2api 已消费标准模型目录，缺口位于 Sub2API `GatewayService.GetAvailableModels` 只聚合账号 `model_mapping`。
- `Channel.SupportedModels()` 已提供 mapping 与 pricing 并集，复用该边界可避免模型特判和重复解析定价结构。
- 将补充限制在已有可调度账号、当前 group 和当前 platform，不扩大模型路由或渠道限制语义。

## Required Evidence

- 定向服务测试覆盖 pricing-only、账号映射并集、平台隔离、通配符过滤和无账号回退。
- 现有 gateway handler 模型目录回归通过。
- broader service/handler tests、gofmt、diff 和精确路径审计通过。
