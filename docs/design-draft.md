# Design Document: add CLI (设计文档)

> 版本: 1.0-draft
> 作者: Designer
> 日期: 2026-06-30
> 状态: 草稿 (待 PM / Architect 评审)
> 项目标识: Project 28
> 前置: PRD v1.0 (Frozen)
> 参考: Project 3 (溢出检测参考), Project 27 (PRD/设计结构模板)
> 仓库: https://github.com/luqz/demo.git
> 分支: 28/delivery

---

## 0. 讨论总览

### 0.1 参与角色与轮次

| 角色 | 参与轮次 | 核心关切 |
|------|---------|---------|
| ProductManager | Kickoff (已完成) | PRD v1.0 定稿，US-05/US-06 升级为 P0 |
| Designer | 当前轮 (draft) | 基于 PRD v1.0 产出设计文档草稿 |
| Architect | 待定 | 交叉审查，产出 Tech Plan |
| Coder | 待定 | 基于设计文档编码 |

### 0.2 关键 PM 裁决对设计的影响

PRD v1.0 已在 Kickoff 中定稿，以下 PM 裁决直接影响设计文档：

| # | 裁决 | PRD 来源 | 设计影响 |
|---|------|---------|----------|
| 1 | `--help`/`--version` 为 P0，使用 `flag` 包 | PRD §2 US-05/US-06 | 引入 flag 包解析，flag.Usage 覆盖为 Designer 定义模板 |
| 2 | 错误消息使用英文 | PRD Changelog | 所有错误/用法提示使用英文，格式统一 |
| 3 | 溢出检测为强制需求 | PRD §3.1, §3.2 | 使用 int64 + 溢出检测算法，不可依赖 Go 原生静默回绕 |
| 4 | stdout 使用 `fmt.Println` | PRD Changelog | 纯数字 + 换行，无前缀/后缀 |
| 5 | 退出码简化 0/1 | PRD Changelog | 退出码 2 移除，所有错误统一退出码 1 |
| 6 | main 文件 ≤ 60 行 | PRD Changelog | 含空行和注释 |
| 7 | 输入长度限制 40 字符 | PRD Changelog | 新增校验步骤，超长输入独立报错 |
| 8 | 版本默认 `"dev"`，ldflags 注入 | PRD Changelog | `var version = "dev"` + `-ldflags "-X main.version=v1.0.0"` |
| 9 | `flag.Usage` 覆盖为 Designer 模板 | PRD Changelog | 帮助文本遵循 Appendix D 格式 |
| 10 | 推后 piped stdin 输入 | PRD D-11 | v1.0 仅接受命令行参数 |

### 0.3 开放问题

| ID | 问题 | 提出者 | 状态 |
|----|------|--------|------|
| Q-01 | input length check 在 flag parse 之前还是之后？flag 包先消费 flags，length check 应作用于位置参数。 | Designer | 待 PM/Architect 确认 |
| Q-02 | `--` 分隔符后的第一个参数为负数时，flag 包行为：`add -- -5 10` → flag.NArg()=2，argv[0]="-5", argv[1]="10"。是否需在设计中显式说明？ | Designer | 待确认（已在 help text 示例中体现） |
| Q-03 | 溢出检测 A×B 符号一致性 vs 简单边界比较？两组方法均可，待 Architect 选定具体实现路径。 | Designer | 待 Architect |

---

## 1. 设计概述

### 1.1 设计范围

`add` 是一个纯命令行工具（CLI），无 GUI、无 Web 界面、无 TUI。设计师的角色聚焦于：命令格式设计、输入输出规范、错误消息文案、帮助/版本信息文案、用户交互流程。

与 Project 3 和 Project 27 的核心差异：Project 28 首次引入 `--help`/`--version` 标志（P0）、溢出检测、参数计数区分（缺失 vs 过多独立消息）、输入长度限制。这些新增需求从 PRD 直接推导，不引入 PRD 范围外的内容。

### 1.2 设计原则

| 原则 | 说明 |
|------|------|
| 零认知负担 | 用户不需要阅读帮助文档就能正确使用。命令格式一眼看懂。 |
| 最小惊讶 | 行为符合 Unix 工具惯例：正常输出到 stdout，错误输出到 stderr，退出码 0/1。 |
| 宁可少做，不做错 | 不猜测用户意图，严格校验输入，给出精确的错误信息。 |
| 错误消息精准化 | 区分「缺失参数」「过多参数」「非法字符」「溢出」「超长」五种错误场景，每种有专属消息。 |
| 一致格式 | 所有错误输出遵循统一模板：`Error: <description>. Usage: add <int_a> <int_b>` |
| 安全优于便利 | 溢出检测为强制需求——宁可报错也不返回错误结果。 |
| 帮助即设计 | `--help` 输出是用户的"第一印象"。精心设计的帮助文本降低学习成本。 |

