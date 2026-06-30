# 技术实施方案 — Project 28: Integer Addition CLI Tool

> **角色**: Architect
> **版本**: 1.0-final
> **日期**: 2026-07-01
> **状态**: Frozen
> **仓库**: https://github.com/luqz/demo.git
> **分支**: 28/delivery
> **输入文档**: PRD v1.0, Context Audit Report, 决策文档 v1.0-final

> **Changelog**:
> v1.0-final — 整合全角色 8 项修正、安全清单、Designer 最终 Help Text、AC -- 裁决。方案冻结，准备实现。

---

## 1. 架构总览

### 1.1 一句话架构

Go 1.21 单二进制，零外部依赖，标准库 `flag` 包解析参数，单 `package main` 拆分为 `main.go`（入口 ≤60 行）、`add.go`（核心计算）、`help.go`（帮助文本常量），`add_test.go` + `main_test.go` 覆盖全部 30 项 AC。

### 1.2 最终决策矩阵

| 决策点 | 决策 | 理由 |
|--------|------|------|
| 语言与版本 | Go 1.21 | PRD 约束 (`go.mod` 声明 `go 1.21`)，单二进制、零依赖、跨平台 |
| CLI 框架 | 标准库 `flag` | PRD 约束，无第三方依赖 |
| 输出方式 | `fmt.Println` (stdout), `fmt.Fprintf(os.Stderr, ...)` (stderr) | PRD 约束 |
| 版本注入 | `-ldflags -X main.version`，默认 `"dev"` | PRD 约束 |
| 错误处理 | `flag.ContinueOnError`，统一由 `main()` 处理退出 | 确保 exit code 1 和错误消息格式一致 |
| 溢出检测 | 手动预检（同号相加前符号比较） | 代码复审易证明，极简依赖。`math/bits.Add64` 备选进附录 |
| 模块结构 | 单一 `package main`，拆分至 `main.go`、`add.go`、`help.go` | 满足 `main.go` ≤ 60 行约束 |
| 测试策略 | 单元测试 (`add_test.go`) + CLI 集成测试 (`main_test.go`) | 覆盖全部 30 项 AC |
| flag.Usage 输出 | `flag.CommandLine.SetOutput(io.Discard)` + `ErrHelp` 分支 `fmt.Print` 到 stdout | 避免 help 双次打印，确保 help 到 stdout |
| 版本标志绑定 | `var showVersion bool` + 两个 `flag.BoolVar(&showVersion, ...)` 调用 | 同一 bool 绑定 `-v` 和 `--version` |
| 溢出错误消息 | `fmt.Errorf("integer overflow: %d + %d exceeds int64 range", a, b)` | 包含操作数值，便于调试 |
| Cross 构建输出 | 输出到 `dist/add-{os}-{arch}[.exe]` | 不丢弃产物，可验证二进制存在 |
| helpText 常量位置 | 独立文件 `help.go` | 避免 `main.go` 超过 60 行 |

### 1.3 无冲突声明

仓库处于初始空状态（仅 `# demo` README），无现有代码、无历史架构、无依赖冲突。所有决策均为新建，无需兼容遗留系统。

---

## 2. 验收条件全量提取

从 PRD §3 提取全部 30 项验收条件。编号 AC-01 至 AC-30。

### 2.1 核心功能验证 (PRD §3.1)

| ID | 命令 | 预期 stdout | 预期 stderr | 预期 exit | 说明 |
|----|------|-------------|-------------|-----------|------|
| AC-01 | `add 2 3` | `5\n` | — | 0 | 基本正数加法 |
| AC-02 | `add 0 0` | `0\n` | — | 0 | 零值边界 |
| AC-03 | `add -5 10` | `5\n` | — | 0 | 第二操作数为负，无需 `--` |
| AC-04 | `add -- -5 10` | `5\n` | — | 0 | `--` 分隔符，第一操作数为负 |
| AC-05 | `add -7 -3` | `-10\n` | — | 0 | 双负数加法 |
| AC-06 | `add 2147483647 1` | `2147483648\n` | — | 0 | 大整数（int32 边界以上，int64 内） |
| AC-07 | `add 9223372036854775807 1` | — | overflow error | 1 | 正溢出（int64 max + 1） |
| AC-08 | `add -9223372036854775808 -1` | — | overflow error | 1 | 负溢出（int64 min + (-1)） |
| AC-09 | `add 9223372036854775807 0` | `9223372036854775807\n` | — | 0 | 正边界 + 0 |
| AC-10 | `add -9223372036854775808 0` | `-9223372036854775808\n` | — | 0 | 负边界 + 0 |

### 2.2 错误路径验证 (PRD §3.2)

