# lib/ — 原生库

本目录按 `{goos}-{goarch}` 组织平台相关的原生库。词典是纯文本，与平台无关，统一放在项目根目录的 `dict/` 下。

程序启动时通过 `runtime.GOOS` / `runtime.GOARCH` 自动拼接库路径，如：
- macOS Intel → `lib/darwin-amd64/libsimple`
- macOS Apple Silicon → `lib/darwin-arm64/libsimple`
- Linux x86_64 → `lib/linux-amd64/libsimple`

## 目录结构

```
lib/
├── darwin-amd64/libsimple.dylib   # macOS Intel
├── darwin-arm64/libsimple.dylib   # macOS Apple Silicon
└── linux-amd64/libsimple.so       # Linux x86_64（需自行编译）
```

词典位于项目根目录 `dict/`（与 `lib/` 平级），详见 [dict/README.md](../dict/README.md)。

## 添加新平台支持

1. 在 [chrwhy/simple](https://github.com/chrwhy/simple) 按说明编译对应平台的 `libsimple` 动态库。
2. 创建 `lib/{goos}-{goarch}/` 目录（如 `lib/linux-amd64/`）。
3. 将编译产物（`libsimple.so` 或 `libsimple.dylib`）放入该目录。
4. 如库文件名不是 `libsimple`，需通过 `SS_LIB_PATH` 环境变量指定完整路径。

## 词典来源

项目根目录 `dict/` 下的词典文件来自 [CppJieba](https://github.com/yanyiwu/cppjieba)，libsimple 通过 `./dict/` 相对路径加载（需从项目根目录运行程序）。