### 1.3 设计约束

| 约束 | 来源 | 说明 |
|------|------|------|
| 纯 CLI，无 GUI/Web/TUI | PRD §1.1 | 所有交互通过终端文本完成 |
| 使用 `flag` 包解析标志 | PRD Changelog | `--help`/`-h`、`--version`/`-v` 通过 flag 包实现 |
| `flag.Usage` 覆盖 | PRD Appendix D | 帮助文本格式由 Designer 指定 |
| 无子命令 | PRD §1.1 | 单一 `add` 命令 |
| 零外部依赖 | PRD §1.2 | 仅 Go 标准库 |
| main.go ≤ 60 行 | PRD Changelog | 含空行和注释 |
| 错误消息英文 | PRD Changelog | 目标用户为开发者 |
| 退出码 0/1 | PRD Changelog | 退出码 2 已移除 |
| stdout 仅结果 | PRD §1.2 | `fmt.Println` 输出纯数字 + 换行 |
| 无 ANSI 颜色 | 继承 Project 3/27 惯例 | 纯文本兼容所有终端 |
| 无配置文件/环境变量 | PRD §4 D-06 | 所有配置通过命令行参数 |
| 无 stdin 输入 | PRD D-11 | v1.0 仅接受命令行参数 |

### 1.4 目标用户

| 用户角色 | 典型场景 | 核心需求 |
|----------|---------|---------|
| 开发者 | 终端快速整数加法 | 避免打开计算器或 Python REPL |
| 脚本编写者 | Shell 脚本中整数加法 | 可靠、可管道化、可捕获输出的 CLI |
| CI 系统 | 构建流水线中数值计算（计数器、偏移量） | 可预测的退出码和 stdout 输出 |

---

## 2. CLI 执行体验

### 2.1 理想路径：计算成功

用户输入两个合法整数，工具输出计算结果。

```
$ add 2 3
5
$ add 0 0
0
$ add -5 10
5
$ add -- -5 10
5
$ add -7 -3
-10
$ add 2147483647 1
2147483648
$ add 9223372036854775807 0
9223372036854775807
$ add -9223372036854775808 0
-9223372036854775808
```

设计要点：
- stdout 仅输出数字 + 换行（`fmt.Println`），无前缀、无标签、无装饰
- 输出可直接被 shell 捕获：`result=$(add 2 3)`
- 当第一个参数为负数时，需使用 `--` 分隔符（`add -- -5 10`），这是 `flag` 包的标准行为

### 2.2 标志路径：帮助信息 (`--help` / `-h`)

```
$ add --help
add — integer addition tool

Usage:
  add [-h] [-v] <int_a> <int_b>

Arguments:
  <int_a>    first integer (int64 range)
  <int_b>    second integer (int64 range)

Options:
  -h, --help       show this help message
  -v, --version    show version number

Examples:
  add 2 3           outputs 5
  add -5 10         outputs 5
  add -- -5 10      use -- separator when the first argument is negative

Notes:
  Negative numbers must directly follow the minus sign (e.g., -5, not - 5).
  When the first argument is negative, use -- to separate flags from args.
```

设计要点：
- 输出至 stdout，退出码 0（非错误路径）
- `--help` 和 `-h` 行为完全相同，均由 `flag` 包自动处理
- 帮助文本通过 `flag.Usage` 覆盖实现（PRD Appendix D）
- 文本不使用 `os.Args[0]` 动态程序名——硬编码 `add`，保证 golden file 可复现

### 2.3 标志路径：版本信息 (`--version` / `-v`)

```
$ add --version          # CI 构建 (ldflags 注入)
add v1.0.0
$ add --version          # 本地 go build (无 ldflags)
add dev
```

设计要点：
- 输出至 stdout，退出码 0（非错误路径）
- `--version` 和 `-v` 行为完全相同
- 版本字符串通过 ldflags 注入：`-ldflags="-X main.version=v1.0.0"`
- 源文件中 `var version = "dev"` 作为默认值
- 格式：`add <version>`（程序名 + 空格 + 版本号），后跟换行

### 2.4 失败路径：参数缺失

当位置参数不足两个（0 个或 1 个）时：

```
$ add
Error: missing arguments, need two integers. Usage: add <int_a> <int_b>
$ add 1
Error: missing arguments, need two integers. Usage: add <int_a> <int_b>
```

设计要点：
- 输出至 stderr，退出码 1
- flag 包先消费所有标志（`--help`、`--version` 等），然后检查 `flag.NArg()`
- 缺失参数和过多参数使用不同的错误消息（精准定位问题）

### 2.5 失败路径：参数过多

当位置参数超过两个时：

```
$ add 1 2 3
Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>
```

