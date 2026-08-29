package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/dabowin/sola/internal/api"
)

func (a *App) registerDailyLotteryHandlers(d *ext.Dispatcher) {
	d.AddHandler(handlers.NewCommand("daily_lottery", a.wrap(a.handleDailyLotteryCommand, a.RateLimit("cmd:daily_lottery", 1))))
}

func (a *App) handleDailyLotteryCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	if requestScope(ctx).Chat.Type != "private" {
		return sendText(b, ctx, "每日额度抽奖请私聊机器人使用。", nil)
	}
	chat, ok, err := a.selectedPrivateChatIfChosen(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return a.showPrivateUserChatList(b, ctx)
	}
	return a.showDailyLotteryCenter(b, ctx, chat)
}

func (a *App) routeDailyLotteryCallback(b *gotgbot.Bot, ctx *ext.Context, payload CallbackPayload) error {
	chatID, err := strconv.ParseInt(strings.TrimSpace(payload.Resource), 10, 64)
	if err != nil || chatID == 0 {
		return sendText(b, ctx, "抽奖目标无效，请重新选择群组。", nil)
	}
	chat, ok, err := a.selectedPrivateChatIfChosen(ctx)
	if err != nil {
		return err
	}
	if !ok || chat.ChatID != chatID {
		return a.showPrivateUserChatList(b, ctx)
	}
	if a.services.DailyLottery == nil {
		return sendText(b, ctx, "每日额度抽奖服务尚未接入。", nil)
	}
	switch payload.Action {
	case "draw":
		if ctx.EffectiveUser == nil {
			return sendText(b, ctx, "无法识别当前用户，请重新打开抽奖面板。", nil)
		}
		member, memberErr := b.GetChatMemberWithContext(requestScope(ctx).Context, chatID, requestScope(ctx).Actor.ID, nil)
		if memberErr != nil || !chatMemberPresent(member) {
			return sendText(b, ctx, "请先加入当前目标群组后再参加抽奖。", nil)
		}
		if a.services.Admin != nil {
			cfg, cfgErr := a.services.Admin.GetConfig(requestScope(ctx).Context, chatID)
			if cfgErr != nil {
				return cfgErr
			}
			if forceSubscribeConfigured(cfg) {
				subscribed, missing, checkErr := a.checkForceSubscription(requestScope(ctx).Context, b, cfg, requestScope(ctx).Actor.ID, true)
				if checkErr != nil {
					return sendText(b, ctx, "订阅状态暂时无法核验，请稍后重试。", nil)
				}
				if !subscribed {
					return sendText(b, ctx, forceSubscribeMuteText(cfg, *ctx.EffectiveUser, missing), &gotgbot.SendMessageOpts{ReplyMarkup: forceSubscribeMarkup(cfg, chatID, requestScope(ctx).Actor.ID)})
				}
			}
		}
		result, err := a.services.DailyLottery.Draw(requestScope(ctx).Context, chatID, requestScope(ctx).Actor.ID)
		if err != nil {
			return sendText(b, ctx, "抽奖失败："+err.Error(), nil)
		}
		status, statusErr := a.services.DailyLottery.Status(requestScope(ctx).Context, chatID, requestScope(ctx).Actor.ID)
		if statusErr != nil {
			if err := sendText(b, ctx, formatDailyLotteryResult(result), nil); err != nil {
				return err
			}
			if result.Result == "won" {
				return sendText(b, ctx, formatDailyLotteryCode(result), nil)
			}
			return nil
		}
		if err := respondText(b, ctx, formatDailyLotteryResult(result), dailyLotteryMarkup(chatID, status)); err != nil {
			return err
		}
		if result.Result == "won" {
			return sendText(b, ctx, formatDailyLotteryCode(result), nil)
		}
		return nil
	case "history":
		return a.showDailyLotteryHistory(b, ctx, chatID)
	case "refresh":
		return a.showDailyLotteryCenter(b, ctx, chat)
	default:
		return a.showDailyLotteryCenter(b, ctx, chat)
	}
}

