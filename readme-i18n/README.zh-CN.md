# Tairitsu

Tairitsu 是一个面向独立安装 ZeroTier 控制器的自托管 Web 控制台，提供初始化、用户账户、网络与成员管理、网络设置，以及控制器网络导入等能力。

## 功能

- 设置向导：控制器连接、SQLite、初始管理员创建
- 用户能力：注册、登录、修改密码、会话管理、管理员转让、管理员创建用户、重置密码、删除用户
- 网络创建、详情查看，以及成员授权 / 拒绝 / 移除
- IPv4、IPv6、Managed Routes、DNS、多播设置
- 导入尚未被 Tairitsu 接管的控制器网络
- 实验性的 `Planet` 生成功能

## 安装

前置条件：

- Tairitsu 可访问的 ZeroTier 控制器
- 控制器 token 文件访问权限
- 一个持久化的 Tairitsu 数据目录
- Docker 或 Podman 等容器运行时

先让 ZeroTier 控制器向 Tairitsu 开放本地控制器 API。编辑 `/var/lib/zerotier-one/local.conf`：

```json
{
  "settings": {
    "allowManagementFrom": [
      "0.0.0.0/0",
      "::/0"
    ]
  }
}
```

修改后重启 ZeroTier 控制器。

使用 Docker 运行 Tairitsu：

```bash
docker run -d \
  --name tairitsu \
  --restart unless-stopped \
  -p 3000:3000 \
  -v /var/lib/zerotier-one:/var/lib/zerotier-one \
  -v /path/to/tairitsu-data:/app/data \
  ghcr.io/gt-610/tairitsu:latest
```

或者使用 Docker Compose：

```yaml
services:
  tairitsu:
    image: ghcr.io/gt-610/tairitsu:latest
    container_name: tairitsu
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - /var/lib/zerotier-one:/var/lib/zerotier-one
      - /path/to/tairitsu-data:/app/data
```

打开 `http://<host>:3000`，按设置向导完成初始化：

1. 填写 ZeroTier 控制器 URL。
2. 填写 token 文件路径。
3. 按需设置 SQLite 数据库路径。
4. 创建初始管理员账户。

## 文档

- [运行维护与支持边界](../docs/OPERATIONS.md)
- [API 文档](../docs/api/Tairitsu_API_Documentation.md)

## 开发

前置条件：

- Go 1.25 或更高版本（需启用 CGO）
- Bun 1.3 或更高版本
- 本地或 Docker 中运行的 ZeroTier 控制器

构建后端：

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build -o ./build/tairitsu ./cmd/tairitsu
```

运行前端：

```bash
cd web
bun install
bun run dev
```

## 许可证

[GNU GPL v3](../LICENSE)

## 法律声明

Tairitsu 不重新分发任何 ZeroTier 控制器代码，而是通过官方 API 与**独立安装**的 ZeroTier 控制器通信，避免落入 ZeroTier 控制器专有代码分发范围。

`Generate Planet` 功能基于 [ztnodeid-go](https://github.com/kmahyyg/ztnodeid-go) 修改，遵循其 GNU GPL v3 许可证，**并非**来自 ZeroTier 本身。

Tairitsu **不是** ZeroTier 的产品，也**不隶属于**、**不受认可于**、**不受支持于** ZeroTier, Inc.
