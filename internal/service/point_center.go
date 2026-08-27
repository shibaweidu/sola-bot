package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dabowin/sola/internal/bot"
	"github.com/dabowin/sola/internal/model"
	"github.com/dabowin/sola/internal/store"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	referralStarted   = "started"
	referralActivated = "activated"
	referralRewarded  = "rewarded"
)

type PointCenterService struct {
	store  *store.Store
	points *PointsService
	botID  int64
}

func NewPointCenterService(st *store.Store, botToken string) *PointCenterService {
	return &PointCenterService{store: st, points: NewPointsService(st), botID: telegramBotIDFromToken(botToken)}
}

func defaultPointCenterConfig(chatID int64) bot.PointCenterConfig {
	return bot.PointCenterConfig{
		ChatID: chatID, InviteEnabled: true, InviterReward: 100, InviteeReward: 20,
		SignEnabled: true, SignReward: 1, ExchangeEnabled: true, ExchangeMinimum: 10,
		ExchangeRate:         1,
		ExchangeInstructions: "打开兑换网站，输入兑换码即可完成免费额度兑换。",
		InviteText:           "🎁 免费领取额度\n\n邀请好友加入群组，完成入群验证后，你将获得积分。\n积分可以兑换额度，无需直接购买，邀请越多，免费额度越多。",
		InvitePageTemplate:   "🎁 邀请免费领额度\n━━━━━━━━━━\n\n邀请好友加入群组，完成入群验证后，你将获得积分。\n积分可以兑换额度，无需直接购买，邀请越多，免费额度越多。\n\n邀请人获得：{inviter_reward} 积分\n被邀请人获得：{invitee_reward} 积分\n\n第一步：把下面的机器人邀请链接分享给好友\n{invite_link}\n\n第二步：好友加入「{group}」，完成频道订阅和入群验证。\n\n验证成功后，双方均可获得积分，积分可兑换免费额度。\n兑换比例：{exchange_rate} 积分 = 1 个额度\n最低兑换：{exchange_minimum} 积分",
		InviteJoinText:       "🚀 请先加入目标群组\n\n完成频道订阅和入群验证后，才能获得积分。\n\n验证成功后，您和邀请人都将获得积分，积分可兑换免费额度。",
		InviteSuccessText:    "🎉 入群成功！\n\n您和邀请人均可获得积分。\n本次获得：+{invitee_points} 积分\n当前积分：{points}\n\n积分可兑换免费额度。",
		PointsText:           "邀请好友获得积分，积分可兑换免费额度。这里查看当前积分、邀请奖励和兑换记录。",
		SignText:             "每天签到可获得积分，积分可用于兑换免费额度。",
		ExchangeText:         "🎁 使用邀请获得的积分兑换免费额度。1 个积分兑换 1 个额度，最低兑换 10 个积分。",
		PurchaseText:         "需要更多额度？可以直接购买额度。",
		ShopText:             "浏览小铺商品，购买更多额度和相关服务。",
		RankText:             "邀请人数最多的前 5 名用户，邀请越多，获得的免费额度越多。",
	}
}

func (s *PointCenterService) GetConfig(ctx context.Context, chatID int64) (bot.PointCenterConfig, error) {
	cfg := defaultPointCenterConfig(chatID)
	if s == nil || s.store == nil || s.store.DB == nil || chatID == 0 {
		return cfg, nil
	}
	var row model.PointCenterConfig
	err := s.store.DB.WithContext(ctx).First(&row, "chat_id = ?", chatID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = modelPointCenterConfig(cfg)
		if createErr := s.store.DB.WithContext(ctx).Create(&row).Error; createErr != nil && !isDuplicateKeyError(createErr) {
			return cfg, createErr
		}
		if err := s.store.DB.WithContext(ctx).First(&row, "chat_id = ?", chatID).Error; err != nil {
			return cfg, err
		}
		return botPointCenterConfig(row), nil
	}
	if err != nil {
		return cfg, err
	}
	return botPointCenterConfig(row), nil
}

