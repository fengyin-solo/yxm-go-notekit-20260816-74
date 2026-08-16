# Bug Reproduction

## Bug 是什么

笔记持久化恢复时丢弃了标签，同时没有重建 notebook 计数缓存，导致重启后的数据视图不一致。

## 如何触发

在 bug 环境中运行：

```bash
go test ./internal/store -run TestNoteStoreReloadPreservesTagsAndNotebookCount -count=20
```

## 错误信息

```text
reloaded tags = []string(nil)
```
