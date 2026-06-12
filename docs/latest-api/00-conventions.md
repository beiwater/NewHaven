# 通用约定

## 认证

除少量公开接口外，最新 API 默认都需要登录后的 Bearer Token。

```http
Authorization: Bearer <token>
```

Token 通常来自登录接口返回的 `data.token`。

## 路由展示规则

为避免把文档拆成多个版本语义，本文档统一使用虚拟路由标识：

- `/V1/`：非加密 / 公共接口
- `/E1/`：加密 / 需要登录接口

文档中展示的路径仅用于阅读与检索，不代表真实后端挂载路径。

## 响应信封

成功响应一般使用统一信封：

```json
{
  "data": {
    "example": true
  },
  "error": null,
  "meta": {
    "request_id": "req-xxx"
  }
}
```

失败响应一般为：

```json
{
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "username is required",
    "details": null
  }
}
```

## 字段命名

- 旧接口与内部 DTO 中常见 `snake_case`
- 市场前端接口里部分字段使用 `camelCase`
- 同一接口内请严格沿用 schema / 示例中的字段名

## 常用错误码

| 状态码 | 错误码 | 说明 |
|--------|--------|------|
| 400 | `BAD_REQUEST` | 请求格式错误 |
| 400 | `VALIDATION_ERROR` | 字段校验失败 |
| 400 | `INSUFFICIENT_FUNDS` | 余额不足 |
| 400 | `INSUFFICIENT_INVENTORY` | 库存不足 |
| 401 | `UNAUTHORIZED` | token 缺失或无效 |
| 403 | `FORBIDDEN` | 无权访问 |
| 404 | `NOT_FOUND` | 资源不存在 |
| 409 | `CONFLICT` | 冲突，例如重复资源或状态不允许 |
| 429 | `RATE_LIMITED` | 频率限制 |
| 500 | `INTERNAL_ERROR` | 服务器内部错误 |

## 编写模板

每个接口建议统一包含：

1. 方法与路径
2. 认证要求
3. 路径参数 / 查询参数
4. 请求示例
5. 响应示例
6. 常见错误
7. 关键字段说明