func (s *PointCenterService) UpdateConfig(ctx context.Context, cfg bot.PointCenterConfig) (bot.PointCenterConfig, error) {
	if cfg.ChatID == 0 {
		return cfg, errors.New("chat_id is required")
	}
	if cfg.InviterReward < 0 || cfg.InviteeReward < 0 || cfg.SignReward < 0 || cfg.ExchangeMinimum < 1 || cfg.ExchangeRate < 1 {
		return cfg, errors.New("奖励、最低兑换积分和兑换比例必须为有效非负数")
	}
	if cfg.ExchangeURL != "" && !validHTTPSURL(cfg.ExchangeURL) {
		return cfg, errors.New("兑换地址必须使用 https")
	}
	if cfg.PurchaseURL != "" && !validHTTPSURL(cfg.PurchaseURL) {
		return cfg, errors.New("购买地址必须使用 https")
	}
	if cfg.ShopURL != "" && !validHTTPSURL(cfg.ShopURL) {
		return cfg, errors.New("小铺地址必须使用 https")
	}
	if cfg.InviteJoinURL != "" && !validHTTPSURL(cfg.InviteJoinURL) {
		return cfg, errors.New("入群链接必须使用 https")
	}
	if len([]rune(cfg.InvitePageTemplate)) > 4000 {
		return cfg, errors.New("邀请页面模板不能超过 4000 个字符")
	}
	if s == nil || s.store == nil || s.store.DB == nil {
		return cfg, nil
	}
	row := modelPointCenterConfig(cfg)
	row.UpdatedAt = time.Now()
	if err := s.store.DB.WithContext(ctx).Save(&row).Error; err != nil {
		return cfg, err
	}
	return botPointCenterConfig(row), nil
}

func (s *PointCenterService) EnsureReferralLink(ctx context.Context, botID, chatID, inviterID int64, botUsername string) (bot.ReferralLinkResult, error) {
	if botID == 0 {
		botID = s.botID
	}
	if botID == 0 || chatID == 0 || inviterID == 0 {
		return bot.ReferralLinkResult{}, errors.New("邀请链接缺少 Bot、群组或用户信息")
	}
	var code model.ReferralCode
	err := s.store.DB.WithContext(ctx).Where("bot_id = ? AND chat_id = ? AND inviter_id = ? AND is_active = ?", botID, chatID, inviterID, true).First(&code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		value, genErr := newReferralCode()
		if genErr != nil {
			return bot.ReferralLinkResult{}, genErr
		}
		code = model.ReferralCode{BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()}, Code: value, BotID: botID, ChatID: chatID, InviterID: inviterID, IsActive: true}
		if err := s.store.DB.WithContext(ctx).Create(&code).Error; err != nil {
			return bot.ReferralLinkResult{}, err
		}
	} else if err != nil {
		return bot.ReferralLinkResult{}, err
	}
	botUsername = strings.TrimPrefix(strings.TrimSpace(botUsername), "@")
	return bot.ReferralLinkResult{Link: fmt.Sprintf("https://t.me/%s?start=ref_%s", botUsername, code.Code), InviterID: inviterID, ChatID: chatID}, nil
}

func (s *PointCenterService) RegisterReferralStart(ctx context.Context, botID int64, rawCode string, inviteeID int64) (bot.ReferralStartResult, error) {
	codeValue := strings.TrimPrefix(strings.TrimSpace(rawCode), "ref_")
	if codeValue == "" || inviteeID == 0 {
		return bot.ReferralStartResult{}, nil
	}
	if botID == 0 {
		botID = s.botID
	}
	var code model.ReferralCode
	if err := s.store.DB.WithContext(ctx).Where("code = ? AND bot_id = ? AND is_active = ?", codeValue, botID, true).First(&code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bot.ReferralStartResult{}, nil
		}
		return bot.ReferralStartResult{}, err
	}
	if code.InviterID == inviteeID {
		return bot.ReferralStartResult{Self: true, ChatID: code.ChatID}, nil
	}
	var existing model.Referral
	if err := s.store.DB.WithContext(ctx).Where("bot_id = ? AND invitee_id = ?", botID, inviteeID).First(&existing).Error; err == nil {
		return bot.ReferralStartResult{Already: true, ChatID: existing.ChatID}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return bot.ReferralStartResult{}, err
	}
	now := time.Now()
	referral := model.Referral{BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now}, CodeID: code.ID, BotID: botID, ChatID: code.ChatID, InviterID: code.InviterID, InviteeID: inviteeID, Status: referralStarted, StartedAt: now}
	if err := s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&referral).Error; err != nil {
			if isDuplicateKeyError(err) {
				return nil
			}
			return err
		}
		return tx.Model(&model.ReferralCode{}).Where("id = ?", code.ID).UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
	}); err != nil {
		return bot.ReferralStartResult{}, err
	}
	return bot.ReferralStartResult{Accepted: true, ChatID: code.ChatID}, nil
}

