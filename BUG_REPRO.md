# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

航班状态流接入一个消费很慢的订阅端后，后台删除这个订阅一直不返回，同一时间的状态发布也逐步堆住；只要订阅端持续读取就不会出现，等发布请求超时后删除操作又会继续。请先不要修改代码，查清订阅注销为什么无法主动解除这段等待，给出涉及的并发时序、资源所有权和影响边界。

## 含 Bug 版本

- 仓库：VanceMichael/go-label-06
- 仓库地址：https://github.com/VanceMichael/go-label-06.git
- parent SHA：a04f5912a8557b7aa946319af333c865ce19a8b8

## 复现步骤

```bash
git clone -- https://github.com/VanceMichael/go-label-06.git bug-repro
cd bug-repro
git checkout --detach a04f5912a8557b7aa946319af333c865ce19a8b8
go test ./internal/stream -run ^TestUnsubscribeReleasesPublisherBlockedBySlowConsumer$ -count=1
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/stream -run ^TestUnsubscribeReleasesPublisherBlockedBySlowConsumer$ -count=1
--- FAIL: TestUnsubscribeReleasesPublisherBlockedBySlowConsumer (0.19s)
    lifecycle_test.go:64: unsubscribe blocked behind a stalled publisher
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/stream	0.242s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/stream -run ^TestUnsubscribeReleasesPublisherBlockedBySlowConsumer$ -count=1
--- FAIL: TestUnsubscribeReleasesPublisherBlockedBySlowConsumer (0.18s)
    lifecycle_test.go:64: unsubscribe blocked behind a stalled publisher
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/stream	0.182s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

根因结论必须准确串起慢订阅端、relay 输出阻塞、inbox 填满、发布路径持有总线读锁等待以及注销路径等待写锁后才能关闭 done 的完整时序，并解释为什么发布上下文超时后注销会恢复、正常消费时不会复现。以 TestUnsubscribeReleasesPublisherBlockedBySlowConsumer 的 -race 红测作为复核证据；目标仓库代码、测试和配置保持零改动，不得给出空泛的“channel 堵塞”结论或实施修复。
