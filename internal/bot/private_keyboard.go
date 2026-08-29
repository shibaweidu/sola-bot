package bot

import (
	"sort"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/dabowin/sola/internal/api"
	"github.com/dabowin/sola/internal/model"
)

func (a *App) registerPrivateKeyboardHandlers(d *ext.Dispatcher) {
	d.AddHandler(handlers.NewMessage(message.All, a.handlePrivateKeyboard))
}

func (a *App) handlePrivateKeyboard(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx == nil || ctx.Message == nil || ctx.Message.From == nil || ctx.Message.From.IsBot || ctx.Message.Chat.Type != "private" {
		return ext.ContinueGroups
	}
	role, err := a.privateMenuRole(ctx)
	if err != nil {
		return err
	}
	items, err := a.privateMenuItems(ctx, role)
	if err != nil {
		return err
	}
	text := strings.TrimSpace(ctx.Message.Text)
	for _, item := range items {
		if !item.Enabled || text != item.RenderLabel() {
			continue
		}
		if item.ActionType == model.BotMenuActionLink {
			return sendText(b, ctx, "打开："+item.RenderLabel(), &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{{Text: item.RenderLabel(), Url: item.ActionValue}}}}})
		}
		if item.ActionType == model.BotMenuActionCustom {
			return a.handleCustomPrivateMenuItem(b, ctx, item)
		}
		if item.ActionKey == "hide_keyboard" {
			return sendText(b, ctx, "已收起私聊菜单。发送 /start 可重新显示。", &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.ReplyKeyboardRemove{RemoveKeyboard: true}})
		}
		return a.handlePrivateMenuAction(b, ctx, role, item.ActionKey)
	}
	return ext.ContinueGroups
}

func (a *App) handleCustomPrivateMenuItem(b *gotgbot.Bot, ctx *ext.Context, item model.BotMenuItem) error {
	chat, selected, err := a.selectedPrivateChatIfChosen(ctx)
	if err != nil {
		return err
	}
	group := ""
	var dashboard PointCenterDashboard
	if selected {
		group = chat.Title
		if a.services.PointCenter != nil && requestScope(ctx).Actor.ID != 0 {
			dashboard, _ = a.services.PointCenter.Dashboard(requestScope(ctx).Context, chat.ChatID, requestScope(ctx).Actor.ID)
		}
	}
	name := strings.TrimSpace(requestScope(ctx).Actor.FirstName)
	if name == "" && ctx.EffectiveUser != nil {
		name = strings.TrimSpace(ctx.EffectiveUser.FirstName + " " + ctx.EffectiveUser.LastName)
	}
	text := renderPrivateMenuText(item.MessageText, name, group, dashboard)
	if model.NormalizeBotMenuLinkMode(item.LinkMode) != model.BotMenuLinkModeButtons {
		return sendText(b, ctx, appendOptionalLink(text, "链接", item.ActionValue), nil)
	}
	markup, err := customMenuInlineMarkup(item.LinkButtonsJSON)
	if err != nil {
		return err
	}
	return sendText(b, ctx, text, &gotgbot.SendMessageOpts{ReplyMarkup: markup})
}

func customMenuInlineMarkup(raw string) (gotgbot.InlineKeyboardMarkup, error) {
	links, err := model.ParseBotMenuInlineLinks(raw)
	if err != nil {
		return gotgbot.InlineKeyboardMarkup{}, err
	}
	sort.SliceStable(links, func(i, j int) bool {
		if links[i].RowIndex != links[j].RowIndex {
			return links[i].RowIndex < links[j].RowIndex
		}
		return links[i].ColumnIndex < links[j].ColumnIndex
	})
	rows := make([][]gotgbot.InlineKeyboardButton, 0)
	currentRow := -1
	for _, link := range links {
		if link.RowIndex != currentRow {
			rows = append(rows, []gotgbot.InlineKeyboardButton{})
			currentRow = link.RowIndex
		}
		rows[len(rows)-1] = append(rows[len(rows)-1], gotgbot.InlineKeyboardButton{Text: link.Label, Url: link.URL})
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}, nil
}

