package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/chatmember"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

const (
	forceSubscribeCacheTTL           = 45 * time.Second
	forceSubscribeNoticeTTL          = 30 * time.Second
	forceSubscribeStateTTL           = 7 * 24 * time.Hour
	defaultForceSubscribeMessage     = "{name}，请先订阅指定频道后才能发言。完成后点击“我已订阅，检查状态”。\n待订阅：{channels}"
	defaultForceSubscribeKickMessage = "{name} 未完成必需频道订阅，已移出群组。\n待订阅：{channels}"
)

func (a *App) registerSubscriptionHandlers(d *ext.Dispatcher) {
	d.AddHandler(handlers.NewMessage(message.NewChatMembers, a.handleForceSubscribeNewMembers))
	// Telegram may emit only a chat_member update for joins performed through
	// an invite link. Keep this in the first dispatcher group so an unverified
	// member is muted before the verification/welcome handlers run.
	d.AddHandler(handlers.NewChatMember(chatmember.All, a.handleForceSubscribeChatMember))
	d.AddHandler(handlers.NewMessage(message.All, a.handleForceSubscribeMessage))
}

func (a *App) handleForceSubscribeChatMember(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx == nil || ctx.ChatMember == nil || a.services.Admin == nil {
		return ext.ContinueGroups
	}
	update := ctx.ChatMember
	if update.Chat.Type != "group" && update.Chat.Type != "supergroup" {
		return ext.ContinueGroups
	}
	if !isNewChatMemberStatus(update.OldChatMember, update.NewChatMember) {
		return ext.ContinueGroups
	}
	member := update.NewChatMember.GetUser()
	if member.IsBot || a.forceSubscribeExempt(b, ctx, member.Id, ChatAdminConfig{ChatID: update.Chat.Id}) {
		return ext.ContinueGroups
	}
	cfg, err := a.services.Admin.GetConfig(requestScope(ctx).Context, update.Chat.Id)
	if err != nil {
		return err
	}
	if !forceSubscribeConfigured(cfg) || a.forceSubscribeExempt(b, ctx, member.Id, cfg) {
		return ext.ContinueGroups
	}
	subscribed, missing, checkErr := a.checkForceSubscription(requestScope(ctx).Context, b, cfg, member.Id, false)
	if checkErr != nil {
		// Do not silently allow a new member through when Telegram cannot check
		// the required channel. Keep the member restricted and show the normal
		// subscription prompt; the callback will report the configuration error.
		log.Printf("force subscription member check failed: chat=%d user=%d error=%v", update.Chat.Id, member.Id, checkErr)
		missing = forceSubscribeTargets(cfg.ForceSubscribeChannels)
		needsVerify := cfg.VerifyEnabled && cfg.VerifyType != "turnstile"
		if err := a.enforceForceSubscription(b, ctx, cfg, member, missing, needsVerify, !needsVerify); err != nil {
			return err
		}
		return ext.EndGroups
	}
	if subscribed {
		return ext.ContinueGroups
	}
	needsVerify := cfg.VerifyEnabled && cfg.VerifyType != "turnstile"
	if err := a.enforceForceSubscription(b, ctx, cfg, member, missing, needsVerify, !needsVerify); err != nil {
		return err
	}
	return ext.EndGroups
}

func isNewChatMemberStatus(oldMember, newMember gotgbot.ChatMember) bool {
	if newMember == nil {
		return false
	}
	switch newMember.GetStatus() {
	case gotgbot.ChatMemberStatusOwner, gotgbot.ChatMemberStatusAdministrator, gotgbot.ChatMemberStatusMember:
		// accepted member statuses
	case gotgbot.ChatMemberStatusRestricted:
		if !newMember.MergeChatMember().IsMember {
			return false
		}
	default:
		return false
	}
	if oldMember == nil {
		return true
	}
	switch oldMember.GetStatus() {
	case gotgbot.ChatMemberStatusLeft, gotgbot.ChatMemberStatusBanned:
		return true
	case gotgbot.ChatMemberStatusRestricted:
		return !oldMember.MergeChatMember().IsMember
	default:
		return false
	}
}