// ResetReferralForTesting clears the selected user's test state. It is
// intentionally exposed only through the owner-protected admin API for local
// testing and should not be used as a production account-management action.
func (s *PointCenterService) ResetReferralForTesting(ctx context.Context, botID, chatID, inviteeID int64) error {
	if chatID == 0 || inviteeID == 0 {
		return errors.New("chat_id and user_id are required")
	}
	if s == nil || s.store == nil || s.store.DB == nil {
		return errors.New("database is not configured")
	}
	if botID == 0 {
		botID = s.botID
	}
	err := s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var referrals []model.Referral
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("bot_id = ? AND chat_id = ? AND (invitee_id = ? OR inviter_id = ?)", botID, chatID, inviteeID, inviteeID).Find(&referrals).Error; err != nil {
			return err
		}
		referralIDs := make([]string, 0, len(referrals))
		for _, referral := range referrals {
			referralIDs = append(referralIDs, referral.ID.String())
		}
		if len(referralIDs) > 0 {
			var rewardLogs []model.PointLog
			if err := tx.Where("chat_id = ? AND (reason IN ?)", chatID, referralReasonValues(referralIDs)).Find(&rewardLogs).Error; err != nil {
				return err
			}
			balances := make(map[int64]int)
			for _, log := range rewardLogs {
				if log.UserID != inviteeID {
					balances[log.UserID] += log.Delta
				}
			}
			for userID, delta := range balances {
				if err := tx.Model(&model.UserPoint{}).Where("chat_id = ? AND user_id = ?", chatID, userID).UpdateColumn("total_points", gorm.Expr("total_points - ?", delta)).Error; err != nil {
					return err
				}
			}
			if err := tx.Where("chat_id = ? AND reason IN ?", chatID, referralReasonValues(referralIDs)).Delete(&model.PointLog{}).Error; err != nil {
				return err
			}
			for _, referral := range referrals {
				if err := tx.Model(&referral).Updates(map[string]any{
					"status": referralStarted, "activated_at": nil, "rewarded_at": nil, "updated_at": time.Now(),
				}).Error; err != nil {
					return err
				}
			}
		}
		// This is a complete local test reset for the selected user: clear all
		// point history and daily sign/exchange state, not only referral rewards.
		if err := tx.Where("chat_id = ? AND user_id = ?", chatID, inviteeID).Delete(&model.PointLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("chat_id = ? AND user_id = ?", chatID, inviteeID).Delete(&model.DailySignRecord{}).Error; err != nil {
			return err
		}
		if err := tx.Where("chat_id = ? AND user_id = ?", chatID, inviteeID).Delete(&model.ExchangeOrder{}).Error; err != nil {
			return err
		}
		return tx.Model(&model.UserPoint{}).Where("chat_id = ? AND user_id = ?", chatID, inviteeID).Update("total_points", 0).Error
	})
	if err != nil {
		return err
	}
	if s.store.Redis != nil {
		_ = s.store.Redis.Del(ctx,
			fmt.Sprintf("referral:success_notice:%d:%d", chatID, inviteeID),
			fmt.Sprintf("join:processed:%d:%d", chatID, inviteeID),
			fmt.Sprintf("force_subscribe:allowed:%d:%d", chatID, inviteeID),
			fmt.Sprintf("force_subscribe:notice:%d:%d", chatID, inviteeID),
			fmt.Sprintf("force_subscribe:muted:%d:%d", chatID, inviteeID),
			fmt.Sprintf("force_subscribe:needs_verify:%d:%d", chatID, inviteeID),
			fmt.Sprintf("force_subscribe:needs_welcome:%d:%d", chatID, inviteeID),
			fmt.Sprintf("unverified:%d:%d", chatID, inviteeID),
		).Err()
	}
	return nil
}

