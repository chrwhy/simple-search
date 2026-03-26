# simple-search

[![Go Report Card](https://goreportcard.com/badge/github.com/chrwhy/simple-search)](https://goreportcard.com/report/github.com/chrwhy/simple-search)
[![Go Reference](https://pkg.go.dev/badge/github.com/chrwhy/simple-search.svg)](https://pkg.go.dev/github.com/chrwhy/simple-search)

基于 **SQLite FTS5** 与 **simple 分词拓展** 的简易全文检索演示程序。使用带拼音能力的 **simple** 分词器扩展检索中文标题与正文，并在交互式 REPL 中查询。

## simple 扩展来源

本仓库中的 **libsimple 动态库**（位于 `lib/{os}-{arch}/libsimple`）来自 **[chrwhy/simple](https://github.com/chrwhy/simple)**：支持中文与拼音的 SQLite FTS5 全文检索扩展。需要其他平台或更新版本时，请在该仓库按说明自行编译，并将产物放到 `lib/{os}-{arch}/` 目录下（如 `lib/linux-amd64/libsimple.so`）。

## 项目做什么

- 在本地 SQLite 数据库中维护两张 FTS5 虚拟表：`docs_name`（标题，拼音分词开启）与 `docs_content`（正文，拼音分词关闭）。
- 启动时用示例数据填充（若表为空）；你在终端里输入自然语言查询，程序用 Jieba 切词并拼成 FTS5 `MATCH` 子句，对标题或正文做检索。
- 支持高亮片段展示，并可选执行原始 SQL 做实验。

技术栈概览：**Go 1.23**、`github.com/mattn/go-sqlite3`（CGO）、**FTS5**、[chrwhy/simple](https://github.com/chrwhy/simple) 提供的 **simple tokenizer**、[chrwhy/simple-parser](https://github.com/chrwhy/simple-parser)（查询解析 SDK，含 Jieba 分词与拼音扩展）。

## 环境要求

- **Go**：1.23 或以上（见 `go.mod`）。
- **CGO**：`go-sqlite3` 需要本机 C 编译器（macOS 上通常已具备 Xcode Command Line Tools）。
- **平台**：当前预置了 macOS（Intel + Apple Silicon）的 `libsimple`。程序会根据 `runtime.GOOS` / `runtime.GOARCH` 自动检测库路径。其他平台（如 Linux）需自行提供对应平台的 `libsimple` 并放到 `lib/{os}-{arch}/` 目录下。

## 快速运行

在项目根目录执行（保证当前工作目录为仓库根目录，以便加载扩展与字典）：

```bash
# 使用 Makefile（推荐）
make run

# 或手动编译运行
go mod download
go build --tags fts5 -o simple-search .
./simple-search
```

若曾用旧表结构生成过数据库，可先删除再启动，以便按新 schema 建表并重新灌数：

```bash
make dev   # 编译 + 删除旧库 + 运行
```

### Makefile 命令

| 命令 | 说明 |
|---|---|
| `make build` | 编译二进制 |
| `make run` | 编译并运行 |
| `make test` | 运行全部测试 |
| `make dev` | 清除旧库后编译运行 |
| `make clean` | 删除二进制和数据库文件 |
| `make fmt` | 格式化代码 |
| `make vet` | 静态检查 |
| `make lint` | fmt + vet |
| `make import FILE=data/extra.json` | 增量导入数据 |

### Docker

```bash
docker build -t simple-search .
docker run -it -v simple-search-db:/app/db simple-search
```

## 环境变量配置

所有配置项均可通过环境变量覆盖（前缀 `SS_`），参见 `.env.example`：

| 环境变量 | 默认值 | 说明 |
|---|---|---|
| `SS_DB_PATH` | `example.db` | SQLite 数据库文件路径 |
| `SS_LIB_PATH` | 自动检测 | libsimple 库路径 |
| `SS_DATA_PATH` | `data/sample.json` | 示例数据文件路径 |
| `SS_NAME_TABLE` | `docs_name` | 标题 FTS5 虚拟表名 |
| `SS_CONTENT_TABLE` | `docs_content` | 内容 FTS5 虚拟表名 |
| `SS_NAME_TOKENIZER` | `simple` | 标题表分词器（含拼音） |
| `SS_CONTENT_TOKENIZER` | `simple 0` | 内容表分词器（无拼音） |

使用 `.env` 文件：

```bash
cp .env.example .env
# 编辑 .env 按需修改，然后：
source .env && ./simple-search
```

## 使用说明

### 命令行参数

| 参数 | 说明 |
|---|---|
| `-query <关键词>` | 按标题搜索并退出（单次模式） |
| `-content <关键词>` | 按内容搜索并退出（单次模式） |
| `-sql <SQL语句>` | 执行原始 SQL 并退出 |
| `-import <文件路径>` | 从 JSON 文件增量导入数据并退出 |

### 交互模式

启动后进入交互菜单：

| 选项 | 说明 |
|---|---|
| **1** | 按 **标题**（`docs_name.name`）搜索；输入会被解析为 `MATCH` 子句。输入 `exit` 返回。 |
| **2** | 按 **正文**（`docs_content.content`）搜索，会被 Jieba 进行分词解析为 `MATCH` 子句。输入 `exit` 返回。 |
| **3** | **原始 SQL**：直接对 `example.db` 执行只读查询（实现为 `Query`），便于调试 FTS5。输入 `exit` 返回。 |
| **4** | 退出程序。 |

表结构（逻辑列，FTS5 内部以全文索引方式存储）：

- `docs_name`：`fid`, `name`, `cate`, `ctime` — `tokenize='simple'`（拼音相关能力打开）。
- `docs_content`：`fid`, `content`, `cate`, `ctime` — `tokenize='simple 0'`。

示例数据与插入逻辑见 `dao/data-loader.go`；建表与查询见 `dao/dao.go`。

## 跨平台支持

程序通过 `runtime.GOOS` / `runtime.GOARCH` 自动拼接库路径（`lib/{goos}-{goarch}/libsimple`）。当前仓库预置了：

- `lib/darwin-amd64/` — macOS Intel
- `lib/darwin-arm64/` — macOS Apple Silicon

如需 Linux 支持，请在 [chrwhy/simple](https://github.com/chrwhy/simple) 编译对应平台的 `libsimple.so`，放到 `lib/linux-amd64/` 目录下即可。

## 项目结构

```
simple-search/
├── main.go              # 入口：CLI 参数 + REPL 交互
├── config/
│   ├── config.go        # 配置加载（.env / 环境变量）+ 校验
│   └── config_test.go   # 配置单元测试
├── dao/
│   ├── dao.go           # 数据库操作：建表、插入、查询
│   ├── data-loader.go   # JSON 数据加载（全量 / 增量）
│   ├── dao_test.go      # sanitizeMatch 测试
│   └── integration_test.go  # 内存 SQLite 集成测试
├── model/
│   └── doc.go           # Doc 数据模型
├── data/
│   └── sample.json      # 示例数据（含边界测试用例）
├── lib/                 # libsimple 原生库（仅 .dylib/.so，按平台组织）
├── dict/                # Jieba 分词词典（纯文本，跨平台共享）
├── Makefile             # 构建 / 测试 / 运行
├── Dockerfile           # 多阶段 Docker 构建
├── .env.example         # 环境变量模板
└── .gitattributes       # 二进制文件标记
```

## 常见问题

- **启动报错找不到扩展**：确认在**项目根目录**运行二进制，且 `lib/{os}-{arch}/libsimple` 存在。可通过 `SS_LIB_PATH` 环境变量指定绝对路径。
- **改表结构后查询异常**：删除 `example.db` 后重新运行，或自行迁移数据（`CREATE VIRTUAL TABLE IF NOT EXISTS` 不会修改已有 FTS 表定义）。
- **依赖下载慢**：可配置 `GOPROXY` 等 Go 模块代理后再执行 `go mod download`。

## 许可证

以仓库内声明为准（若未单独声明，请向维护者确认）。