| ID | 命令 | 预期 stderr (golden string) | 预期 exit | 说明 |
|----|------|-----------------------------|-----------|------|
| AC-11 | `add` (no args) | `Error: missing arguments, need two integers. Usage: add <int_a> <int_b>\n` | 1 | 零参数 |
| AC-12 | `add 1` (one arg) | 同 AC-11 | 1 | 单参数 |
| AC-13 | `add 1 2 3` (three args) | `Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>\n` | 1 | 三参数 |
| AC-14 | `add abc 1` | `Error: "abc" is not a valid integer. Usage: add <int_a> <int_b>\n` | 1 | 第一个参数非整数 |
| AC-15 | `add 1 xyz` | `Error: "xyz" is not a valid integer. Usage: add <int_a> <int_b>\n` | 1 | 第二个参数非整数 |
| AC-16 | `add 1.5 2` | `Error: "1.5" is not a valid integer. Usage: add <int_a> <int_b>\n` | 1 | 浮点数 |
| AC-17 | `add <100-char> 1` | `Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>\n` | 1 | 输入超长 |

### 2.3 帮助与版本验证 (PRD §3.3)

| ID | 命令 | 预期 stdout | 预期 exit | 说明 |
|----|------|-------------|-----------|------|
| AC-18 | `add --help` | 完整帮助文本 (见 §6.4 Designer 最终稿) | 0 | 长选项帮助 |
| AC-19 | `add -h` | 同 AC-18 | 0 | 短选项帮助 |
| AC-20 | `add --version` (CI build) | `add v1.0.0\n` | 0 | CI 构建版本号 |
| AC-21 | `add -v` | 同 AC-20 | 0 | 短选项版本 |
| AC-22 | `add --version` (local build) | `add dev\n` | 0 | 本地构建默认版本 |

### 2.4 代码与构建验证 (PRD §3.4)

| ID | 检查项 | 验证方式 |
|----|--------|----------|
| AC-23 | `go.mod` 声明 `go 1.21` | `grep "^go 1.21" go.mod` 匹配 |
| AC-24 | main 源文件 ≤ 60 行 | `grep -cve '^\s*$' -e '^\s*//' main.go` ≤ 60（排除空行和注释） |
| AC-25 | `add_test.go` 存在，6 个核心单元测试 | `go test -run TestAdd -v` 6 个子测试全部 PASS |
| AC-26 | `main_test.go` 存在，CLI 集成测试覆盖 AC-01 至 AC-22 | `go test -run TestCLI -v` 全部 PASS |
| AC-27 | `go vet ./...` 无报错 | exit 0，无 warning |
| AC-28 | 5 平台交叉编译通过 | `GOOS=X GOARCH=Y go build` 各平台 exit 0 |
| AC-29 | `go mod verify` 通过 | exit 0 |
| AC-30 | 生产构建命令可用 | `CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=v1.0.0" -trimpath -o add` exit 0 |

### 2.5 退出码规范 (PRD §3.5)

| 退出码 | 含义 | 覆盖的 AC |
|--------|------|-----------|
| 0 | 成功，结果输出到 stdout | AC-01~06, AC-09~10, AC-18~22 |
| 1 | 用户输入错误（参数缺失、格式错误、溢出、输入超长、flag 解析错误） | AC-07~08, AC-11~17 |

> 退出码 2 保留给未来运行时错误，v1.0 不实现不测试。

---

## 3. 验收条件 → 实现覆盖映射

以下说明全部 30 项 AC 如何在单一全栈实现中被覆盖。Coder 按顺序交付文件，无并行拆分。

### 3.1 核心功能 (AC-01 至 AC-10)

覆盖于 `add.go` 中的 `add(a, b int64) (int64, error)` 函数。

| AC | 实现覆盖 |
|----|----------|
| AC-01 | `add(2, 3)` → `5, nil`。基本正数加法路径。 |
| AC-02 | `add(0, 0)` → `0, nil`。零值不触发任何溢出条件。 |
| AC-03 | `add(-5, 10)` → `5, nil`。异号加法不触发溢出检测（只有同号才预检）。 |
| AC-04 | `add(-5, 10)` → `5, nil`。`flag` 包自动处理 `--` 分隔符，`add()` 接收相同输入。 |
| AC-05 | `add(-7, -3)` → `-10, nil`。双负同号，但 `-7 > math.MinInt64 - (-3)`（即 `-7 > -9223372036854775805`），不触发负溢出。 |
| AC-06 | `add(2147483647, 1)` → `2147483648, nil`。正数同号但未溢出。 |
| AC-07 | `add(9223372036854775807, 1)` → `0, overflowError`。正溢出预检：`a > math.MaxInt64 - b` 为 true。 |
| AC-08 | `add(-9223372036854775808, -1)` → `0, overflowError`。负溢出预检：`a < math.MinInt64 - b` 为 true。 |
| AC-09 | `add(9223372036854775807, 0)` → `9223372036854775807, nil`。零值操作数不触发同号预检。 |
| AC-10 | `add(-9223372036854775808, 0)` → `-9223372036854775808, nil`。同上。 |