func referralReasonValues(ids []string) []string {
	reasons := make([]string, 0, len(ids)*2)
	for _, id := range ids {
		reasons = append(reasons, "invite_success:"+id, "invite_welcome:"+id)
	}
	return reasons
}

func (s *PointCenterService) ActivateReferral(ctx context.Context, botID, chatID, inviteeID int64) (bot.ReferralRewardResult, error) {
	if botID == 0 {
		botID = s.botID
	}
	var result bot.ReferralRewardResult
	err := s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var referral model.Referral
		q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("bot_id = ? AND invitee_id = ?", botID, inviteeID)
		if chatID != 0 {
			q = q.Where("chat_id = ?", chatID)
		}
		if err := q.First(&referral).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		cfg, err := s.getConfigTx(ctx, tx, referral.ChatID)
		if err != nil {
			return err
		}
		if !cfg.InviteEnabled {
			return nil
		}
		if referral.Status != referralStarted {
			if referral.Status == referralRewarded {
				result = bot.ReferralRewardResult{
					AlreadyRewarded: true,
					InviterID:       referral.InviterID,
					InviteeID:       referral.InviteeID,
					ChatID:          referral.ChatID,
					InviterReward:   cfg.InviterReward,
					InviteeReward:   cfg.InviteeReward,
				}
			}
			return nil
		}
		now := time.Now()
		if cfg.InviterReward > 0 {
			if err := addPointsTx(tx, referral.ChatID, referral.InviterID, cfg.InviterReward, fmt.Sprintf("invite_success:%s", referral.ID)); err != nil {
				return err
			}
		}
		if cfg.InviteeReward > 0 {
			if err := addPointsTx(tx, referral.ChatID, referral.InviteeID, cfg.InviteeReward, fmt.Sprintf("invite_welcome:%s", referral.ID)); err != nil {
				return err
			}
		}
		if err := tx.Model(&referral).Updates(map[string]any{"status": referralRewarded, "activated_at": now, "rewarded_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		result = bot.ReferralRewardResult{Rewarded: true, InviterID: referral.InviterID, InviteeID: referral.InviteeID, ChatID: referral.ChatID, InviterReward: cfg.InviterReward, InviteeReward: cfg.InviteeReward}
		return nil
	})
	return result, err
}

func (s *PointCenterService) Sign(ctx context.Context, chatID, userID int64) (bot.SignResult, error) {
	if chatID == 0 || userID == 0 {
		return bot.SignResult{}, errors.New("签到目标无效")
	}
	cfg, err := s.GetConfig(ctx, chatID)
	if err != nil {
		return bot.SignResult{}, err
	}
	if !cfg.SignEnabled {
		return bot.SignResult{}, errors.New("本群签到功能已关闭")
	}
	day := time.Now().In(chinaLocation()).Format("2006-01-02")
	var result bot.SignResult
	err = s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record := model.DailySignRecord{ChatID: chatID, UserID: userID, SignDate: day, Reward: cfg.SignReward, CreatedAt: time.Now()}
		if err := tx.Create(&record).Error; err != nil {
			if isDuplicateKeyError(err) {
				return errors.New("今天已经签到")
			}
			return err
		}
		if cfg.SignReward > 0 {
			if err := addPointsTx(tx, chatID, userID, cfg.SignReward, "daily_sign"); err != nil {
				return err
			}
		}
		result = bot.SignResult{Signed: true, Reward: cfg.SignReward}
		return nil
	})
	return result, err
}

