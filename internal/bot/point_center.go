package bot

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

const defaultInvitePageTemplate = `🎁 邀请免费领额度
━━━━━━━━━━

邀请好友加入群组，完成入群验证后，你将获得积分。
积分可以兑换额度，无需直接购买，邀请越多，免费额度越多。

邀请人获得：{inviter_reward} 积分
被邀请人获得：{invitee_reward} 积分

第一步：把下面的机器人邀请链接分享给好友
{invite_link}

第二步：好友加入「{group}」，完成频道订阅和入群验证。

验证成功后，双方均可获得积分，积分可兑换免费额度。
兑换比例：{exchange_rate} 积分 = 1 个额度
最低兑换：{exchange_minimum} 积分`

type invitePageTemplateData struct {
	InviteLink      string
	Group           string
	InviterReward   int
	InviteeReward   int
	ExchangeMinimum int
	ExchangeRate    int
	InviteJoinURL   string
}

func formatInvitePageTemplate(template string, data invitePageTemplateData) string {
	if strings.TrimSpace(template) == "" {
		template = defaultInvitePageTemplate
	}
	return strings.TrimSpace(strings.NewReplacer(
		"{invite_link}", data.InviteLink,
		"{group}", data.Group,
		"{inviter_reward}", strconv.Itoa(data.InviterReward),
		"{invitee_reward}", strconv.Itoa(data.InviteeReward),
		"{exchange_minimum}", strconv.Itoa(data.ExchangeMinimum),
		"{exchange_rate}", strconv.Itoa(data.ExchangeRate),
		"{invite_join_url}", data.InviteJoinURL,
	).Replace(template))
}

func (a *App) routePointCenterCallback(b *gotgbot.Bot, ctx *ext.Context, payload CallbackPayload) error {
	if payload.Action == "check_join" {
		return a.checkReferralJoin(b, ctx, payload)
	}
	chat, ok, err := a.selectedPrivateChatIfChosen(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return a.showPrivateUserChatList(b, ctx)
	}
	if payload.Action != "exchange" {
		return answerCallback(b, ctx, "未知积分中心操作")
	}
	points, err := strconv.Atoi(payload.Resource)
	if err != nil || points <= 0 {
		return answerCallback(b, ctx, "兑换数量无效")
	}
	result, err := a.services.PointCenter.Exchange(requestScope(ctx).Context, chat.ChatID, requestScope(ctx).Actor.ID, points)
	if err != nil {
		// Keep the exchange panel intact and send the reason as a new message. Editing
		// the panel made Telegram clients appear to ignore the button when an error
		// (insufficient points or inventory) occurred.
		_ = answerCallback(b, ctx, "兑换失败："+err.Error())
		return sendText(b, ctx, "兑换失败："+err.Error(), nil)
	}
	_ = answerCallback(b, ctx, "兑换成功")
	if err := sendText(b, ctx, formatExchangeResult(result), nil); err != nil {
		return err
	}
	// Keep the code in its own short message so mobile users can long-press and
	// copy it without selecting the surrounding redemption instructions.
	return sendText(b, ctx, formatExchangeCode(result), nil)
}

