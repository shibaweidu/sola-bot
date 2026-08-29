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
			_ = answerCallback(b, ctx, "无法识别当前用户")
			return sendText(b, ctx, "无法识别当前用户，请重新打开抽奖面板。", nil)
		}
		member, memberErr := b.GetChatMemberWithContext(requestScope(ctx).Context, chatID, requestScope(ctx).Actor.ID, nil)
		if memberErr != nil || !chatMemberPresent(member) {
			_ = answerCallback(b, ctx, "请先加入目标群组")
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
					_ = answerCallback(b, ctx, "订阅状态暂时无法核验")
					return sendText(b, ctx, "订阅状态暂时无法核验，请稍后重试。", nil)
				}
				if !subscribed {
					_ = answerCallback(b, ctx, "请先完成频道订阅")
					return sendText(b, ctx, forceSubscribeMuteText(cfg, *ctx.EffectiveUser, missing), &gotgbot.SendMessageOpts{ReplyMarkup: forceSubscribeMarkup(cfg, chatID, requestScope(ctx).Actor.ID)})
				}
			}
		}
		result, err := a.services.DailyLottery.Draw(requestScope(ctx).Context, chatID, requestScope(ctx).Actor.ID)
		if err != nil {
			_ = answerCallback(b, ctx, "抽奖未完成")
			return sendText(b, ctx, "抽奖失败："+err.Error(), nil)
		}
		// A draw callback creates a new result message so every attempt remains
		// visible in the conversation. Remove only the old message's buttons to
		// prevent stale panels from triggering duplicate-looking draws.
		if ctx.CallbackQuery != nil && ctx.EffectiveChat != nil && ctx.EffectiveMessage != nil {
			_, _, _ = b.EditMessageReplyMarkupWithContext(requestScope(ctx).Context, &gotgbot.EditMessageReplyMarkupOpts{
				ChatId: ctx.EffectiveChat.Id, MessageId: ctx.EffectiveMessage.MessageId,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{}},
			})
		}
		_ = answerCallback(b, ctx, "抽奖结果已生成")
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
		if err := sendText(b, ctx, formatDailyLotteryResult(result), dailyLotteryMarkup(chatID, status)); err != nil {
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
		fmt.Sprintf("今日免费次数：%d/%d", minInt(status.UsedAttempts, status.DailyAttempts), status.DailyAttempts),
		fmt.Sprintf("免费剩余次数：%d", status.Remaining),
	}
	if status.PaidEnabled {
		lines = append(lines, fmt.Sprintf("积分抽奖：已开启（免费次数用完后，每次扣除 %d 积分，已抽 %d 次）", status.PaidCostPoints, status.PaidAttempts))
	} else {
		lines = append(lines, "积分抽奖：未开启")
	}
	lines = append(lines, "", "抽奖结果为随机结果，以实际发放结果为准。")
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
				result = "已中奖"
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
	costText := "本次免费抽奖"
	if result.CostPoints > 0 {
		costText = fmt.Sprintf("本次已消耗：%d 积分", result.CostPoints)
	}
	if result.Result != "won" {
		return fmt.Sprintf("🎲 本次未中奖\n\n%s\n今日免费剩余次数：%d", costText, result.Remaining)
	}
	lines := []string{
		"🎉 恭喜你中奖！",
		"本次抽奖已完成，兑换码将单独发送。",
		costText,
	}
	if result.RedeemURL != "" {
		lines = append(lines, "兑换地址："+result.RedeemURL)
	}
	lines = append(lines, "", fmt.Sprintf("今日免费剩余次数：%d", result.Remaining))
	return strings.Join(lines, "\n")
}

func formatDailyLotteryCode(result DailyLotteryDrawResult) string {
	return fmt.Sprintf("🔑 兑换码\n━━━━━━━━━━\n\n%s\n\n请长按上方兑换码复制。", strings.TrimSpace(result.Code))
}

func dailyLotteryMarkup(chatID int64, status DailyLotteryStatus) *gotgbot.SendMessageOpts {
	resource := strconv.FormatInt(chatID, 10)
	rows := make([][]gotgbot.InlineKeyboardButton, 0, 2)
	if status.Enabled && (status.Remaining > 0 || status.PaidEnabled) && dailyLotteryHasAvailablePrize(status) {
		label := "🎲 免费抽一次"
		if status.Remaining == 0 {
			label = fmt.Sprintf("💎 %d 积分再抽一次", status.PaidCostPoints)
		}
		rows = append(rows, []gotgbot.InlineKeyboardButton{{Text: label, CallbackData: CallbackData("daily_lottery", "draw", resource)}})
	}
	rows = append(rows, []gotgbot.InlineKeyboardButton{
		{Text: "🔄 刷新次数", CallbackData: CallbackData("daily_lottery", "refresh", resource)},
		{Text: "📋 抽奖记录", CallbackData: CallbackData("daily_lottery", "history", resource)},
	})
	return &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func dailyLotteryHasAvailablePrize(status DailyLotteryStatus) bool {
	for _, prize := range status.Prizes {
		if prize.Enabled && prize.AvailableCode > 0 {
			return true
		}
	}
	return false
}
