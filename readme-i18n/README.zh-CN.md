# Tairitsu

Tairitsu 是一个面向独立安装 ZeroTier 控制器的自托管 Web 控制台，重点覆盖日常主线管理：初始化、登录与注册、网络与成员管理、网络设置、导入接管，以及用户治理。

它当前面向 **SQLite、单实例、自托管** 部署场景，不追求做成像 `ztnet` 那样更厚的平台型系统；`Planet` 生成也仍然保留为实验性能力。

## 当前范围

- 设置向导：控制器连接、SQLite、初始管理员创建
- 用户能力：注册、登录、修改密码、会话管理、管理员转让、管理员创建用户、重置密码、删除用户
- 网络 owner 视角下的网络与成员管理
- IPv4、IPv6、Managed Routes、DNS、多播设置
- 导入尚未被 Tairitsu 接管的控制器网络

## 访问模型

- 每个 ZeroTier 网络在 Tairitsu 中都恰好有一个 `owner`
- 网络 `owner` 负责该网络及其成员管理
- 平台 `admin` 负责初始化、用户治理和网络导入
- 平台 `admin` 不会自动获得所有已归属网络的读写权限

## 部署

当前公开部署主路径是 Docker / Podman：

- 镜像：`ghcr.io/gt-610/tairitsu:latest`
- 当前发布镜像目标：`linux/amd64`
- 控制器要求：独立安装的 ZeroTier 控制器，并通过 `local.conf` 开放 API 访问

详细文档：

- [安装与部署](../docs/INSTALLATION.md)
- [运行维护与支持边界](../docs/OPERATIONS.md)
- [API 文档](../docs/api/Tairitsu_API_Documentation.md)

手动主机安装可以用于开发，但当前公开文档主路径仍是 Docker / Podman。

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

发布前基线：

- `go test ./...`
- `bun test`
- `bun run lint`
- `bun run build`
- `docker build .`

## 许可证

[GNU GPL v3](../LICENSE)

## 法律声明

Tairitsu 不重新分发任何 ZeroTier 控制器代码，而是通过官方 API 与**独立安装**的 ZeroTier 控制器通信，避免落入 ZeroTier 控制器专有代码分发范围。

`Generate Planet` 功能基于 [ztnodeid-go](https://github.com/kmahyyg/ztnodeid-go) 修改，遵循其 GNU GPL v3 许可证，**并非**来自 ZeroTier 本身。

Tairitsu **不是** ZeroTier 的产品，也**不隶属于**、**不受认可于**、**不受支持于** ZeroTier, Inc.
