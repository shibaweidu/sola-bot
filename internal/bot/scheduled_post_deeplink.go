package bot

import (
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

var scheduledPostDeepLinkActions = map[string]string{
	"s": "sign",
	"i": "invite_rewards",
	"p": "points",
	"r": "rank",
	"e": "exchange",
}

func (a *App) handleScheduledPostActionDeepLink(b *gotgbot.Bot, ctx *ext.Context, payload string) error {
	if ctx == nil || ctx.Message == nil || ctx.Message.From == nil || ctx.Message.Chat.Type != "private" {
		return ext.EndGroups
	}
	action, chatID, ok := parseScheduledPostActionDeepLink(payload)
	if !ok || a.services.PointCenter == nil {
		return sendText(b, ctx, "功能入口无效，请从群组消息重新打开。", nil)
	}

	member, err := b.GetChatMemberWithContext(requestScope(ctx).Context, chatID, ctx.Message.From.Id, nil)
	if err != nil || !chatMemberPresent(member) {
		return sendText(b, ctx, "请先加入来源群组后再使用该功能。", nil)
	}
	if err := a.setSelectedChatID(requestScope(ctx).Context, ctx.Message.From.Id, chatID); err != nil {
		return err
	}
	chat, ok, err := a.selectedPrivateChatIfChosen(ctx)
	if err != nil {
		return err
	}
	if !ok || chat.ChatID != chatID {
		return sendText(b, ctx, "来源群组当前未绑定，暂时无法使用该功能。", nil)
	}
	ref := ChatRef{ID: chat.ChatID, Type: chat.ChatType, Title: chat.Title, Username: chat.Username, InviteLink: chat.InviteLink}
	if action == "exchange" {
		cfg, cfgErr := a.services.PointCenter.GetConfig(requestScope(ctx).Context, chatID)
		if cfgErr != nil {
			return cfgErr
		}
		if !cfg.ExchangeEnabled {
			return sendText(b, ctx, "当前群组暂未开启积分兑换。", nil)
		}
	}

	switch action {
	case "sign":
		err = a.showPointCenterSign(b, ctx, ref)
	case "invite_rewards":
		err = a.showPointCenterInvite(b, ctx, ref)
	case "points":
		err = a.showPointCenterDashboard(b, ctx, ref)
	case "rank":
		err = a.showPointCenterRank(b, ctx, ref)
	case "exchange":
		err = a.showPointCenterExchange(b, ctx, ref)
	}
	if err != nil {
		return err
	}
	return a.syncPrivateKeyboard(b, ctx)
}

func parseScheduledPostActionDeepLink(payload string) (string, int64, bool) {
	parts := strings.SplitN(strings.TrimPrefix(strings.TrimSpace(payload), "pa_"), "_", 2)
	if len(parts) != 2 {
		return "", 0, false
	}
	action, ok := scheduledPostDeepLinkActions[parts[0]]
	chatID, err := strconv.ParseInt(parts[1], 10, 64)
	return action, chatID, ok && err == nil && chatID != 0
}
