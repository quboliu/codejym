# 填空练习设计

本文档描述 CodeJYM 的填空练习第一版设计。领域术语以根目录 [CONTEXT.md](../../CONTEXT.md) 为准；关键架构边界见相关 ADR。

## 目标

在现有描红练习之外，引入独立的填空练习模式。用户仍从训练组和文件进入练习，但可以在同一个练习页选择“描红”或“填空”。

填空练习基于可复用的填空模板。模板由本地规则或 LLM 辅助生成，但是否入池由系统验收决定。模板池按源文件内容版本复用，达到上限后不再生成新模板。

## 非目标

- 第一版不做手动重生成模板池。
- 第一版不做模板列表管理 UI。
- 第一版不做多 attempt 历史。
- 第一版不让前端直接调用 LLM。
- 第一版不把 LLM 输出当作最终事实来源。
- 第一版不做语义等价判题，只做源码精确匹配。

## 相关 ADR

- [ADR 0001: Model Source Access for Fill-in Template Generation](../adr/0001-model-source-access-for-fill-in-template-generation.md)
- [ADR 0002: In-Process Language Analysis for Fill-in Checks](../adr/0002-in-process-language-analysis-for-fill-in-checks.md)
- [ADR 0003: Serialize Fill-in Template Replenishment per Source File Version](../adr/0003-serialize-fill-in-template-replenishment-per-source-file-version.md)
- [ADR 0004: Server-side Model Credentials](../adr/0004-server-side-model-credentials.md)

## 核心模型

### 练习模式

填空练习是独立练习模式，不是描红练习的显示开关。

- 描红练习按完整源码逐字推进 cursor。
- 填空练习按空输入和按空校验。
- 两个模式共享文件选择、布局和练习页入口。
- 两个模式拥有不同的会话模型。

### 源文件内容版本

填空模板绑定到源文件内容版本，而不是只绑定路径。

建议身份：

```text
userId + assetId + relPath + contentHash
```

同一路径内容变更后进入新的模板池。旧模板不应自动适用于新内容。

### 模板池

每个源文件内容版本拥有一个模板池。

规则：

- 默认模板上限为 8。
- 进入填空练习时懒生成，不在上传时批量生成。
- 模板池为空时，进入流程等待第一个可用模板。
- 模板池已有 active 模板时，先进入练习，再后台补齐。
- 一次补齐最多 3 次候选尝试。
- 达到上限后停止 LLM 调用和本地补齐。
- 同一个模板池同一时间只允许一个补齐任务。

## 模板生命周期

模板状态：

- `candidate`: 候选或待验收，不给用户练。
- `active`: 已通过验收，可被自动选择。
- `retired`: 不再用于新练习，但已有会话仍可查看和恢复。

不直接删除低质量模板，以免破坏历史会话。

## 模板数据结构

模板级字段建议：

```ts
interface FillInTemplate {
  id: string
  userId: string
  assetId: string
  relPath: string
  contentHash: string
  language: string
  status: 'candidate' | 'active' | 'retired'
  difficulty: 'easy' | 'medium' | 'hard'
  intent: string
  generationMethod: 'model' | 'fallback'
  promptVersion?: string
  provider?: string
  model?: string
  scores: TemplateScores
  audit: TemplateAuditTrail
  blanks: FillInBlank[]
  createdAt: string
  updatedAt: string
}
```

空级字段建议：

```ts
interface FillInBlank {
  id: string
  startOffset: number
  endOffset: number
  answer: string
  lineStart: number
  lineEnd: number
  kind:
    | 'identifier'
    | 'call_argument'
    | 'condition'
    | 'field_access'
    | 'literal'
    | 'return_value'
    | 'error_handling'
    | 'comment'
    | 'other'
  valueScore: number
  difficultyContribution: number
  hint?: string
  rationale?: string
}
```

不保存跨语言不稳定的完整 AST。通用 offsets、kind、scores、rationale 足够支持渲染、校验、审计和去重。

## 空的设计原则

一个空应覆盖连续语义片段，而不是随机字符碎片。

高价值挖空对象：

