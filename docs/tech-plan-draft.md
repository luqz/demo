# 技术实施方案（草案）— Project 28: Integer Addition CLI Tool

> **角色**: Architect
> **版本**: 1.0-draft
> **日期**: 2026-07-01
> **仓库**: https://github.com/luqz/demo.git
> **分支**: 28/delivery
> **输入文档**: PRD v1.0, Context Audit Report

---

## 1. 架构总览

### 1.1 一句话架构

单一 Go 二进制文件，零外部依赖，`flag` 包解析参数，`main.go` 承载全部逻辑（含溢出检测），辅助测试文件独立。

### 1.2 关键决策

| 决策点 | 选择 | 依据 |
|--------|------|------|
| 语言 | Go 1.21 | PRD 明确要求 `go.mod` 声明 `go 1.21`；单二进制、零依赖、跨平台 |
| CLI 框架 | 标准库 `flag` | PRD 明确要求 `flag` 包，无第三方依赖 |
| 输出方式 | `fmt.Println` | PRD 明确要求：stdout 仅输出数字 + 换行 |
| 版本注入 | `-ldflags -X main.version` | PRD 明确要求；默认 `"dev"` |
| 错误处理 | `flag.ContinueOnError` | PRD 明确要求 flag parse error 统一 exit 1 |
| 溢出检测 | 手动比较（预检） | 无 `checked` 包依赖；用加法前符号预判 |
| 模块结构 | 单一 package `main` | 项目极简，无需子包；`add()` 函数可内联或抽至 `add.go` |
| 测试策略 | 单元测试 + CLI 集成测试 | `add_test.go` 测 `add()` 函数，`main_test.go` 用 `exec.Command` 测全 AC |
| 构建策略 | CGO_ENABLED=0，strip + trimpath | PRD 明确要求；确保最小二进制体积与可移植性 |
| 文件行数 | main 源文件 ≤ 60 行 | PRD 3.4 明确约束；排除空行和注释 |

### 1.3 无冲突声明

仓库处于初始空状态（仅 `# demo` README），无现有代码、无历史架构、无依赖冲突。所有决策均为新建，无需兼容任何遗留系统。

---

## 2. 验收条件全量提取

从 PRD §3 提取全部 30 项验收条件，按类别编号。

### 2.1 核心功能验证 (3.1)

| ID | 命令 | 预期 stdout | 预期 stderr | 预期 exit | 关键点 |
|----|------|-------------|-------------|-----------|--------|
| AC-01 | `add 2 3` | `5\n` | — | 0 | 基本正数加法 |
| AC-02 | `add 0 0` | `0\n` | — | 0 | 零值边界 |
| AC-03 | `add -5 10` | `5\n` | — | 0 | 第二操作数为负 |
| AC-04 | `add -- -5 10` | `5\n` | — | 0 | `--` 分隔符，第一操作数为负 |
| AC-05 | `add -7 -3` | `-10\n` | — | 0 | 双负数加法 |
| AC-06 | `add 2147483647 1` | `2147483648\n` | — | 0 | 大整数（int32 边界以上，int64 内） |
| AC-07 | `add 9223372036854775807 1` | — | overflow error | 1 | 正溢出（int64 max + 1） |
| AC-08 | `add -9223372036854775808 -1` | — | overflow error | 1 | 负溢出（int64 min + (-1)） |
| AC-09 | `add 9223372036854775807 0` | `9223372036854775807\n` | — | 0 | 正边界精确值 |
| AC-10 | `add -9223372036854775808 0` | `-9223372036854775808\n` | — | 0 | 负边界精确值 |

### 2.2 错误路径验证 (3.2)

| ID | 命令 | 预期 stderr | 预期 exit | 关键点 |
|----|------|-------------|-----------|--------|
| AC-11 | `add` | `Error: missing arguments, need two integers. Usage: add <int_a> <int_b>\n` | 1 | 零参数 |
| AC-12 | `add 1` | 同上 | 1 | 单参数 |
| AC-13 | `add 1 2 3` | `Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>\n` | 1 | 三参数 |
| AC-14 | `add abc 1` | `Error: "abc" is not a valid integer. Usage: add <int_a> <int_b>\n` | 1 | 第一个参数非整数 |
| AC-15 | `add 1 xyz` | `Error: "xyz" is not a valid integer. Usage: add <int_a> <int_b>\n` | 1 | 第二个参数非整数 |
| AC-16 | `add 1.5 2` | `Error: "1.5" is not a valid integer. Usage: add <int_a> <int_b>\n` | 1 | 浮点数 |
| AC-17 | `add <100-char> 1` | `Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>\n` | 1 | 输入超长 |

