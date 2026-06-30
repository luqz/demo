## 审查结论: REJECT

判定理由: 分支 13/agent/coder 在最新提交 `b78b1ac` 下累计有三项未解决问题。最关键的：`run-smoke` 入口脚本缺失——该文件在修复提交 `9aae2b6` 中恢复后，于提交 `d78fae7` 中被再次删除。README Quick Start 和 `.gitlab-ci.yml.example` 均引用 `./run-smoke`，缺失导致测试套件无法按文档方式交付。此外，`.coverage` 和 `coverage.json` 仍被 git 追踪（虽已加入 .gitignore），`tests/test_detail.py` 仍存在私有属性直接访问。QA 测试确认 7/8 用例通过、1 跳过，代码核心逻辑正确，上述问题均为可快速修复的非代码级缺陷。

---

## 审查范围：覆盖的文件数和测试数

| 维度 | 数量 |
|------|------|
| 审查分支 | 13/agent/coder (HEAD: b78b1ac) |
| 三点 diff (origin/main...HEAD) | 51 文件 (+11917/-57) |
| 源代码文件 (src/) | 7 个模块 |
| 测试文件 (tests/) | 5 个模块 |
| 配置/脚本/文档 | 15 个 |
| 测试用例总数 | 8 (TC-LOGIN-01 ~ TC-DETAIL-02) |
| QA 报告通过率 | 7 PASS / 0 FAIL / 1 SKIP / 0 ERROR |
| 代码覆盖率 | 52% (218/418 语句) |

---

## 代码审查：逐文件正确性分析、回归、测试覆盖

### src/config.py (289 行)
- 正确性: PASS。`SmokeError` 自定义异常、`extract_password()` 阅后即焚（`os.environ.pop`）、`validate_target_url()` 四层 SSRF 校验、`Config.from_env()` 环境变量加载、`Config.__repr__()` 密码脱敏（`***`）。导入已修复为 `from ipaddress import ip_address`（仅保留实际使用符号）。
- 回归风险: 低。独立模块。
- 测试覆盖: 68%。`_is_internal_ip()` 和 `_get_ip_for_host()` 0% 覆盖（需真实 DNS），符合 E2E smoke test 特征。

### src/http_client.py (219 行)
- 正确性: PASS。`HttpClient` 封装 `requests.Session`，`HttpResponse` dataclass 记录 `initial_status`/`final_status`，5 种环境错误精确分类。`set_session()` 和 `set_base_url()` 公开方法已添加，取代对私有属性的直接访问。
- 回归风险: 低。
- 测试覆盖: 71%。错误处理分支预期低覆盖。

### src/reporter.py (374 行)
- 正确性: PASS。`classify_result()` 通过 `SmokeError` 字符串匹配区分 FAIL/ERROR；`_sanitize_token()` 长度启发式正则脱敏 token（40+ 字符 base64-like）；`TerminalReporter` 5 区域 ANSI 彩色输出；`JsonReporter` schema v1.0 JSON。非 TTY 自动关闭颜色。
- 回归风险: 低。
- 测试覆盖: 28%。reporter 由 conftest.py pytest hooks 间接调用，coverage 工具难以追踪。

### src/health_check.py (137 行)
- 正确性: PASS。`check_production_guard()` 8 个非生产关键词白名单，`health_check()` HTTP GET liveness + 5 种错误分类。FORCE_PRODUCTION guard 正确。
- 回归风险: 低。
- 测试覆盖: 40%。生产 guard 和错误分支预期低覆盖。

### src/adapters/auth_strategy.py (63 行)
- 正确性: PASS。`AuthState` dataclass 和 `AuthStrategy` Protocol 定义正确。参数签名使用 `**kwargs` 过渡设计（PRD §4.1 明确）。
- 测试覆盖: 100%（Protocol 方法体 `...` 被排除）。

