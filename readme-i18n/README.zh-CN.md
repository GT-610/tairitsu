# Tairitsu

**注意：此项目仍在开发中。当前优先目标是面向单实例自托管部署的 SQLite 优先 MVP。**

Tairitsu 是 ZeroTier 控制器的 Web 管理界面实现，提供友好的 GUI 来管理 ZeroTier 网络、成员和配置，由 Golang 后端（通过官方 API 与 ZeroTier 控制器通信）和 React 前端组成。

## 功能特性

- **网络管理**：创建、编辑和删除 ZeroTier 网络
- **成员管理**：管理网络成员、授权设备并分配 IP
- **配置控制**：配置网络设置，包括 IP 范围、路由和规则
- **实时状态**：监控网络和成员状态
- **SQLite 优先 MVP**：一期仅正式支持 SQLite
- **安全认证**：基于 JWT 的安全访问认证
- **响应式设计**：Material Design 现代化响应式界面
- **管理工具**：支持导入现有 ZeroTier 网络，并在需要时生成自定义 Planet 文件

## 一期支持矩阵

- 正式支持：SQLite、单实例自托管 ZeroTier 控制器、小规模管理员场景
- 开发中：更完整的账户设置与 Planet 工具链
- 当前不作为一期承诺：MySQL、PostgreSQL

## 当前运行模型

- 后端已经收敛为显式依赖装配路径，统一组装配置、数据库、ZeroTier client、service、handler 与 middleware。
- setup 阶段路由与运行时路由按应用状态隔离，不再依赖单独的“重载路由”步骤。
- 前后端按同版本一起分发，因此内部 API 会直接清理无用兼容层，而不是长期保留兼容接口。

## 部署

### Docker / Podman

1. 在 ZeroTier 主目录中创建 `local.conf` 文件（通常位于 `/var/lib/zerotier-one`）。如果已存在 `local.conf` 文件，请跳过此步骤。

2. 在 ZeroTier 的 `local.conf` 中配置 `allowManagementFrom`：

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

   这一步会让 ZeroTier 控制器可从任何 IP 地址访问。

   如果需要更严格的访问控制，也可以这样改，但是需要确保 Tairitsu 容器能够访问该 IP：

   ```json
   {
      "settings": {
         "allowManagementFrom": [
               "<本地IP网段>",
         ]
      }
   }
   ```

   修改配置后，重启 ZeroTier 容器。

3. **运行 Tairitsu 容器**

   ```bash
   docker run -d \
       --name tairitsu \
       -p 3000:3000 \
       -v /var/lib/zerotier-one:/var/lib/zerotier-one \
       -v path/to/tairitsu/data:/app/data \
       ghcr.io/gt-610/tairitsu:latest
   ```

   或使用 Compose：

   ```yaml
   services:
     tairitsu:
       image: ghcr.io/gt-610/tairitsu:latest
       ports:
         - 3000:3000
       volumes:
         - /var/lib/zerotier-one:/var/lib/zerotier-one
         - path/to/tairitsu/data:/app/data
   ```

### 手动安装

还没写完。

### 开发

#### 前置条件

- Go 1.25 或更高版本（需启用 CGO）
- Bun 1.3 或更高版本
- 本地或 Docker 中运行的 ZeroTier 控制器

#### 后端
```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build -o ./build/tairitsu ./cmd/tairitsu
```

#### 前端
```bash
cd web
bun install
bun run dev
```

前端开发服务器默认在 3000 端口启动，并将 API 请求代理到后端服务器。

生产构建命令：

```bash
cd web
bun run build
```

附加内部文档：

- [一期问题地图](../docs/phase1-audit.md)
- [ZeroTier Docker 手工联调清单](../docs/testing/zerotier-docker-checklist.md)

## 贡献

欢迎贡献！有问题就交 Issue 或发 PR。

## 许可证

[GNU GPL v3](LICENSE)。

## 法律声明

自 1.16.0 版本起，ZeroTier 的控制器组件采用[商业性的、专有的非开源许可协议](https://github.com/zerotier/ZeroTierOne/blob/main/nonfree/LICENSE.md)。Tairitsu 从未重新分发任何 ZeroTier 控制器代码，完全符合 ZeroTier 的许可条款。

### ZeroTier 许可证合规性
Tairitsu 是 ZeroTier 网络的独立管理界面。本项目**不包含、分发或修改**任何 ZeroTier 源代码或二进制文件。

该软件通过官方 API 与**独立安装**的 ZeroTier 控制器通信。用户必须在遵守许可的前提下，部署自己的 ZeroTier 控制器。

“生成 Planet” 功能基于 [ztnodeid-go](https://github.com/kmahyyg/ztnodeid-go) 修改，遵循其 GNU GPL v3 许可证，**并非**来自 ZeroTier 本身。

---

Tairitsu **不是** ZeroTier 的产品。它**不隶属于** ZeroTier 公司，也未经其认可或支持。