### 2.3 帮助与版本验证 (3.3)

| ID | 命令 | 预期 stdout | 预期 exit | 关键点 |
|----|------|-------------|-----------|--------|
| AC-18 | `add --help` | 完整帮助文本 (见 PRD Appendix D) | 0 | 长选项帮助 |
| AC-19 | `add -h` | 同 AC-18 | 0 | 短选项帮助 |
| AC-20 | `add --version` (CI build) | `add v1.0.0\n` | 0 | CI 构建版本号 |
| AC-21 | `add -v` | 同 AC-20 | 0 | 短选项版本 |
| AC-22 | `add --version` (local build) | `add dev\n` | 0 | 本地构建默认版本 |

### 2.4 代码与构建验证 (3.4)

| ID | 检查项 | 验证方式 |
|----|--------|----------|
| AC-23 | `go.mod` 声明 `go 1.21` | `grep "^go 1.21" go.mod` |
| AC-24 | main 源文件 ≤ 60 行 | `wc -l main.go`（排除空行和注释） |
| AC-25 | `add_test.go` 存在，6 个单元测试用例 | `go test -run TestAdd -v` 通过 6 个子测试 |
| AC-26 | `main_test.go` 存在，CLI 集成测试覆盖 AC-01 至 AC-22 | `go test -run TestCLI -v` 全部通过 |
| AC-27 | `go vet ./...` 无报错 | `go vet ./...` exit 0 |
| AC-28 | 5 平台交叉编译通过 | `GOOS=X GOARCH=Y go build` 各平台 exit 0 |
| AC-29 | `go mod verify` 通过 | `go mod verify` exit 0 |
| AC-30 | 生产构建命令可用 | `CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=v1.0.0" -trimpath -o add` exit 0 |

---

## 3. 验收条件 → 实现覆盖映射

以下说明每个 AC 如何在单一 Coder 全栈实现链中被覆盖。Coder 在一次完整实现中交付所有文件，无需拆分子任务。

### 3.1 核心功能 (AC-01 至 AC-10)

| AC | 代码覆盖点 | 实现方式 |
|----|-----------|----------|
| AC-01~AC-06, AC-09~AC-10 | `add(a, b int64) (int64, error)` 函数 | `add()` 函数执行 `a + b`，在加法前做溢出预检。AC-09/AC-10 加零不触发溢出。AC-04 由 `flag` 包自动处理 `--` 分隔符。 |
| AC-07 | 溢出检测：正溢出 | 加法前检查：`if a > 0 && b > 0 && a > math.MaxInt64 - b { return 0, overflowErr }` |
| AC-08 | 溢出检测：负溢出 | 加法前检查：`if a < 0 && b < 0 && a < math.MinInt64 - b { return 0, overflowErr }` |

溢出检测不依赖第三方包，用纯 Go 的符号比较在加法前预判，确保零依赖约束。

### 3.2 错误路径 (AC-11 至 AC-17)

| AC | 代码覆盖点 | 实现方式 |
|----|-----------|----------|
| AC-11, AC-12 | `main()` 中 `flag.Args()` 长度检查 | 解析 flags 后，`len(flag.Args()) < 2` → 打印缺失参数错误，退出 |
| AC-13 | 同上 | `len(flag.Args()) > 2` → 打印过多参数错误，退出 |
| AC-14, AC-15, AC-16 | `strconv.ParseInt` 错误处理 | 对 `flag.Arg(0)` 和 `flag.Arg(1)` 分别调用 `strconv.ParseInt(s, 10, 64)`；返回值 `err != nil` 则格式化错误消息，带引号包裹原始输入 |
| AC-17 | 输入长度预检 | 在 `ParseInt` 之前检查 `len(arg) > 40`，超长则打印长度错误，退出 |

所有错误消息格式严格遵循 PRD Appendix C 的 golden strings，包含 `Error: ` 前缀、英文描述、双引号包裹无效输入、末尾 `Usage: add <int_a> <int_b>`。

