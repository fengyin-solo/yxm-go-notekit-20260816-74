# Bug Reproduction

## Bug 是什么

服务层校验错误在传到 HTTP 层时丢失了可识别的错误链，状态码映射无法识别 invalid input，最终返回 500。

## 如何触发

在 bug 环境中运行：

```bash
go test ./internal/httpapi -run TestWrappedValidationErrorReturnsBadRequest -count=20
```

## 错误信息

```text
status = 500, want 400; body={"error":"invalid input: title is required"}
```