func (a *App) handleForceSubscribeNewMembers(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx == nil || ctx.Message == nil || a.services.Admin == nil {
		return ext.ContinueGroups
	}
	scope := requestScope(ctx)
	if scope.Chat.Type != "group" && scope.Chat.Type != "supergroup" {
		return ext.ContinueGroups
	}
	cfg, err := a.services.Admin.GetConfig(scope.Context, scope.Chat.ID)
	if err != nil {
		return err
	}
	if !forceSubscribeConfigured(cfg) {
		return ext.ContinueGroups
	}
	blocked := false
	for _, member := range ctx.Message.NewChatMembers {
		if member.IsBot || a.forceSubscribeExempt(b, ctx, member.Id, cfg) {
			continue
		}
		subscribed, missing, checkErr := a.checkForceSubscription(scope.Context, b, cfg, member.Id, false)
		if checkErr != nil {
			log.Printf("force subscription check failed: chat=%d user=%d error=%v", scope.Chat.ID, member.Id, checkErr)
			missing = forceSubscribeTargets(cfg.ForceSubscribeChannels)
			blocked = true
			needsVerify := cfg.VerifyEnabled && cfg.VerifyType != "turnstile"
			if err := a.enforceForceSubscription(b, ctx, cfg, member, missing, needsVerify, !needsVerify); err != nil {
				return err
			}
			continue
		}
		if subscribed {
			continue
		}
		blocked = true
		needsVerify := cfg.VerifyEnabled && cfg.VerifyType != "turnstile"
		if err := a.enforceForceSubscription(b, ctx, cfg, member, missing, needsVerify, !needsVerify); err != nil {
			return err
		}
	}
	if blocked {
		return ext.EndGroups
	}
	return ext.ContinueGroups
}

func (a *App) handleForceSubscribeMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx == nil || ctx.Message == nil || ctx.Message.From == nil || a.services.Admin == nil {
		return ext.ContinueGroups
	}
	scope := requestScope(ctx)
	if scope.Chat.Type != "group" && scope.Chat.Type != "supergroup" {
		return ext.ContinueGroups
	}
	msg := ctx.Message
	if msg.From.IsBot || len(msg.NewChatMembers) > 0 || a.isMessageFromTelegramAdmin(b, ctx) {
		return ext.ContinueGroups
	}
	cfg, err := a.services.Admin.GetConfig(scope.Context, scope.Chat.ID)
	if err != nil {
		return err
	}
	if !forceSubscribeConfigured(cfg) || a.forceSubscribeExempt(b, ctx, msg.From.Id, cfg) {
		return ext.ContinueGroups
	}
	subscribed, missing, err := a.checkForceSubscription(scope.Context, b, cfg, msg.From.Id, false)
	if err != nil {
		log.Printf("force subscription message check failed: chat=%d user=%d error=%v", scope.Chat.ID, msg.From.Id, err)
		missing = forceSubscribeTargets(cfg.ForceSubscribeChannels)
		_, _ = b.DeleteMessageWithContext(scope.Context, msg.Chat.Id, msg.MessageId, nil)
		if enforceErr := a.enforceForceSubscription(b, ctx, cfg, *msg.From, missing, false, false); enforceErr != nil {
			return enforceErr
		}
		return ext.EndGroups
	}
	if subscribed {
		if err := a.restoreForceSubscribeMute(scope.Context, b, scope.Chat.ID, msg.From.Id); err != nil {
			return err
		}
		return ext.ContinueGroups
	}
	_, _ = b.DeleteMessageWithContext(scope.Context, msg.Chat.Id, msg.MessageId, nil)
	if err := a.enforceForceSubscription(b, ctx, cfg, *msg.From, missing, false, false); err != nil {
		return err
	}
	return ext.EndGroups
}

func (a *App) forceSubscribeExempt(b *gotgbot.Bot, ctx *ext.Context, userID int64, cfg ChatAdminConfig) bool {
	if userID == 0 {
		return true
	}
	if cfg.VerifyWhitelist != "" && isInWhitelist(cfg.VerifyWhitelist, userID) {
		return true
	}
	scope := requestScope(ctx)
	if scope.Chat.ID == 0 || b == nil {
		return false
	}
	member, err := b.GetChatMemberWithContext(scope.Context, scope.Chat.ID, userID, nil)
	if err != nil {
		return false
	}
	switch member.GetStatus() {
	case "creator", "administrator":
		return true
	default:
		return false
	}
}

func forceSubscribeConfigured(cfg ChatAdminConfig) bool {
	return cfg.ForceSubscribeEnabled && len(forceSubscribeTargets(cfg.ForceSubscribeChannels)) > 0
}

func forceSubscribeTargets(raw string) []string {
	parts := strings.FieldsFunc(strings.ReplaceAll(raw, "\r\n", "\n"), func(r rune) bool {
		return r == '\n' || r == ',' || r == ';'
	})
	seen := make(map[string]struct{}, len(parts))
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		target := normalizeForceSubscribeTarget(part)
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		result = append(result, target)
	}
	return result
}