设计要点：
- 输出至 stderr，退出码 1
- 与缺失参数场景区分，帮助用户快速定位问题

### 2.6 失败路径：输入过长

当任一参数超过 40 字符时：

```
$ add $(python3 -c 'print("1"*100)') 1
Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>
```

设计要点：
- 输出至 stderr，退出码 1
- 长度检查在整数解析之前执行（防止对超长字符串调用 ParseInt）
- 限制为 40 个字符（PRD 明确规定）
- 先报告第一个超长参数（从左到右校验）

### 2.7 失败路径：非法整数

当任一参数无法被 `strconv.ParseInt` 解析为 int64 时：

```
$ add abc 1
Error: "abc" is not a valid integer. Usage: add <int_a> <int_b>
$ add 1 xyz
Error: "xyz" is not a valid integer. Usage: add <int_a> <int_b>
$ add 1.5 2
Error: "1.5" is not a valid integer. Usage: add <int_a> <int_b>
```

设计要点：
- 输出至 stderr，退出码 1
- 无效参数用 ASCII 双引号包裹（`"abc"`），与 PRD Appendix C 一致
- 使用 Go `%q` 格式化动词或手动拼接 `"` + arg + `"`
- 仅报告第一个无效参数（从左到右校验）

### 2.8 失败路径：整数溢出

当两个 int64 整数相加结果超出 int64 范围时：

```
$ add 9223372036854775807 1
Error: integer overflow: 9223372036854775807 + 1 exceeds int64 range. Usage: add <int_a> <int_b>
$ add -9223372036854775808 -1
Error: integer overflow: -9223372036854775808 + -1 exceeds int64 range. Usage: add <int_a> <int_b>
```

设计要点：
- 输出至 stderr，退出码 1
- 溢出检测为强制需求（PRD 明确规定，与 Project 3/27 不同）
- 溢出检测算法：检查结果符号是否与两操作数一致，或使用边界前置检查
- 消息中包含具体的操作数，帮助用户理解溢出原因

### 2.9 失败路径：flag 解析错误

当传入未知标志时（`flag.ContinueOnError` 模式）：

```
$ add --unknown 1 2
（flag 包输出错误信息到 stderr，退出码 1）
```

设计要点：
- `flag.CommandLine.Init("add", flag.ContinueOnError)` —— 不调用 `os.Exit`
- flag 包自动生成错误消息，无需自定义
- 退出码 1（在 flag 解析错误后由 main 返回）

### 2.10 输出矩阵总览

| 场景 | stdout | stderr | 退出码 |
|------|--------|--------|--------|
| 正常计算 | 计算结果（纯数字 + \n） | 无 | 0 |
| `--help` / `-h` | 帮助文本（见 §2.2） | 无 | 0 |
| `--version` / `-v` | `add <version>\n` | 无 | 0 |
| 参数缺失 (0 或 1 个) | 无 | `Error: missing arguments, need two integers. Usage: add <int_a> <int_b>\n` | 1 |
| 参数过多 (≥3 个) | 无 | `Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>\n` | 1 |
| 输入过长 (>40 字符) | 无 | `Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>\n` | 1 |
| 非法整数（非数字等） | 无 | `Error: "<input>" is not a valid integer. Usage: add <int_a> <int_b>\n` | 1 |
| 整数溢出 | 无 | `Error: integer overflow: <a> + <b> exceeds int64 range. Usage: add <int_a> <int_b>\n` | 1 |
| flag 解析错误 | 无 | flag 包默认错误输出 | 1 |

### 2.11 终端输出区域结构

| 区域 | 触发条件 | 流向 | 内容 |
|------|---------|------|------|
| 结果输出 | 计算成功 | stdout | 纯数字 + \n |
| 帮助输出 | `--help`/`-h` | stdout | 结构化帮助文本 |
| 版本输出 | `--version`/`-v` | stdout | `add <version>\n` |
| 错误输出 | 校验失败 | stderr | `Error: <description>. Usage: ...` |

无分隔线、无标题、无页脚、无摘要。零装饰。

### 2.12 ANSI 颜色规范

不使用。唯一输出是纯数字（stdout）或错误消息（stderr），无颜色标记。与 PRD "无彩色输出" 惯例一致。

---

## 3. 错误处理设计

### 3.1 退出码语义

| 退出码 | 含义 | 触发条件 |
|--------|------|----------|
| 0 | 成功 | 计算完成（结果输出至 stdout）、帮助信息输出、版本信息输出 |
| 1 | 输入错误 | 参数缺失、参数过多、输入过长、非法整数、溢出、flag 解析错误 |

> 退出码 2 已从 PRD 移除，未来预留用于运行时错误。

### 3.2 错误消息模板