溢出检测公式（两组条件，使用 `math.MaxInt64` 和 `math.MinInt64`）：

```
正溢出: a > 0 && b > 0 && a > math.MaxInt64 - b
负溢出: a < 0 && b < 0 && a < math.MinInt64 - b
```

溢出错误消息（包含操作数值）：

```
"fmt.Errorf("integer overflow: %d + %d exceeds int64 range", a, b)"
```

### 3.2 错误路径 (AC-11 至 AC-17)

覆盖于 `main.go` 中的参数校验逻辑 + `parseArg` 辅助函数。

| AC | 实现覆盖 |
|----|----------|
| AC-11 | `flag.Args()` 长度为 0 → `printError("missing arguments, need two integers")` → `os.Exit(1)` |
| AC-12 | `flag.Args()` 长度为 1 → 同 AC-11 |
| AC-13 | `flag.Args()` 长度 > 2 → `printError("too many arguments, only two integers accepted")` → `os.Exit(1)` |
| AC-14 | `parseArg("abc")` → `len("abc")` ≤ 40，但 `strconv.ParseInt` 返回 error → `printError(`"abc" is not a valid integer`)` |
| AC-15 | `parseArg("xyz")` → 同 AC-14，`strconv.ParseInt` 返回 error |
| AC-16 | `parseArg("1.5")` → `strconv.ParseInt("1.5", 10, 64)` 返回 error（非整数格式）→ `printError(`"1.5" is not a valid integer`)` |
| AC-17 | `parseArg("111...100chars")` → `len(arg) > 40` → 短路返回 `errInputTooLong` → `printError("input too long, maximum 40 characters per argument")` |

`parseArg` 函数实现（置于 `main.go`）：

```
func parseArg(raw string) (int64, error) {
    if len(raw) > 40 {
        return 0, errInputTooLong
    }
    n, err := strconv.ParseInt(raw, 10, 64)
    if err != nil {
        return 0, fmt.Errorf("%q is not a valid integer", raw)
    }
    return n, nil
}
```

使用 `%q` 格式化而非手动拼接双引号，天然处理控制字符转义（安全清单 S4）。

`printError` 函数：

```
func printError(msg string) {
    fmt.Fprintf(os.Stderr, "Error: %s. Usage: add <int_a> <int_b>\n", msg)
}
```

### 3.3 帮助与版本 (AC-18 至 AC-22)

覆盖于 `main.go` 的 flag 配置 + `help.go` 的 `helpText` 常量。

| AC | 实现覆盖 |
|----|----------|
| AC-18 | `flag.CommandLine.Init("add", flag.ContinueOnError)` + `flag.CommandLine.SetOutput(io.Discard)`。Parse 返回 `flag.ErrHelp` 时 `fmt.Print(helpText)` 输出帮助到 stdout，`os.Exit(0)`。`helpText` 常量在 `help.go`，内容为 §6.4 Designer 最终稿。 |
| AC-19 | `-h` 和 `--help` 由 flag 包自动映射到同一处理路径，行为同 AC-18。 |
| AC-20 | `var showVersion bool` + `flag.BoolVar(&showVersion, "v", false, "show version number")` + `flag.BoolVar(&showVersion, "version", false, "show version number")`。解析后 `showVersion == true` → `fmt.Printf("add %s\n", version)` + `os.Exit(0)`。CI 构建通过 `-ldflags="-X main.version=v1.0.0"` 注入。 |
| AC-21 | `-v` 和 `--version` 共享同一 `showVersion` bool，行为同 AC-20。 |
| AC-22 | `var version = "dev"`。无 ldflags 时默认值生效。`add --version` → `add dev\n`。 |

#### flag 配置关键要点

```
// init() 或 main() 开头
flag.CommandLine.Init("add", flag.ContinueOnError)
flag.CommandLine.SetOutput(io.Discard)   // 禁止 flag 包自行输出错误/帮助
flag.Usage = func() {
    fmt.Print(helpText)                  // helpText 到 stdout（仅当 flag 自动调用时）
}

func main() {
    var showVersion bool
    flag.BoolVar(&showVersion, "v", false, "show version number")
    flag.BoolVar(&showVersion, "version", false, "show version number")

    if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
        if err == flag.ErrHelp {
            fmt.Print(helpText)          // stdout，一次打印
            os.Exit(0)
        }
        printError(err.Error())          // 其他 flag 错误 → stderr, exit 1
        os.Exit(1)
    }

    if showVersion {
        fmt.Printf("add %s\n", version)
        os.Exit(0)
    }
    // ... 参数校验和计算
}
```

**`SetOutput(io.Discard)` 的作用**：阻止 flag 包在 `ErrHelp` 场景下自动向 stderr 打印帮助文本。`flag.Usage` 保留作为兜底（仅 flag 内部自动触发时调用），主控流走 `ErrHelp` 分支显式 `fmt.Print(helpText)`，保证帮助输出仅一次到 stdout。

