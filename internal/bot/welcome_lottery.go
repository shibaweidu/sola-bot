package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func (a *App) handleDailyLotteryDeepLink(b *gotgbot.Bot, ctx *ext.Context, payload string) error {
	if ctx == nil || ctx.Message == nil || ctx.Message.From == nil || ctx.Message.Chat.Type != "private" {
		return ext.EndGroups
	}
	rawID := strings.TrimPrefix(payload, "dl_")
	chatID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || chatID == 0 || a.services.DailyLottery == nil {
		return sendText(b, ctx, "每日抽奖入口无效，请从群组欢迎语重新打开。", nil)
	}
	scope := requestScope(ctx)
	member, err := b.GetChatMemberWithContext(scope.Context, chatID, scope.Actor.ID, nil)
	if err != nil || !chatMemberPresent(member) {
		return sendText(b, ctx, "请先加入来源群组后再参加每日额度抽奖。", nil)
	}
	if a.services.Admin != nil {
		cfg, cfgErr := a.services.Admin.GetConfig(scope.Context, chatID)
		if cfgErr == nil && forceSubscribeConfigured(cfg) {
			subscribed, missing, checkErr := a.checkForceSubscription(scope.Context, b, cfg, scope.Actor.ID, true)
			if checkErr != nil {
				return sendText(b, ctx, "订阅状态暂时无法核验，请稍后重试。", nil)
			}
			if !subscribed {
				return sendText(b, ctx, forceSubscribeMuteText(cfg, *ctx.Message.From, missing), &gotgbot.SendMessageOpts{ReplyMarkup: forceSubscribeMarkup(cfg, chatID, scope.Actor.ID)})
			}
		}
	}
	if err := a.setSelectedChatID(scope.Context, scope.Actor.ID, chatID); err != nil {
		return err
	}
	chat, ok, err := a.selectedPrivateChatIfChosen(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return sendText(b, ctx, "来源群组尚未绑定到当前账号，请先在后台绑定该群组。", nil)
	}
	if err := a.showDailyLotteryCenter(b, ctx, chat); err != nil {
		return err
	}
	return a.syncPrivateKeyboard(b, ctx)
}

func dailyLotteryDeepLinkHelp(chatID int64) string {
	return fmt.Sprintf("dl_%d", chatID)
}