| 类别 | 错误消息模板 | 触发条件 | 检测方式 |
|------|------------|---------|---------|
| 参数缺失 | `Error: missing arguments, need two integers. Usage: add <int_a> <int_b>` | `flag.NArg() < 2` | flag 包 NArg() |
| 参数过多 | `Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>` | `flag.NArg() > 2` | flag 包 NArg() |
| 输入过长 | `Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>` | `len(arg) > 40` | len() |
| 非法整数 | `Error: "<input>" is not a valid integer. Usage: add <int_a> <int_b>` | `strconv.ParseInt()` 返回 error | ParseInt 返回值 |
| 整数溢出 | `Error: integer overflow: <a> + <b> exceeds int64 range. Usage: add <int_a> <int_b>` | 结果符号与操作数不一致 | 算法检测 |

### 3.3 错误消息统一格式

所有错误消息遵循统一模板：

```
Error: <specific description>. Usage: add <int_a> <int_b>
```

- 前缀 `Error: `（英文）
- 具体描述（英文，区分不同错误场景）
- 句号分隔
- `Usage: add <int_a> <int_b>` 后缀（固定，帮助用户纠正输入）
- 程序名 `add` 硬编码（不使用 `os.Args[0]`，保证 golden file 静态可复现）

### 3.4 校验优先级

校验顺序固定，遵循「结构先于内容，安全先于计算」原则：

1. **Flag 解析** — `flag.Parse()`，`flag.ContinueOnError` 模式
2. **参数计数** — `flag.NArg()`，区分 0 缺失 / 过多
3. **输入长度** — 每个参数 ≤ 40 字符
4. **整数解析** — `strconv.ParseInt(arg, 10, 64)`，检查 error
5. **溢出检测** — 加法前/后检查是否超出 int64 范围
6. **计算输出** — 执行加法，fmt.Println 输出结果

设计决策：先检查结构（个数），再检查安全（长度），再检查内容（格式），最后检查计算边界（溢出）。每步发现第一个错误即终止。

---

## 4. 用户交互流程

### 4.1 主流程图

```
用户输入命令
    │
    ▼
┌─────────────────────────────────┐
│ flag 解析                        │
│ flag.CommandLine.Init(...)      │
│ flag.Parse()                    │
│ flag.ContinueOnError            │
└───────────┬─────────────────────┘
            │
     ┌──────┴──────┐
     │ 失败        │ 成功
     ▼             ▼
┌────────────┐  ┌─────────────────────────┐
│ stderr:    │  │ --help / -h 触发了？     │
│ flag 错误   │  └───────────┬─────────────┘
│ exit 1     │              │
└────────────┘       ┌──────┴──────┐
                      │ 是          │ 否
                      ▼             ▼
              ┌────────────┐  ┌─────────────────────────┐
              │ stdout:    │  │ --version / -v 触发了？  │
              │ 帮助文本    │  └───────────┬─────────────┘
              │ exit 0     │              │
              └────────────┘       ┌──────┴──────┐
                                    │ 是          │ 否
                                    ▼             ▼
                            ┌────────────┐  ┌───────────────────────┐
                            │ stdout:    │  │ flag.NArg() == 2 ?    │
                            │ add <ver>  │  └───────────┬───────────┘
                            │ exit 0     │              │
                            └────────────┘       ┌──────┴──────┐
                                          ┌──────┴──────┐      │
                                          │ < 2 (缺失)  │      │ 是
                                          ▼             ▼      │
                                   ┌────────────┐ ┌──────┐     │
                                   │ stderr:    │ │ > 2  │     │
                                   │ missing    │ │ 过多  │     │
                                   │ arguments  │ │      │     │
                                   │ exit 1     │ └──┬───┘     │
                                   └────────────┘    │        │
                                                     ▼        │
                                              ┌────────────┐  │
                                              │ stderr:    │  │
                                              │ too many   │  │
                                              │ arguments  │  │
                                              │ exit 1     │  │
                                              └────────────┘  │
                                                              ▼
                                              ┌─────────────────────────────┐
                                              │ len(argv[0]) ≤ 40 ?         │
                                              │ len(argv[1]) ≤ 40 ?         │
                                              └───────────┬─────────────────┘
                                                          │
                                                   ┌──────┴──────┐
                                                   │ 否          │ 是
                                                   ▼             ▼
                                            ┌────────────┐  ┌─────────────────────────┐
                                            │ stderr:    │  │ ParseInt(argv[0], 10, 64)│
                                            │ input too  │  └───────────┬─────────────┘
                                            │ long       │              │
                                            │ exit 1     │       ┌──────┴──────┐
                                            └────────────┘       │ 失败        │ 成功
                                                                  ▼             ▼
                                                          ┌────────────┐  ┌─────────────────────────┐
                                                          │ stderr:    │  │ ParseInt(argv[1], 10, 64)│
                                                          │ not a valid│  └───────────┬─────────────┘
                                                          │ integer    │              │
                                                          │ exit 1     │       ┌──────┴──────┐
                                                          └────────────┘       │ 失败        │ 成功
                                                                                ▼             ▼
                                                                        ┌────────────┐  ┌─────────────────────┐
                                                                        │ stderr:    │  │ 溢出检测             │
                                                                        │ not a valid│  │ a + b 超出 int64?    │
                                                                        │ integer    │  └───────────┬─────────┘
                                                                        │ exit 1     │              │
                                                                        └────────────┘       ┌──────┴──────┐
                                                                                              │ 是          │ 否
                                                                                              ▼             ▼
                                                                                      ┌────────────┐  ┌────────────┐
                                                                                      │ stderr:    │  │ stdout:    │
                                                                                      │ overflow   │  │ a + b      │
                                                                                      │ exit 1     │  │ exit 0     │
                                                                                      └────────────┘  └────────────┘
```

