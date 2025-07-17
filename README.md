# kiro2cc

一个让claude-code使用kiro免费claude模型的工具。它能自动帮你管理kiro授权、启动本地代理并配置好环境变量。

## 1. 安装

### macOS / Linux
```bash
curl -sSL https://raw.githubusercontent.com/deepwind/kiro2cc/main/install.sh | bash
```
> 脚本会将 `kiro2cc` 安装到 `$HOME/.local/bin`。请确保这个目录在你的 `PATH` 中。

### Windows
1.  从 [Releases 页面](https://github.com/deepwind/kiro2cc/releases/latest)下载最新的 `.exe` 文件。
2.  重命名为 `kiro2cc.exe` 并放到一个在系统 `Path` 环境变量里的目录。

## 2. 使用

安装 `claude` 后，用 `kiro2cc claude` 来启动 `claude-code`。

```bash
# 启动 claude-code
kiro2cc claude

# 也可以传递参数
kiro2cc claude --version
```

`kiro2cc` 会在后台自动处理好一切。

## 3. 停止

当你用完后，可以停止后台服务。

```bash
kiro2cc stop
```

---

### 其他命令

- `kiro2cc server --daemon`: 在后台启动服务。
- `kiro2cc refresh`: 手动刷新 token。
- `kiro2cc read`: 查看当前 token 状态。

本项目使用 MIT 许可证。
