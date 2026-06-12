# 用户 / 公司

本文件覆盖与玩家账号、公司资料、故事进度相关的核心接口。

## GET /E1/players/me/companies/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `MyCompaniesResponse`

### 响应示例

```json
{
  "companies": [
    {
      "id": 1,
      "name": "NewHaven Corp",
      "money": 50000,
      "level": 5
    }
  ]
}
```

### 字段说明

- `companies`: 当前登录玩家拥有的公司列表
- `money`: 公司现金余额
- `level`: 公司等级

## GET /E1/companies/{companyId}/

认证: 需要
路径参数: `companyId`
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `CompanyProfileResponse`

### 响应示例

```json
{
  "authCompany": {
    "id": 1,
    "name": "NewHaven Corp",
    "money": 50000,
    "level": 5,
    "xp": 1200
  },
  "authUser": {
    "id": 42,
    "username": "newplayer"
  },
  "levelInfo": {
    "level": 5,
    "xp": 1200,
    "xpToNext": 2000
  },
  "preferences": {
    "theme": "dark",
    "notifications": true
  }
}
```

### 常见错误

- `403 FORBIDDEN`: 访问了不属于当前账号的公司资料

## PATCH /E1/companies/me/story-progress/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `StoryProgressRequest`

### 请求示例

```json
{
  "storyId": "tutorial_intro",
  "stepId": "step_build_factory",
  "status": "completed"
}
```

### 响应示例

```json
{
  "status": "completed",
  "stepId": "step_build_factory"
}
```

### 字段说明

- `status`: `not_started` / `in_progress` / `completed` / `skipped`
- `storyId`: 故事线标识
- `stepId`: 当前步骤标识

### 常见错误

- `400 VALIDATION_ERROR`: `storyId`、`stepId` 或 `status` 不合法

