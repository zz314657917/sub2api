---
type: contract-review
status: approved
task_id: key-fast-policy-20261105
verdict: PASS
base_commit: 0f6b739881ccf9e4ad423f464aeedcc9abd7f48a
contract_sha256: 7003aeb62b2186516058dabc306c61d16edda2d6f83d26336b958afc76497e61
previous_pass_contract_sha256: f54335d5d9cc8199c6b4de1b52aa57d8193de94e19b31c6442f1894d48b1cbce
previous_failed_contract_sha256: 00ed474f56fb52cabee67237f3eb914290b6f522830dddde21b5329e896f35d1
review_round: 3
reviewer: independent-contract-reviewer
reviewer_role_model: gpt-6.1-sol (Agent Matrix requirement)
last_verified: 2026-11-05
---

### PASS: key-fast-policy-20261105

## R3 最小增补复审

**PASS。本次最小增补合理，未发现阻断。** 本轮只批准更新已存在测试的版本预期，不代替实现或独立 QA 验收。

### Findings

- 最终合同第 110 行将 `backend/internal/service/openai_policy_group_fast_20261103_test.go` 加入 Backend Developer 白名单；第 147–149 行明确仅允许 snapshot version 断言从 20 改为 21，并保留全部 group force、duplicate、permission 和其他回归断言。
- 已只读确认该测试第 141 行仍为 `require.Equal(t, 20, snapshot.Version)`；当前 `api_key_auth_cache_impl.go:17` 已将快照版本升为 21。R2 已批准成功标准第 50 行要求递增认证快照版本，使旧缓存重新加载策略。因此将该断言精确改为 `require.Equal(t, 21, snapshot.Version)` 是同步既定版本契约，不是删去断言、改为宽松匹配或掩盖功能回归。
- 该测试后续仍验证 JSON 往返、主 Group/RouteGroups 的 ForceOpenAIFast/Hydrated、LongContextPricingEnabled、旧 version=19 的拒绝以及 duplicate 保留行为。本增补不授权更改这些断言、被测业务逻辑或其他旧测试。
- 与归档 R2 合同比较，差异恰为上述一条 Allowed Path 与一条三行约束；其他合同字节完全不变。原三态、优先级、可信 context、HTTP/WS、持久化/mask/cache、托管密钥、ent、冻结与独立 QA 门禁继续有效。

### Executed Checks

- 读取当前最终合同全文及 Agent Matrix，HEAD 仍精确等于 `base_commit`。
- 归档 `E:/codex-runtime/sub2api/key-fast-policy-20261105/contract-r2-approved.md` 的实际 SHA256 为 `f54335d5d9cc8199c6b4de1b52aa57d8193de94e19b31c6442f1894d48b1cbce`。
- 在内存中移除本次两处增补，重建字节的 SHA256 与上述 R2 完全相同；没有修改或重新序列化任何合同文件。
- 当前最终合同完整 SHA256 为 `7003aeb62b2186516058dabc306c61d16edda2d6f83d26336b958afc76497e61`，包含现有 approved/PASS 元数据。
- 归档 `E:/codex-runtime/sub2api/key-fast-policy-20261105/contract-review-r2-PASS.md` 与本轮更新前报告哈希一致，实际 SHA256 为 `c22a4808310229f550625cc0577ae1919383cb7de07e996768b7de5cb477a9cf`。原 R1/R2 证据保持原位。
- 仅只读查看目标测试相关段落、版本定义及使用点；未改业务或测试文件，未运行测试、构建、数据库、浏览器或全库审查。唯一写入为本报告。

### Unverified Risks / Recommendation

- 主控报告的新 default/unit 479 通过及既有 GroupFast 失败，本轮未独立运行核验，不作为产品 QA PASS。Developer 应按此次授权仅改一个数值，并由独立 QA 重跑相关 GroupFast 与本任务回归，记录实际命令、计数和原失败证据。
- **可按 R3 合同授权 Developer 将该断言 20→21；不得改变其他断言。审批仅对 frontmatter 的完整 SHA256 `7003aeb62b2186516058dabc306c61d16edda2d6f83d26336b958afc76497e61` 有效，任何合同字节变化都须独立重审。**
- 以下 R2/R1 内容仅保留为历史记录；其中“本轮”、旧哈希和合同行号均属于对应历史轮次，当前审批以本节及 frontmatter 为准。

---

## R2 历史记录：PASS（原证据保留，当前审批由 R3 取代）

**PASS。上轮唯一阻断 R1 已关闭，没有其他已知阻断。**