### 3.4 代码与构建 (AC-23 至 AC-30)

覆盖于项目文件与 Makefile。

| AC | 覆盖方式 |
|----|----------|
| AC-23 | `go.mod` 写入 `module add` + `go 1.21`。`go mod tidy -compat=1.21` 在 CI 中显式执行。 |
| AC-24 | `main.go` 控制在 ≤ 60 行（排除空行和注释）。`add()` 在 `add.go`，`helpText` 在 `help.go`，`parseArg` 在 `main.go`。行数预算充足。验证命令：`grep -cve '^\s*$' -e '^\s*//' main.go`。 |
| AC-25 | `add_test.go`：10 个 `t.Run` 子测试（6 核心 + 4 补充 T1-T4），table-driven 风格。详见 §5.2。 |
| AC-26 | `main_test.go`：使用 `TestMain` 模式构建二进制一次，所有 CLI 测试复用。覆盖 AC-01 至 AC-22（含补充用例 T5-T6）。详见 §5.3。 |
| AC-27 | 实现完成后执行 `go vet ./...`。Makefile `vet` 目标调用。 |
| AC-28 | Makefile `cross` 目标循环 5 组 `GOOS`/`GOARCH`，输出到 `dist/add-{os}-{arch}[.exe]`。 |
| AC-29 | `go mod verify` 作为 CI 前置步骤执行。 |
| AC-30 | Makefile `build` 目标：`CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -trimpath -o add`。 |

### 3.5 AC -- 分隔符裁决（Architect 裁定）

PRD §3.1 中 4 条 AC 的命令字符串是示意性的。Go `flag` 包在 `os.Args` 中将 `-5` 解析为 flag `-5`（未定义），导致错误。实际测试传参如下——PRD 文本保持不动：

| AC ID | PRD 命令字符串 | `main_test.go` 实际传参 | 说明 |
|-------|---------------|------------------------|------|
| AC-03 | `add -5 10` | `["10", "-5"]` | 负第二操作数，调换顺序无需 `--` |
| AC-05 | `add -7 -3` | `["--", "-7", "-3"]` | 双负，第一参数需 `--` |
| AC-08 | `add -9223372036854775808 -1` | `["--", "-9223372036854775808", "-1"]` | 负溢出，第一参数需 `--` |
| AC-10 | `add -9223372036854775808 0` | `["--", "-9223372036854775808", "0"]` | 负边界，第一参数需 `--` |

**QA 条件**：`main_test.go` 中对这 4 个 case 必须加注释，明确标注 "PRD 命令为示意，实现中使用 `--` 分隔符"。

---

## 4. 技术架构详细设计

### 4.1 包/模块结构

```
add/                              ← 仓库根目录（单一模块）
├── main.go                       ← 入口: flag 解析、参数校验、调用 add()、输出 (≤60行)
├── add.go                        ← 纯函数: add(a,b int64) + 溢出检测
├── help.go                       ← helpText 常量（Designer 最终稿）
├── add_test.go                   ← 单元测试: add() 10 个用例 (6核心 + 4补充)
├── main_test.go                  ← CLI 集成测试: exec.Command 覆盖 AC-01~22 + 补充
├── go.mod                        ← module add, go 1.21
├── go.sum                        ← 空或仅含 go 版本哈希（零依赖）
├── Makefile                      ← build/test/vet/cross/checksum/release/clean
├── .gitignore                    ← 忽略 add, dist/, *.exe
├── .github/workflows/ci.yml      ← CI 管道: lint → test → cross-build → release
├── LICENSE                       ← MIT
├── README.md                     ← 项目简介、安装、使用示例
└── docs/
    ├── prd.md                    ← PRD v1.0
    ├── tech-plan.md              ← 本文档
    ├── context-audit.md          ← 上下文审计报告
    └── reports/                  ← 各角色报告目录
```

### 4.2 main.go 流程（伪代码）

