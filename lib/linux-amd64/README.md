# linux-amd64 平台占位

此目录用于放置 Linux x86_64 平台的 `libsimple.so` 及 Jieba 词典。

## 使用方法

1. 在 [chrwhy/simple](https://github.com/chrwhy/simple) 编译 Linux 版本的 `libsimple`：
   ```bash
   # 克隆并编译（需在 Linux 环境下）
   git clone https://github.com/chrwhy/simple.git
   cd simple
   make
   ```
2. 将编译产物 `libsimple.so` 复制到此目录。
3. 复制 `dict/` 词典目录（可从 `darwin-amd64/dict/` 或 `darwin-arm64/dict/` 直接复制，词典与平台无关）。
4. 运行程序即可自动检测。
