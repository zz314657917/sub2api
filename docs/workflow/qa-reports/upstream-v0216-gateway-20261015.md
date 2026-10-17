### PASS: upstream-v0216-gateway-20261015

Findings：未发现阻断问题。独立审阅完整六文件 diff（含新增文件）；既有测试函数全文不变。nil/non-OpenAI guard 先于自动桥接开关，显式 tools/instructions 保留，OpenAI account/channel/global 优先级不变。tool-result 的 base64/URL、文本图像顺序、名称绑定、is_error、多结果均有精确 parts 断言。精确缺缓存 403 在 policy/重试/账号写入前透传；OAuth wrapper/单元素数组、headers/requestID、2 MiB/readErr 负例和普通 403/overload 策略均覆盖。最终分支使用循环 confirmed bool，不对失败 probe 重新分类。

| 批次 | PASS 顶层 / 含子测试 | FAIL / SKIP |
| --- | --- | --- |
| canonical-default | 66 / 572 | 0 / 0 |
| canonical-packages | 262 / 387 | 0 / 0 |
| canonical-unit | 9 / 106 | 0 / 0 |
| exact-policy | 2 / 8 | 0 / 0 |
| new-exact-once | 9 / 131 | 0 / 0 |

全部命令退出 0；九个新增顶层用例分别精确执行一次，两条漏匹配 policy 用例另跑。生产 go build -p 1 ./...、六路径只读 gofmt -l、git diff --check 均 PASS。

范围：HEAD=de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d；相对 test-compat 同 base 候选，仅合同六路径不同（含一个新增文件）。其余 4665 个 tracked 文件原始 SHA-256 与 test-compat 候选逐一一致，未按换行白名单排除。test-compat 六文件前后与 seal 完全一致。QA 前后 4671 个候选文件原始 SHA-256、HEAD/index、status、untracked/文件标志和 prerequisite seal 文件均无漂移；Developer DONE 六文件哈希相符。

证据：E:\codex-runtime\upstream-v0216-20261015\gateway-qa-2026-10-17T09-31-04-579Z-results.json（精确命令/计数/逐例名字/日志路径）；E:\codex-runtime\upstream-v0216-20261015\gateway-qa-2026-10-17T09-31-04-579Z-scope-before.json、E:\codex-runtime\upstream-v0216-20261015\gateway-qa-2026-10-17T09-31-04-579Z-scope-after.json；E:\codex-runtime\upstream-v0216-20261015\gateway-qa-2026-10-17T09-31-04-579Z-full.diff。Seal：E:/codex-runtime/upstream-v0216-20261015/gateway-sealed-files.json（平面字典 exact 六路径 -> lowercase SHA-256）。

风险：证据限 canonical 本地测试/mock 转发和生产编译，不代表真实 provider、数据库或部署验收。早期 Developer fixture 编译/不支持 MiniMax 假设失败日志保留，最终候选以本轮独立证据为准。未修改业务/测试/依赖或主树源码，无 overlay/隐藏/skip、stage/commit/push/deploy/子代理；可交主控按 seal 继续集成门禁。
