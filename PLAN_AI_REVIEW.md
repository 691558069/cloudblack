# 离线 AI 审核系统 — 实现计划

> 创建时间：2026-06-08
> 目标：实现离线 AI 辅助审核，自动拒绝垃圾/广告，自动通过多人举报，模糊记录进人工队列

---

## 一、总体架构

```
用户提交
  ↓
[1] 本地规则拦截（字数/广告/重复） — 已有
  ↓
[2] 行为数据检查（多人举报/高频查询） — 纯 SQL
  ↓
[3] 离线时段判断（是否处于 AI 工作时段）
    ├─ 否 → 进入人工审核队列（现有流程）
    └─ 是 → 继续
  ↓
[4] AI 文本分析（异步调用 OpenAI 格式 API）
    ├─ AI 评分 0-20 + 本地规则命中 → 自动拒绝
    ├─ AI 评分 ≥85 + 行为证据充分 → 自动通过
    └─ 其他 → 标记"AI待审"，低优先级排队人工
  ↓
[5] 管理员上班后查看"AI 离线记录"页面
    ├─ 抽查 AI 自动通过的记录
    ├─ 一键纠正误判
    └─ 调整阈值
```

---

## 二、数据库变更

### 2.1 新增表：ai_settings（AI 审核配置）

```sql
CREATE TABLE IF NOT EXISTS ai_settings (
    id INTEGER PRIMARY KEY,
    enable_offline_ai INTEGER DEFAULT 0,          -- 总开关
    offline_start TEXT DEFAULT '23:00',           -- 自动开启时段
    offline_end TEXT DEFAULT '08:00',             -- 自动关闭时段
    auto_reject_confidence INTEGER DEFAULT 20,    -- 自动拒绝阈值（AI 评分 ≤ 此值）
    auto_approve_confidence INTEGER DEFAULT 85,   -- 自动通过阈值（AI 评分 ≥ 此值）
    min_ip_count INTEGER DEFAULT 3,               -- 多人举报最低独立 IP 数
    min_query_count INTEGER DEFAULT 50,           -- 高查询热度阈值
    api_provider TEXT DEFAULT 'openai',           -- 供应商标识
    api_key TEXT,                                 -- 密钥（明文存储，后台管理）
    api_model TEXT DEFAULT 'gpt-4o-mini',           -- 模型名称
    api_endpoint TEXT DEFAULT '',                 -- 自定义接口地址（空 = 官方）
    updated_at TEXT DEFAULT (datetime('now'))
);
```

**初始化：** 程序启动时自动插入 id=1 的默认记录（如果表为空）。

### 2.2 新增表：ai_review_logs（AI 审核记录）

```sql
CREATE TABLE IF NOT EXISTS ai_review_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,                    -- 关联 cloudblack_records.id
    action TEXT NOT NULL,                          -- submit / admin_add
    ai_result TEXT NOT NULL,                       -- auto_approve / auto_reject / manual_review
    ai_score INTEGER DEFAULT 0,                    -- AI 评分 0-100
    ai_reason TEXT,                                -- AI 给出的理由
    raw_response TEXT,                              -- AI 完整 JSON 返回（调试）
    behavior_ip_count INTEGER DEFAULT 0,           -- 该 QQ 72h 内独立 IP 数
    behavior_query_count INTEGER DEFAULT 0,        -- 该 QQ 30 天内查询次数
    final_status TEXT DEFAULT 'pending',           -- pending / confirmed / corrected
    corrected_by INTEGER,                          -- 管理员 ID（如果纠正）
    corrected_at TEXT,                              -- 纠正时间
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_ai_logs_record ON ai_review_logs(record_id);
CREATE INDEX IF NOT EXISTS idx_ai_logs_created ON ai_review_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_ai_logs_status ON ai_review_logs(final_status);
```

---

## 三、后端模块设计

### 3.1 新文件：ai.go（AI 调用模块）