### 4.2 分支数量

| 出口类型 | 数量 | 说明 |
|----------|------|------|
| 成功 (exit 0) | 3 | 正常计算、帮助信息、版本信息 |
| 失败 (exit 1) | 7 | flag 解析错误、参数缺失、参数过多、输入过长、非法整数(A)、非法整数(B)、溢出 |
| 合计 | 10 | 出口 |

---

## 5. 边缘情况设计

### 5.1 整数溢出检测

与 Project 3/27 的核心区别：Project 28 必须检测溢出并报错，不可依赖 Go 原生 `int` 静默回绕。

方案选择（待 Architect 确认实现路径）：

**方案 A — 符号一致性检测**：
```
if a > 0 && b > 0 && result < 0 → positive overflow
if a < 0 && b < 0 && result > 0 → negative overflow
```

**方案 B — 边界前置检查**：
```
if b > 0 && a > math.MaxInt64 - b → positive overflow
if b < 0 && a < math.MinInt64 - b → negative overflow
```

两种方案均需使用 `int64` 类型（非 Go `int`），通过 `strconv.ParseInt(s, 10, 64)` 解析。

### 5.2 边界值行为

| 输入 | 行为 | 说明 |
|------|------|------|
| `9223372036854775807 0` | 输出 `9223372036854775807`，exit 0 | int64 最大值 + 0 = 正常 |
| `-9223372036854775808 0` | 输出 `-9223372036854775808`，exit 0 | int64 最小值 + 0 = 正常 |
| `9223372036854775807 1` | 溢出错误，exit 1 | 正溢出 |
| `-9223372036854775808 -1` | 溢出错误，exit 1 | 负溢出 |
| `9223372036854775807 -1` | 输出 `9223372036854775806`，exit 0 | 边界正常运算 |
| `-9223372036854775808 1` | 输出 `-9223372036854775807`，exit 0 | 边界正常运算 |

### 5.3 输入长度限制

| 输入 | len | 行为 |
|------|-----|------|
| `"42"` | 2 | 通过，继续解析 |
| `"-9223372036854775808"` | 20 | 通过（int64 最小值，20 字符） |
| `"9223372036854775807"` | 19 | 通过（int64 最大值，19 字符） |
| `"1" * 41` | 41 | 拒绝：`input too long` |
| `"1" * 40` | 40 | 通过（边界值），继续解析 |

### 5.4 `--` 分隔符行为

当第一个位置参数为负数时，`flag` 包会将其误解析为 flag。解决方案：

```
$ add -5 10        # 报错：flag 包将 -5 视为未知 flag
$ add -- -5 10     # 正确：-- 后 -5 被识别为位置参数，输出 5
```

这是 `flag` 包的标准行为，在帮助文本中已说明。不需要自定义 flag 解析逻辑来绕过此限制。

### 5.5 特殊数值处理

| 输入 | ParseInt 行为 | 设计表现 |
|------|--------------|---------|
| `+5` | 解析为 `5` | 正常计算 |
| `-0` | 解析为 `0` | 正常计算 |
| `0005` | 解析为 `5` | 正常计算 |
| `0xFF` | 返回 error | `Error: "0xFF" is not a valid integer` |
| `1e3` | 返回 error | `Error: "1e3" is not a valid integer` |
| `""` (空) | 返回 error | `Error: "" is not a valid integer` |

### 5.6 跨平台换行符

- Go `fmt.Println` 在 Windows 输出 CRLF (`\r\n`)，在 Unix 输出 LF (`\n`)
- 纯数字输出无兼容性问题——shell 脚本 `$()` 捕获自动处理尾部换行符
- 帮助文本和错误消息同样使用 `fmt.Println`/`fmt.Fprintln`，换行符行为一致

---

## 6. 帮助文本设计

### 6.1 完整帮助文本