### src/adapters/mock_auth.py (73 行)
- 正确性: PASS。`MockAuth.authenticate()` 状态码优先判定（2xx/3xx → PASS），fallback 到关键词匹配。`EnvironmentError` 转换为 `CONNECTION_ERROR`。
- 测试覆盖: 59%。

### tests/conftest.py (360 行)
- 正确性: PASS。Session-scoped fixtures 按依赖链正确编排。`_map_test_name()` 函数已提取，消除原先两处硬编码 if-else 链重复。CI 环境 `--allow-insecure-target` 拒绝逻辑正确。Teardown 含 `config.clear_password()` 和 logout。
- 回归风险: 低。

### tests/mock_server.py (295 行)
- 正确性: PASS。实现完整 mock 契约：POST /login（有效/无效/空凭据）、POST /projects（已认证/未认证）、GET /projects/{id}、POST /logout。

### tests/test_login.py (114 行) — 4 条用例
- 正确性: PASS。TC-LOGIN-01 验证 `auth_session.is_authenticated`；TC-LOGIN-02 使用 CSPRNG 生成无效密码（`secrets.token_hex`）；TC-LOGIN-03 空凭据非 2xx 即 PASS；TC-LOGIN-04 URL 重定向检查优先于状态码。

### tests/test_project.py (82 行) — 2 条用例
- 正确性: PASS。TC-PROJECT-01 使用 `client.set_session()` 和 `client.set_base_url()` 公开方法（修复了此前私有属性访问问题）。重复导入已移除。TC-PROJECT-02 未登录创建被正确拒绝。
- 验证: 原有 C1（重复导入）、C2（私有属性访问）均已修复。

### tests/test_detail.py (84 行) — 2 条用例
- 正确性: PASS。TC-DETAIL-01 `TEST_EXISTING_PROJECT_ID` 未配置 → SKIP，404 → `SmokeError` 标记 ERROR。TC-DETAIL-02 使用保证不存在的 UUID 做未登录探测。
- 观察: 第 30-31 行仍直接访问 `client._session` 和 `client._base_url` 私有属性，未使用 `set_session()`/`set_base_url()` 公开方法。与 test_project.py 的修复不一致。

### run-smoke 入口脚本
- 状态: 缺失。分支 HEAD `b78b1ac` 中该文件不存在。提交 `9aae2b6` 恢复后，提交 `d78fae7` 再次将其删除。README 和 `.gitlab-ci.yml.example` 均引用 `./run-smoke`。

### .gitignore
- 状态: 已更新。包含 `.coverage`、`coverage.json`、`_debug_*.py`、`_verify_*.py`、`_bootstrap_*.py` 5 个新规则。修复已应用。

---

## 安全审查：逐项列出

### 密钥/凭据泄露
无发现。代码中无硬编码密码、API key 或 token。`extract_password()` 通过 `os.environ.pop()` 阅后即焚，`Config.__repr__()` 返回 `***`。安全扫描对所有源文件返回零真实密钥匹配（config.py 中的 `your-test-password` 是错误提示模板文本，test_login.py 中的 `password = "wrong-"` 是测试用的随机密码生成）。

### 硬编码密钥
无发现。

### 依赖漏洞
`requirements.txt` 使用 `pip-compile --generate-hashes` 生成，每行依赖均锁定 sha256 hash。CI 模板包含独立的 `pip-audit` job。

### 认证/授权缺陷
无发现。`AuthStrategy` Protocol 定义认证抽象层，`MockAuth` 正确实现。`auth_session` fixture 失败时下游用例自动 SKIP。

### 输入校验缺失
充分。`validate_target_url()` 四层 SSRF 防护（HTTPS 强制 / 裸 IP 拒绝 / localhost 拒绝 / 内网 CIDR 拒绝 + DNS rebinding 二次校验）。`--allow-insecure-target` 逐级放行，CI 环境代码级拒绝。`FORCE_PRODUCTION` guard 防止 CI 误打生产。