### 3.3 帮助与版本 (AC-18 至 AC-22)

| AC | 代码覆盖点 | 实现方式 |
|----|-----------|----------|
| AC-18, AC-19 | `flag.Usage` 覆盖 | 在 `init()` 或 `main()` 开头设置 `flag.Usage = func() { ... }`，输出 PRD Appendix D 的完整帮助文本。`flag` 包自动将 `-h`/`--help` 映射到 `flag.Usage`。 |
| AC-20, AC-21 | `flag.BoolVar` 版本标志 | 声明 `var showVersion = flag.Bool("v", false, ...)` + `flag.Bool("version", false, ...)`，解析后若为 true 则 `fmt.Printf("add %s\n", version)` 并 `os.Exit(0)`。 |
| AC-22 | 默认版本字符串 | 声明 `var version = "dev"`（package-level），CI 构建时通过 `-ldflags="-X main.version=v1.0.0"` 注入覆盖。 |

### 3.4 代码与构建 (AC-23 至 AC-30)

| AC | 覆盖点 | 实现方式 |
|----|--------|----------|
| AC-23 | `go.mod` | 文件内容写入 `module add` + `go 1.21` |
| AC-24 | main 源文件行数 | 保持 `main.go` ≤ 60 行（排除空行和注释）；若逻辑膨胀则抽出 `add.go` 单独放置 `add()` 函数 |
| AC-25 | `add_test.go` | 6 个 `t.Run()` 子测试：normal positive, normal negative, mixed sign, positive overflow, negative overflow, boundary exact。使用 table-driven test 风格。 |
| AC-26 | `main_test.go` | 22+ 个子测试，每个用 `exec.Command` 构建并运行二进制，捕获 stdout/stderr/exit code，逐条比对 golden output |
| AC-27 | `go vet` | 代码完成后执行 `go vet ./...` |
| AC-28 | 交叉编译 | Makefile `cross` 目标循环 `GOOS`/`GOARCH` 组合 |
| AC-29 | `go mod verify` | 作为 `make test` 或 CI 前置步骤 |
| AC-30 | 构建命令 | Makefile `build` 目标包含完整 ldflags，CI 中直接调用 |

---

## 4. 技术架构详细设计

### 4.1 包/模块结构

```
add/                          ← 仓库根目录（仅此一个模块）
├── main.go                   ← 入口：flag 解析、参数校验、调用 add()、输出
├── add.go                    ← 纯函数：add(a, b int64) (int64, error)
├── add_test.go               ← 单元测试：add() 函数 6 个用例
├── main_test.go              ← CLI 集成测试：exec.Command 覆盖全部 AC
├── go.mod                    ← module add, go 1.21
├── go.sum                    ← 空或仅含 go 版本哈希（零依赖）
├── Makefile                  ← build/test/vet/cross/clean
├── README.md                 ← 项目简介、安装、使用示例
└── docs/
    ├── prd.md                ← 本 PRD（产品经理交付）
    ├── context-audit.md      ← 上下文审计报告
    ├── tech-plan-draft.md    ← 本文档
    └── reports/              ← 各角色报告目录
```

### 4.2 main.go 结构（伪代码 / 流程）

```
1. package main
2. import ("flag", "fmt", "os", "strconv")
3. var version = "dev"
4. func init() { flag.Usage = customUsage }
5. func main() {
6.     showVersion := flag.Bool("v", false, ...) + flag.Bool("version", false, ...)
7.     flag.Parse()                                   // flag.ContinueOnError 保证 Parse 错误时继续
8.     if *showVersion { fmt.Printf("add %s\n", version); os.Exit(0) }
9.     args := flag.Args()
10.    if len(args) < 2 { printError(missingArgs); os.Exit(1) }
11.    if len(args) > 2 { printError(tooManyArgs); os.Exit(1) }
12.    a, err := parseArg(args[0])
13.    b, err := parseArg(args[1])
14.    result, err := add(a, b)
15.    if err != nil { printError(overflowError(a, b)); os.Exit(1) }
16.    fmt.Println(result)
17. }
```

行数估算：约 50 行（含空行和注释），符合 ≤ 60 行约束。

### 4.3 add.go 结构

