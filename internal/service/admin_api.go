package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/dabowin/sola/internal/api"
	"github.com/dabowin/sola/internal/bot"
	"github.com/dabowin/sola/internal/model"
	"gorm.io/gorm"
)

type adminAPIService struct {
	admin      *AdminService
	moderation *ModerationService
	botToken   string
}

func (s *adminAPIService) GetConfig(ctx context.Context, chatID int64) (*api.ChatAdminConfig, error) {
	if s == nil || s.admin == nil {
		cfg := defaultAdminConfig(chatID)
		return adminConfigToAPI(cfg, defaultModerationConfig(chatID)), nil
	}
	cfg, err := s.admin.GetConfig(ctx, chatID)
	if err != nil {
		return nil, err
	}
	moderationCfg := defaultModerationConfig(chatID)
	if s.moderation != nil {
		moderationCfg, err = s.moderation.GetConfig(ctx, chatID)
		if err != nil {
			return nil, err
		}
	}
	return adminConfigToAPI(cfg, moderationCfg), nil
}

func (s *adminAPIService) UpdateConfig(ctx context.Context, chatID int64, req api.ChatAdminConfigUpdateRequest) (*api.ChatAdminConfig, error) {
	if s == nil || s.admin == nil {
		cfg := defaultAdminConfig(chatID)
		applyAPIAdminPatch(&cfg, req)
		moderationCfg := defaultModerationConfig(chatID)
		applyModerationConfigPatch(&moderationCfg, moderationPatchFromAPI(req))
		return adminConfigToAPI(cfg, moderationCfg), nil
	}
	cfg, err := s.admin.UpdateConfig(ctx, chatID, bot.ChatAdminConfigPatch{
		WelcomeText:                 req.WelcomeText,
		WelcomeEnabled:              req.WelcomeEnabled,
		WelcomeDeleteSecs:           req.WelcomeDeleteSeconds,
		VerifyEnabled:               req.VerifyEnabled,
		VerifyType:                  req.VerifyType,
		VerifyTimeout:               req.VerifyTimeout,
		VerifyQuestion:              req.VerifyQuestion,
		VerifyOptions:               req.VerifyOptions,
		VerifyCorrectIndex:          req.VerifyCorrectIndex,
		VerifyDifficulty:            req.VerifyDifficulty,
		WarnLimit:                   req.WarnLimit,
		ForceSubscribeEnabled:       req.ForceSubscribeEnabled,
		ForceSubscribeChannels:      req.ForceSubscribeChannels,
		ForceSubscribeAction:        req.ForceSubscribeAction,
		ForceSubscribeMessage:       req.ForceSubscribeMessage,
		ForceSubscribeKickMessage:   req.ForceSubscribeKickMessage,
		ForceSubscribeChannelLabels: req.ForceSubscribeChannelLabels,
	})
	if err != nil {
		return nil, err
	}
	moderationCfg := defaultModerationConfig(chatID)
	if s.moderation != nil {
		moderationCfg, err = s.moderation.UpdateConfig(ctx, chatID, moderationPatchFromAPI(req))
		if err != nil {
			return nil, err
		}
	}
	return adminConfigToAPI(cfg, moderationCfg), nil
}

func (s *adminAPIService) ListBans(ctx context.Context, chatID int64, query api.CommonListQuery) ([]api.BanLog, error) {
	if s == nil || s.admin == nil {
		return []api.BanLog{}, nil
	}
	items, err := s.admin.ListBans(ctx, chatID, query.Limit)
	if err != nil {
		return nil, err
	}
	out := make([]api.BanLog, 0, len(items))
	for _, item := range items {
		out = append(out, adminBanToAPI(item))
	}
	return out, nil
}

func (s *adminAPIService) Ban(ctx context.Context, req api.AdminBanRequest) error {
	if s == nil || s.admin == nil {
		return nil
	}
	if err := ensureOwnedTelegramChatID(ctx, s.admin.store, req.OwnerUserID, req.ChatID); err != nil {
		return err
	}
	return s.admin.RecordBan(ctx, req.ChatID, req.UserID, req.BannedBy, req.Reason)
}

