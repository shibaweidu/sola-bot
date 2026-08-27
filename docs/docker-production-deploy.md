# GitHub 自动构建与 Docker 生产部署

本方案由 GitHub Actions 构建镜像并推送到 GitHub Container Registry（GHCR）。生产服务器只需要 Docker、`docker-compose.prod.yml`、`scripts/deploy.sh` 和 `.env`，不需要安装 Go、Node.js，也不需要保存项目源码。

## 发布的镜像

镜像前缀默认为：

```text
ghcr.io/shibaweidu/sola-bot
```

每次构建会发布以下镜像：

```text
ghcr.io/shibaweidu/sola-bot/api
ghcr.io/shibaweidu/sola-bot/bot
ghcr.io/shibaweidu/sola-bot/worker
ghcr.io/shibaweidu/sola-bot/web
ghcr.io/shibaweidu/sola-bot/migrate
```

标签规则：

- `main`：发布 `latest` 和 `sha-xxxxxxx`。
- 其他已配置分支：发布分支名和 `sha-xxxxxxx`。
- Pull Request：运行测试和镜像构建检查，但不推送镜像。

## 第一次初始化服务器

推荐 Ubuntu 22.04 或 24.04。先安装 Docker，并让部署用户可以直接执行 Docker：

```bash
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker "$USER"
```

重新登录 SSH 后创建部署目录：

```bash
sudo mkdir -p /opt/sola-bot/scripts
sudo chown -R "$USER:$USER" /opt/sola-bot
cd /opt/sola-bot
```

创建生产 `.env`。可以从仓库的 `.env.example` 开始，但必须替换 Bot Token、管理员密码、数据库密码和 JWT 密钥：

```bash
curl -fsSL https://raw.githubusercontent.com/shibaweidu/sola-bot/main/.env.example -o .env
nano .env
chmod 600 .env
```

生产镜像配置：

```env
SOLA_IMAGE_REGISTRY=ghcr.io/shibaweidu/sola-bot
SOLA_IMAGE_TAG=latest
SOLA_HTTP_BIND=0.0.0.0
SOLA_HTTP_PORT=80
```

`.env` 只保留在服务器，不要提交到 GitHub。

## 配置 GitHub 自动部署

在 GitHub 仓库的 `Settings -> Environments` 创建 `production` 环境，然后在该环境或仓库的 Actions Secrets 中配置：

| Secret | 说明 |
|---|---|
| `SERVER_HOST` | 服务器 IP 或域名 |
| `SERVER_PORT` | SSH 端口，留空时使用 22 |
| `SERVER_USER` | 可以执行 Docker 的 SSH 用户 |
| `SERVER_SSH_KEY` | SSH 私钥完整内容 |
| `SERVER_KNOWN_HOSTS` | 服务器 SSH Host Key |
| `DEPLOY_PATH` | 部署目录，例如 `/opt/sola-bot` |

获取服务器 Host Key：

```bash
ssh-keyscan -H -p 22 your-server.example.com
```

将完整输出保存为 `SERVER_KNOWN_HOSTS`。不要关闭 SSH Host Key 校验。

配置完成后，每次代码推送或合并到 `main`，工作流会：

1. 运行 Go 测试、Go Vet 和前端构建。
2. 构建并推送五个镜像。
3. 上传生产 Compose 和部署脚本。
4. 登录 GHCR，拉取本次 SHA 镜像。
5. 运行数据库迁移并重启服务。
6. 等待 PostgreSQL、Redis、API、Bot、Worker 和 Nginx 全部健康。

如果没有配置完整的 SSH Secrets，工作流只发布镜像，自动部署步骤会安全跳过。

## 手动更新服务器

服务器上的日常更新命令：

```bash
cd /opt/sola-bot
./scripts/deploy.sh
```

脚本会执行镜像拉取、数据库迁移、容器更新和健康检查。

如果 GHCR Package 是私有的，先使用具有 `read:packages` 权限的 GitHub Token 登录：

```bash
echo 'YOUR_GITHUB_TOKEN' | docker login ghcr.io -u YOUR_GITHUB_USER --password-stdin
```

## 部署功能分支

功能分支 `feature/fangzhang-edition` 会发布同名标签，可用于上线前测试：

```bash
cd /opt/sola-bot
./scripts/deploy.sh feature-fangzhang-edition
```

## 回滚

每次构建都会保留不可变的提交标签。查看 GitHub Actions 中的标签后指定旧 SHA：

```bash
cd /opt/sola-bot
./scripts/deploy.sh sha-0ca991c
```

数据库迁移只向前执行。涉及数据库结构的版本回滚前，应确认旧程序能兼容已经执行的新迁移。

## 检查与排错

```bash
cd /opt/sola-bot
docker compose --env-file .env -f docker-compose.prod.yml ps
docker compose --env-file .env -f docker-compose.prod.yml logs --tail=100 bot
docker compose --env-file .env -f docker-compose.prod.yml logs --tail=100 api
```

数据存放在 Docker Named Volumes `sola-bot_pgdata` 和 `sola-bot_redis_data` 中。更新时不要运行 `docker compose down -v`，否则会删除数据库卷。
