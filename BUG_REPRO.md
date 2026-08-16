# Bug Reproduction

## Bug 是什么

笔记创建和列表路径没有保留调用方传入的 context，导致已取消的请求仍会继续执行存储写入。

## 如何触发

在 bug 环境中运行：

```bash
go test ./internal/service -run TestNoteCreateHonorsCanceledContext -count=20
```

## 错误信息

```text
Create with canceled context returned <nil>, want context.Canceled
```