- 有意义的函数名、变量名、类型名、接口名。
- API 调用和关键参数。
- 条件表达式，而不是 `if` 本身。
- 循环边界和迭代表达式，而不是 `for` 本身。
- 错误处理关键路径。
- 字段访问链和结构体字段。
- 自定义类型、泛型参数、接口约束。
- 关键字面量，如状态码、配置 key、正则、SQL 片段、协议常量。

低价值挖空对象：

- 单独的语言关键字，如 `if`、`for`、`return`、`func`、`int`。
- 纯标点、括号、分号、逗号。
- 空白、缩进、换行。
- 太短且无语义的 token。
- 机械样板代码。

## 模板生成

### 进入填空练习

`Enter Fill-in Practice` 是一个后端业务动作。它负责：

1. 读取当前源文件内容版本。
2. 获取或创建模板池。
3. 如果需要，触发模板补齐。
4. 自动选择一个 active 模板。
5. 获取或创建用户对该模板的填空会话。
6. 返回可渲染练习数据，不返回未揭示答案。

### 本地规则降级

LLM 不可用或候选全部拒绝时，系统必须仍可生成基础模板。

本地规则优先挖：

- 长标识符。
- 函数调用参数。
- 字符串或数字常量。
- 条件表达式中的关键比较项。
- 字段访问链。

降级模板也必须通过练习价值检查和多样性检查。

### LLM 辅助生成

LLM 只产出结构化候选模板，不直接入池。

要求：

- 输出严格 JSON。
- blanks 必须使用源码 offset。
- offset 必须能精确还原原文片段。
- range 不能越界、不能重叠、不能破坏 UTF-8 边界。
- 输出包含 intent、difficulty、value rationale、avoidance rationale、self scores。

后端对 JSON 做 schema 校验和源文校验。不接受自然语言方案作为模板。

### 模型上下文

后端可以把当前源文件内容版本发送给模型，但必须满足：

- 只发送当前文件，不发送整仓库。
- 不发送用户凭证、token、无关文件树。
- 对超大文件做截断或分块。
- 可配置关闭模型源码访问。

历史模板只提供摘要，不提供完整历史答案。

历史摘要可包含：

- template id。
- blanks 的行号范围。
- blank kind 分布。
- intent。
- difficulty。
- 覆盖特征。
- 相似度摘要。

## 模型配置

第一版支持用户级全局模型配置。

默认 provider：

- DeepSeek。

支持目标：

- DeepSeek。
- OpenAI-compatible GPT。
- Anthropic Claude。

配置入口放在用户设置弹窗或设置区，不放进练习主流程。

设置项：

- provider。
- model。
- base URL。
- API key。
- 是否允许将当前文件内容发送给模型。
- 测试连接。
- 删除 key。

密钥规则：

- 开发或部署默认 key 只能来自外部 secret 配置。
- 默认 key 不进入 Git 仓库。
- 用户自带 key 服务端加密保存。
- 前端只显示 masked key。
- 模型调用只发生在后端。
- 没有用户 key 且没有外部默认 key 时，只使用本地降级模板。

## 模板验收

模板入池使用硬门槛加加权评分。

### 硬门槛

硬门槛失败时直接拒绝，LLM 高分不能覆盖。

硬门槛包括：

- JSON schema 合法。
- offset 合法且能精确匹配源码片段。
- range 不重叠、不越界。
- UTF-8 边界合法。
- 挖空比例合理。
- 空数量合理。
- 单个空长度合理。
- 不以空白或标点为主。
- 练习价值最低分达标。
- 与历史模板最高相似度不超过阈值。

### 练习价值检查

练习价值检查用于拒绝低价值挖空。

第一版应语言感知优先，文本降级。

语言感知第一批覆盖：

- Go。
- TypeScript。
- JavaScript。
- Python。
- Rust。

其他语言使用通用文本和 token 规则，再结合 LLM 评分。

### 多样性检查

多样性检查用于避免模板池重复。

指标包括：

- 字符 range overlap 或 Jaccard。
- 行集合 overlap。
- blank kind 分布差异。
- 空长度分布差异。
- 结构目标差异，如函数签名、调用表达式、条件、错误处理、注释。
- 难度密度差异。
- LLM 对语义重复度和练习目标差异的结构化评分。