- 最终合同第 108–109 行已将 `backend/internal/service/openai_gateway_chat_completions_raw.go` 与 `backend/internal/service/openai_gateway_responses_chat_fallback.go` 加入 Backend Developer 白名单。
- 第 144–145 行将这两文件的新增授权限定为从最终策略输出重取 tier metadata，禁止计费或协议重写。
- 第 65–67 行明确要求 raw Chat 与 Responses-to-raw-Chat 的流式/非流式覆盖，以及原先非 nil 的 flex/priority 替换、key/global filter 删除后从最终输出取得 metadata。因此 R1 的路径与可测试性缺口均已补齐。
- 本轮实际读取最终合同原始字节，多次 SHA256 均为 `f54335d5d9cc8199c6b4de1b52aa57d8193de94e19b31c6442f1894d48b1cbce`，包含现有 approved/PASS 元数据。
- 在内存中逆转本轮四处修订后，SHA256 精确还原为上轮 `00ed474f56fb52cabee67237f3eb914290b6f522830dddde21b5329e896f35d1`；其余已审合同字节未变。未修改或重新序列化合同。
- 六个补充 prerequisite 的候选文件 SHA256 均匹配 `untracked-prerequisites.json` 与 `before.json.untrackedPrerequisites`。它们是冻结的旧任务输入，不因本报告成为本任务交付。原编译失败日志仍保留；本轮不宣称纠正后 baseline 或产品测试已通过。
- 本轮仅复核合同修订、原始文件哈希与上述冻结输入，未重做全库审查，未运行产品测试或构建。

**审批仅对 frontmatter 所记录的完整最终合同 SHA256 有效。任何合同字节变化都须独立重审。** 可以进入该合同允许的隔离实现；本 PASS 不代表实现、独立 QA、迁移或部署通过。

首轮 FAIL 已归档至 `E:/codex-runtime/sub2api/key-fast-policy-20261105/contract-review-r1-FAIL.md`，归档文件 SHA256 为 `1a0b6716c24df2e51a582f51242c2588ccf97b04b4f1edd05ca0114c546c7e95`，本轮核对其与更新前报告逐字节一致。以下全文保留为**首轮历史记录**，其中 FAIL、待修订和旧合同行号均已被上面的本轮结论取代，不构成当前待办。

---

## 首轮历史记录：FAIL（已由本轮 PASS 取代）

## Findings

**一项阻断，当前合同不能进入实现。** 三态语义、优先级、鉴权、存储和独立验收总体可测试，但产品文件白名单不能完整实现第 5 条成功标准。本轮为合同静态审查，没有把既存代码缺口称为已运行复现。

### R1 / P1：两个真实 HTTP 分支的最终 tier 元数据修复不在 Allowed Paths

- 合同第 58–62 行要求所有 HTTP/WS 入口的出站 tier、路由 hint 和计费元数据反映最终策略输出；第 102 行仅允许修改 `backend/internal/service/openai_gateway_service.go` 这一网关产品文件，第 125–126 行禁止其他未列出的产品路径。
- `backend/internal/service/openai_gateway_chat_completions.go:118–121` 会将 APIKey 且不使用 Responses 的账号交给 `forwardAsRawChatCompletions`。这包含 `account.Platform == PlatformOpenAI` 的账号，不能作为“其他平台不受影响”排除。
- `backend/internal/service/openai_gateway_chat_completions_raw.go:82` 在执行策略前从原 body 捕获局部变量 `serviceTier`；第 107–116 行调用共享策略并采用新出站 body，却不重取 tier。第 270–272 行把旧变量传入流式/非流式处理，第 451、545 行写入结果 `ServiceTier`。因此即使在允许的共享 helper 中正确实现 key force/filter，仍会出现：
  - 客户端不带 tier、key=force、group 不强制、global=pass：出站 priority，结果 tier 仍为 nil。
  - 客户端 priority、key=filter、group 不强制、global=pass：出站没有 tier，结果仍为 priority。
  - 客户端 flex、key=force、global=pass：出站 priority，结果仍为 flex。
- `backend/internal/service/openai_gateway_service.go:2902–2903` 另有实际 Responses→raw Chat 分流。`backend/internal/service/openai_gateway_responses_chat_fallback.go:57` 同样先捕获原 tier，第 97 行才应用策略，第 105–107 行只在原值 nil 时重取。原值为 flex/default/priority 时，force 替换或 filter 删除后仍可能保留原元数据；第 320、453、470、501 行使用该变量。
- 仅修改共享 helper 无法正常更新上述调用方已经保存的局部变量；改价格、计费所有权或绕开现有兼容分流也不符合本合同。

