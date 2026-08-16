# Bug Reproduction

## Bug 是什么

笔记本删除流程只把关联笔记移入软删除状态，没有永久清理，导致后续仍可通过 ID 或 include_deleted 查询到残留笔记。

## 如何触发

在 bug 环境中运行：

```bash
go test ./internal/service -run TestNotebookDeleteRemovesNotesPermanently -count=20
```

## 错误信息

```text
deleted notebook left note readable: <nil>
```