func renderPrivateMenuText(template, name, group string, dashboard PointCenterDashboard) string {
	return strings.NewReplacer(
		"{name}", name,
		"{group}", group,
		"{points}", strconv.FormatInt(dashboard.CurrentPoints, 10),
		"{today_invite}", strconv.FormatInt(dashboard.TodayInvitePoints, 10),
		"{successful_invites}", strconv.FormatInt(dashboard.SuccessfulInvites, 10),
		"{exchanged_amount}", strconv.FormatInt(dashboard.ExchangedAmount, 10),
	).Replace(strings.TrimSpace(template))
}

func appendOptionalLink(text, label, rawURL string) string {
	text = strings.TrimSpace(text)
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return text
	}
	if text == "" {
		return label + "：" + rawURL
	}
	return text + "\n\n" + label + "：" + rawURL
}

func (a *App) privateMenuItems(ctx *ext.Context, role string) ([]model.BotMenuItem, error) {
	if a.services.Menu == nil {
		return nil, nil
	}
	return a.services.Menu.List(requestScope(ctx).Context, role)
}

func (a *App) privateMenuRole(ctx *ext.Context) (string, error) {
	chat, ok, err := a.selectedPrivateChatIfChosen(ctx)
	if err != nil {
		return model.BotMenuRoleMember, err
	}
	if !ok || requestScope(ctx).Actor.ID == 0 {
		return model.BotMenuRoleMember, nil
	}
	if a.services.Access == nil {
		return model.BotMenuRoleAdmin, nil
	}
	isAdmin, err := a.services.Access.IsAdmin(requestScope(ctx).Context, chat.ChatID, requestScope(ctx).Actor.ID)
	if err != nil {
		return model.BotMenuRoleMember, err
	}
	if isAdmin {
		return model.BotMenuRoleAdmin, nil
	}
	return model.BotMenuRoleMember, nil
}

func (a *App) selectedPrivateChatIfChosen(ctx *ext.Context) (api.ChatBinding, bool, error) {
	scope := requestScope(ctx)
	selectedID, ok := a.getSelectedChatID(scope.Context, scope.Actor.ID)
	if !ok {
		selectedID = 0
	}
	chats, err := a.listPrivateUserChats(ctx)
	if err != nil {
		return api.ChatBinding{}, false, err
	}
	for _, chat := range chats {
		if chat.ChatID == selectedID {
			return chat, true, nil
		}
	}
	if len(chats) == 1 {
		if err := a.setSelectedChatID(scope.Context, scope.Actor.ID, chats[0].ChatID); err != nil {
			return api.ChatBinding{}, false, err
		}
		return chats[0], true, nil
	}
	return api.ChatBinding{}, false, nil
}