func (a *App) showDailyLotteryCenter(b *gotgbot.Bot, ctx *ext.Context, chat api.ChatBinding) error {
	if a.services.DailyLottery == nil {
		return sendText(b, ctx, "每日额度抽奖服务尚未接入。", nil)
	}
	status, err := a.services.DailyLottery.Status(requestScope(ctx).Context, chat.ChatID, requestScope(ctx).Actor.ID)
	if err != nil {
		return err
	}
	lines := []string{
		"🎁 每日额度抽奖",
		"━━━━━━━━━━",
		fmt.Sprintf("当前目标：%s", chatTitle(chat)),
		fmt.Sprintf("今日次数：%d/%d", status.UsedAttempts, status.DailyAttempts),
		fmt.Sprintf("剩余次数：%d", status.Remaining),
	}
	if status.CostPoints == 0 {
		lines = append(lines, "抽奖成本：免费")
	} else {
		lines = append(lines, fmt.Sprintf("抽奖成本：每次 %d 积分", status.CostPoints))
	}
	lines = append(lines, "", "当前奖池：")
	prizeCount := 0
	availableCount := 0
	for _, prize := range status.Prizes {
		if !prize.Enabled {
			continue
		}
		prizeCount++
		if prize.AvailableCode > 0 {
			availableCount += prize.AvailableCode
		}
		lines = append(lines, fmt.Sprintf("%d 额度 · 权重 %d/1000 · 可用 %d", prize.Amount, prize.Weight, prize.AvailableCode))
	}
	if !status.Enabled {
		lines = append(lines, "", "每日额度抽奖尚未开启，请联系管理员。")
	} else if prizeCount == 0 {
		lines = append(lines, "", "暂无奖池配置，请联系管理员。")
	} else if availableCount == 0 {
		lines = append(lines, "", "当前暂无可用兑换码，请稍后再试。")
	} else {
		lines = append(lines, "", "中奖兑换码将通过私聊发送。")
	}
	return respondText(b, ctx, strings.Join(lines, "\n"), dailyLotteryMarkup(chat.ChatID, status))
}

func (a *App) showDailyLotteryHistory(b *gotgbot.Bot, ctx *ext.Context, chatID int64) error {
	items, err := a.services.DailyLottery.History(requestScope(ctx).Context, chatID, requestScope(ctx).Actor.ID, 20)
	if err != nil {
		return err
	}
	lines := []string{"📋 抽奖记录", "━━━━━━━━━━"}
	if len(items) == 0 {
		lines = append(lines, "暂无抽奖记录。")
	} else {
		for _, item := range items {
			result := "未中奖"
			if item.Result == "won" {
				result = fmt.Sprintf("中奖 %d 额度", item.Amount)
			}
			lines = append(lines, fmt.Sprintf("%s 第%d次：%s", item.DrawDate, item.AttemptNo, result))
		}
	}
	status, statusErr := a.services.DailyLottery.Status(requestScope(ctx).Context, chatID, requestScope(ctx).Actor.ID)
	if statusErr != nil {
		return sendText(b, ctx, strings.Join(lines, "\n"), nil)
	}
	return respondText(b, ctx, strings.Join(lines, "\n"), dailyLotteryMarkup(chatID, status))
}

func formatDailyLotteryResult(result DailyLotteryDrawResult) string {
	if result.Result != "won" {
		return fmt.Sprintf("🎲 本次未中奖\n\n已使用：%d/3 次\n剩余次数：%d", result.AttemptNo, result.Remaining)
	}
	lines := []string{
		"🎉 恭喜你中奖！",
		fmt.Sprintf("本次获得：%d 额度", result.Amount),
	}
	if result.RedeemURL != "" {
		lines = append(lines, "兑换地址："+result.RedeemURL)
	}
	if result.Guaranteed {
		lines = append(lines, "", "本次为第 3 次保底中奖。")
	}
	lines = append(lines, "", fmt.Sprintf("已使用：%d/3 次 · 剩余：%d 次", result.AttemptNo, result.Remaining))
	return strings.Join(lines, "\n")
}

func formatDailyLotteryCode(result DailyLotteryDrawResult) string {
	return fmt.Sprintf("🔑 兑换码\n━━━━━━━━━━\n\n%s\n\n请长按上方兑换码复制。", strings.TrimSpace(result.Code))
}

func dailyLotteryMarkup(chatID int64, status DailyLotteryStatus) *gotgbot.SendMessageOpts {
	resource := strconv.FormatInt(chatID, 10)
	rows := make([][]gotgbot.InlineKeyboardButton, 0, 2)
	if status.Enabled && status.Remaining > 0 && dailyLotteryHasAvailablePrize(status) {
		rows = append(rows, []gotgbot.InlineKeyboardButton{{Text: "🎲 抽一次", CallbackData: CallbackData("daily_lottery", "draw", resource)}})
	}
	rows = append(rows, []gotgbot.InlineKeyboardButton{
		{Text: "🔄 刷新次数", CallbackData: CallbackData("daily_lottery", "refresh", resource)},
		{Text: "📋 抽奖记录", CallbackData: CallbackData("daily_lottery", "history", resource)},
	})
	return &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}}
}

func dailyLotteryHasAvailablePrize(status DailyLotteryStatus) bool {
	for _, prize := range status.Prizes {
		if prize.Enabled && prize.AvailableCode > 0 {
			return true
		}
	}
	return false
}