LLM 可以参与评分，但系统固定阈值决定是否通过。

### 加权评分

硬门槛通过后再算加权分。

建议初始权重：

- 系统练习价值分：35%。
- 系统多样性分：25%。
- LLM 练习价值分：20%。
- LLM 多样性分：10%。
- 难度合理性分：10%。

总分低于阈值拒绝。

### 审计信息

每次候选尝试都应保存审计信息。

审计内容：

- generation method。
- provider 和 model。
- prompt version。
- content hash。
- LLM 原始结构化输出。
- 系统计算分。
- LLM 自评分。
- 最相似历史模板和相似度。
- 验收结果。
- 拒绝原因。
- accepted/rejected 时间。

第一版不需要向用户展示审计信息。

## 会话模型

第一版每个用户对每个填空模板只有一条可恢复填空会话。

会话记录：

- 当前每个空的输入。
- 每个空状态。
- 错误次数。
- 是否看过答案。
- 完成状态。
- 更新时间。

空状态建议：

- `empty`
- `incorrect`
- `correct`
- `revealed`

练习完成结果：

- `independent_completion`: 所有空都自己答对。
- `assisted_completion`: 至少一个空看过答案。

`revealed` 不算自主答对。自动选模板时，revealed 多或错误多的模板应更容易再次安排。

## 答案校验

填空答案采用源码精确匹配。

规则：

- 后端保存 blank answer。
- 前端默认不拿未揭示答案。
- 前端提交 blank input 到后端。
- 后端进行精确匹配。
- 语义等价不算对。
- 大小写、标点、空格默认都必须一致。
- UI 可以提示差异，但不能让前端拥有完整答案。

显示答案：

- 通过 reveal API 返回单个 blank answer。
- 后端记录该空为 revealed。
- 该空不算独立答对。

## 模板选择

进入填空练习时第一版自动选择模板，不先做模板列表。

选择优先级：

1. 用户未完成过的 active 模板。
2. 错误率高、revealed 多或学习需求高的模板。
3. 最久未练的模板。
4. 已完成模板中的 least-recently practiced。

用户可以“换一个模板”，但只在已有 active 模板里切换，不触发新生成。

## API 设计

### 进入填空练习

```http
POST /api/fill-in/enter
```

请求：

```json
{
  "assetId": "asset_123",
  "path": "src/main.go"
}
```

响应不包含未揭示答案：

```json
{
  "template": {
    "id": "tmpl_123",
    "difficulty": "medium",
    "intent": "练习错误处理和 API 调用参数",
    "generationMethod": "model",
    "provider": "deepseek"
  },
  "source": {
    "assetId": "asset_123",
    "path": "src/main.go",
    "language": "go",
    "content": "..."
  },
  "blanks": [
    {
      "id": "blank_1",
      "startOffset": 120,
      "endOffset": 141,
      "lineStart": 8,
      "lineEnd": 8,
      "kind": "condition",
      "hint": "错误分支判断",
      "status": "empty",
      "currentInput": "",
      "errorCount": 0,
      "revealed": false
    }
  ],
  "session": {
    "id": "fill_sess_123",
    "status": "in_progress",
    "completedBlanks": 0,
    "totalBlanks": 1
  }
}
```

### 提交答案

```http
POST /api/fill-in/sessions/{sessionId}/answers/{blankId}
```

请求：

```json
{
  "input": "err != nil"
}
```

响应：

```json
{
  "blankId": "blank_1",
  "correct": true,
  "status": "correct",
  "errorCount": 0,
  "sessionStatus": "in_progress"
}
```

错误时不返回答案。

### 显示答案

```http
POST /api/fill-in/sessions/{sessionId}/blanks/{blankId}/reveal
```

响应：

```json
{
  "blankId": "blank_1",
  "answer": "err != nil",
  "status": "revealed"
}
```

### 重置进度

```http
POST /api/fill-in/sessions/{sessionId}/reset
```

### 换一个模板

```http
POST /api/fill-in/sessions/{sessionId}/switch-template
```

只在已有 active 模板中切换，不触发新模板生成。

## 持久化建议

