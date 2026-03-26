# scripts/

此目录用于放置构建和部署辅助脚本。

## 关于原生库文件

`lib/` 目录下的 `.dylib` / `.so` 文件是预编译的原生库，已直接包含在仓库中。

对于正式项目或库文件较大的情况，建议使用 Git LFS 管理：

```bash
# 安装 git-lfs
brew install git-lfs  # macOS
# 或
apt install git-lfs   # Ubuntu/Debian

# 初始化
git lfs install

# 追踪原生库文件
git lfs track "lib/**/*.dylib"
git lfs track "lib/**/*.so"

# 提交 .gitattributes
git add .gitattributes
git commit -m "Track native libraries with Git LFS"
```

## 编译 libsimple

如需自行编译最新版本的 libsimple，请参考 [chrwhy/simple](https://github.com/chrwhy/simple) 仓库的说明。