func normalizeForceSubscribeTarget(raw string) string {
	target := strings.TrimSpace(raw)
	if target == "" {
		return ""
	}
	if strings.HasPrefix(target, "https://t.me/") || strings.HasPrefix(target, "http://t.me/") {
		parsed, err := url.Parse(target)
		if err != nil || parsed.Host != "t.me" || parsed.Path == "" {
			return ""
		}
		target = strings.TrimPrefix(parsed.Path, "/")
		if strings.HasPrefix(target, "+") {
			return ""
		}
		if i := strings.IndexByte(target, '/'); i >= 0 {
			target = target[:i]
		}
		return "@" + strings.TrimPrefix(target, "@")
	}
	if strings.HasPrefix(target, "@") {
		return "@" + strings.TrimPrefix(strings.TrimSpace(target), "@")
	}
	if _, err := strconv.ParseInt(target, 10, 64); err == nil {
		return target
	}
	return ""
}

func forceSubscribeChannelLabels(raw string) map[string]string {
	labels := make(map[string]string)
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			parts = strings.SplitN(line, "=", 2)
		}
		if len(parts) != 2 {
			continue
		}
		target := normalizeForceSubscribeTarget(parts[0])
		label := strings.TrimSpace(parts[1])
		if target != "" && label != "" {
			labels[target] = label
		}
	}
	return labels
}

func forceSubscribeChannelLabel(cfg ChatAdminConfig, target string) string {
	if label := forceSubscribeChannelLabels(cfg.ForceSubscribeChannelLabels)[target]; label != "" {
		return label
	}
	return target
}

func forceSubscribeDisplayNames(cfg ChatAdminConfig, targets []string) []string {
	names := make([]string, 0, len(targets))
	for _, target := range targets {
		names = append(names, forceSubscribeChannelLabel(cfg, target))
	}
	return names
}

func forceSubscribeTargetURL(target string) string {
	if strings.HasPrefix(target, "@") {
		return "https://t.me/" + strings.TrimPrefix(target, "@")
	}
	return ""
}

func forceSubscribeCacheKey(chatID, userID int64) string {
	return fmt.Sprintf("force_subscribe:allowed:%d:%d", chatID, userID)
}

func forceSubscribeNoticeKey(chatID, userID int64) string {
	return fmt.Sprintf("force_subscribe:notice:%d:%d", chatID, userID)
}

func forceSubscribeMutedKey(chatID, userID int64) string {
	return fmt.Sprintf("force_subscribe:muted:%d:%d", chatID, userID)
}

func forceSubscribeNeedsVerifyKey(chatID, userID int64) string {
	return fmt.Sprintf("force_subscribe:needs_verify:%d:%d", chatID, userID)
}

func forceSubscribeNeedsWelcomeKey(chatID, userID int64) string {
	return fmt.Sprintf("force_subscribe:needs_welcome:%d:%d", chatID, userID)
}

func (a *App) checkForceSubscription(ctx context.Context, b *gotgbot.Bot, cfg ChatAdminConfig, userID int64, refresh bool) (bool, []string, error) {
	targets := forceSubscribeTargets(cfg.ForceSubscribeChannels)
	if len(targets) == 0 {
		return true, nil, nil
	}
	cacheKey := forceSubscribeCacheKey(cfg.ChatID, userID)
	if !refresh && a.services.Redis != nil {
		if value, err := a.services.Redis.Get(ctx, cacheKey).Result(); err == nil && value == "1" {
			return true, nil, nil
		}
	}
	missing := make([]string, 0)
	for _, target := range targets {
		member, err := requestSubscriptionMember(ctx, b, target, userID)
		if err != nil {
			return false, nil, fmt.Errorf("check required subscription %s: %w", target, err)
		}
		if !member.IsMember {
			missing = append(missing, target)
		}
	}
	if len(missing) == 0 && a.services.Redis != nil {
		_ = a.services.Redis.Set(ctx, cacheKey, "1", forceSubscribeCacheTTL).Err()
	}
	return len(missing) == 0, missing, nil
}

type subscriptionMemberResult struct {
	Status   string `json:"status"`
	IsMember bool   `json:"is_member"`
}