func (a *App) showPointCenterInvite(b *gotgbot.Bot, ctx *ext.Context, chat ChatRef) error {
	cfg, err := a.services.PointCenter.GetConfig(requestScope(ctx).Context, chat.ID)
	if err != nil {
		return err
	}
	if !cfg.InviteEnabled {
		return sendText(b, ctx, "当前群组暂未开启邀请奖励。", nil)
	}
	link, err := a.services.PointCenter.EnsureReferralLink(requestScope(ctx).Context, b.User.Id, chat.ID, requestScope(ctx).Actor.ID, b.User.Username)
	if err != nil {
		return err
	}
	shareText := strings.TrimSpace(cfg.InviteText)
	if shareText == "" {
		shareText = "🎁 免费领取额度\n\n邀请好友加入群组，完成入群验证后，你将获得积分。\n积分可以兑换额度，无需直接购买，邀请越多，免费额度越多。"
	}
	shareURL := "https://t.me/share/url?" + url.Values{"url": {link.Link}, "text": {shareText}}.Encode()
	joinURL := strings.TrimSpace(cfg.InviteJoinURL)
	if joinURL == "" {
		joinURL = chatInviteURL(chat)
	}
	groupName := strings.TrimSpace(chat.Title)
	if groupName == "" {
		if username := strings.TrimPrefix(strings.TrimSpace(chat.Username), "@"); username != "" {
			groupName = "@" + username
		} else {
			groupName = "目标群组"
		}
	}
	text := formatInvitePageTemplate(cfg.InvitePageTemplate, invitePageTemplateData{
		InviteLink: link.Link, Group: groupName, InviterReward: cfg.InviterReward,
		InviteeReward: cfg.InviteeReward, ExchangeMinimum: cfg.ExchangeMinimum,
		ExchangeRate: cfg.ExchangeRate, InviteJoinURL: joinURL,
	})
	rows := [][]gotgbot.InlineKeyboardButton{{{Text: "📤 分享邀请链接", Url: shareURL}}}
	if joinURL != "" {
		rows = append(rows, []gotgbot.InlineKeyboardButton{{Text: "🚀 加入目标群组", Url: joinURL}})
	}
	return sendText(b, ctx, text, &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}})
}

func chatInviteURL(chat ChatRef) string {
	if raw := strings.TrimSpace(chat.InviteLink); strings.HasPrefix(raw, "https://t.me/") {
		return raw
	}
	if username := strings.TrimPrefix(strings.TrimSpace(chat.Username), "@"); username != "" {
		return "https://t.me/" + username
	}
	return ""
}

func (a *App) sendReferralJoinPrompt(b *gotgbot.Bot, ctx *ext.Context, chatID int64) error {
	return a.sendReferralJoinPromptWithStatus(b, ctx, chatID, "")
}

func (a *App) sendReferralJoinPromptWithStatus(b *gotgbot.Bot, ctx *ext.Context, chatID int64, status string) error {
	if chatID == 0 {
		return nil
	}
	cfg, err := a.services.PointCenter.GetConfig(requestScope(ctx).Context, chatID)
	if err != nil {
		return err
	}
	chats, err := a.listPrivateUserChats(ctx)
	if err != nil {
		return err
	}
	var chat ChatRef
	for _, binding := range chats {
		if binding.ChatID != chatID {
			continue
		}
		chat = ChatRef{ID: binding.ChatID, Type: binding.ChatType, Title: binding.Title, Username: binding.Username, InviteLink: binding.InviteLink}
		break
	}
	joinURL := strings.TrimSpace(cfg.InviteJoinURL)
	if joinURL == "" {
		joinURL = chatInviteURL(chat)
	}
	joinText := strings.TrimSpace(cfg.InviteJoinText)
	if joinText == "" {
		joinText = "🚀 请先加入目标群组\n\n完成频道订阅和入群验证后，才能获得积分。\n\n验证成功后，您和邀请人都将获得积分，积分可兑换免费额度。"
	}
	if title := strings.TrimSpace(chat.Title); title != "" {
		joinText = fmt.Sprintf("%s\n\n目标群组：%s", joinText, title)
	}
	if strings.TrimSpace(status) != "" {
		joinText = strings.TrimSpace(status) + "\n\n" + joinText
	}
	if joinURL == "" {
		joinText += "\n\n当前群组暂无可用的公开入群链接，请联系管理员创建或在后台配置目标群组入群链接。"
		if ctx != nil && ctx.EffectiveUser != nil {
			return sendText(b, ctx, joinText, &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{{Text: "✅ 我已加入，检查状态", CallbackData: CallbackData("pointcenter", "check_join", strconv.FormatInt(chatID, 10), strconv.FormatInt(ctx.EffectiveUser.Id, 10))}}}}})
		}
		return sendText(b, ctx, joinText, nil)
	}
	rows := [][]gotgbot.InlineKeyboardButton{{{Text: "🚀 加入目标群组", Url: joinURL}}}
	if ctx != nil && ctx.EffectiveUser != nil {
		rows = append(rows, []gotgbot.InlineKeyboardButton{{Text: "✅ 我已加入，检查状态", CallbackData: CallbackData("pointcenter", "check_join", strconv.FormatInt(chatID, 10), strconv.FormatInt(ctx.EffectiveUser.Id, 10))}})
	}
	return sendText(b, ctx, joinText, &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}})
}

