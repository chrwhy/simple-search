# simple-search

基于 **SQLite FTS5** 与 **simple 分词拓展** 的简易全文检索演示程序。使用带拼音能力的 **simple** 分词器扩展检索中文标题与正文，并在交互式 REPL 中查询。

## simple 扩展来源

本仓库中的 `**libsimple` 动态库**（如 `libsimple-osx-x64/libsimple`）来自 **[chrwhy/simple](https://github.com/chrwhy/simple)**：支持中文与拼音的 SQLite FTS5 全文检索扩展。需要其他平台或更新版本时，请在该仓库按说明自行编译，并将产物路径配置到 `dao/dao.go` 的 `InitDB` 中。

## 项目做什么

- 在本地 SQLite 数据库中维护两张 FTS5 虚拟表：`docs_name`（标题，拼音分词开启）与 `docs_content`（正文，拼音分词关闭）。
- 启动时用示例数据填充（若表为空）；你在终端里输入自然语言查询，程序用 Jieba 切词并拼成 FTS5 `MATCH` 子句，对标题或正文做检索。
- 支持高亮片段展示，并可选执行原始 SQL 做实验。

技术栈概览：**Go 1.23**、`github.com/mattn/go-sqlite3`（CGO）、**FTS5**、[chrwhy/simple](https://github.com/chrwhy/simple) 提供的 `**simple` tokenizer**（本仓库预置 `libsimple-osx-x64`）、`github.com/chrwhy/open-pinyin`（拼音 parser）、`github.com/yanyiwu/gojieba`（结巴分词）。

## 环境要求

- **Go**：1.23 或以上（见 `go.mod`）。
- **CGO**：`go-sqlite3` 需要本机 C 编译器（macOS 上通常已具备 Xcode Command Line Tools）。
- **平台**：当前代码里加载的 `simple` 扩展路径为 `**./libsimple-osx-x64/libsimple`**，适用于 **macOS x64**（Apple Silicon 上若通过 Rosetta 运行 x64 二进制需自行验证）。其他系统需替换为对应平台的 `libsimple` 并修改 `dao/dao.go` 中的扩展路径。

## 快速运行

在项目根目录执行（保证当前工作目录为仓库根目录，以便加载扩展与字典）：

```bash
go mod download
go build --tags fts5 -o gosimple .
./gosimple
```

若曾用旧表结构生成过数据库，可先删除再启动，以便按新 schema 建表并重新灌数：

```bash
rm -f example.db
./gosimple
```

仓库提供了脚本 `dev_run.sh`，会编译、删除 `example.db` 并启动程序：

```bash
chmod +x dev_run.sh
./dev_run.sh
```

## 使用说明

启动后进入交互菜单：


| 选项    | 说明                                                                            |
| ----- |-------------------------------------------------------------------------------|
| **1** | 按 **标题**（`docs_name.name`）搜索；输入会被解析为 `MATCH` 子句。 输入 `exit` 返回。                |
| **2** | 按 **正文**（`docs_content.content`）搜索，会被 Jieba 进行分词解析为 `MATCH` 子句。输入 `exit` 返回。 |
| **3** | **原始 SQL**：直接对 `example.db` 执行只读查询（实现为 `Query`），便于调试 FTS5。输入 `exit` 返回。       |
| **4** | 退出程序。                                                                         |


表结构（逻辑列，FTS5 内部以全文索引方式存储）：

- `docs_name`：`fid`, `name`, `cate`, `ctime` — `tokenize='simple'`（拼音相关能力打开）。
- `docs_content`：`fid`, `content`, `cate`, `ctime` — `tokenize='simple 0'`。

示例数据与插入逻辑见 `dao/data-loader.go`；建表与查询见 `dao/dao.go`。

## 常见问题

- **启动报错找不到扩展**：确认在**项目根目录**运行二进制，且 `libsimple-osx-x64/libsimple` 存在；非 macOS 或非 x64 需自行提供扩展并改 `InitDB` 中的路径。
- **改表结构后查询异常**：删除 `example.db` 后重新运行，或自行迁移数据（`CREATE VIRTUAL TABLE IF NOT EXISTS` 不会修改已有 FTS 表定义）。
- **依赖下载慢**：可配置 `GOPROXY` 等 Go 模块代理后再执行 `go mod download`。

## 许可证

以仓库内声明为准（若未单独声明，请向维护者确认）。