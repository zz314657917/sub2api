### PASS: upstream-gap-fixes-20261103

## Findings

本轮以用户指定的独立最终 QA recovery 角色接手原 QA 已冻结的 v2 探针，按仓库矩阵 gpt-6.1-sol 执行；此前仅审合同及报告格式，未编写源码、Developer 测试或独立探针。本轮没有发现本批次新增缺陷。

R2 freeze SHA256：
d5c30bd3e62d46e983db13614c901d6aab049862d486f4a279facef9090b463b

父合同 b731e0c443ccc0b41f5e85f3a1b4fe8cc6ef0e09db8e3fa4b992540c84c72ce0、增补 a51b65f646f67a0a38dfda32bbdd63b2a1729c07b17ab5f20618cb79c8a9bb5c、增补独立 review c02282bfc5f3b4a07bd1138b9d05a8d93f675be1168c73a065e2d21f7bb6d227 均匹配。十一源文件与候选/QA 副本一致，原九个 R1 哈希不变；六份 v2 探针与原 manifest 一致。

两个旧测试 diff 仅包含批准的 scheduler 自愈/未来 reset 双子例，以及 settlement matrix 统计倍率增加 accountGate；原断言、循环、case 名和容差保留。

## Checks

以下六组由本轮在独立 QA 副本亲自运行，退出码全部 0，无 fail/skip：

| 运行 | PASS 叶子 |
| --- | ---: |
| 独立 service unit，含原八项兼容 guards | 73 |
| 独立 service default | 65 |
| 独立 repository unit | 12 |
| 实际 mounted Vue，--no-cache | 7 |
| 原完整 service regression unit | 671 |
| 原完整 service regression default | 620 |

Go 固定 1.26.10，GOROOT/GOTOOLCHAIN 为进程内设置，GOPROXY/GOSUMDB=off、-mod=readonly、-count=1。命令/参数/时间/退出码见 qa-preparation/probes-v2/qa-recovery-r2-*.receipt.json；本轮实际日志汇总见 qa-recovery-r2-green-results.json。

完整回归使用未过滤的 existing-regression-regex.txt。独立比较原日志：248/207 顶层声明均保留，原叶子仅 scheduler 一项拆成两项，旧 17 项全部转绿。独立案例去除 default 重复后为 84 项，加八项 guards 为 92 项。

复用并重新解析原日志、核对哈希的证据：v2 原生产 RED 为 service 22 失败、repository 4 失败、mounted 4 失败（共 30 个不同失败）；Developer 对齐测试的原生产 overlay 双标签均 17 失败。原 R1 fixture 失败、v1/v2 探针和失败日志均原样保留，未用旧 fixture 失败充当新行为通过。

复用 Controller R2 检查并核验其 27 项证据哈希：backend build/typecheck/lint/frontend build 通过；新增 Go 为 77 个不同叶子各重复三次、231 次通过；合同三份 Vue spec 为 37 通过及同一个旧 create-query 失败。本轮未重复这些构建或将其冒充亲自执行。原 v2 40 项证据哈希也全部匹配。

## 保护及未测边界

- 本轮未修改 main/候选/QA 的源码或测试，未分配角色、重查网关、读取凭据、安装依赖、提交、部署或启动浏览器。六组执行进程均已结束。
- QA 与候选代码目录逐文件核对无额外差异，额外文件仅六份冻结探针；延后的图片价格四文件不变。与 main 的额外比较差异已归因：九项仅 LF/CRLF、两项为原受保护 Pelican 脏文件，其 main 原始哈希均未变。
- 复核已有包代码前后快照无 hash/link 变化；当前核对 70,740 文件元数据及 3,187 链接，对差异项和关键测试运行包重新哈希 1,124 项，无代码漂移。仅 vue 全局类型文件 mtime 变化而 hash 相同。没有重建全量 package manifest；当前 Vitest 生成缓存与实际 v2 前后快照一致，早于本轮开始。
- 四项旧失败继续保留：TestCalculateCostWithLongContext_PreservesAndSplitsImageInputCost；TestOpenAIQuotaQueryUsageSendsCodexHeadersAndUsesEagerProxy；TestOpenAIQuotaServiceNotConfigured（nil dereference panic）；KeysView create query / does not overwrite a touched default group when rates load after the dialog opens。完整选择 GREEN 不代表广域 Go/Vue 全套 GREEN。
- 未测 live DB、provider、浏览器或真实并发负载；SQLmock/事务替身与 mounted Vue 是本地验证。main 业务尚未由本轮合入，集成后的 main 门禁由 Controller 执行。

## Verdict

PASS，范围为冻结 R2 的六项本地修复及已批准测试对齐，可交 Controller 做保护性精确集成。十一源、六探针、合同/review 及证据身份由 qa-recovery-r2-evidence.json 与 qa-recovery-r2-seal.json 绑定；不构成发布/部署或线上验收通过。