### 不安全配置
发现 2 项:
1. run-smoke 入口脚本缺失: 分支 HEAD 中 `run-smoke` 不存在。该文件在提交 `9aae2b6` 中恢复后，于提交 `d78fae7` 被再次删除。README Quick Start 和 `.gitlab-ci.yml.example` smoke-test job 均引用 `./run-smoke`，缺失导致测试套件无法按文档方式交付。
2. Build 产物仍被 git 追踪: `.coverage` 和 `coverage.json` 已加入 `.gitignore` 但仍被 git 追踪（`git ls-files` 返回非空）。需执行 `git rm --cached` 移除追踪。

---

## 修复建议

### 代码问题

1. run-smoke 入口脚本缺失 — 文件: run-smoke（缺失）— 最新提交 d78fae7 再次删除了 run-smoke（17 行 shell wrapper），README 和 CI 配置均依赖此入口。建议: 从提交 9aae2b6 恢复 run-smoke 文件，内容为 17 行 shell wrapper 使用 `set -euo pipefail` 调用 pytest。

2. tests/test_detail.py:30-31 私有属性访问 — 直接访问 `client._session = http` 和 `client._base_url = config.target_url`。建议: 改用 `client.set_session(http)` 和 `client.set_base_url(config.target_url)`，与 test_project.py 修复方案一致。

### 安全问题

1. 入口脚本缺失导致部署失败 — 文件: run-smoke（缺失）— `./run-smoke` 是 PRD §7 第一项交付物，README 和 CI 配置均依赖它。建议: 从提交 9aae2b6 恢复 run-smoke 文件。

2. Build 产物仍被 git 追踪 — 文件: .coverage, coverage.json — 虽已加入 .gitignore 但仍被 git 追踪，磁盘上存在。建议: 执行 `git rm --cached .coverage coverage.json`；从磁盘删除这 2 个文件。

---

## 测试结果

| 指标 | 值 |
|------|-----|
| 测试用例总数 | 8 |
| PASS | 7 (TC-LOGIN-01 ~ TC-DETAIL-02) |
| FAIL | 0 |
| SKIP | 1 (TC-DETAIL-01: TEST_EXISTING_PROJECT_ID not configured) |
| ERROR | 0 |
| 退出码 | 0 |
| 执行耗时 | 0.03s (mock 环境) |
| 整体代码覆盖率 | 52% (218/418 语句) |
| Lint (ruff) | 14 个 cosmetic 问题，无逻辑/安全错误 |
| 依赖安装 | 9 个包全部成功 |
| 依赖 hash 锁定 | 所有包均含 sha256 hash |
| 密钥扫描 | 零真实密钥发现 |
| 响应体 sanitize | 通过（token 字段脱敏） |
| TARGET_URL SSRF 防护 | 通过 |
| 会话吊销 | 通过（logout 端点被调用） |

测试报告来源: docs/reports/test-report.md（QA 验证，7 PASS / 0 FAIL / 1 SKIP）。

---

## 审查问题清单

- [ ] 文件: run-smoke（缺失） | 类别: security | 问题: 提交 d78fae7 再次删除了 run-smoke 入口脚本（17 行 shell wrapper），README 和 CI 配置均依赖 ./run-smoke，测试套件无法按文档执行。此前在 9aae2b6 中已恢复，被后续提交再次删除 | 建议: 从提交 9aae2b6 恢复 run-smoke 文件，内容为使用 set -euo pipefail 的 17 行 shell wrapper
- [ ] 文件: .coverage, coverage.json | 类别: security | 问题: 已加入 .gitignore 但仍被 git 追踪（git ls-files 返回非空），磁盘上存在 build 产物文件 | 建议: 执行 git rm --cached .coverage coverage.json；从磁盘删除这 2 个文件
- [ ] 文件: tests/test_detail.py:30-31 | 类别: code | 问题: 直接访问 HttpClient._session 和 _base_url 私有属性，与 test_project.py 的修复方案不一致（后者已改用 set_session() 和 set_base_url() 公开方法） | 建议: 改用 client.set_session(http) 和 client.set_base_url(config.target_url)