```
add — integer addition tool

Usage:
  add [-h] [-v] <int_a> <int_b>

Arguments:
  <int_a>    first integer (int64 range)
  <int_b>    second integer (int64 range)

Options:
  -h, --help       show this help message
  -v, --version    show version number

Examples:
  add 2 3           outputs 5
  add -5 10         outputs 5
  add -- -5 10      use -- separator when the first argument is negative

Notes:
  Negative numbers must directly follow the minus sign (e.g., -5, not - 5).
  When the first argument is negative, use -- to separate flags from args.
```

### 6.2 帮助文本设计原则

| 原则 | 说明 |
|------|------|
| 结构清晰 | 标题 → 用法 → 参数 → 选项 → 示例 → 注意事项，信息密度从高到低 |
| 程序名硬编码 | 使用 `add` 而非 `os.Args[0]`，保证 golden file 可复现 |
| 参数范围明确 | `int64 range` 表述，不列出具体数值范围（太冗长） |
| 示例实用 | 三个示例覆盖：正常输入、负数作为第二参数、负数作为第一参数（`--` 分隔符） |
| 注意事项简短 | 仅两条关键陷阱，不罗列 PRD 全部约束 |

### 6.3 帮助文本触发

- `add --help` → stdout 输出帮助文本，exit 0
- `add -h` → 等同于 `--help`
- flag 包自动处理 `--help`/`-h` 信号，调用 `flag.Usage` 函数

---

## 7. 版本信息设计

### 7.1 版本输出格式

```
add v1.0.0    # CI 构建（ldflags 注入）
add dev        # 本地构建（默认值）
```

格式：`add <version>\n`

- 程序名硬编码 `add`
- 空格分隔
- 版本号：CI 通过 ldflags 注入（如 `v1.0.0`），本地默认 `dev`
- 后跟换行

### 7.2 实现方式

```go
var version = "dev"

// main:
// flag.StringVar / flag.BoolVar for version flag
// On --version: fmt.Printf("add %s\n", version); os.Exit(0)
```

构建命令：`CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=v1.0.0" -trimpath -o add`

---

## 8. 文件与目录结构

### 8.1 目标目录树

```
project-28/
├── main.go              # CLI 入口及全部逻辑 (≤ 60 行)
├── add.go               # add() 函数：溢出检测 + 加法逻辑
├── add_test.go          # add() 单元测试（6 个用例）
├── main_test.go         # CLI 集成测试（覆盖全部 AC）
├── go.mod               # Go 模块定义 (go 1.21, 零外部依赖)
├── go.sum               # 校验和 (由 go mod tidy 自动生成)
├── Makefile             # 构建目标: build, test, vet, cross, clean
├── README.md            # 项目说明
└── docs/
    ├── prd.md           # 产品需求文档 (PRD v1.0)
    ├── design.md        # 设计文档 (本文件)
    ├── design-draft.md  # 设计文档草稿 (当前文件)
    └── context-audit.md # 项目上下文审计报告
```

### 8.2 关键设计选择

| 设计选择 | 理由 |
|---------|------|
| `main.go` + `add.go` 分离 | `add.go` 包含纯函数 `add(a, b int64) (int64, error)`，便于单独测试 |
| `add_test.go` 独立 | 单元测试只测 `add()` 函数，6 个用例覆盖所有溢出场景 |
| `main_test.go` 集成测试 | 通过 `exec.Command` 测试完整 CLI 行为，覆盖 PRD §3.1-3.3 全部 AC |
| 零外部依赖 | `go.mod` 仅声明 `go 1.21`，无 `require` 块 |
| Makefile | 封装 `build`/`test`/`vet`/`cross`/`clean`，PRD §E 列为 P1 交付物 |

### 8.3 与 Project 3/27 的结构差异

| 维度 | Project 3 | Project 27 | Project 28 |
|------|-----------|------------|------------|
| 源文件结构 | 单一 main.go | 单一 main.go | main.go + add.go 分离 |
| 测试文件 | main_test.go | main_test.go | add_test.go + main_test.go |
| Makefile | 无 | 无 | 有（P1 交付物） |
| main.go 行数限制 | 未明确 | ≤ 50 行 | ≤ 60 行 |

---

## 9. 配置体验

### 9.1 配置方式

无配置。工具不接受配置文件、环境变量、或额外的命令行标志（除 `--help`/`--version`）。所有行为通过位置参数控制。与 PRD §4 D-06 "无配置文件/环境变量"一致。

### 9.2 省略理由

本工具是极简 CLI 加法计算器，无需要配置的运行时行为。以下配置域在本设计中明确不适用：

| 配置域 | 省略理由 |
|--------|---------|
| 环境变量 | PRD §4 D-06 明确排除 |
| 配置文件 | PRD §4 D-06 明确排除 |
| 额外标志位 | 仅 `--help`/`--version`（P0），无其他标志 |
| 输出格式切换 | 唯一输出为纯数字，无格式变体 |
| 语言切换 | 仅英文，PRD §4 D-09 无 i18n |