**建议：** 由 Planner 在 Backend Developer 白名单加入以下两个现有文件，仅授权“从最终策略输出重取 tier 并传递给现有结果/计费链”的最小修改，保持价格和计费所有权不变：

- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`

验收明确覆盖这两个分支的流式与非流式，至少包含上述三种输入以及 global filter/block。断言实际 transport body 与返回/记录的 tier 一致；global block 不触发上游调用。新增测试可继续放在已经批准的 `backend/internal/service/*_key_fast_20261105_test.go`，不必扩大测试目录。本报告没有修改合同或这两个产品文件。

## Gate Checks

| 项目 | 结论与履约检查点 |
| --- | --- |
| 需求可测试性 | 三态、默认、更新省略、报错、HTTP/WS 和表单操作均可形成判定；R1 的路径缺口必须先修订。 |
| 三态与未知值 | 合同第 39–53 行明确 exact values、默认/旧空记录 passthrough、未知及非字符串 400、force 覆盖缺省及未知 tier、filter 只删归一化 fast/priority。应区分 key policy 与客户端 service_tier，不把 ultrafast 纳入 key filter。 |
| 优先级 | 顺序应为可信 key 策略 → group force 覆盖 → 原 global evaluator 最终裁决。global=pass 仍允许 group/key 作用；global filter/block 必须作用于 force 后的 priority。不得将 global pass 解释成跳过 key/group。 |
| 真实鉴权 context | `api_key_auth.go:348–358` 的 `SetAPIKeyContext` 是现有有效密钥入口，也被 Google 鉴权及模型/首响应重路由复用。新增策略必须写入 `c.Request.Context()` 的类型安全 key，不能只放 Gin 字典或从 body/header 读取。当前允许修改 middleware 与 ctxkey，边界足够。 |
| HTTP 入口 | native Responses 有 `openai_gateway_service.go:3388` 内联分支，不能只改 raw helper；OAuth passthrough、Chat→Responses、Messages 均已有共享 helper 调用。raw Chat 和 Responses→raw Chat 为 R1 漏项。路由别名复用相同 handler，但 QA 应登记覆盖映射。 |
| 两种 WS 与后续帧 | normal WS 在 `openai_ws_forwarder.go:3065` 对每轮调用共享 helper，随后采用策略输出；passthrough WS 在 adapter 第 407、668 行分别处理首帧与后续帧，第 154、182、684 行从策略后输出更新每轮 metadata。可通过允许修改的共享 helper 接入，无证据表明必须扩大这两个 WS 文件的修改范围。省略 tier 的后续 response.create 仍需执行 force；其他事件逐字节保持。 |
| 持久化与 update mask | schema、生成文件、DTO、mapper、repository、service inputs/domain 均有单一 Backend owner。`api_key_repo.go:183–204` 是鉴权专用 Select 投影，必须加入字段；普通 get/list 成功不能代替此断言。`APIKeyUpdateFields` 与 repo 条件写入必须新增独立 mask；只改 policy 不得回写 quota_used 或限流累计值。 |
| 鉴权缓存 | 当前 snapshot version=20，`applyAuthCacheEntry` 拒绝不匹配版本。合同明确递增版本，并要求 snapshot 双向往返及更新失效。现有 Update 在成功后调用 `InvalidateAuthCacheByKey`，可保持其所有权；必须实际验证 L1/L2 和旧快照 miss 后重新加载新策略。 |
| 托管密钥 | `api_key_service.go:1195–1206` 先校验 owner 和 Cafe 限制；`isCafeManagedAPIKeySafeUpdate` 第 1477–1488 行是安全更新白名单。新字段必须加入禁止普通修改的条件，同时保留原有允许的安全编辑。仅校验普通 key 的 owner 不足以覆盖此约束。 |
| ent 生成边界 | schema 与 8 个对应生成文件已列出，runtime defaults 在 `backend/ent/runtime/runtime.go`，根 `runtime.go` 只有 stitching 注释。建议追加 schema 字段以降低索引扰动；只接受该字段必要生成差异，核对 defaults/validators、scan/assign、columns、mutation、create/update/upsert 和 migrate 描述。不得运行带依赖改写的全量生成后直接交付噪声。 |
| 迁移 | 本轮数字迁移最高为 254，255 当前未占用；写前重查要求合理。这里只审 SQL/source 边界，不执行数据库迁移。 |
| 前端与独立验收 | Backend/Frontend 写集分离；mounted create/save/edit/reset、序列化、实际构建页、冻结候选、独立 QA 与最终 hash 校验均有明确要求。已准备的模拟 CRUD fixture 可用于页面验收，其结果不能代替鉴权/数据库/真实 gateway 分支测试。 |

## 建议的验收判定

以下为现有成功标准的具体化；除 R1 的两文件授权外，不要求增加产品范围。

| key | group force | global 对最终 priority 的动作 | 客户端 tier | 期望 |
| --- | --- | --- | --- | --- |
| force | false | pass | 缺省/flex/default/ultrafast/未知 | 出站 priority；metadata=priority |
| filter | false | pass | fast / priority，含大小写及首尾空白归一化 | 删除 tier；metadata 不保留 priority |
| filter | false | pass | flex/default/ultrafast/其他值 | 新 key filter 不改写；保留入口既有规范化行为 |
| filter | true | pass | 任意 | group 胜出，priority |
| force 或任意 | true 或 false | filter | 最终 tier=priority | 删除 priority，不在后续阶段重新注入 |
| force 或 group force | 任意 | block | 最终 tier=priority | 沿用既有 HTTP/WS 拒绝协议，零上游调用 |
| 任意 | 任意 | 任意 | 实际账号为非 OpenAI | 新 key 策略无作用；不改变既有其他策略行为 |

- JSON 边界应包含遗漏、合法三值、非法字符串、null、数字、布尔、数组及对象；更新遗漏保留，显式 passthrough 可从 force/filter 清回默认。不要把旧数据库空值兼容当作接收所有非法 API 输入的理由。
- 用真实 middleware→request context→gateway 数据流证明可信策略生效，并以伪造 body/header 做负向对照；覆盖重路由后 context、cache hit 与旧版本 reload。
- Cafe seat/membership 托管 key 的 policy 修改均应被现有限制拒绝，普通合法编辑行为保持；policy-only update 的 SQL mask 不包含用量累计列。
- WS 至少包含同连接首帧/多轮、省略 tier、非 response.create、global block 与取消/错误路径，验证逐轮 metadata 不沿用前一轮。
- 新测试实际执行数必须非零；`-tags unit` 与默认构建分别记录。保留并复现基线失败，不能删断言或把未匹配测试算作通过。

## Executed Checks

- 在指定候选目录读取 status 当前任务段、Agent Matrix、spec 当前功能段和合同全文；核实当前阶段为 `contract-draft`，HEAD 精确等于合同 base_commit。
- 只读检查上述 middleware、DTO、service/repository、cache、schema/generated runtime、实际网关分流与 WS metadata 链，以及前端 API/表单、package.json、Vitest 入口和现有相关测试。
- 通过多次 `Get-FileHash -Algorithm SHA256` 检查合同原始文件；包括已存在的 approved/PASS 元数据，完整 SHA256 始终为 frontmatter 所列值。
- 以进程环境 `GOROOT=E:/tools/go1.26.10/go`、`GOTOOLCHAIN=local` 执行 `E:/tools/go1.26.10/go/bin/go.exe version`，输出 `go version go1.26.10 windows/amd64`。未调用默认 Go 或下载工具链。
- 确认 `frontend/node_modules` 是指向主目录现有依赖的 junction；没有安装、更新或改写该依赖目录。
- 核对当前最高数字迁移与既有 dirty 路径清单。本轮无产品增量供实现 diff 评审；唯一写入目标为本报告。
- 未运行产品测试、构建、ent 生成、浏览器、fixture、provider 调用、数据库或迁移；本报告不授予产品 QA 或上线通过。

## Unverified Risks

- 主控提供的 before.json 全量文件/dirty-input 冻结、后续组合候选与最终交付 manifest，仍须由 Controller/独立 QA 校验；本轮没有重新执行全库审查。
- Agent Matrix 与合同要求的角色模型均为 `gpt-6.1-sol`；实际派工身份由主控核对。本报告不以同一模型名称替代角色独立性证据。
- 未来实现的缓存失效时序、数据库往返、真实 HTTP/WS 分支与页面操作尚未验收。模拟 CRUD fixture 的准备状态来自主控说明，本轮未执行。

## Recommendation / Approval

修订 R1 的 Backend allowlist 并明确两个兼容分支的最终 metadata 断言后，重新提交完整最终合同审查。其余已检查的范围没有发现新的必须修改项。

**本轮 verdict=FAIL，不批准派工。** 合同内预置的 `status: approved` / `review_verdict: PASS` 不能代替独立 PASS。frontmatter 的 `contract_sha256` 仅标识本轮被拒绝的精确合同字节；本报告不为修改后的合同自动补签。后续若判定 PASS，审批也只能对所记录的完整 SHA256 有效，任何合同字节变化均须重审。