func (a *App) handlePrivateMenuAction(b *gotgbot.Bot, ctx *ext.Context, role, action string) error {
	chat, ok, err := a.selectedPrivateChatIfChosen(ctx)
	if err != nil {
		return err
	}
	needsTarget := map[string]bool{
		"points": true, "sign": true, "rank": true, "lottery": true, "daily_lottery": true,
		"private_console": true, "admin_center": true, "admin_config": true, "scheduled_posts": true,
		"invite_rewards": true, "exchange": true, "purchase": true, "shop": true,
	}
	if needsTarget[action] && !ok {
		publicAction := map[string]bool{
			"points": true, "sign": true, "rank": true, "lottery": true, "daily_lottery": true,
			"invite_rewards": true, "exchange": true, "purchase": true, "shop": true,
		}
		if publicAction[action] {
			return a.showPrivateUserChatList(b, ctx)
		}
		return a.showPrivateChatList(b, ctx)
	}
	if strings.HasPrefix(action, "admin_") || action == "private_console" || action == "scheduled_posts" {
		if role != model.BotMenuRoleAdmin || !a.privateTargetIsAdmin(ctx, chat.ChatID) {
			return sendText(b, ctx, "需要当前目标的管理员权限。", nil)
		}
	}
	switch action {
	case "points":
		return a.showPointCenterDashboard(b, ctx, ChatRef{ID: chat.ChatID, Type: chat.ChatType, Title: chat.Title, Username: chat.Username})
	case "sign":
		return a.showPointCenterSign(b, ctx, ChatRef{ID: chat.ChatID, Type: chat.ChatType, Title: chat.Title, Username: chat.Username})
	case "rank":
		return a.showPointCenterRank(b, ctx, ChatRef{ID: chat.ChatID, Type: chat.ChatType, Title: chat.Title, Username: chat.Username})
	case "invite_rewards":
		return a.showPointCenterInvite(b, ctx, ChatRef{ID: chat.ChatID, Type: chat.ChatType, Title: chat.Title, Username: chat.Username, InviteLink: chat.InviteLink})
	case "exchange":
		return a.showPointCenterExchange(b, ctx, ChatRef{ID: chat.ChatID, Type: chat.ChatType, Title: chat.Title, Username: chat.Username})
	case "purchase":
		return a.showPointCenterPurchase(b, ctx, ChatRef{ID: chat.ChatID, Type: chat.ChatType, Title: chat.Title, Username: chat.Username})
	case "shop":
		return a.showPointCenterShop(b, ctx, ChatRef{ID: chat.ChatID, Type: chat.ChatType, Title: chat.Title, Username: chat.Username})
	case "lottery":
		return a.showPrivateLotteryCenter(b, ctx, chat)
	case "daily_lottery":
		return a.showDailyLotteryCenter(b, ctx, chat)
	case "help":
		return a.handleHelp(b, ctx)
	case "info":
		return a.handleInfo(b, ctx)
	case "private_console":
		return a.showPrivateConsole(b, ctx)
	case "admin_center":
		return a.showPrivateAdminCenter(b, ctx, chat)
	case "admin_config":
		return a.showPrivateAdminConfig(b, ctx, chat)
	case "scheduled_posts":
		return a.showScheduledPostsForChat(b, ctx, chat.ChatID, privateConsoleMarkup(chat))
	default:
		return sendText(b, ctx, "该菜单动作暂不可用。", nil)
	}
}

func (a *App) privateTargetIsAdmin(ctx *ext.Context, chatID int64) bool {
	if a.services.Access == nil {
		return true
	}
	scope := requestScope(ctx)
	ok, err := a.services.Access.IsAdmin(scope.Context, chatID, scope.Actor.ID)
	return err == nil && ok
}

func (a *App) privateKeyboardOpts(ctx *ext.Context, role string) (*gotgbot.SendMessageOpts, error) {
	items, err := a.privateMenuItems(ctx, role)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return &gotgbot.SendMessageOpts{}, nil
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].RowIndex != items[j].RowIndex {
			return items[i].RowIndex < items[j].RowIndex
		}
		return items[i].ColumnIndex < items[j].ColumnIndex
	})
	rows := make([][]gotgbot.KeyboardButton, 0)
	currentRow := -1
	for _, item := range items {
		if !item.Enabled {
			continue
		}
		if item.RowIndex != currentRow {
			rows = append(rows, []gotgbot.KeyboardButton{})
			currentRow = item.RowIndex
		}
		rows[len(rows)-1] = append(rows[len(rows)-1], gotgbot.KeyboardButton{Text: item.RenderLabel()})
	}
	return &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
		Keyboard:              rows,
		ResizeKeyboard:        true,
		IsPersistent:          true,
		InputFieldPlaceholder: "请选择功能",
	}}, nil
}

func (a *App) syncPrivateKeyboard(b *gotgbot.Bot, ctx *ext.Context) error {
	role, err := a.privateMenuRole(ctx)
	if err != nil {
		return err
	}
	opts, err := a.privateKeyboardOpts(ctx, role)
	if err != nil {
		return err
	}
	if opts.ReplyMarkup == nil {
		return nil
	}
	return sendText(b, ctx, "\u2063", opts)
}