**结构：**
```go
// AISettings 从 ai_settings 表读取的配置
type AISettings struct { ... }

// GetAISettings() 读取 ai_settings id=1
// GetDistinctIPCount(qq, hours) 统计 N 小时内独立 IP 数
// GetQueryCount(qq, days) 统计 N 天内查询次数

// BuildPrompt(...) 构建 OpenAI 格式 Prompt
// CallAIReview(...) 调用外部 API
//   - 支持自定义 endpoint
//   - 支持所有 OpenAI 格式 API
//   - 超时 10 秒
//   - 失败返回 nil，不阻塞流程

// PerformOfflineReview(recordID, qq, reason, tags, ...) 核心审核逻辑
//   - 检查时段
//   - 检查行为数据
//   - 调用 AI
//   - 决策
//   - 执行（自动通过/拒绝/标记）
//   - 写日志

// GetModelsFromAPI(endpoint, apiKey) 获取模型列表
//   - GET {endpoint}/models
//   - 返回 []string
//   - 超时 8 秒
```

### 3.2 修改文件：config.go

- 新增 `ensureColumn` 调用（如果 ai_settings / ai_review_logs 表已存在但缺少列）
- 新增 `InitAISettings()` 初始化默认配置
- 已有函数 `GetSetting` / `SetSetting` 不动

### 3.3 修改文件：api.go + web.go

**在 submit 成功写入数据库后，启动 goroutine 异步调用 AI：**
```go
go func() {
    PerformOfflineReview(recordID, qqNum, reason, tags, nickname, accounts)
}()
```

**注意：** 异步调用不阻塞用户返回，用户看到"提交成功，等待审核"。

### 3.4 修改文件：admin.go

**新增路由（Operator 级别，所有管理员可见）：**
```
GET  /admin/ai_settings          AI 设置页面
POST /admin/ai_settings          保存 AI 设置
GET  /admin/ai_review_logs       AI 离线记录列表
GET  /admin/api/models           代理获取模型列表（AJAX）
POST /admin/api/correct_ai       纠正 AI 决策（AJAX）
```

**侧边栏新增链接：** AI设置、AI离线记录

---

## 四、前端交互设计

### 4.1 AI 设置页面（/admin/ai_settings）

**表单分组：**

| 分组 | 字段 |
|------|------|
| 离线模式 | 开关、启用时段（开始-结束） |
| 阈值 | 自动拒绝阈值(0-100)、自动通过阈值(0-100)、多人举报IP数、查询热度阈值 |
| API 配置 | 供应商选择、自定义接口地址、API密钥(password)、模型 |

**模型选择器交互：**
```
页面加载时：
  检查 api_endpoint + api_key 是否已填
    ├─ 已填 → 自动 GET /admin/api/models → 填充 select 下拉框
    └─ 未填 → 显示普通 input

用户修改 API 地址/密钥后：
  点击"刷新模型列表"按钮
    ├─ 成功 → 更新 select，显示"获取成功，共 X 个模型"
    └─ 失败 → 显示 input，红色提示"无法获取，请手动输入"

下拉框始终包含一个"手动输入"选项
```

### 4.2 AI 离线记录页面（/admin/ai_review_logs）

**表格列：**
| 列 | 说明 |
|----|------|
| ID | 日志 ID |
| 时间 | created_at |
| QQ | 被审核的 QQ |
| AI 评分 | 0-100 |
| 行为数据 | "3IP/查询120次" |
| AI 决策 | auto_approve / auto_reject / manual_review |
| AI 理由 | 一句话 |
| 最终状态 | pending / confirmed / corrected |
| 操作 | 确认 / 撤销 / 查看详情 |

**筛选条件：**
- 日期范围
- AI 决策类型（自动通过/自动拒绝/转人工）
- 最终状态（待确认/已确认/已纠正）
- QQ 号搜索

**批量操作：** 勾选多条 → 批量确认 / 批量撤销

**详情弹窗：** 点击"查看详情"显示：
- 完整提交信息
- AI raw_response（JSON）
- 行为数据明细
- 管理员纠正记录

### 4.3 仪表板新增卡片

```
离线 AI 审核统计
├─ 今日自动通过: X 条
├─ 今日自动拒绝: X 条
├─ 待人工审核: X 条
└─ 待抽查: X 条
```

---

## 五、核心审核逻辑（决策矩阵）

```
AI 评分 0-20：
  → 自动拒绝（垃圾/广告/灌水）

AI 评分 21-84：
  → 转人工（AI 不确定）

AI 评分 85-100：
  ├─ 行为证据充分（IP ≥ 3 或 查询 ≥ 50）
  │   → 自动通过
  └─ 行为证据不足
      → 转人工（AI 可信但缺乏交叉验证）
```