func (s *PointCenterService) Dashboard(ctx context.Context, chatID, userID int64) (bot.PointCenterDashboard, error) {
	var dashboard bot.PointCenterDashboard
	if err := s.store.DB.WithContext(ctx).Model(&model.UserPoint{}).Where("chat_id = ? AND user_id = ?", chatID, userID).Select("COALESCE(total_points, 0)").Scan(&dashboard.CurrentPoints).Error; err != nil {
		return dashboard, err
	}
	start := time.Now().In(chinaLocation())
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()).UTC()
	end := start.Add(24 * time.Hour)
	if err := s.store.DB.WithContext(ctx).Model(&model.PointLog{}).Where("chat_id = ? AND user_id = ? AND created_at >= ? AND created_at < ? AND reason LIKE ?", chatID, userID, start, end, "invite_success:%").Select("COALESCE(SUM(delta), 0)").Scan(&dashboard.TodayInvitePoints).Error; err != nil {
		return dashboard, err
	}
	if err := s.store.DB.WithContext(ctx).Model(&model.Referral{}).Where("chat_id = ? AND inviter_id = ? AND status = ?", chatID, userID, referralRewarded).Count(&dashboard.SuccessfulInvites).Error; err != nil {
		return dashboard, err
	}
	if err := s.store.DB.WithContext(ctx).Model(&model.ExchangeOrder{}).Where("chat_id = ? AND user_id = ? AND status NOT IN ?", chatID, userID, []string{"cancelled", "refunded"}).Select("COALESCE(SUM(amount), 0)").Scan(&dashboard.ExchangedAmount).Error; err != nil {
		return dashboard, err
	}
	day := time.Now().In(chinaLocation()).Format("2006-01-02")
	var count int64
	if err := s.store.DB.WithContext(ctx).Model(&model.DailySignRecord{}).Where("chat_id = ? AND user_id = ? AND sign_date = ?", chatID, userID, day).Count(&count).Error; err != nil {
		return dashboard, err
	}
	dashboard.SignedToday = count > 0
	return dashboard, nil
}

func (s *PointCenterService) Exchange(ctx context.Context, chatID, userID int64, points int) (bot.ExchangeResult, error) {
	if points <= 0 {
		return bot.ExchangeResult{}, errors.New("兑换积分必须大于 0")
	}
	cfg, err := s.GetConfig(ctx, chatID)
	if err != nil {
		return bot.ExchangeResult{}, err
	}
	if !cfg.ExchangeEnabled {
		return bot.ExchangeResult{}, errors.New("积分兑换功能已关闭")
	}
	if points < cfg.ExchangeMinimum {
		return bot.ExchangeResult{}, fmt.Errorf("最低兑换 %d 个积分", cfg.ExchangeMinimum)
	}
	amount := points * cfg.ExchangeRate
	var result bot.ExchangeResult
	err = s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var point model.UserPoint
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("chat_id = ? AND user_id = ?", chatID, userID).First(&point).Error; err != nil {
			return errors.New("积分不足")
		}
		if point.TotalPoints < int64(points) {
			return errors.New("积分不足")
		}
		var code model.ExchangeCode
		now := time.Now()
		q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("status = ? AND amount = ? AND (expires_at IS NULL OR expires_at > ?)", "available", amount, now).Order("created_at ASC")
		if err := q.First(&code).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("当前兑换码库存不足")
			}
			return err
		}
		if code.ExpiresAt != nil && code.ExpiresAt.Before(now) {
			return errors.New("当前兑换码已过期，请稍后重试")
		}
		if err := tx.Model(&point).Where("total_points >= ?", points).Updates(map[string]any{"total_points": gorm.Expr("total_points - ?", points), "updated_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.PointLog{UserID: userID, ChatID: chatID, Delta: -points, Reason: fmt.Sprintf("exchange_debit:%d", points), CreatedAt: now}).Error; err != nil {
			return err
		}
		order := model.ExchangeOrder{BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now}, UserID: userID, ChatID: chatID, PointsSpent: points, Amount: amount, CodeID: code.ID, Status: "assigned", RedeemURL: code.RedeemURL}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		if err := tx.Model(&code).Updates(map[string]any{"status": "assigned", "assigned_user": userID, "assigned_chat": chatID, "assigned_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		redeemURL := code.RedeemURL
		if redeemURL == "" {
			redeemURL = cfg.ExchangeURL
		}
		order.RedeemURL = redeemURL
		if err := tx.Model(&order).Update("redeem_url", redeemURL).Error; err != nil {
			return err
		}
		result = bot.ExchangeResult{OrderID: order.ID.String(), Code: code.Code, Amount: amount, Points: points, RedeemURL: redeemURL, Instructions: cfg.ExchangeInstructions}
		return nil
	})
	return result, err
}