func requestSubscriptionMember(ctx context.Context, b *gotgbot.Bot, target string, userID int64) (subscriptionMemberResult, error) {
	chatID := any(target)
	if numeric, err := strconv.ParseInt(target, 10, 64); err == nil {
		chatID = numeric
	}
	raw, err := b.RequestWithContext(ctx, "getChatMember", map[string]any{
		"chat_id": chatID,
		"user_id": userID,
	}, nil)
	if err != nil {
		return subscriptionMemberResult{}, err
	}
	var member subscriptionMemberResult
	if err := json.Unmarshal(raw, &member); err != nil {
		return subscriptionMemberResult{}, err
	}
	member.IsMember = member.IsMember || member.Status == "creator" || member.Status == "administrator" || member.Status == "member"
	return member, nil
}

func (a *App) enforceForceSubscription(b *gotgbot.Bot, ctx *ext.Context, cfg ChatAdminConfig, user gotgbot.User, missing []string, needsVerify bool, needsWelcome bool) error {
	scope := requestScope(ctx)
	if cfg.ForceSubscribeAction == "kick" {
		if _, err := b.SendMessageWithContext(scope.Context, scope.Chat.ID, forceSubscribeKickText(cfg, user, missing), &gotgbot.SendMessageOpts{ReplyMarkup: forceSubscribeMarkup(cfg, scope.Chat.ID, user.Id)}); err != nil {
			log.Printf("force subscription notice failed: chat=%d user=%d error=%v", scope.Chat.ID, user.Id, err)
		}
		return a.kickUnverifiedMember(b, scope.Chat.ID, user.Id)
	}
	if _, err := b.RestrictChatMemberWithContext(scope.Context, scope.Chat.ID, user.Id, mutePermissions(), &gotgbot.RestrictChatMemberOpts{UseIndependentChatPermissions: true}); err != nil {
		log.Printf("force subscription mute failed; kicking user: chat=%d user=%d error=%v", scope.Chat.ID, user.Id, err)
		return a.kickUnverifiedMember(b, scope.Chat.ID, user.Id)
	}
	shouldNotice := true
	if a.services.Redis != nil {
		if value, getErr := a.services.Redis.Get(scope.Context, forceSubscribeNoticeKey(scope.Chat.ID, user.Id)).Result(); getErr == nil && value != "" {
			shouldNotice = false
		}
		_ = a.services.Redis.Set(scope.Context, forceSubscribeMutedKey(scope.Chat.ID, user.Id), "1", forceSubscribeStateTTL).Err()
		if needsVerify {
			_ = a.services.Redis.Set(scope.Context, forceSubscribeNeedsVerifyKey(scope.Chat.ID, user.Id), "1", forceSubscribeStateTTL).Err()
		}
		if needsWelcome {
			_ = a.services.Redis.Set(scope.Context, forceSubscribeNeedsWelcomeKey(scope.Chat.ID, user.Id), "1", forceSubscribeStateTTL).Err()
		}
		if shouldNotice {
			if _, err := a.services.Redis.Set(scope.Context, forceSubscribeNoticeKey(scope.Chat.ID, user.Id), "1", forceSubscribeNoticeTTL).Result(); err != nil {
				log.Printf("force subscription notice throttle failed: chat=%d user=%d error=%v", scope.Chat.ID, user.Id, err)
			}
		}
	}
	if shouldNotice {
		_, err := b.SendMessageWithContext(scope.Context, scope.Chat.ID, forceSubscribeMuteText(cfg, user, missing), &gotgbot.SendMessageOpts{ReplyMarkup: forceSubscribeMarkup(cfg, scope.Chat.ID, user.Id)})
		return err
	}
	return nil
}