```
package main

import ("flag"; "fmt"; "io"; "os"; "strconv")

var version = "dev"

var (
    errInputTooLong = errors.New("input too long, maximum 40 characters per argument")
)

func init() {
    flag.CommandLine.Init("add", flag.ContinueOnError)
    flag.CommandLine.SetOutput(io.Discard)
    flag.Usage = func() { fmt.Print(helpText) }
}

func main() {
    var showVersion bool
    flag.BoolVar(&showVersion, "v", false, "show version number")
    flag.BoolVar(&showVersion, "version", false, "show version number")

    if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
        if err == flag.ErrHelp {
            fmt.Print(helpText)
            os.Exit(0)
        }
        printError(err.Error())
        os.Exit(1)
    }

    if showVersion {
        fmt.Printf("add %s\n", version)
        os.Exit(0)
    }

    args := flag.Args()
    if len(args) < 2 {
        printError("missing arguments, need two integers")
        os.Exit(1)
    }
    if len(args) > 2 {
        printError("too many arguments, only two integers accepted")
        os.Exit(1)
    }

    a, err := parseArg(args[0])
    if err != nil {
        printError(err.Error())
        os.Exit(1)
    }
    b, err := parseArg(args[1])
    if err != nil {
        printError(err.Error())
        os.Exit(1)
    }

    result, err := add(a, b)
    if err != nil {
        printError(err.Error())
        os.Exit(1)
    }

    fmt.Println(result)
}

func parseArg(raw string) (int64, error) {
    if len(raw) > 40 {
        return 0, errInputTooLong
    }
    n, err := strconv.ParseInt(raw, 10, 64)
    if err != nil {
        return 0, fmt.Errorf("%q is not a valid integer", raw)
    }
    return n, nil
}

func printError(msg string) {
    fmt.Fprintf(os.Stderr, "Error: %s. Usage: add <int_a> <int_b>\n", msg)
}
```

行数估算：约 55 行（含空行和注释），符合 ≤ 60 行约束。

### 4.3 add.go

```
package main

import (
    "fmt"
    "math"
)

func add(a, b int64) (int64, error) {
    // 正溢出预检
    if a > 0 && b > 0 && a > math.MaxInt64-b {
        return 0, fmt.Errorf("integer overflow: %d + %d exceeds int64 range", a, b)
    }
    // 负溢出预检
    if a < 0 && b < 0 && a < math.MinInt64-b {
        return 0, fmt.Errorf("integer overflow: %d + %d exceeds int64 range", a, b)
    }
    return a + b, nil
}
```

行数估算：约 15 行。零外部依赖。

### 4.4 help.go

包含 Designer 最终稿的 `helpText` 常量。内容见 §6.4。

### 4.5 错误消息生成策略

所有错误消息遵循统一格式：

```
Error: <description>. Usage: add <int_a> <int_b>\n
```

- `printError()` 为唯一错误输出点，`os.Exit(1)` 统一由 `main()` 调用
- `parseArg()` 的无效整数错误使用 `%q` 格式化原始输入（安全清单 S4）
- `add()` 的溢出错误包含操作数值（修正 #2）
- AC 测试中对比 golden string 时，引用 `package main` 中的非导出常量，消除同步风险

### 4.6 版本注入机制

```
var version = "dev"              // package-level, default

// CI 构建:
// CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=v1.0.0" -trimpath -o add

// 本地构建（无 ldflags）:
// go build → ./add --version → "add dev"
```

### 4.7 跨平台兼容性

| 平台 | 注意事项 |
|------|----------|
| Linux/amd64 | 默认构建目标 |
| Linux/arm64 | 仅交叉编译验证 |
| Darwin/amd64 | Intel Mac |
| Darwin/arm64 | Apple Silicon；`CGO_ENABLED=0` 确保不依赖 Xcode |
| Windows/amd64 | Go 标准库 `os.Stderr`/`os.Stdout`/`\n` 行为一致 |

标准库 `flag`、`fmt`、`os`、`strconv`、`math` 在所有目标平台行为一致，无需条件编译。

---

## 5. 测试设计

### 5.1 add_test.go — 单元测试

10 个 `t.Run` 子测试，table-driven 风格。直接调用 `add(a, b int64)`，不依赖 CLI。

| # | 子测试名称 | a | b | 预期 result | 预期 err | 说明 |
|---|-----------|----|----|-------------|---------|------|
| 1 | NormalPositive | 2 | 3 | 5 | nil | AC-01 |
| 2 | NormalNegative | -7 | -3 | -10 | nil | AC-05 |
| 3 | MixedSign | -5 | 10 | 5 | nil | AC-03 |
| 4 | PositiveOverflow | 9223372036854775807 | 1 | 0 | non-nil | AC-07 |
| 5 | NegativeOverflow | -9223372036854775808 | -1 | 0 | non-nil | AC-08 |
| 6 | BoundaryExact | 9223372036854775807 | 0 | 9223372036854775807 | nil | AC-09 |
| 7 | T1: NegativeZero | -0 | 5 | 5 | nil | 补充：-0 合法输入 |
| 8 | T2: InputExactly40 | (40-char "1"s) | 1 | ParseInt 正常 | nil | 补充：40 字符边界合法 |
| 9 | T3: MixedSignNoOverflow1 | math.MaxInt64 | -1 | math.MaxInt64-1 | nil | 补充：异号不溢出（大正+小负） |
| 10 | T4: MixedSignNoOverflow2 | math.MinInt64 | 1 | math.MinInt64+1 | nil | 补充：异号不溢出（大负+小正） |

> T1-T4 来自全角色共识修正 #6。T2 验证 `input = 40 chars` 不触发长度错误；T3-T4 验证异号加法不触发溢出误判。

