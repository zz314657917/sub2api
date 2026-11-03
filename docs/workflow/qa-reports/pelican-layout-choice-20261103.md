### PASS: pelican-layout-choice-20261103

本地实现并验收两种鹈鹕作品展示方式。工具栏的“上下排列 / 横向浏览”
记住个人浏览器选择，首次进入或无效偏好默认上下排列。两种布局共用一份
分组、卡片和多轮结果，切换保留搜索、模型/状态/分组筛选和分页。

## Findings

未发现本次四文件增量的明确阻断问题。相对任务开始时的实际 dirty 输入，
仅修改 PelicanTestsView.vue、PelicanLazyArtwork.vue 和各自测试。
已有未提交多轮作品、共享分组排序、雷达、管理设置及计划入口均保留。

- 恢复响应式分组列，组内作品上下排列；横向保留独立滚动和导航。
- 切换只重建预览消费者，保留作品缓存及结果卡片身份，不新增 list/settings
  请求。存储读取/写入失败时仍可正常切换；不修改服务端共享设置。
- 合并同帧 scroll/resize/observer 几何工作和横向导航测量，卸载取消排队帧
  和导航计时器。隐藏页面、权限失效仍立即清空预览。
- 保留实际可见才挂载 iframe、240px 预取、客户端三并发、缓存去重和安全
  沙箱；未改 API、后端、清洗器、依赖或价格。

## Executed Checks

- 独立合同 R2 PASS：contract SHA256
  `ffd1be85520e6cc68e78cd016e55735e7f0b1129b6aab9907558d9a6aea663b0`；
  R1 两项验收命令/证据目录问题及原 FAIL 报告保留。
- 冻结1048前端输入，1044非目标原字节一致；独立 QA 绑定整个
  source-freeze.json SHA256
  `527b209ff046646afe8e3023cff001113dcaac398ef9e15cc438712e1d479e84`。
- 原始基线113项通过；Developer、独立QA、fresh main均13套件125项通过。
  三阶段 typecheck、四文件只读 ESLint、各自外部 Vite 构建通过；主目录
  四路径 git diff --check 无空白错误。
- QA 使用自身构建的静态资源：24项布局/存储/几何、97项功能、5项390px
  纵向滚动检查通过。1440/390、明暗、两布局共八组合无整页横向溢出。
  精确旧轮放大、重播、过滤、分页、15秒刷新、失败重试、403清空通过。
- 同70个lazy host的确定性事件探针：results-region同帧60次scroll的位置
  读取8400→70，window4200→70，五帧10500→350。仅证明重复几何工作减少，
  不将其称为FPS、CPU或生产速度指标。
- QA压力测试原始remote峰值6、异步网络通知峰值9的FAIL证据保留。
  逐trace归因确认原生XHR send至loadend峰值3；所有85个remote重叠区间
  中未取消XHR不超过3，59次取消均最终ERR_ABORTED/CDP canceled。
  异步通知中的成功loadend也独立核对，6项生命周期归因检查通过。
- 精确四路径集成后，12376原始文件/HEAD/暂存条目保护通过；fresh main
  测试、类型、lint、构建通过。浏览器/daemon/fixture清理为0，不触碰其他会话。
- fresh main构建与独立QA实际浏览器验收构建逐文件哈希对比，259文件完全
  一致，差异0；见 built-identity.json。

## Unverified Risks

浏览器使用本地合成多轮fixture；未部署、未更新运行中网站或容器，未执行
真实后台写入、provider/DB验收。隐藏页面用确定性事件检查。已取消请求
在远端的物理结束及浏览器通知存在延迟，不承诺远端物理并发始终不超过3。
构建既有大chunk/Browserslist/混合导入提示保留。

## Evidence / Recommendation

可继续本地评审。未提交、推送或部署。

- Independent review: ../contract-reviews/pelican-layout-choice-20261103-review.md
- Runtime: E:/codex-runtime/sub2api/pelican-layout-choice-20261103/
- Developer: developer-report.md
- Independent QA: qa/qa-report.md、qa/qa-result.json
- Fresh main: main-checks.json、main-*.log、main-dist/
- Browser images: qa/frozen-r2-vertical-1440-dark.png、
  qa/frozen-vertical-390-scrolled-dark.png 及同目录其他布局/主题截图。