func (a *App) checkReferralJoin(b *gotgbot.Bot, ctx *ext.Context, payload CallbackPayload) error {
	if ctx == nil || ctx.CallbackQuery == nil || ctx.EffectiveUser == nil || len(payload.Arguments) < 1 || a.services.Admin == nil {
		return answerCallback(b, ctx, "检查入口已失效")
	}
	if ctx.EffectiveChat == nil || ctx.EffectiveChat.Type != "private" {
		return answerCallback(b, ctx, "请在与机器人的私聊中检查")
	}
	chatID, err := strconv.ParseInt(payload.Resource, 10, 64)
	if err != nil || chatID == 0 {
		return answerCallback(b, ctx, "目标群组无效")
	}
	if expected, err := strconv.ParseInt(payload.Arguments[0], 10, 64); err == nil && expected != ctx.EffectiveUser.Id {
		return answerCallback(b, ctx, "这不是你的入群检查")
	}
	member, err := b.GetChatMemberWithContext(requestScope(ctx).Context, chatID, ctx.EffectiveUser.Id, nil)
	if err != nil || !chatMemberPresent(member) {
		return answerCallback(b, ctx, "暂未检测到你已加入目标群组，请先加入后再检查")
	}
	cfg, err := a.services.Admin.GetConfig(requestScope(ctx).Context, chatID)
	if err != nil {
		return err
	}
	if forceSubscribeConfigured(cfg) {
		subscribed, missing, checkErr := a.checkForceSubscription(requestScope(ctx).Context, b, cfg, ctx.EffectiveUser.Id, true)
		if checkErr != nil {
			return answerCallback(b, ctx, "订阅频道暂时无法核验，请稍后重试")
		}
		if !subscribed {
			text := forceSubscribeMuteText(cfg, *ctx.EffectiveUser, missing)
			_, _ = b.RestrictChatMemberWithContext(requestScope(ctx).Context, chatID, ctx.EffectiveUser.Id, mutePermissions(), &gotgbot.RestrictChatMemberOpts{UseIndependentChatPermissions: true})
			if a.services.Redis != nil {
				_ = a.services.Redis.Set(requestScope(ctx).Context, forceSubscribeMutedKey(chatID, ctx.EffectiveUser.Id), "1", forceSubscribeStateTTL).Err()
				_ = a.services.Redis.Set(requestScope(ctx).Context, forceSubscribeNeedsWelcomeKey(chatID, ctx.EffectiveUser.Id), "1", forceSubscribeStateTTL).Err()
			}
			_, _ = b.SendMessageWithContext(requestScope(ctx).Context, ctx.EffectiveChat.Id, text, &gotgbot.SendMessageOpts{ReplyMarkup: forceSubscribeMarkup(cfg, chatID, ctx.EffectiveUser.Id)})
			return answerCallback(b, ctx, "请先完成频道订阅")
		}
	}
	if cfg.VerifyEnabled && cfg.VerifyType != "turnstile" && a.services.Redis != nil {
		if value, getErr := a.services.Redis.Get(requestScope(ctx).Context, fmt.Sprintf("unverified:%d:%d", chatID, ctx.EffectiveUser.Id)).Result(); getErr == nil && value == "1" {
			return answerCallback(b, ctx, "请先完成群组入群验证")
		}
	}
	if cfg.VerifyEnabled && cfg.VerifyType != "turnstile" && member.GetStatus() == gotgbot.ChatMemberStatusRestricted && !member.MergeChatMember().CanSendMessages {
		return answerCallback(b, ctx, "请先完成群组入群验证")
	}
	if a.services.PointCenter == nil {
		return answerCallback(b, ctx, "积分服务尚未接入")
	}
	reward, err := a.services.PointCenter.ActivateReferral(requestScope(ctx).Context, b.User.Id, chatID, ctx.EffectiveUser.Id)
	if err != nil {
		return err
	}
	if reward.Rewarded {
		pointCfg, _ := a.services.PointCenter.GetConfig(requestScope(ctx).Context, chatID)
		a.sendReferralSuccessNotice(b, ctx, *ctx.EffectiveUser, chatID, pointCfg, reward)
		_ = a.postWelcomeMessage(b, ctx, chatID, cfg, *ctx.EffectiveUser)
		return answerCallback(b, ctx, fmt.Sprintf("已确认入群，获得 %d 积分", reward.InviteeReward))
	}
	if reward.AlreadyRewarded {
		pointCfg, _ := a.services.PointCenter.GetConfig(requestScope(ctx).Context, chatID)
		a.sendReferralSuccessNotice(b, ctx, *ctx.EffectiveUser, chatID, pointCfg, reward)
	}
	dashboard, _ := a.services.PointCenter.Dashboard(requestScope(ctx).Context, chatID, ctx.EffectiveUser.Id)
	return answerCallback(b, ctx, fmt.Sprintf("已确认入群，当前积分：%d", dashboard.CurrentPoints))
}

