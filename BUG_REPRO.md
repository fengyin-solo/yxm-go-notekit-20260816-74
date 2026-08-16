# Bug Reproduction

## Bug 是什么

多标签筛选没有要求笔记同时包含所有请求标签，标签规范化也在创建、查询和服务层之间不一致。

## 如何触发

在 bug 环境中运行：

```bash
go test ./internal/service -run TestListRequiresEveryRequestedTag -count=20
```

## 错误信息

```text
multi-tag filter returned 2 notes
```