```
package main

import "math"

func add(a, b int64) (int64, error) {
    // 正溢出检测
    if a > 0 && b > 0 && a > math.MaxInt64 - b {
        return 0, overflowError
    }
    // 负溢出检测
    if a < 0 && b < 0 && a < math.MinInt64 - b {
        return 0, overflowError
    }
    return a + b, nil
}
```

行数估算：约 15 行。无 `math/big`，无第三方包。

### 4.4 parseArg 辅助函数

```
func parseArg(s string) (int64, error) {
    if len(s) > 40 {
        return 0, errInputTooLong   // sentinel error
    }
    return strconv.ParseInt(s, 10, 64)
}
```

行数估算：约 8 行。可放在 `main.go` 或独立 `parse.go` 中。为保持 `main.go` ≤ 60 行，建议放入 `main.go` 末尾或独立文件。

### 4.5 错误消息生成

所有错误消息遵循统一模板。定义 helper：

```
func printError(msg string) {
    fmt.Fprintf(os.Stderr, "Error: %s. Usage: add <int_a> <int_b>\n", msg)
}
```

或直接为每种错误类型硬编码完整 golden string。PRD 对错误字符串有严格要求（精确匹配），因此采用字符串常量 + `fmt.Fprintf` 模式更安全。

### 4.6 版本机制

```
var version = "dev"              // package-level, mutable via ldflags

// main() 中:
versionFlag := flag.Bool("v", false, "show version number")
flag.BoolVar(versionFlag, "version", false, "show version number")
// ...
if *versionFlag {
    fmt.Printf("add %s\n", version)
    os.Exit(0)
}
```

`flag` 包自动处理 `-v` 与 `--version` 的别名绑定。使用 `BoolVar` 的方式是将同一个 `*bool` 绑定到两个 flag name。

### 4.7 错误退出策略

所有错误统一使用 `os.Exit(1)`。`flag` 配合 `flag.ContinueOnError`（详见下文 flag 配置），出错不自动退出，由 `main()` 统一处理和退出。这样可以确保所有错误消息格式一致（PRD 要求的标准格式，而非 `flag` 默认格式）。

#### flag 配置要点

```
flag.CommandLine.Init("add", flag.ContinueOnError)
flag.CommandLine.SetOutput(os.Stderr)
```

注意：`flag.Usage` 覆盖输出到 stderr 的默认行为——PRD 要求 `--help` 输出到 stdout。需要仔细处理：覆盖 `flag.Usage` 后，手动将 help 文本输出到 stdout，然后 exit 0。

实际上 `flag` 包在遇到 `-h`/`--help` 时会自动调用 `flag.Usage()` 然后 `os.Exit(2)`（默认错误处理）或返回错误（`ContinueOnError`）。使用 `ContinueOnError` 时，Parse 返回 `flag.ErrHelp`，可以在 main 中捕获并输出帮助到 stdout + exit 0。

推荐方案：

```
flag.CommandLine.Init("add", flag.ContinueOnError)
flag.CommandLine.SetOutput(os.Stderr)
flag.Usage = func() {
    fmt.Print(helpText)   // 输出到 stdout
}
// ...
if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
    if err == flag.ErrHelp {
        fmt.Print(helpText)
        os.Exit(0)
    }
    // other flag errors → exit 1
    printError(err.Error())
    os.Exit(1)
}
```

### 4.8 跨平台兼容性

| 平台 | 注意事项 |
|------|----------|
| Linux/amd64 | 默认构建目标，无特殊处理 |
| Linux/arm64 | 仅交叉编译验证 |
| Darwin/amd64 | Intel Mac |
| Darwin/arm64 | Apple Silicon；`CGO_ENABLED=0` 确保不依赖 Xcode |
| Windows/amd64 | `os.Stderr`/`os.Stdout` 行为一致；换行符 `\n` 在 Windows 终端正常显示 |

Go 标准库的 `flag`、`fmt`、`os`、`strconv` 在所有目标平台行为一致，无需平台条件编译。

---

## 5. 唯一实现 Job 说明

### 5.1 实现模型

本项目采用单 Coder、单实现链的全栈模型。无需前后端拆分、无需并行工作流。Coder 在一个完整实现链中交付所有产物。

### 5.2 Job 范围