func chatMemberPresent(member gotgbot.ChatMember) bool {
	if member == nil {
		return false
	}
	switch member.GetStatus() {
	case gotgbot.ChatMemberStatusOwner, gotgbot.ChatMemberStatusAdministrator, gotgbot.ChatMemberStatusMember:
		return true
	case gotgbot.ChatMemberStatusRestricted:
		return member.MergeChatMember().IsMember
	default:
		return false
	}
}

func (a *App) showPointCenterDashboard(b *gotgbot.Bot, ctx *ext.Context, chat ChatRef) error {
	cfg, err := a.services.PointCenter.GetConfig(requestScope(ctx).Context, chat.ID)
	if err != nil {
		return err
	}
	dashboard, err := a.services.PointCenter.Dashboard(requestScope(ctx).Context, chat.ID, requestScope(ctx).Actor.ID)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("💎 我的积分\n━━━━━━━━━━\n当前积分：%d\n今日邀请积分：%d\n累计邀请成功：%d 人\n已兑换额度：%d\n\n%s", dashboard.CurrentPoints, dashboard.TodayInvitePoints, dashboard.SuccessfulInvites, dashboard.ExchangedAmount, fallbackText(cfg.PointsText, "查看当前积分、今日邀请奖励和历史兑换额度。"))
	return sendText(b, ctx, text, nil)
}

func (a *App) showPointCenterSign(b *gotgbot.Bot, ctx *ext.Context, chat ChatRef) error {
	result, err := a.services.PointCenter.Sign(requestScope(ctx).Context, chat.ID, requestScope(ctx).Actor.ID)
	if err != nil {
		return sendText(b, ctx, err.Error(), nil)
	}
	dashboard, _ := a.services.PointCenter.Dashboard(requestScope(ctx).Context, chat.ID, requestScope(ctx).Actor.ID)
	return sendText(b, ctx, fmt.Sprintf("📅 签到成功，+%d 积分。\n当前积分：%d", result.Reward, dashboard.CurrentPoints), nil)
}