溢出错误验证：检查 `err != nil` 并且错误消息包含 `"integer overflow"` 前缀，不硬编码完整 golden string（由 `add()` 内部动态生成含操作数值）。

### 5.2 main_test.go — CLI 集成测试

使用 `TestMain` 模式：`TestMain` 中 `go build` 生成一次二进制，所有子测试通过 `exec.Command(binary, args...)` 复用。

| 类别 | 用例数 | AC 覆盖 |
|------|--------|---------|
| 核心功能 | 10 | AC-01 至 AC-10 |
| 错误路径 | 7 | AC-11 至 AC-17 |
| 帮助与版本 | 5 | AC-18 至 AC-22 |
| 补充用例 | 2 | T5, T6（见下） |
| **合计** | **24** | AC-01 至 AC-22 + 补充 |

补充用例（来自修正 #7）：

| ID | args | 预期 | 说明 |
|----|------|------|------|
| T5 | `["--", "-v"]` | stdout: `add dev\n`, exit 0 | `--` 后 flag 不解析，`-v` 被当作 positional arg，但由于只有一个参数 → 缺失参数错误。**注意**：需精确分析 flag 行为后再写预期。 |
| T6 | `["-v", "2", "3"]` | stdout 同时包含版本号和结果或版本号优先 | flag 和 positional 混合用法，验证 flag 优先级。按 Designer Notes: "Flag and positional arguments can be mixed"，结合 flag 包默认行为（flag 在第一个非 flag 参数前解析），传 `-v 2 3` 时 flag 解析到 `-v` 即停止处理后续 flag（因为遇到非 flag 参数 `2`），但 `-v` 本身是 boolean flag 不需要值，所以正常触发版本输出 + exit 0。 |

> Coder 实现 T5-T6 前需在代码注释中分析 flag 包行为，确保预期值正确。

`--` 分隔符裁决（AC-03/05/08/10）的应用：

测试代码中对 4 个涉及负第一参数的 case 使用实际传参（见 §3.5），并附带注释说明 PRD 示意性命令 vs 实现差异。

测试常量子引用：`main_test.go` 中直接引用 `package main` 中的 `helpText` 等非导出常量，避免 golden string 硬编码重复，消除同步风险。

---

## 6. 关键修正与补充规范

### 6.1 8 项全角色共识修正

| # | 修正项 | 内容 |
|---|--------|------|
| 1 | flag.ErrHelp 双次打印 | `flag.CommandLine.SetOutput(io.Discard)` + 仅在 `ErrHelp` 分支 `fmt.Print(helpText)` 一次到 stdout |
| 2 | overflow 错误消息缺操作数值 | `add()` 返回 `fmt.Errorf("integer overflow: %d + %d exceeds int64 range", a, b)` |
| 3 | helpText 常量超 60 行约束 | 移到独立文件 `help.go` |
| 4 | Makefile cross 输出到 /dev/null | 改为输出到 `dist/add-{os}-{arch}[.exe]` |
| 5 | flag.BoolVar 绑定 | `var showVersion bool` + 两个 `flag.BoolVar(&showVersion, ...)` 调用 |
| 6 | add_test.go 补充边缘用例 | T1-T4（-0, =40字符, 异号不溢出×2） |
| 7 | main_test.go 补充边界用例 | T5-T6（`--` 后 flag 不解析, flag 优先级） |
| 8 | 文件清单追加 | `.gitignore`, `.github/workflows/ci.yml`, `LICENSE` (MIT) |

### 6.2 安全检查清单

纳入集成交付检查（§8.3），S1-S4：

| # | 检查项 | 验证方式 |
|---|--------|----------|
| S1 | `-trimpath` 生效 | `strings <binary> | grep -c '/'` 应返回 0 或极小值 |
| S2 | 静态链接确认 | `file <binary>` 应包含 `statically linked` |
| S3 | 超长输入不 panic 不 OOM | `add $(python3 -c 'print("A"*1000)') 1` → stderr 报告 too long, exit 1 |
| S4 | 控制字符不污染终端 | `parseArg` 中使用 `%q` 格式化，天然转义控制字符 |

### 6.3 Makefile 目标

```makefile
BINARY  := add
VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)
DIST    := dist

.PHONY: build test vet cross checksum release clean

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY)

test:
	go test -v ./...

vet:
	go vet ./...

cross:
	@mkdir -p $(DIST)
	GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -trimpath -o $(DIST)/add-linux-amd64
	GOOS=linux   GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -trimpath -o $(DIST)/add-linux-arm64
	GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -trimpath -o $(DIST)/add-darwin-amd64
	GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -trimpath -o $(DIST)/add-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -trimpath -o $(DIST)/add-windows-amd64.exe

checksum:
	cd $(DIST) && sha256sum add-* > SHA256SUMS

release: cross checksum

clean:
	rm -f $(BINARY)
	rm -rf $(DIST)
```