func forceSubscribeMarkup(cfg ChatAdminConfig, chatID, userID int64) gotgbot.InlineKeyboardMarkup {
	rows := make([][]gotgbot.InlineKeyboardButton, 0)
	row := make([]gotgbot.InlineKeyboardButton, 0, 2)
	for _, target := range forceSubscribeTargets(cfg.ForceSubscribeChannels) {
		if targetURL := forceSubscribeTargetURL(target); targetURL != "" {
			row = append(row, gotgbot.InlineKeyboardButton{Text: "订阅 " + forceSubscribeChannelLabel(cfg, target), Url: targetURL})
			if len(row) == 2 {
				rows = append(rows, row)
				row = nil
			}
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	rows = append(rows, []gotgbot.InlineKeyboardButton{{Text: "我已订阅，检查状态", CallbackData: CallbackData("subscribe", "check", strconv.FormatInt(chatID, 10), strconv.FormatInt(userID, 10))}})
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func forceSubscribeMuteText(cfg ChatAdminConfig, user gotgbot.User, missing []string) string {
	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if name == "" {
		name = strconv.FormatInt(user.Id, 10)
	}
	return formatForceSubscribeMessage(cfg.ForceSubscribeMessage, cfg, name, missing, defaultForceSubscribeMessage)
}

func forceSubscribeKickText(cfg ChatAdminConfig, user gotgbot.User, missing []string) string {
	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if name == "" {
		name = strconv.FormatInt(user.Id, 10)
	}
	return formatForceSubscribeMessage(cfg.ForceSubscribeKickMessage, cfg, name, missing, defaultForceSubscribeKickMessage)
}

func formatForceSubscribeMessage(template string, cfg ChatAdminConfig, name string, missing []string, fallback string) string {
	if strings.TrimSpace(template) == "" {
		template = fallback
	}
	text := strings.ReplaceAll(template, "{name}", name)
	text = strings.ReplaceAll(text, "{channels}", strings.Join(forceSubscribeDisplayNames(cfg, missing), "、"))
	return text
}

func (a *App) handleForceSubscribeCallback(b *gotgbot.Bot, ctx *ext.Context, payload CallbackPayload) error {
	if ctx == nil || ctx.CallbackQuery == nil || a.services.Admin == nil || len(payload.Arguments) < 1 {
		return answerCallback(b, ctx, "订阅检查已失效")
	}
	userID, err := strconv.ParseInt(payload.Arguments[0], 10, 64)
	if err != nil || ctx.CallbackQuery.From.Id != userID {
		return answerCallback(b, ctx, "这不是你的订阅检查")
	}
	scope := requestScope(ctx)
	chatID := scope.Chat.ID
	if payload.Resource != "" {
		if parsed, parseErr := strconv.ParseInt(payload.Resource, 10, 64); parseErr == nil && parsed != 0 {
			chatID = parsed
		}
	}
	cfg, err := a.services.Admin.GetConfig(scope.Context, chatID)
	if err != nil {
		return err
	}
	if !forceSubscribeConfigured(cfg) {
		return answerCallback(b, ctx, "当前群组未启用强制订阅")
	}
	if a.services.Redis != nil {
		_ = a.services.Redis.Del(scope.Context, forceSubscribeCacheKey(chatID, userID))
	}
	subscribed, missing, err := a.checkForceSubscription(scope.Context, b, cfg, userID, true)
	if err != nil {
		return answerCallback(b, ctx, "订阅频道暂时无法核验，请联系群管理员")
	}
	if !subscribed {
		return answerCallback(b, ctx, "仍有频道未订阅："+strings.Join(missing, "、"))
	}
	if err := a.restoreForceSubscribeMute(scope.Context, b, chatID, userID); err != nil {
		return err
	}
	needsVerify := false
	needsWelcome := false
	if a.services.Redis != nil {
		if value, getErr := a.services.Redis.Get(scope.Context, forceSubscribeNeedsVerifyKey(chatID, userID)).Result(); getErr == nil && value == "1" {
			needsVerify = true
			_ = a.services.Redis.Del(scope.Context, forceSubscribeNeedsVerifyKey(chatID, userID))
		}
		if value, getErr := a.services.Redis.Get(scope.Context, forceSubscribeNeedsWelcomeKey(chatID, userID)).Result(); getErr == nil && value == "1" {
			needsWelcome = true
			_ = a.services.Redis.Del(scope.Context, forceSubscribeNeedsWelcomeKey(chatID, userID))
		}
	}
	if needsVerify && cfg.VerifyEnabled && ctx.EffectiveUser != nil {
		if err := a.startVerificationChallenge(b, ctx, cfg, *ctx.EffectiveUser); err != nil {
			return err
		}
	}
	if ctx.CallbackQuery.Message != nil {
		_, _ = ctx.CallbackQuery.Message.Delete(b, nil)
	}
	if needsWelcome && ctx.EffectiveUser != nil {
		_ = a.postWelcomeMessage(b, ctx, chatID, cfg, *ctx.EffectiveUser)
	}
	return answerCallback(b, ctx, "订阅已确认，可以继续使用")
}

func (a *App) restoreForceSubscribeMute(ctx context.Context, b *gotgbot.Bot, chatID, userID int64) error {
	if a.services.Redis == nil {
		return nil
	}
	value, err := a.services.Redis.Get(ctx, forceSubscribeMutedKey(chatID, userID)).Result()
	if err != nil || value != "1" {
		return nil
	}
	if _, err := b.RestrictChatMemberWithContext(ctx, chatID, userID, fullPermissions(), &gotgbot.RestrictChatMemberOpts{UseIndependentChatPermissions: true}); err != nil {
		return err
	}
	return a.services.Redis.Del(ctx, forceSubscribeMutedKey(chatID, userID)).Err()
}