func (a *App) showPointCenterExchange(b *gotgbot.Bot, ctx *ext.Context, chat ChatRef) error {
	cfg, err := a.services.PointCenter.GetConfig(requestScope(ctx).Context, chat.ID)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("🎁 积分兑换免费额度\n━━━━━━━━━━\n%s\n\n最低兑换：%d 积分\n兑换比例：%d 积分 = %d 个额度\n\n请选择兑换数量：", fallbackText(cfg.ExchangeText, "🎁 使用邀请获得的积分兑换免费额度。1 个积分兑换 1 个额度，最低兑换 10 个积分。"), cfg.ExchangeMinimum, 1, cfg.ExchangeRate)
	if strings.TrimSpace(cfg.ExchangeInstructions) != "" {
		text += "\n\n兑换方式：\n" + strings.TrimSpace(cfg.ExchangeInstructions)
	}
	text = appendOptionalLink(text, "兑换地址", cfg.ExchangeURL)
	// Do not present options that are below the configured minimum. Always keep a
	// useful fallback option so a misconfigured minimum still gives the user a
	// clear, actionable button.
	options := []int{10, 20, 50, 100}
	buttons := make([]gotgbot.InlineKeyboardButton, 0, len(options))
	for _, points := range options {
		if points < cfg.ExchangeMinimum {
			continue
		}
		buttons = append(buttons, gotgbot.InlineKeyboardButton{Text: fmt.Sprintf("%d 积分", points), CallbackData: CallbackData("pointcenter", "exchange", strconv.Itoa(points))})
	}
	if len(buttons) == 0 {
		buttons = append(buttons, gotgbot.InlineKeyboardButton{Text: fmt.Sprintf("%d 积分", cfg.ExchangeMinimum), CallbackData: CallbackData("pointcenter", "exchange", strconv.Itoa(cfg.ExchangeMinimum))})
	}
	rows := make([][]gotgbot.InlineKeyboardButton, 0, (len(buttons)+1)/2)
	for i := 0; i < len(buttons); i += 2 {
		end := i + 2
		if end > len(buttons) {
			end = len(buttons)
		}
		rows = append(rows, buttons[i:end])
	}
	return sendText(b, ctx, text, &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}})
}

func (a *App) showPointCenterPurchase(b *gotgbot.Bot, ctx *ext.Context, chat ChatRef) error {
	cfg, err := a.services.PointCenter.GetConfig(requestScope(ctx).Context, chat.ID)
	if err != nil {
		return err
	}
	return sendText(b, ctx, appendOptionalLink(fallbackText(cfg.PurchaseText, "购买额度说明。"), "购买链接", cfg.PurchaseURL), nil)
}

func (a *App) showPointCenterShop(b *gotgbot.Bot, ctx *ext.Context, chat ChatRef) error {
	cfg, err := a.services.PointCenter.GetConfig(requestScope(ctx).Context, chat.ID)
	if err != nil {
		return err
	}
	return sendText(b, ctx, appendOptionalLink(fallbackText(cfg.ShopText, "小铺商品说明。"), "小铺链接", cfg.ShopURL), nil)
}

func (a *App) showPointCenterRank(b *gotgbot.Bot, ctx *ext.Context, chat ChatRef) error {
	cfg, err := a.services.PointCenter.GetConfig(requestScope(ctx).Context, chat.ID)
	if err != nil {
		return err
	}
	entries, err := a.services.PointCenter.TopInviters(requestScope(ctx).Context, chat.ID, 5)
	if err != nil {
		return err
	}
	lines := []string{"🏆 积分榜", "━━━━━━━━━━", fallbackText(cfg.RankText, "邀请人数最多的前 5 名用户。")}
	if len(entries) == 0 {
		lines = append(lines, "", "暂无有效邀请记录。")
	}
	for i, entry := range entries {
		name := fmt.Sprintf("用户 %d", entry.UserID)
		if member, memberErr := b.GetChatMemberWithContext(requestScope(ctx).Context, chat.ID, entry.UserID, nil); memberErr == nil {
			user := member.GetUser()
			if strings.TrimSpace(user.Username) != "" {
				name = "@" + user.Username
			} else if strings.TrimSpace(user.FirstName) != "" {
				name = user.FirstName
			}
		}
		lines = append(lines, fmt.Sprintf("%d. %s - %d 人", i+1, name, entry.Count))
	}
	return sendText(b, ctx, strings.Join(lines, "\n"), nil)
}

func formatExchangeResult(result ExchangeResult) string {
	text := fmt.Sprintf("兑换成功\n━━━━━━━━━━\n兑换额度：%d\n消耗积分：%d\n\n%s", result.Amount, result.Points, fallbackText(result.Instructions, "请按照兑换网站说明使用兑换码。"))
	return appendOptionalLink(text, "兑换地址", result.RedeemURL)
}

func formatExchangeCode(result ExchangeResult) string {
	return fmt.Sprintf("🔑 兑换码\n━━━━━━━━━━\n\n%s\n\n请长按上方兑换码复制。", strings.TrimSpace(result.Code))
}

func fallbackText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