**行为证据不足但 AI 评分高 → 不自动通过**，这是保守策略，宁可排队也不要误判。

---

## 六、Prompt 设计

```
你是一位云黑名单审核助手。你的任务是对用户提交的云黑记录进行可信度评分。

【重要原则】
- 文本可以被伪造，不要只看"写得好不好"
- 警惕模板化描述（多个提交文本高度相似）
- 警惕纯情绪发泄（"这个人是骗子"但没有具体信息）
- 警惕"听说"、"别人说的"等二手信息
- 警惕广告、网址、加群等无关内容

【提交信息】
QQ号: {qq}
严重程度: {severity}/5
标签: {tags}
原因: {reason}
关联账号: {accounts}

【评分规则】
0-20: 明显垃圾/广告/灌水/与云黑无关
21-40: 描述模糊、缺乏证据、纯情绪
41-60: 描述一般，有一定信息但不够详细
61-85: 描述详细，有具体金额/平台/时间/经过
86-100: 描述完整、结构化、有跨平台证据

【输出格式】（JSON）
{
  "score": 整数0-100,
  "result": "auto_reject" | "manual_review" | "auto_approve",
  "reason": "一句话中文解释"
}
```

---

## 七、实现进度

| # | 任务 | 文件 | 状态 | 备注 |
|---|------|------|------|------|
| 1 | 创建 ai_settings + ai_review_logs 表 | config.go | ⬜ 未开始 | |
| 2 | 初始化默认 AI 配置 | config.go | ⬜ 未开始 | |
| 3 | 创建 ai.go（AI 调用模块） | ai.go | ⬜ 未开始 | 新文件 |
| 4 | 实现 GetModelsFromAPI | ai.go | ⬜ 未开始 | |
| 5 | 实现 BuildPrompt + CallAIReview | ai.go | ⬜ 未开始 | |
| 6 | 实现 PerformOfflineReview（核心逻辑） | ai.go | ⬜ 未开始 | |
| 7 | 在 api.go submit 中异步调用 AI | api.go | ⬜ 未开始 | |
| 8 | 在 web.go submit 中异步调用 AI | web.go | ⬜ 未开始 | |
| 9 | 在 admin.go adminAddPost 中异步调用 AI | admin.go | ⬜ 未开始 | |
| 10 | 新增 AI 设置路由和页面 | admin.go | ⬜ 未开始 | |
| 11 | 新增 /admin/api/models 接口 | admin.go | ⬜ 未开始 | |
| 12 | 新增 AI 离线记录路由和页面 | admin.go | ⬜ 未开始 | |
| 13 | 新增纠正 AI 决策接口 | admin.go | ⬜ 未开始 | |
| 14 | 仪表板新增 AI 统计卡片 | admin.go | ⬜ 未开始 | |
| 15 | 侧边栏新增 AI 相关链接 | admin.go | ⬜ 未开始 | |
| 16 | 编译测试 | — | ⬜ 未开始 | |
| 17 | 本地完整测试 | — | ⬜ 未开始 | |

---

## 八、注意事项

1. **API Key 安全**：明文存储在 SQLite 中（因为是单机部署，不涉及网络传输），但后台页面输入框用 type="password"
2. **异步调用不阻塞**：AI 调用放在 goroutine 中，用户提交后立即返回
3. **失败兜底**：AI API 超时/失败/返回异常 → 记录 error 到 ai_review_logs，记录留在人工队列
4. **时段判断**：使用服务器本地时间（config.go 已设置 timezone = Asia/Shanghai）
5. **模型列表缓存**：内存缓存 5 分钟，刷新按钮强制重新获取
6. **OpenAI 格式兼容性**：请求体严格遵循 {model, messages, temperature=0.3, max_tokens=500} 格式
7. **不要在代码中写死任何 API endpoint 或 key**：全部从 ai_settings 表读取

---

## 九、完成标准

- [ ] 后台 AI 设置页面可正常保存/读取配置
- [ ] 模型列表可自动获取并下拉选择
- [ ] 离线时段内提交自动触发 AI 审核
- [ ] AI 自动拒绝的垃圾记录不进入人工队列
- [ ] AI 自动通过的记录管理员可抽查/纠正
- [ ] AI 离线记录页面支持筛选/分页/批量操作
- [ ] 编译通过，无 panic
- [ ] 本地测试：模拟提交 10 条，检查 AI 日志是否正确记录