### 6.4 Designer 最终 Help Text（`help.go` 中 `helpText` 常量的唯一来源）

```
add — computes the sum of two integers in int64 range.

Usage:
  add [-h] [-v] <int_a> <int_b>

Arguments:
  <int_a>    first integer (int64 range)
  <int_b>    second integer (int64 range)

Options:
  -h, --help       show this help message
  -v, --version    show version number

Examples:
  add 2 3            outputs 5
  add 10 -5          outputs 5
  add -7 -3          outputs -10
  add -- -5 10       outputs 5  (use -- when first argument is negative)

Notes:
  Negative numbers must directly follow the minus sign (e.g., -5, not - 5).
  When the first argument is negative, use -- to separate flags from args.
  Flag and positional arguments can be mixed (e.g., add -v 2 3).
  Exit codes: 0 on success, 1 on input error.
```

> 此文本为 Designer Round 3 交付最终稿，与 PRD Appendix D 存在差异（标题措辞、示例增量、Notes 扩充）。以此为准。当 PRD 中 `flag.Usage` 要求 "matches Designer's help template" 时，即指此文本。

---

## 7. 唯一实现 Job 说明

### 7.1 实现模型

单 Coder、单链全栈实现。无需前后端拆分、无需并行工作流。Coder 在一个完整实现链中交付所有产物（源代码、测试、构建文件、CI 配置、文档、LICENSE）。

### 7.2 Job 范围

| 类别 | 文件 |
|------|------|
| 源代码 | `main.go`, `add.go`, `help.go` |
| 测试 | `add_test.go` (10 用例), `main_test.go` (24 用例) |
| 构建 | `go.mod`, `Makefile` |
| CI | `.github/workflows/ci.yml` |
| 配置 | `.gitignore` |
| 许可证 | `LICENSE` (MIT) |
| 文档 | `README.md` 重写 |
| 目录 | 确保 `docs/` 结构完整 |

### 7.3 实现顺序（单链）

```
Step 1:  go.mod                      → 模块初始化 (go 1.21)
Step 2:  add.go                      → add() 纯函数 + 溢出检测
Step 3:  add_test.go                 → 10 个单元测试，确保 add() 正确 (RED-GREEN-REFACTOR)
Step 4:  help.go                     → helpText 常量 (Designer 最终稿)
Step 5:  main.go                     → flag 解析 + 参数校验 + 错误消息 + 帮助/版本 + parseArg + printError (≤60行)
Step 6:  main_test.go                → 24 个 CLI 集成测试，覆盖 AC-01~22 + T5-T6
Step 7:  Makefile                    → build/test/vet/cross/checksum/release/clean
Step 8:  .gitignore                  → 忽略 add, dist/, *.exe
Step 9:  .github/workflows/ci.yml    → CI 管道 (lint → test → cross-build → release)
Step 10: LICENSE                     → MIT
Step 11: README.md                   → 项目文档重写
Step 12: go vet + cross + go mod verify → 构建质量门
Step 13: 集成交付检查                → 全量验证（见 §8）
```

Step 3 在 Step 5 之前执行，确保核心计算逻辑在集成前已验证通过（TDD 原则）。

---

## 8. 集成交付检查清单

Coder 完成实现后，按以下清单逐项验证。

### 8.1 构建维度

- [ ] `go mod verify` 返回 exit 0
- [ ] `go vet ./...` 返回 exit 0，无 warning
- [ ] `grep "^go 1.21" go.mod` 匹配 (AC-23)
- [ ] `grep -cve '^\s*$' -e '^\s*//' main.go` ≤ 60 (AC-24)
- [ ] `make build VERSION=v1.0.0` 成功生成 `add` 二进制 (AC-30)
- [ ] `make cross` 全部 5 平台通过 (AC-28)
- [ ] `strings add | grep -c '/'` 返回 0 或极小值 (S1: -trimpath 生效)
- [ ] `file add` 确认静态链接 (S2: statically linked)
- [ ] `go mod verify` (AC-29)

### 8.2 测试维度

- [ ] `go test -v -run TestAdd` 10 个子测试全部 PASS (AC-25 + T1-T4)
- [ ] `go test -v -run TestCLI` 覆盖 AC-01 至 AC-22 + T5-T6，全部 PASS (AC-26)
- [ ] 对 AC-03/05/08/10 的测试代码中，4 处含 `--` 分隔符注释

### 8.3 功能维度（手动验证关键 AC）

- [ ] AC-01 至 AC-10：核心功能
- [ ] AC-11 至 AC-17：错误路径，精确比对 golden error string
- [ ] AC-18 至 AC-22：帮助与版本
- [ ] S3：`add $(python3 -c 'print("A"*1000)') 1` → stderr too long, exit 1，无 panic 无 OOM

### 8.4 文档维度