func (s *PointCenterService) TopInviters(ctx context.Context, chatID int64, limit int) ([]bot.InviteRankEntry, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	var rows []struct {
		UserID int64
		Count  int64
	}
	err := s.store.DB.WithContext(ctx).Model(&model.Referral{}).Select("inviter_id as user_id, COUNT(*) as count").Where("chat_id = ? AND status = ?", chatID, referralRewarded).Group("inviter_id").Order("count DESC, inviter_id ASC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]bot.InviteRankEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, bot.InviteRankEntry{UserID: row.UserID, Count: row.Count})
	}
	return out, nil
}

func (s *PointCenterService) getConfigTx(ctx context.Context, tx *gorm.DB, chatID int64) (bot.PointCenterConfig, error) {
	var row model.PointCenterConfig
	if err := tx.WithContext(ctx).First(&row, "chat_id = ?", chatID).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return bot.PointCenterConfig{}, err
		}
		cfg := defaultPointCenterConfig(chatID)
		row = modelPointCenterConfig(cfg)
		if err := tx.WithContext(ctx).Create(&row).Error; err != nil && !isDuplicateKeyError(err) {
			return cfg, err
		}
		return cfg, nil
	}
	return botPointCenterConfig(row), nil
}

func addPointsTx(tx *gorm.DB, chatID, userID int64, delta int, reason string) error {
	now := time.Now()
	row := model.UserPoint{UserID: userID, ChatID: chatID, TotalPoints: int64(delta), UpdatedAt: now}
	if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "chat_id"}}, DoUpdates: clause.Assignments(map[string]any{"total_points": gorm.Expr("user_points.total_points + ?", delta), "updated_at": now})}).Create(&row).Error; err != nil {
		return err
	}
	return tx.Create(&model.PointLog{UserID: userID, ChatID: chatID, Delta: delta, Reason: reason, CreatedAt: now}).Error
}

func modelPointCenterConfig(cfg bot.PointCenterConfig) model.PointCenterConfig {
	return model.PointCenterConfig{ChatID: cfg.ChatID, InviteEnabled: cfg.InviteEnabled, InviterReward: cfg.InviterReward, InviteeReward: cfg.InviteeReward, SignEnabled: cfg.SignEnabled, SignReward: cfg.SignReward, ExchangeEnabled: cfg.ExchangeEnabled, ExchangeMinimum: cfg.ExchangeMinimum, ExchangeRate: cfg.ExchangeRate, ExchangeURL: cfg.ExchangeURL, ExchangeInstructions: cfg.ExchangeInstructions, PurchaseURL: cfg.PurchaseURL, PurchaseText: cfg.PurchaseText, ShopURL: cfg.ShopURL, ShopText: cfg.ShopText, InviteText: cfg.InviteText, InvitePageTemplate: cfg.InvitePageTemplate, InviteJoinURL: cfg.InviteJoinURL, InviteJoinText: cfg.InviteJoinText, InviteSuccessText: cfg.InviteSuccessText, PointsText: cfg.PointsText, SignText: cfg.SignText, ExchangeText: cfg.ExchangeText, RankText: cfg.RankText}
}

func botPointCenterConfig(row model.PointCenterConfig) bot.PointCenterConfig {
	return bot.PointCenterConfig{ChatID: row.ChatID, InviteEnabled: row.InviteEnabled, InviterReward: row.InviterReward, InviteeReward: row.InviteeReward, SignEnabled: row.SignEnabled, SignReward: row.SignReward, ExchangeEnabled: row.ExchangeEnabled, ExchangeMinimum: row.ExchangeMinimum, ExchangeRate: row.ExchangeRate, ExchangeURL: row.ExchangeURL, ExchangeInstructions: row.ExchangeInstructions, PurchaseURL: row.PurchaseURL, PurchaseText: row.PurchaseText, ShopURL: row.ShopURL, ShopText: row.ShopText, InviteText: row.InviteText, InvitePageTemplate: row.InvitePageTemplate, InviteJoinURL: row.InviteJoinURL, InviteJoinText: row.InviteJoinText, InviteSuccessText: row.InviteSuccessText, PointsText: row.PointsText, SignText: row.SignText, ExchangeText: row.ExchangeText, RankText: row.RankText}
}

func newReferralCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

func chinaLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*60*60)
	}
	return loc
}

func validHTTPSURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && strings.EqualFold(u.Scheme, "https") && u.Host != ""
}