| 维度 | 内容 |
|------|------|
| 源代码 | `main.go` (≤60行), `add.go`, 相关辅助函数 |
| 测试 | `add_test.go` (6 单元测试), `main_test.go` (22+ CLI 测试) |
| 构建 | `go.mod`, `Makefile` (5 目标: build/test/vet/cross/clean) |
| 文档 | `README.md` 重写 |
| 目录 | 确保 `docs/` 结构完整 |

### 5.3 实现顺序（单链）

```
Step 1: go.mod            → 模块初始化
Step 2: add.go            → add() 纯函数 + 溢出检测
Step 3: add_test.go       → 6 个单元测试，确保 add() 正确
Step 4: main.go           → flag 解析 + 参数校验 + 错误消息 + 帮助/版本
Step 5: main_test.go      → CLI 集成测试，覆盖全部 AC-01 至 AC-22
Step 6: Makefile          → 构建目标
Step 7: README.md         → 项目文档
Step 8: go vet + cross    → 质量门
Step 9: 集成交付检查      → 全量验证（见 §7）
```

Step 3 在 Step 4 之前执行，确保核心计算逻辑在集成前已验证通过（Test-Driven 原则：先测核心，再测入口）。

---

## 6. 文件级设计细节

### 6.1 main.go 约束分析

目标：≤ 60 行（排除空行和注释），同时覆盖 flag 解析、参数校验（缺失/过多/非整数/超长）、溢出检测调用、帮助输出、版本输出、错误消息生成。

策略：
- `add()` 函数从 `main.go` 移到 `add.go`，减省约 15 行
- `parseArg()` 辅助函数保留在 `main.go` 或独立文件
- `printError()` 辅助函数精简为一行
- 帮助文本定义为 const string，不参与行数计算（PRD 规定排除注释和空行，字符串字面量应正常计数；但帮助文本天然多行，需折中）

若帮助文本过长导致 `main.go` 超 60 行，将帮助文本常量移至独立文件 `help.go`。

### 6.2 add.go

```go
package main

import "math"

func add(a, b int64) (int64, error) {
    if a > 0 && b > 0 && a > math.MaxInt64-b {
        return 0, errOverflow
    }
    if a < 0 && b < 0 && a < math.MinInt64-b {
        return 0, errOverflow
    }
    return a + b, nil
}
```

### 6.3 add_test.go 用例表

| 子测试名称 | 输入 a | 输入 b | 预期结果 | 预期错误 |
|------------|--------|--------|----------|----------|
| NormalPositive | 2 | 3 | 5 | nil |
| NormalNegative | -7 | -3 | -10 | nil |
| MixedSign | -5 | 10 | 5 | nil |
| PositiveOverflow | 9223372036854775807 | 1 | 0 | non-nil |
| NegativeOverflow | -9223372036854775808 | -1 | 0 | non-nil |
| BoundaryExact | 9223372036854775807 | 0 | 9223372036854775807 | nil |

### 6.4 main_test.go CLI 测试模式

每个 CLI 测试遵循统一模式：

```go
func TestCLI_Add(t *testing.T) {
    binary := buildBinary(t)    // go build -o /tmp/test-add
    tests := []struct{
        name string
        args []string
        wantStdout string
        wantStderr string
        wantExit int
    }{
        {name: "AC-01", args: []string{"2","3"}, wantStdout: "5\n", wantExit: 0},
        // ... 22+ cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := exec.Command(binary, tt.args...)
            stdout, _ := cmd.Output()       // or CombinedOutput for stderr
            // assert stdout/stderr/exit
        })
    }
}
```

注意事项：
- 次测试前 `go build` 一次，复用二进制
- 使用 `cmd.Output()` 获取 stdout，但 exit code 需通过 `cmd.Run()` + `cmd.ProcessState.ExitCode()` 获取
- 负参数（如 `-5`）需使用 `"--", "-5"` 模板，或直接传 `"-5"` 给 `exec.Command`（`exec.Command` 不经过 shell 分词，直接传字符串不会触发 flag 解析歧义）

### 6.5 Makefile 目标

```makefile
BINARY := add
VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test vet cross clean

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY)

test:
	go test -v ./...

vet:
	go vet ./...

cross:
	GOOS=linux   GOARCH=amd64 go build -o /dev/null
	GOOS=linux   GOARCH=arm64 go build -o /dev/null
	GOOS=darwin  GOARCH=amd64 go build -o /dev/null
	GOOS=darwin  GOARCH=arm64 go build -o /dev/null
	GOOS=windows GOARCH=amd64 go build -o /dev/null

clean:
	rm -f $(BINARY)
```

