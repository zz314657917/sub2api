### PASS: video-market-official-x7-20261104

## Findings

- 本地模型广场目录已按 APIMart `original_price` 美元数值与展示倍率 7 更新；
  涉及七个视频分组、十三个模型变体、57 个规格。v15 公开目录原来启用的
  39 个规格继续启用；未启用的 Kling、Wan、Veo 保持关闭。
- 新安装默认目录与 v11-v15 旧目录迁移均覆盖；Mini、Seedance 2.5 沿用已有
  卡片和规格 ID。仅更新报价和倍率，保留已有说明、行启停、排序和绑定。
  Veo 3.1 Fast 依官方原单位显示为每次，其余视频规格为每秒。
- 新增迁移测试发现中文标题规范化可能把豆包图像分组误匹配为视频；
  已加视频类别限制并收紧标题匹配，保留失败日志，原测试断言全部通过。

## Executed Checks

- 2026-11-04 北京时间直接读取 APIMart 公开接口
  `https://apimart.ai/api/pricing/models/all`，HTTP 200；涉及十三个模型的
  `fixed_prices` 与前次快照一致。采用 `fixed_prices.items[].original_price`，
  未使用会员价格或折后价。
- Go 1.26.10，设置对应 GOROOT 与 GOTOOLCHAIN=local，执行
  `go test -mod=readonly -overlay=<evidence>/snapshot-overlay.json -tags=unit ./internal/service -run ModelMarket -count=1 -json`。
  仓库 17 个测试与 12 个子测试、额外一个隔离快照探针均通过，失败 0。
  探针通过真实 `SettingService.GetModelMarketCatalog` 读取此前 v15
  公开目录快照，不访问应用数据库。
- 运行 `node output/playwright/video-price-20261103-7f41/verify-prices.mjs`。
  实际 Go 输出的默认目录和迁移目录，各 57/57 规格逐项与最新官方原价、
  单位、倍率一致；调用现有 Vue 页面的真实展示函数验证金额。
  原目录 21 个分组的说明、启停、规格、排序、绑定均保持，非视频报价保持。
  展示计算使用账号组倍率 1；选择其他账号组时沿用页面原有额外倍率。
- `npm.cmd run test:run -- src/views/public/__tests__/ModelPlazaView.spec.ts src/__tests__/public-pages.spec.ts`：
  2 文件，16 测试通过。
- gofmt 和两源码文件的 `git diff --check` 通过。全树 diff-check 另有既有
  `knowledge/tasks/current-task.md` 两处 Markdown 行尾空白，未修改这些行。
- 00:28 北京时间核对其他 18 个现有 dirty 文件 SHA-256 与续做基线相同；
  HEAD 保持 `77b338b71929966688ddc5528783a902572346b3`。其后共享任务快照
  有并行会话更新，补丁保留新内容，只在文件开头增加本任务记录。

### 2026-11-04 提交前复核

- 用户授权“整理下 提交”。以 `516012c948539369d0f25a76c634c78f2f62a8dc`
  为基线，通过临时 Git index 构造仅包含两份报价源码/测试和本报告的候选，
  将 4928 个文件检出至独立快照，使用基线原有依赖；共享暂存区不纳入其他改动。
- 两份报价源码 SHA-256 与前次通过的版本一致，分别为 `0eafb45b`、
  `95334346`。独立快照使用 Go 1.26.10 与 `-mod=readonly -tags=unit`，
  重跑 17 项目录测试、12 个子测试及一个隔离快照探针，失败 0。
- 独立快照的两份前端公开页面测试 16/16 通过；使用本日已保存的 APIMart
  `original_price` 快照复核，默认目录及 v15 迁移目录均 57/57 报价匹配，
  迁移目录保持原有 39 项可见规格和 21 个分组配置。
- 4928 个快照文件在测试前后逐文件 SHA-256 一致；gofmt 与候选 diff-check
  通过。随后仅补充本报告，不修改已测试的业务源码。
- 提交准备证据：
  `E:/codex-runtime/sub2api/video-market-official-x7-20261104/publication-20261104/`，
  重点为 `candidate.json`、`validation-summary.json`、`go-tests.jsonl`、
  `frontend-tests.json` 和 `price-verification.json`。

## Unverified Risks

- 未部署、更新容器或写入线上配置；当前网站尚未应用本地源码。
- 未改请求扣费、账单、路由或 `/v1/models` 可用模型清单；报价调整不能
  代表实际结算逻辑也已变更。未做付费视频调用。
- 未启动浏览器；本轮展示核对为 Vue 原函数及现有 mounted 测试，
  不宣称线上浏览器验收或独立 QA。
- 默认 PATH 指向 Go 1.26.3；首次自动下载及混用 GOROOT 的验证失败属于
  工具链环境，最终使用本机 Go 1.26.10 完成。

## Recommendation

本地报价修改和提交前定向验证完成，可按用户授权创建限定三份文件的本地提交。
未推送、更新容器或部署；线上尚未应用。

证据目录：`output/playwright/video-price-20261103-7f41/`。
重点文件：`price-verification.json`、`final-validation-summary.json`、
`go-model-market-20261104.jsonl`、`apimart-catalog-20261104.json`。