func (s *adminAPIService) Mute(ctx context.Context, req api.AdminMuteRequest) error {
	if s == nil || s.admin == nil {
		return nil
	}
	if err := ensureOwnedTelegramChatID(ctx, s.admin.store, req.OwnerUserID, req.ChatID); err != nil {
		return err
	}
	return s.admin.MuteMember(ctx, s.botToken, req.ChatID, req.UserID, time.Duration(req.DurationSeconds)*time.Second, req.Reason)
}

func (s *adminAPIService) Unmute(ctx context.Context, req api.AdminUnmuteRequest) error {
	if s == nil || s.admin == nil {
		return nil
	}
	if err := ensureOwnedTelegramChatID(ctx, s.admin.store, req.OwnerUserID, req.ChatID); err != nil {
		return err
	}
	return s.admin.UnmuteMember(ctx, s.botToken, req.ChatID, req.UserID)
}

func (s *adminAPIService) Unban(ctx context.Context, chatID int64, userID int64, ownerUserID string) error {
	if s == nil || s.admin == nil {
		return nil
	}
	if err := ensureOwnedTelegramChatID(ctx, s.admin.store, ownerUserID, chatID); err != nil {
		return err
	}
	return s.admin.RecordUnban(ctx, chatID, userID, 0)
}

func (s *adminAPIService) ListWarns(ctx context.Context, chatID int64, userID int64) ([]api.WarnRecord, error) {
	if s == nil || s.admin == nil {
		return []api.WarnRecord{}, nil
	}
	items, err := s.admin.ListWarns(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	out := make([]api.WarnRecord, 0, len(items))
	for _, item := range items {
		out = append(out, adminWarnToAPI(item))
	}
	return out, nil
}

func (s *adminAPIService) ExportUserRows(ctx context.Context, query api.ExportUserQuery) ([]api.ExportUserRow, error) {
	if s == nil || s.admin == nil || s.admin.store == nil || s.admin.store.DB == nil {
		return []api.ExportUserRow{}, nil
	}
	db := s.admin.store.DB.WithContext(ctx).Model(&model.UserPoint{})
	ownedIDs, err := ownedTelegramChatIDs(ctx, s.admin.store, query.OwnerUserID)
	if err != nil {
		return nil, err
	}
	db = scopeTelegramChatID(db, "chat_id", query.ChatID, ownedIDs)
	var points []model.UserPoint
	if err := db.Order("total_points desc").Limit(10000).Find(&points).Error; err != nil {
		if isMissingTableError(err) {
			return []api.ExportUserRow{}, nil
		}
		return nil, err
	}
	users, err := usersByTelegramID(ctx, s.admin.store.DB, userPointIDs(points))
	if err != nil {
		return nil, err
	}
	warnCounts, err := activeWarnCounts(ctx, s.admin.store.DB, points)
	if err != nil {
		return nil, err
	}

	rows := make([]api.ExportUserRow, 0, len(points))
	for _, point := range points {
		row := api.ExportUserRow{
			UserID:      point.UserID,
			Username:    strconv.FormatInt(point.UserID, 10),
			DisplayName: fmt.Sprintf("User %d", point.UserID),
			ChatID:      point.ChatID,
			TotalPoints: point.TotalPoints,
			Level:       levelNameForPoints(ctx, s.admin.store.DB, point.ChatID, point.TotalPoints),
			Status:      "active",
			JoinedAt:    point.UpdatedAt,
			LastSeenAt:  point.UpdatedAt,
		}
		if user, ok := users[point.UserID]; ok {
			applyExportUserRow(&row, user)
		}
		row.WarnCount = int(warnCounts[chatUserKey{chatID: point.ChatID, userID: point.UserID}])
		if query.Status != "" && !strings.EqualFold(row.Status, strings.TrimSpace(query.Status)) {
			continue
		}
		if !matchesExportUserKeyword(row, query.Keyword) {
			continue
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (s *adminAPIService) BatchUserAction(ctx context.Context, req api.BatchUserRequest) (*api.BatchUserResult, error) {
	result := &api.BatchUserResult{Failed: []string{}}
	if s == nil || s.admin == nil {
		return result, nil
	}
	if len(req.UserIDs) == 0 {
		return result, fmt.Errorf("no users selected")
	}
	if len(req.UserIDs) > 200 {
		return result, fmt.Errorf("single batch limit is 200 users")
	}
	if err := ensureOwnedTelegramChatID(ctx, s.admin.store, req.OwnerUserID, req.ChatID); err != nil {
		return result, err
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "batch_" + action
	}
	switch action {
	case "ban":
		for _, userID := range req.UserIDs {
			if err := s.admin.RecordBan(ctx, req.ChatID, userID, 0, reason); err != nil {
				result.Failed = append(result.Failed, fmt.Sprintf("%d: %v", userID, err))
				continue
			}
			result.SuccessCount++
		}
	case "mute":
		duration := time.Duration(req.DurationSeconds) * time.Second
		if duration <= 0 {
			duration = 30 * time.Minute
		}
		for _, userID := range req.UserIDs {
			if err := s.admin.MuteMember(ctx, s.botToken, req.ChatID, userID, duration, reason); err != nil {
				result.Failed = append(result.Failed, fmt.Sprintf("%d: %v", userID, err))
				continue
			}
			result.SuccessCount++
		}
	case "adjust_points":
		points := NewPointsService(s.admin.store)
		for _, userID := range req.UserIDs {
			if err := points.Adjust(ctx, req.ChatID, userID, req.Delta, reason); err != nil {
				result.Failed = append(result.Failed, fmt.Sprintf("%d: %v", userID, err))
				continue
			}
			result.SuccessCount++
		}
	default:
		return result, fmt.Errorf("unsupported action %s", req.Action)
	}
	return result, nil
}

type chatUserKey struct {
	chatID int64
	userID int64
}

func applyExportUserRow(row *api.ExportUserRow, user model.User) {
	if row == nil {
		return
	}
	if user.Username != nil && strings.TrimSpace(*user.Username) != "" {
		row.Username = *user.Username
	}
	if strings.TrimSpace(user.DisplayName) != "" {
		row.DisplayName = user.DisplayName
	}
	if strings.TrimSpace(user.Status) != "" {
		row.Status = normalizeUserStatus(user.Status)
	}
	row.JoinedAt = user.CreatedAt
	if user.LastLoginAt != nil {
		row.LastSeenAt = *user.LastLoginAt
	} else if !user.UpdatedAt.IsZero() {
		row.LastSeenAt = user.UpdatedAt
	}
}

func activeWarnCounts(ctx context.Context, db *gorm.DB, points []model.UserPoint) (map[chatUserKey]int64, error) {
	out := map[chatUserKey]int64{}
	if db == nil || len(points) == 0 {
		return out, nil
	}
	chatIDs := make([]int64, 0, len(points))
	userIDs := make([]int64, 0, len(points))
	seenChats := map[int64]struct{}{}
	seenUsers := map[int64]struct{}{}
	for _, point := range points {
		if _, ok := seenChats[point.ChatID]; !ok {
			seenChats[point.ChatID] = struct{}{}
			chatIDs = append(chatIDs, point.ChatID)
		}
		if _, ok := seenUsers[point.UserID]; !ok {
			seenUsers[point.UserID] = struct{}{}
			userIDs = append(userIDs, point.UserID)
		}
	}
	var counts []struct {
		ChatID int64
		UserID int64
		Count  int64
	}
	err := db.WithContext(ctx).
		Model(&model.WarnRecord{}).
		Select("chat_id, user_id, COUNT(*) AS count").
		Where("cleared = ? AND chat_id IN ? AND user_id IN ?", false, chatIDs, userIDs).
		Group("chat_id, user_id").
		Scan(&counts).Error
	if err != nil {
		if isMissingTableError(err) {
			return out, nil
		}
		return nil, err
	}
	for _, item := range counts {
		out[chatUserKey{chatID: item.ChatID, userID: item.UserID}] = item.Count
	}
	return out, nil
}

func adminConfigToAPI(cfg bot.ChatAdminConfig, moderationCfg model.ChatModerationConfig) *api.ChatAdminConfig {
	return &api.ChatAdminConfig{
		ChatID:                      cfg.ChatID,
		WelcomeText:                 cfg.WelcomeText,
		WelcomeEnabled:              cfg.WelcomeEnabled,
		WelcomeDeleteSeconds:        cfg.WelcomeDeleteSecs,
		VerifyEnabled:               cfg.VerifyEnabled,
		VerifyType:                  cfg.VerifyType,
		VerifyTimeout:               cfg.VerifyTimeout,
		VerifyQuestion:              cfg.VerifyQuestion,
		VerifyOptions:               cfg.VerifyOptions,
		VerifyCorrectIndex:          cfg.VerifyCorrectIndex,
		VerifyDifficulty:            cfg.VerifyDifficulty,
		WarnLimit:                   cfg.WarnLimit,
		ForceSubscribeEnabled:       cfg.ForceSubscribeEnabled,
		ForceSubscribeChannels:      cfg.ForceSubscribeChannels,
		ForceSubscribeAction:        cfg.ForceSubscribeAction,
		ForceSubscribeMessage:       cfg.ForceSubscribeMessage,
		ForceSubscribeKickMessage:   cfg.ForceSubscribeKickMessage,
		ForceSubscribeChannelLabels: cfg.ForceSubscribeChannelLabels,
		BlockLinks:                  moderationCfg.BlockLinks,
		LinkWhitelist:               moderationCfg.LinkWhitelist,
		LinkBlacklist:               moderationCfg.LinkBlacklist,
		BlockForwards:               moderationCfg.BlockForwards,
		BlockMedia:                  moderationCfg.BlockMedia,
		KeywordFilterEnabled:        moderationCfg.KeywordFilterEnabled,
		SpamScoreThreshold:          moderationCfg.SpamScoreThreshold,
		AiFilterEnabled:             moderationCfg.AiFilterEnabled,
		RestrictUnverified:          moderationCfg.RestrictUnverified,
	}
}

func applyAPIAdminPatch(cfg *bot.ChatAdminConfig, req api.ChatAdminConfigUpdateRequest) {
	if req.WelcomeText != nil {
		cfg.WelcomeText = *req.WelcomeText
	}
	if req.WelcomeEnabled != nil {
		cfg.WelcomeEnabled = *req.WelcomeEnabled
	}
	if req.WelcomeDeleteSeconds != nil {
		cfg.WelcomeDeleteSecs = *req.WelcomeDeleteSeconds
	}
	if req.VerifyEnabled != nil {
		cfg.VerifyEnabled = *req.VerifyEnabled
	}
	if req.VerifyType != nil {
		cfg.VerifyType = *req.VerifyType
	}
	if req.VerifyTimeout != nil {
		cfg.VerifyTimeout = *req.VerifyTimeout
	}
	if req.VerifyQuestion != nil {
		cfg.VerifyQuestion = *req.VerifyQuestion
	}
	if req.VerifyOptions != nil {
		cfg.VerifyOptions = *req.VerifyOptions
	}
	if req.VerifyCorrectIndex != nil {
		cfg.VerifyCorrectIndex = *req.VerifyCorrectIndex
	}
	if req.VerifyDifficulty != nil {
		cfg.VerifyDifficulty = *req.VerifyDifficulty
	}
	if req.WarnLimit != nil {
		cfg.WarnLimit = *req.WarnLimit
	}
	if req.ForceSubscribeEnabled != nil {
		cfg.ForceSubscribeEnabled = *req.ForceSubscribeEnabled
	}
	if req.ForceSubscribeChannels != nil {
		cfg.ForceSubscribeChannels = *req.ForceSubscribeChannels
	}
	if req.ForceSubscribeAction != nil {
		cfg.ForceSubscribeAction = *req.ForceSubscribeAction
	}
	if req.ForceSubscribeMessage != nil {
		cfg.ForceSubscribeMessage = *req.ForceSubscribeMessage
	}
	if req.ForceSubscribeKickMessage != nil {
		cfg.ForceSubscribeKickMessage = *req.ForceSubscribeKickMessage
	}
	if req.ForceSubscribeChannelLabels != nil {
		cfg.ForceSubscribeChannelLabels = *req.ForceSubscribeChannelLabels
	}
}

func moderationPatchFromAPI(req api.ChatAdminConfigUpdateRequest) ModerationConfigPatch {
	return ModerationConfigPatch{
		VerifyEnabled:        req.VerifyEnabled,
		VerifyType:           req.VerifyType,
		VerifyTimeoutSeconds: req.VerifyTimeout,
		VerifyQuestion:       req.VerifyQuestion,
		VerifyOptions:        req.VerifyOptions,
		VerifyCorrectIndex:   req.VerifyCorrectIndex,
		WarnLimit:            req.WarnLimit,
		BlockLinks:           req.BlockLinks,
		LinkWhitelist:        req.LinkWhitelist,
		LinkBlacklist:        req.LinkBlacklist,
		BlockForwards:        req.BlockForwards,
		BlockMedia:           req.BlockMedia,
		KeywordFilterEnabled: req.KeywordFilterEnabled,
		SpamScoreThreshold:   req.SpamScoreThreshold,
		AiFilterEnabled:      req.AiFilterEnabled,
		RestrictUnverified:   req.RestrictUnverified,
		WelcomeText:          req.WelcomeText,
		WelcomeDeleteSeconds: req.WelcomeDeleteSeconds,
	}
}

func adminBanToAPI(record bot.BanLog) api.BanLog {
	return api.BanLog{
		ID:         record.ID,
		UserID:     record.UserID,
		ChatID:     record.ChatID,
		Reason:     record.Reason,
		BannedBy:   record.BannedBy,
		BannedAt:   record.BannedAt,
		UnbannedAt: record.UnbannedAt,
	}
}

func adminWarnToAPI(record bot.WarnRecord) api.WarnRecord {
	return api.WarnRecord{
		ID:        record.ID,
		UserID:    record.UserID,
		ChatID:    record.ChatID,
		Reason:    record.Reason,
		WarnedBy:  record.WarnedBy,
		CreatedAt: record.CreatedAt,
		Cleared:   record.Cleared,
	}
}

func (s *AdminService) MuteMember(ctx context.Context, botToken string, chatID int64, userID int64, duration time.Duration, reason string) error {
	if s == nil || s.store == nil {
		return fmt.Errorf("admin service is not configured")
	}
	botToken = strings.TrimSpace(botToken)
	if botToken == "" {
		return fmt.Errorf("telegram bot token is not configured")
	}
	tgBot, err := gotgbot.NewBot(botToken, nil)
	if err != nil {
		return err
	}
	until := time.Now().Add(duration).Unix()
	_, err = tgBot.RestrictChatMemberWithContext(ctx, chatID, userID, mutePermissions(), &gotgbot.RestrictChatMemberOpts{UntilDate: until, UseIndependentChatPermissions: true})
	return err
}

func (s *AdminService) UnmuteMember(ctx context.Context, botToken string, chatID int64, userID int64) error {
	if s == nil || s.store == nil {
		return fmt.Errorf("admin service is not configured")
	}
	botToken = strings.TrimSpace(botToken)
	if botToken == "" {
		return fmt.Errorf("telegram bot token is not configured")
	}
	tgBot, err := gotgbot.NewBot(botToken, nil)
	if err != nil {
		return err
	}
	_, err = tgBot.RestrictChatMemberWithContext(ctx, chatID, userID, fullPermissions(), &gotgbot.RestrictChatMemberOpts{UseIndependentChatPermissions: true})
	return err
}

func mutePermissions() gotgbot.ChatPermissions {
	return gotgbot.ChatPermissions{}
}

func fullPermissions() gotgbot.ChatPermissions {
	yes := true
	return gotgbot.ChatPermissions{
		CanSendMessages:       true,
		CanSendAudios:         true,
		CanSendDocuments:      true,
		CanSendPhotos:         true,
		CanSendVideos:         true,
		CanSendVideoNotes:     true,
		CanSendVoiceNotes:     true,
		CanSendPolls:          true,
		CanSendOtherMessages:  true,
		CanAddWebPagePreviews: true,
		CanReactToMessages:    &yes,
		CanChangeInfo:         true,
		CanInviteUsers:        true,
		CanPinMessages:        true,
		CanManageTopics:       &yes,
	}
}

var _ api.ChatAdminService = (*adminAPIService)(nil)

func matchesExportUserKeyword(row api.ExportUserRow, keyword string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	return strings.Contains(strings.ToLower(row.Username), keyword) ||
		strings.Contains(strings.ToLower(row.DisplayName), keyword) ||
		strings.Contains(strconv.FormatInt(row.UserID, 10), keyword)
}

func levelNameForPoints(ctx context.Context, db *gorm.DB, chatID int64, points int64) string {
	if db == nil {
		return ""
	}
	var level model.LevelConfig
	err := db.WithContext(ctx).
		Where("chat_id = ? AND min_points <= ?", chatID, points).
		Order("min_points desc").
		First(&level).Error
	if err != nil || strings.TrimSpace(level.Label) == "" {
		return ""
	}
	return level.Label
}