---

## 10. 安全 UX 设计

### 10.1 省略理由

本工具处理纯整数运算，不涉及密码、密钥、令牌、外部网络请求、文件系统访问。以下安全域在本设计中明确不适用：

| 安全域 | 省略理由 |
|--------|---------|
| 密码/密钥生命周期 | 无敏感数据 |
| 响应体截断/脱敏 | 无外部请求 |
| 生产环境保护 | 无生产/测试环境区分 |
| 输出 sanitization | 纯数字输出，无可注入内容 |

### 10.2 安全相关设计决策

| 决策 | 说明 |
|------|------|
| 输入长度限制 40 字符 | 防止超大字符串传入 ParseInt，避免潜在 DoS（虽然 Go 标准库已足够健壮） |
| 溢出检测 | 保证计算正确性——不会静默返回错误结果，这在 CI 计数器和 shell 脚本中至关重要 |
| `strconv.ParseInt` 而非 `Atoi` | `ParseInt` 明确限制 bitSize=64，避免 `int` 类型在 32 位平台的范围差异 |

---

## 11. 移除的功能

以下功能在 PRD 中明确排除，未纳入本设计：

| 排除项 | 原因 |
|--------|------|
| 浮点数/小数加法 (D-01) | 仅整数 |
| 交互模式 / REPL (D-02) | 纯 CLI，单次调用 |
| 减法/乘法/除法 (D-03) | 单一职责 |
| 三个及以上操作数 (D-04) | 严格两个整数 |
| 结果格式化 / 逗号/十六进制 (D-05) | 纯十进制 |
| 配置文件 / 环境变量 (D-06) | 零配置 |
| 大整数 / 任意精度 (D-07) | int64 限制 |
| 文件输入 (D-08) | 仅命令行参数 |
| 本地化 / i18n (D-09) | 仅英文 |
| 安装脚本 / 包管理发布 (D-10) | 仅源码 |
| stdin 管道输入 (D-11) | 推后至 v1.1 |

---

## 12. 设计决策日志

| ID | 决策 | 理由 | 来源 |
|----|------|------|------|
| D-01 | 使用 `flag` 包实现 `--help`/`--version` | PRD US-05/US-06 升级为 P0，Changelog 明确要求 `flag` 包 | PM Kickoff (PRD) |
| D-02 | `flag.ContinueOnError` 模式 | 不调用 `os.Exit(2)`，统一退出码 0/1 | PM Kickoff (PRD §3.5) |
| D-03 | `flag.Usage` 覆盖为 Designer 模板 | PRD Appendix D 明确指定帮助文本格式 | PM Kickoff (PRD Appendix D) |
| D-04 | 版本默认 `"dev"`，ldflags 注入 | PRD Changelog 明确要求 | PM Kickoff (PRD) |
| D-05 | 缺失参数和过多参数使用不同错误消息 | 相较于 Project 3/27 的单一用法提示更精准，帮助用户快速定位 | Designer (从 PRD AC 推导) |
| D-06 | 所有错误消息带 `Usage: add <int_a> <int_b>` 后缀 | PRD Appendix C 统一格式 | PM Kickoff (PRD §C) |
| D-07 | 错误消息中程序名硬编码 `add` | 保证 golden file 静态可复现，避免 `go run` 临时路径干扰 | Designer (继承 P27 D-02) |
| D-08 | stdout 仅输出数字 | 使输出可被 shell 脚本直接捕获 | Designer (遵循 Unix 惯例) |
| D-09 | stderr 承载所有错误输出 | Unix 惯例 | Designer |
| D-10 | 退出码 0/1 二进制 | PRD 明确移除退出码 2 | PM Kickoff (PRD §3.5) |
| D-11 | 输入长度限制 40 字符 | PRD Changelog 明确要求 | PM Kickoff (PRD) |
| D-12 | 溢出检测使用 int64 + 算法检测 | PRD §3.1 明确要求覆盖正溢出和负溢出 | PM Kickoff (PRD) |
| D-13 | `strconv.ParseInt` 替代 `strconv.Atoi` | 需要 bitSize=64 明确语义以进行溢出检测；Atoi 返回 `int` 无法保证跨平台 | Designer |
| D-14 | 仅英文错误消息 | 目标用户为开发者，英文通用 | Designer (PRD D-09) |
| D-15 | 无 ANSI 颜色输出 | 纯文本兼容所有终端 | Designer |
| D-16 | `main.go` + `add.go` 分离 | 纯函数 `add()` 独立测试，main.go 仅负责 CLI 编排 | Designer |
| D-17 | `add_test.go` 独立于 `main_test.go` | 单元测试（6 用例）与集成测试（全 AC）职责分离 | Designer |
| D-18 | `--help` 输出到 stdout 且 exit 0 | 帮助信息不是错误；Unix 惯例 `--help` 返回 0 | Designer (Unix 惯例) |
| D-19 | `--version` 输出到 stdout 且 exit 0 | 版本信息不是错误 | Designer (Unix 惯例) |
| D-20 | 版本输出格式 `add <version>` | 简洁一致，与 help 文本中程序名风格呼应 | Designer |
| D-21 | CGO_ENABLED=0 静态编译 | 确保跨平台可移植性 | Designer (PRD §1.2) |
| D-22 | Makefile 为 P1 交付物 | PRD Appendix E 明确列出 | PRD §E |

