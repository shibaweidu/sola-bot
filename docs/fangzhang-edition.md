# 方丈版二开说明

本文记录 `feature/fangzhang-edition` 分支围绕 Telegram 群运营的二开范围、配置入口和上线检查项。

## 能力范围

| 能力 | 当前实现 | 配置入口 |
| --- | --- | --- |
| 新人进群验证 | 按钮、数字验证码、选择题、Telegram Poll、数学题、Cloudflare Turnstile | Web「群组配置」或 `/set_verify`、`/verify_toggle` |
| 新人欢迎词 | 可开关，支持 `{name}`，可设置 0-3600 秒自动删除 | Web「群组配置」或 `/set_welcome`、`/welcome_toggle`、`/set_welcome_delete` |
| 强制订阅 | 支持多个公开频道/数字频道 ID；未订阅可持续禁言或移出，支持按钮复核、提示文案和频道昵称 | Web「群组配置」或 `/force_subscribe` |
| 垃圾消息拦截 | 关键词动作、链接白/黑名单、转发/媒体锁、未验证用户限制、风险分和 AI 二审 | Web「群组配置」与「关键词规则」 |
| 自动回复 | 包含/精确/正则匹配，支持删除触发消息 | Web「自动回复」或 `/add_reply`、`/del_reply`、`/replies` |
| 定时发送 | 单次、循环、富媒体、按钮、置顶、自动删除 | Web「定时任务」或 `/post_create`、`/posts` |
| 群成员管理 | 封禁、解封、禁言、解禁、踢出、警告、管理员升降级、头衔 | Web「封禁管理」或群管命令 |
| 群消息管理 | 单条删除、批量清理、关键词处理、内容锁、违规记录和审计 | `/del`、`/purge` 与 Web 管理后台 |

## 本分支补齐的关键链路

1. `captcha` 不再回退为按钮验证。机器人会生成六位数字验证码，限制尝试次数，成功后恢复发言权限并发送欢迎消息。
2. 关闭新人验证时仍会发送欢迎消息；Turnstile 已在入群申请阶段完成验证，不再进群后二次弹出按钮验证。
3. 自动回复拥有独立消息 Handler，不依赖积分模块是否启用，并在垃圾消息审核通过后才执行。
4. 链接、转发和媒体锁为直接拦截规则。开启后命中即删除并写入违规与审计记录，不再只是增加风险分。
5. `/api/admin/config/:chatID` 同时读写新人配置和垃圾拦截配置，Web 后台可完整管理这些开关。
6. 新增 migration `000024_welcome_controls`，持久化欢迎消息开关和自动删除时间。
7. 新增 migration `000025_force_subscription`，持久化强制订阅频道和未订阅处理策略。
8. 新增 migration `000026_force_subscription_text`，持久化未订阅/移出提示文案和频道显示名称映射。

## 强制订阅文案与频道昵称

在 Web 后台进入「群组设置 → 强制订阅」即可编辑以下内容：

- **未订阅提示文案**：禁言模式下发送的消息。
- **移出群组提示文案**：`kick` 模式下发送的消息。
- **频道显示名称**：每行使用 `频道目标 | 显示名称`，例如 `@your_channel | 考拉 AI 官方频道`；也兼容 `=` 分隔。频道目标可填写公开频道用户名或数字 ID，映射只影响按钮和提示文案中的显示，不影响 Telegram 校验。

两种文案支持变量 `{name}`（用户名称）和 `{channels}`（当前未订阅频道显示名称，多个名称用顿号连接）。字段留空时使用默认文案。示例：

```text
@your_channel | 考拉 AI 官方频道
-1001234567890 | VIP 资源频道
```

## Telegram 前置设置

机器人必须是群管理员，并至少拥有：删除消息、限制成员、封禁成员、邀请用户权限。需要使用管理员升降级功能时，还要允许机器人添加管理员。

在 BotFather 中关闭 Privacy Mode，否则机器人无法读取普通群消息，垃圾拦截、自动回复和消息积分都只能看到命令。Turnstile 模式还要求群组开启「加入前需管理员批准」，并配置 Mini App URL 与 Cloudflare 密钥。

强制订阅模式要求机器人能调用必订阅频道的 `getChatMember`。因此机器人必须被加入每个必订阅频道并设为管理员；公开频道建议使用 `@username` 或 `https://t.me/username`，私有频道使用数字频道 ID（如 `-1001234567890`）。

## 本地启动

Windows 下可直接双击项目根目录的 `start.bat`；脚本会保留已有 `.env`，自动启动数据库、缓存、API、后台和 Worker，并在检测到有效 Bot Token 后启动 Bot。代码更新后可使用 `start.bat -Build` 强制重建镜像。

```powershell
Copy-Item .env.example .env
# 编辑 .env，至少填写 SOLA_BOT_TOKEN、SOLA_JWT_SECRET 和数据库密码
docker compose up -d --build
```

Telegram 接入顺序：从 [@BotFather](https://t.me/BotFather) 创建 Bot 并把 Token 写入 `SOLA_BOT_TOKEN`；将 Bot 加入测试群并设为管理员，授予删除消息、限制成员、封禁成员和邀请用户权限；在 BotFather 执行 `/setprivacy` 选择 `Disable`。强制订阅还需要把 Bot 加入每个必订阅频道并设为频道管理员。Polling 模式不需要公网域名，启动后可通过 `docker compose --env-file .env logs -f bot` 确认 `telegram bot connected`。

管理后台默认通过 `SOLA_HTTP_PORT` 暴露。首次上线先在测试群验证以下流程：

1. 六种验证方式各完成一次成功、失败和超时流程。
2. 分别关闭验证、关闭欢迎消息，确认两个开关互不影响。
3. 开启链接/转发/媒体锁，确认管理员豁免、白名单放行和普通成员拦截。
4. 创建自动回复，确认垃圾消息不会触发回复，正常消息只回复一次。
5. 创建一次性和循环定时任务，确认 Worker 发送、置顶和自动删除。
6. 验证封禁、禁言、踢人、单删和批量删除，并检查审计日志。

## 验证命令

```powershell
go test ./...
go vet ./...
go build ./...
Set-Location web
npm ci
npm run build
npm run build:mini
```

真实 Telegram API 链路需要有效 Bot Token 和测试群，不能由离线单元测试替代。
