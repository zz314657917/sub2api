### PASS: upstream-v0216-test-compat-20261015

Findings：未发现阻断问题。独立审阅六文件完整 diff；既有测试/断言/业务输入未弱化，13 个 email 测试函数全文与 HEAD 相同。helper 按相邻 Ent fixture 的 Balance=1、Concurrency=1、RoleUser、StatusActive 初始化，保留调用输入。Update 对照生产 UserUpdateFields 逐字段更新，保留未选字段与邮箱索引；AdjustBalance/SetBalance 意外调用直接 panic。两条 proxy 更新路径及计费、countTokens 断言保留。

| 批次 | PASS 顶层 / 含子测试 | FAIL / SKIP |
| --- | --- | --- |
| impacted | 10 / 33 | 0 / 0 |
| related | 66 / 131 | 0 / 0 |
| email | 13 / 13 | 0 / 0 |
| default | 21 / 364 | 0 / 0 |
| gemini-extra | 2 / 8 | 0 / 0 |

全部合同命令独立执行退出 0；default 按授权仅运行一次 -v；补跑两个漏匹配 Gemini 用例。生产 go build -p 1 ./...、六文件只读 gofmt -l、git diff --check 全部 PASS。

范围：HEAD=de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d。4670 个 tracked 文件中，仅六个允许测试文件与 HEAD Git blob 不同，其余 4664 文件通过 git hash-object 原始或标准 clean hash 对 HEAD blob 核验（含源码/依赖/其他测试；core.autocrlf=true，分别处理既存 HEAD CRLF 和 checkout 换行转换）。QA 前后全部 4670 文件原始 SHA-256、HEAD、index 原始 SHA-256、status、untracked 和文件标志均不变。未使用 overlay、隐藏文件、skip，未修改测试或生产文件、未暂存/提交/推送/部署/开子代理。

日志/计数/命令：E:\codex-runtime\upstream-v0216-20261015\compat-qa-2026-10-17T09-09-48-044Z-results.json；范围证据：E:\codex-runtime\upstream-v0216-20261015\compat-qa-2026-10-17T09-09-48-044Z-scope-initial.json、E:\codex-runtime\upstream-v0216-20261015\compat-qa-2026-10-17T09-09-48-044Z-scope-final.json；完整 diff：E:\codex-runtime\upstream-v0216-20261015\compat-qa-2026-10-17T09-09-48-044Z-full.diff。初次核验脚本错误使用 no-filters，将既存 CRLF 判为差异，在测试前停止；修正检查脚本后重新完整执行，原失败日志保留。

Seal：E:/codex-runtime/upstream-v0216-20261015/test-compat-sealed-files.json（平面字典，仅六个 exact repo path -> lowercase SHA-256）。

补充原始字节证据：其余 4664 个 tracked 文件逐一与同 base 的 pristine gateway 原始 SHA-256 比对，零差异；见 E:/codex-runtime/upstream-v0216-20261015/compat-qa-pristine-raw-bytes-proof.json。Python 文件两树和 HEAD 的原始 SHA-256 均为 43878828651706dd6725e5d4717bb0093e83134a8a96c805ff111f9a0282ae83；两树该文件 Git 状态干净、attributes text/eol 均 unspecified、i/crlf w/crlf。详见 compat-qa-checkout-eol-proof.json；非白名单排除。前置检查误报仅属 harness，原失败日志保留。

风险：仅验收本合同 test-only 兼容前置；未扩大为完整服务回归、race、真实 provider/外部数据库或部署验收。可交主控按 seal 继续后续门禁。