---

## 13. 附录 A：与 Project 3 和 Project 27 的差异对比

Project 28 是同一领域（整数加法 CLI 工具）的第三次交付。以下对比三个项目的关键设计差异：

| 维度 | Project 3 | Project 27 | Project 28 |
|------|-----------|------------|------------|
| `--help`/`-h` | 不实现（源码无分支） | 不实现（PRD D-01 排除） | 实现（P0，flag 包） |
| `--version`/`-v` | 不实现 | 不实现 | 实现（P0，flag 包 + ldflags） |
| flag 包 | 不使用（os.Args 直读） | 不使用（os.Args 直读） | 使用（flag 包） |
| 溢出检测 | 不检测（静默回绕） | 不检测（静默回绕） | 强制检测（正溢出 + 负溢出） |
| 整数类型 | `int` (Atoi) | `int` (Atoi) | `int64` (ParseInt) |
| 错误消息区分度 | 无区分（统一用法提示） | 无区分（统一用法提示） | 区分：缺失/过多/格式/溢出/过长 |
| 程序名 | 硬编码 `add` | 静态 `<int> <int>` | 硬编码 `add` |
| 输入长度限制 | 无 | 无 | 40 字符 |
| stdout 方式 | fmt.Println | fmt.Println | fmt.Println |
| 退出码 | 0/1 | 0/1 | 0/1（退出码 2 已移除） |
| main 行数 | 未明确 | ≤ 50 | ≤ 60 |
| 源文件结构 | 单一 main.go | 单一 main.go | main.go + add.go |
| 测试文件 | 单一 main_test.go | 单一 main_test.go | add_test.go + main_test.go |
| Makefile | 无 | 无 | 有（P1） |
| go 版本 | go 1.26.4 | go 1.21 | go 1.21 |
| stdin 管道输入 | 无 | 无 | 推后至 v1.1 (D-11) |

**复用的设计元素**（与 Project 3/27 相同）：
- stdout 仅输出数字 + 换行
- stderr 以 `Error: ` 为前缀
- 退出码 0/1
- 英文错误消息
- 无配置文件、环境变量、ANSI 颜色、交互模式
- CGO_ENABLED=0 静态编译
- 从左到右校验，仅报告第一个错误

---

## 14. 附录 B：错误消息 Golden Strings

以下 golden strings 用于集成测试断言（来自 PRD Appendix C）：

| # | 场景 | stderr 精确字符串 |
|---|------|-------------------|
| 1 | 参数缺失 | `Error: missing arguments, need two integers. Usage: add <int_a> <int_b>\n` |
| 2 | 参数过多 | `Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>\n` |
| 3 | 非法整数 | `Error: "abc" is not a valid integer. Usage: add <int_a> <int_b>\n` |
| 4 | 整数溢出 | `Error: integer overflow: 9223372036854775807 + 1 exceeds int64 range. Usage: add <int_a> <int_b>\n` |
| 5 | 输入过长 | `Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>\n` |

---

## 15. 附录 C：帮助文本 Spec

```text
add — integer addition tool

Usage:
  add [-h] [-v] <int_a> <int_b>

Arguments:
  <int_a>    first integer (int64 range)
  <int_b>    second integer (int64 range)

Options:
  -h, --help       show this help message
  -v, --version    show version number

Examples:
  add 2 3           outputs 5
  add -5 10         outputs 5
  add -- -5 10      use -- separator when the first argument is negative

Notes:
  Negative numbers must directly follow the minus sign (e.g., -5, not - 5).
  When the first argument is negative, use -- to separate flags from args.
```

实现方式：在 `main.go` 的 `init()` 或 `main()` 开头设置 `flag.Usage = func() { fmt.Print(helpText) }`，其中 `helpText` 为上述常量字符串。

---

> 设计文档草稿完成。待 PM 确认 Q-01（input length check 顺序）、Q-02（`--` 行为是否需显式说明）、Q-03（溢出检测实现路径选定），以及 Architect 对文件结构（add.go 分离）和测试文件分工的审查。