- [ ] `README.md` 已重写，包含项目简介、安装方式 (`go install`)、使用示例、构建说明
- [ ] `docs/` 目录包含 `prd.md`、`tech-plan.md`、`context-audit.md`

### 8.5 交付维度

- [ ] 所有变更在 `28/delivery` 分支
- [ ] Git 提交信息清晰描述变更
- [ ] `.gitignore` 排除 `add`、`dist/`、`*.exe`
- [ ] `LICENSE` 文件存在

---

## 9. 风险与缓解

| 风险 | 影响 | 缓解 |
|------|------|------|
| `main.go` 因未知原因超过 60 行 | AC-24 失败 | 已将 `add()` 移至 `add.go`，`helpText` 移至 `help.go`。`main.go` 仅含 flag/参数/输出逻辑，预算充足 |
| 错误消息 Golden String 不匹配 | AC-11~17 测试失败 | `main_test.go` 直接引用 `package main` 中常量，零同步风险 |
| `flag` 包对 `--` 行为与预期不一致 | AC-03/05/08/10 失败 | 已在 §3.5 制定裁决，`main_test.go` 使用实际传参 + 注释 |
| `strconv.ParseInt` 对超长字符串行为 | AC-17 可能被 ParseInt 而非长度检查捕获 | `parseArg` 先检查 `len > 40`，短路返回长度错误 |
| CI 管道配置错误 | 构建失败 | Makefile 作为构建唯一真相来源，CI 只调用 `make` 目标 |
| Go 版本兼容性 | 构建失败 | `go mod tidy -compat=1.21` 在 CI 中显式执行 |
| 超长输入导致 panic/OOM | 程序崩溃 | `parseArg` 中 `> 40` 检查在第一行执行，S3 手动验证 |

---

## 10. 附录

### 10.1 溢出检测数学证明

对于 int64 加法 `a + b`：

- 正溢出条件：`a > 0 && b > 0 && a > MaxInt64 - b`
- 负溢出条件：`a < 0 && b < 0 && a < MinInt64 - b`

证明：

- MaxInt64 = 2^63 - 1 = 9223372036854775807
- MinInt64 = -2^63 = -9223372036854775808
- 若 `a + b > MaxInt64`，则 `a > MaxInt64 - b`（给定 `b > 0`，减法不溢出）
- 若 `a + b < MinInt64`，则 `a < MinInt64 - b`（给定 `b < 0`，减法不溢出）

此公式在同号场景下安全无歧义，且无需实际执行溢出加法后再检测。

### 10.2 math/bits.Add64 备选方案（v2 候选）

Go 1.12+ 标准库 `math/bits.Add64(a, b, carry)` 返回 `(sum, carryOut)`，天然支持溢出检测：

```
sum, carry := bits.Add64(uint64(a), uint64(b), 0)
if carry != 0 { ... }
```

v1.0 不采用的原因：手动符号比较同样正确，代码复审更容易证明，且不引入对 `math/bits` 包的额外认知负担。在 v2 如需扩展为无符号或更大整数范围时，`math/bits` 方案为优先候选。

### 10.3 已关闭的讨论项

- **flag.BoolVar 是否存在**：存在，Go 1.0 起就有。Designer 早期判断错误已纠正。
- **math/bits.Add64**：技术正确，v1.0 不采用，备注为 v2 候选。
- **flag.Usage 行为**：讨论已闭环，`SetOutput(io.Discard)` + `ErrHelp` 分支方案明确。
- **`28/agent/coder` 分支文档**：已完成远程合并到 `28/delivery`。

### 10.4 交付物清单

| 文件 | 描述 | 优先级 |
|------|------|--------|
| `main.go` | 入口：flag 解析、参数校验、输出 | P0 |
| `add.go` | `add()` 函数 + 溢出检测 | P0 |
| `help.go` | `helpText` 常量 (Designer 最终稿) | P0 |
| `add_test.go` | 单元测试 (10 个用例) | P0 |
| `main_test.go` | CLI 集成测试 (24 个用例) | P0 |
| `go.mod` | Go 模块定义 (go 1.21) | P0 |
| `Makefile` | 构建目标 (build/test/vet/cross/checksum/release/clean) | P1 |
| `.github/workflows/ci.yml` | CI 管道配置 | P1 |
| `.gitignore` | 忽略 `add`, `dist/`, `*.exe` | P1 |
| `LICENSE` | MIT 许可证 | P1 |
| `README.md` | 项目文档 | P0 |
| `docs/prd.md` | PRD v1.0 | P0 |
| `docs/tech-plan.md` | 本文档 | P0 |
| `docs/context-audit.md` | 上下文审计报告 | P0 |

---

> **方案冻结。实现开始。**
>
> Coder 在一个全栈实现链中按 §7.3 顺序交付所有产物（Step 1→13），无需并行拆分。实现完成后按 §8 清单逐项验证。