---

## 7. 集成交付检查清单

Coder 完成实现后，按以下清单逐项验证：

### 7.1 构建维度

- [ ] `go mod verify` 返回 exit 0
- [ ] `go vet ./...` 返回 exit 0，无 warning
- [ ] `make build VERSION=v1.0.0` 成功生成 `add` 二进制
- [ ] `make cross` 全部 5 平台通过
- [ ] `file add` 确认静态链接（`statically linked`）
- [ ] main 源文件行数 ≤ 60（`grep -cve '^\s*$' -e '^\s*//' main.go` ≤ 60）

### 7.2 功能维度

- [ ] AC-01 至 AC-10：核心功能全部通过（手动运行或 `go test`）
- [ ] AC-11 至 AC-17：错误路径全部通过，精确比对 golden error string
- [ ] AC-18 至 AC-22：帮助与版本输出正确

### 7.3 测试维度

- [ ] `go test -v -run TestAdd` 6 个子测试全部 PASS
- [ ] `go test -v -run TestCLI` 覆盖所有 AC-01 至 AC-22，全部 PASS
- [ ] 测试覆盖率 `go test -cover` ≥ 90%（期望 100%，但不强求）

### 7.4 文档维度

- [ ] README.md 已重写，包含项目简介、安装方式（`go install`）、使用示例、构建说明
- [ ] `docs/` 目录包含 prd.md、context-audit.md、tech-plan-draft.md

### 7.5 交付维度

- [ ] 所有变更在 `28/delivery` 分支
- [ ] Git 提交信息清晰描述变更内容
- [ ] 最终 `git status` 无未跟踪的必要文件

---

## 8. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| `main.go` 超过 60 行 | 违反 AC-24 | 中 | 将帮助文本常量、`parseArg` 辅助函数抽至独立文件 |
| `flag` 包对 `--` 的行为与预期不一致 | AC-04 失败 | 低 | Go 1.21 `flag` 包稳定行为：`--` 后所有内容视为 positional args；在 `main_test.go` 中显式覆盖此 case |
| `strconv.ParseInt` 对超长字符串行为 | AC-17 可能被 ParseInt 的 `ErrSyntax` 覆盖而非长度错误 | 低 | 在 `ParseInt` 之前先检查 `len(s) > 40`，短路返回长度错误 |
| 交叉编译平台兼容性 | Linux/arm64 在 macOS 上交叉编译需额外配置 | 低 | CGO_ENABLED=0 消除 C 依赖；`go tool dist list` 确认支持 |
| 错误消息字符串精确匹配 | 测试失败因空格/标点差异 | 中 | 使用字符串常量，测试中直接引用同一常量，避免硬编码 golden string 重复 |

---

## 9. 附录

### 9.1 错误常量定义（建议）

```go
const (
    helpText = `add — integer addition tool

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
`

    errMsgMissing  = "missing arguments, need two integers"
    errMsgTooMany  = "too many arguments, only two integers accepted"
    errMsgTooLong  = "input too long, maximum 40 characters per argument"
    errMsgOverflow = "integer overflow"
)
```

### 9.2 溢出检测数学证明

对于 int64 加法 `a + b`：
- 正溢出：`a > 0 && b > 0 && a > MaxInt64 - b`
- 负溢出：`a < 0 && b < 0 && a < MinInt64 - b`

证明：
- MaxInt64 = 2^63 - 1 = 9223372036854775807
- MinInt64 = -2^63 = -9223372036854775808
- 若 `a + b > MaxInt64`，则 `a > MaxInt64 - b`（给定 `b > 0`，减法不溢出）
- 若 `a + b < MinInt64`，则 `a < MinInt64 - b`（给定 `b < 0`，减法不溢出）

此公式在 `a > 0, b > 0` 和 `a < 0, b < 0` 场景下安全无歧义，且无需实际执行溢出加法后再检测。

---

> **本草案由架构师撰写，供 Coder 直接作为技术实现指南使用。Coder 在实现过程中如需调整细节，应记录偏差并更新本文档。**