表或等价结构建议：

- `source_file_versions`
  - user_id, asset_id, rel_path, content_hash, language, size。
- `fill_in_templates`
  - source file version, status, difficulty, intent, generation method, provider/model, scores, audit。
- `fill_in_blanks`
  - template_id, offsets, answer, line range, kind, scores, hint, rationale。
- `fill_in_sessions`
  - user_id, template_id, status, completion outcome。
- `fill_in_blank_answers`
  - session_id, blank_id, current input, status, error count, revealed flag。
- `user_model_configs`
  - provider, model, base URL, encrypted key, source access setting, masked key metadata。

实现可以用 JSONB 合并 blanks 或 answers，但必须保留可校验、可迁移、可审计的结构。

## 前端 UI

练习页增加模式切换：

- 描红。
- 填空。

填空区：

- 显示源代码上下文。
- 被挖空的语义片段显示为输入框。
- Tab 和 Shift+Tab 在空之间移动。
- Enter 校验当前空。
- 错误时保留输入。
- 支持显示答案。
- 支持重置当前模板。
- 支持换一个模板。

练习页轻量显示：

- 模板难度。
- 模板目标。
- 模板来源：本地 / DeepSeek / GPT / Claude。
- 当前完成进度。

模型配置放在用户设置，不放进练习主流程。

## 实施切片

### 1. 数据模型和 API 骨架

- 增加源文件内容版本 hash。
- 增加模板池、模板、空、填空会话。
- 实现 enter、answer、reveal、reset、switch-template。
- 先不接 LLM。

### 2. 本地规则模板生成

- 实现 fallback template。
- 实现语言感知检查骨架和文本降级。
- 至少跑通 Go、TypeScript/JavaScript、Python、Rust 的第一版识别路径或降级路径。
- 打通完整填空体验。

### 3. 前端填空模式 UI

- 练习页增加模式切换。
- 实现空输入、提交、显示答案、重置、换模板。
- 保持描红模式不受影响。

### 4. 可配置 LLM 供应商

- 用户设置页增加模型配置。
- 默认 provider 为 DeepSeek。
- 支持 OpenAI-compatible 和 Anthropic Claude。
- 用户 key 服务端加密保存。
- 开发或部署默认 key 只来自外部 secret，不进入仓库。
- 后端统一调用模型。

### 5. LLM 候选模板生成和验收

- 严格结构化 JSON 输出。
- prompt 版本管理。
- 模型源码访问开关。
- 历史模板摘要。
- 价值、多样性、难度评分。
- 审计 trail。
- 模板池补齐。

### 6. 质量调优

- 调整练习价值阈值。
- 调整多样性阈值。
- 调整加权评分。
- 优化 prompt。
- 强化 Rust、Go、TS/JS、Python 的语言感知检查。
- 优化模板选择策略。

## 风险和缓解

### 低价值模板

风险：模板挖关键字、标点或样板代码。

缓解：prompt 明确禁止，系统练习价值检查硬门槛拒绝。

### 重复模板

风险：LLM 生成看似不同但挖空位置和目标重复的模板。

缓解：历史摘要辅助模型避重，系统多样性检查硬门槛拒绝。

### LLM 不可用

风险：首次进入填空练习卡死。

缓解：本地 fallback template。

### 源码外发

风险：部署方或用户不允许把源码发给模型。

缓解：模型源码访问可关闭；没有可用 key 或关闭后走 fallback template。

### 密钥泄露

风险：用户 key 或开发 key 暴露。

缓解：key 不进 Git；前端不调用模型；用户 key 服务端加密保存；API 只返回 masked key。

### 并发生成

风险：多个请求同时补齐导致重复模板、超额模板和额外成本。

缓解：按源文件内容版本串行化补齐，模板数量检查和入池事务化。

## 待实现前需要细化

- 具体数据库迁移方案。
- 用户模型 key 的加密密钥来源和轮换策略。
- DeepSeek、OpenAI-compatible、Anthropic 的 provider adapter 接口。
- 语言感知 parser 依赖选择。
- 初始阈值和评分公式的测试样例。
- fill-in API 的鉴权、限流和错误码。
