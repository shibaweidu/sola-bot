package service

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/dabowin/sola/internal/bot"
	"github.com/dabowin/sola/internal/model"
	"github.com/dabowin/sola/internal/store"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const dailyLotteryAttempts = 3

type DailyLotteryService struct{ store *store.Store }

func NewDailyLotteryService(st *store.Store) *DailyLotteryService {
	return &DailyLotteryService{store: st}
}

func (s *DailyLotteryService) GetConfig(ctx context.Context, chatID int64) (bot.DailyLotteryConfig, error) {
	cfg := bot.DailyLotteryConfig{ChatID: chatID, DailyAttempts: dailyLotteryAttempts, GuaranteeOnLast: true}
	if s == nil || s.store == nil || s.store.DB == nil {
		return cfg, nil
	}
	var row model.DailyLotteryConfig
	err := s.store.DB.WithContext(ctx).Where("chat_id = ?", chatID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cfg, nil
	}
	if err != nil {
		return bot.DailyLotteryConfig{}, err
	}
	return dailyLotteryConfigToBot(row), nil
}

func (s *DailyLotteryService) UpdateConfig(ctx context.Context, cfg bot.DailyLotteryConfig) (bot.DailyLotteryConfig, error) {
	if cfg.ChatID == 0 {
		return bot.DailyLotteryConfig{}, errors.New("chat_id is required")
	}
	if cfg.CostPoints < 0 {
		return bot.DailyLotteryConfig{}, errors.New("抽奖积分不能为负数")
	}
	if cfg.PaidEnabled && cfg.CostPoints <= 0 {
		return bot.DailyLotteryConfig{}, errors.New("开启积分抽奖后，每次消耗积分必须大于 0")
	}
	if s == nil || s.store == nil || s.store.DB == nil {
		cfg.DailyAttempts = dailyLotteryAttempts
		return cfg, nil
	}
	now := time.Now()
	row := model.DailyLotteryConfig{ChatID: cfg.ChatID, Enabled: cfg.Enabled, DailyAttempts: dailyLotteryAttempts, CostPoints: cfg.CostPoints, PaidEnabled: cfg.PaidEnabled, GuaranteeOnLast: cfg.GuaranteeOnLast, CreatedAt: now, UpdatedAt: now}
	if err := s.store.DB.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "chat_id"}}, DoUpdates: clause.AssignmentColumns([]string{"enabled", "daily_attempts", "cost_points", "paid_enabled", "guarantee_on_last", "updated_at"})}).Create(&row).Error; err != nil {
		return bot.DailyLotteryConfig{}, err
	}
	return dailyLotteryConfigToBot(row), nil
}

func (s *DailyLotteryService) ListPrizes(ctx context.Context, chatID int64) ([]bot.DailyLotteryPrize, error) {
	if s == nil || s.store == nil || s.store.DB == nil {
		return []bot.DailyLotteryPrize{}, nil
	}
	var rows []model.DailyLotteryPrize
	if err := s.store.DB.WithContext(ctx).Where("chat_id = ?", chatID).Order("amount asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]bot.DailyLotteryPrize, 0, len(rows))
	for _, row := range rows {
		item := bot.DailyLotteryPrize{ID: row.ID, ChatID: row.ChatID, Amount: row.Amount, Weight: row.Weight, Enabled: row.Enabled}
		available, err := s.availableCount(ctx, chatID, row.Amount)
		if err != nil {
			return nil, err
		}
		item.AvailableCode = available
		items = append(items, item)
	}
	return items, nil
}

func (s *DailyLotteryService) ReplacePrizes(ctx context.Context, chatID int64, prizes []bot.DailyLotteryPrize) ([]bot.DailyLotteryPrize, error) {
	if chatID == 0 {
		return nil, errors.New("chat_id is required")
	}
	seen := make(map[int]bool, len(prizes))
	totalWeight := 0
	for _, prize := range prizes {
		if prize.Amount <= 0 {
			return nil, errors.New("奖品额度必须大于 0")
		}
		if seen[prize.Amount] {
			return nil, errors.New("奖池额度不能重复")
		}
		seen[prize.Amount] = true
		if prize.Weight <= 0 || prize.Weight > 1000 {
			return nil, errors.New("中奖权重必须在 1 到 1000 之间")
		}
		if prize.Enabled {
			totalWeight += prize.Weight
		}
	}
	if len(prizes) == 0 {
		return nil, errors.New("至少配置一个奖池")
	}
	if totalWeight <= 0 {
		return nil, errors.New("至少启用一个有效奖池")
	}
	if totalWeight > 1000 {
		return nil, errors.New("启用奖池权重总和不能超过 1000")
	}
	if s == nil || s.store == nil || s.store.DB == nil {
		return prizes, nil
	}
	for _, prize := range prizes {
		available, err := s.availableCount(ctx, chatID, prize.Amount)
		if err != nil {
			return nil, err
		}
		if available == 0 {
			return nil, fmt.Errorf("额度 %d 暂无可用兑换码", prize.Amount)
		}
	}
	err := s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("chat_id = ?", chatID).Delete(&model.DailyLotteryPrize{}).Error; err != nil {
			return err
		}
		for _, prize := range prizes {
			row := model.DailyLotteryPrize{ChatID: chatID, Amount: prize.Amount, Weight: prize.Weight, Enabled: prize.Enabled, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.ListPrizes(ctx, chatID)
}

func (s *DailyLotteryService) Status(ctx context.Context, chatID, userID int64) (bot.DailyLotteryStatus, error) {
	cfg, err := s.GetConfig(ctx, chatID)
	if err != nil {
		return bot.DailyLotteryStatus{}, err
	}
	date := time.Now().In(chinaLocation()).Format("2006-01-02")
	used := 0
	if s != nil && s.store != nil && s.store.DB != nil {
		var count int64
		if err := s.store.DB.WithContext(ctx).Model(&model.DailyLotteryAttempt{}).Where("chat_id = ? AND user_id = ? AND draw_date = ?", chatID, userID, date).Count(&count).Error; err != nil {
			return bot.DailyLotteryStatus{}, err
		}
		used = int(count)
	}
	prizes, err := s.ListPrizes(ctx, chatID)
	if err != nil {
		return bot.DailyLotteryStatus{}, err
	}
	paidAttempts := maxInt(used-dailyLotteryAttempts, 0)
	return bot.DailyLotteryStatus{ChatID: chatID, UserID: userID, DrawDate: date, DailyAttempts: dailyLotteryAttempts, UsedAttempts: used, Remaining: maxInt(dailyLotteryAttempts-used, 0), CostPoints: cfg.CostPoints, PaidEnabled: cfg.PaidEnabled, PaidCostPoints: cfg.CostPoints, PaidAttempts: paidAttempts, Enabled: cfg.Enabled, Prizes: prizes}, nil
}

func (s *DailyLotteryService) Draw(ctx context.Context, chatID, userID int64) (bot.DailyLotteryDrawResult, error) {
	if userID == 0 || chatID == 0 {
		return bot.DailyLotteryDrawResult{}, errors.New("无法识别抽奖用户或群组")
	}
	if s == nil || s.store == nil || s.store.DB == nil {
		return bot.DailyLotteryDrawResult{}, errors.New("抽奖服务尚未接入数据库")
	}
	date := time.Now().In(chinaLocation()).Format("2006-01-02")
	var result bot.DailyLotteryDrawResult
	err := s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cfg model.DailyLotteryConfig
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("chat_id = ?", chatID).First(&cfg).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("每日额度抽奖尚未配置")
			}
			return err
		}
		if !cfg.Enabled {
			return errors.New("每日额度抽奖未开启")
		}
		var attempts []model.DailyLotteryAttempt
		if err := tx.Where("chat_id = ? AND user_id = ? AND draw_date = ?", chatID, userID, date).Order("attempt_no asc").Find(&attempts).Error; err != nil {
			return err
		}
		attemptNo := len(attempts) + 1
		paidDraw := attemptNo > dailyLotteryAttempts
		if paidDraw && !cfg.PaidEnabled {
			return errors.New("今日免费次数已用完，积分抽奖未开启")
		}
		available, err := dailyLotteryAvailablePrizes(tx, chatID)
		if err != nil {
			return err
		}
		if len(available) == 0 {
			return errors.New("当前奖池暂无可用兑换码，请稍后再试")
		}
		guaranteed := cfg.GuaranteeOnLast && attemptNo == dailyLotteryAttempts && previousAttemptsLost(attempts)
		selected, won, err := chooseDailyPrize(available, guaranteed)
		if err != nil {
			return err
		}
		costPoints := 0
		if paidDraw {
			costPoints = cfg.CostPoints
		}
		if costPoints > 0 {
			if err := adjustLotteryPointsTx(tx, chatID, userID, -costPoints, fmt.Sprintf("daily_lottery:%s:%d", date, attemptNo)); err != nil {
				return err
			}
		}
		result = bot.DailyLotteryDrawResult{ChatID: chatID, UserID: userID, DrawDate: date, AttemptNo: attemptNo, CostPoints: costPoints, Guaranteed: guaranteed}
		if !won {
			if err := tx.Create(&model.DailyLotteryAttempt{ChatID: chatID, UserID: userID, DrawDate: date, AttemptNo: attemptNo, Result: "lost", CostPoints: costPoints}).Error; err != nil {
				return err
			}
			result.Result = "lost"
			result.Remaining = maxInt(dailyLotteryAttempts-attemptNo, 0)
			return nil
		}
		var code model.DailyLotteryCode
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("chat_id = ? AND status = ? AND amount = ? AND (expires_at IS NULL OR expires_at > ?)", chatID, "available", selected.Amount, time.Now()).Order("created_at asc")
		if err := query.First(&code).Error; err != nil {
			return errors.New("该额度兑换码刚刚被领完，请重新抽奖")
		}
		assignedAt := time.Now()
		if err := tx.Model(&model.DailyLotteryCode{}).Where("id = ? AND status = ?", code.ID, "available").Updates(map[string]any{"status": "assigned", "assigned_user": userID, "assigned_at": assignedAt, "updated_at": assignedAt}).Error; err != nil {
			return err
		}
		codeID := uuid.UUID(code.ID)
		if err := tx.Create(&model.DailyLotteryAttempt{ChatID: chatID, UserID: userID, DrawDate: date, AttemptNo: attemptNo, Result: "won", Amount: selected.Amount, CodeID: &codeID, CostPoints: costPoints}).Error; err != nil {
			return err
		}
		result.Result = "won"
		result.Amount = selected.Amount
		result.Code = code.Code
		result.RedeemURL = code.RedeemURL
		result.Remaining = maxInt(dailyLotteryAttempts-attemptNo, 0)
		return nil
	})
	return result, err
}

func (s *DailyLotteryService) History(ctx context.Context, chatID, userID int64, limit int) ([]bot.DailyLotteryAttemptRecord, error) {
	if limit <= 0 || limit > 30 {
		limit = 30
	}
	if s == nil || s.store == nil || s.store.DB == nil {
		return []bot.DailyLotteryAttemptRecord{}, nil
	}
	var rows []model.DailyLotteryAttempt
	if err := s.store.DB.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chatID, userID).Order("created_at desc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]bot.DailyLotteryAttemptRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, bot.DailyLotteryAttemptRecord{DrawDate: row.DrawDate, AttemptNo: row.AttemptNo, Result: row.Result, Amount: row.Amount, CostPoints: row.CostPoints})
	}
	return items, nil
}

func dailyLotteryAvailablePrizes(tx *gorm.DB, chatID int64) ([]bot.DailyLotteryPrize, error) {
	var prizes []model.DailyLotteryPrize
	if err := tx.Where("chat_id = ? AND enabled = ?", chatID, true).Order("amount asc").Find(&prizes).Error; err != nil {
		return nil, err
	}
	items := make([]bot.DailyLotteryPrize, 0, len(prizes))
	for _, prize := range prizes {
		var count int64
		if err := tx.Model(&model.DailyLotteryCode{}).Where("chat_id = ? AND status = ? AND amount = ? AND (expires_at IS NULL OR expires_at > ?)", chatID, "available", prize.Amount, time.Now()).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			prizeItem := bot.DailyLotteryPrize{ID: prize.ID, ChatID: prize.ChatID, Amount: prize.Amount, Weight: prize.Weight, Enabled: prize.Enabled, AvailableCode: int(count)}
			items = append(items, prizeItem)
		}
	}
	return items, nil
}

func (s *DailyLotteryService) availableCount(ctx context.Context, chatID int64, amount int) (int, error) {
	var count int64
	if s == nil || s.store == nil || s.store.DB == nil {
		return 0, nil
	}
	err := s.store.DB.WithContext(ctx).Model(&model.DailyLotteryCode{}).Where("chat_id = ? AND status = ? AND amount = ? AND (expires_at IS NULL OR expires_at > ?)", chatID, "available", amount, time.Now()).Count(&count).Error
	return int(count), err
}

func chooseDailyPrize(prizes []bot.DailyLotteryPrize, guaranteed bool) (bot.DailyLotteryPrize, bool, error) {
	if len(prizes) == 0 {
		return bot.DailyLotteryPrize{}, false, errors.New("当前奖池暂无可用兑换码")
	}
	if guaranteed {
		index, err := randomInt(len(prizes))
		return prizes[index], true, err
	}
	total := 0
	for _, prize := range prizes {
		total += prize.Weight
	}
	roll, err := randomInt(1000)
	if err != nil {
		return bot.DailyLotteryPrize{}, false, err
	}
	if roll >= total {
		return bot.DailyLotteryPrize{}, false, nil
	}
	cursor := 0
	for _, prize := range prizes {
		cursor += prize.Weight
		if roll < cursor {
			return prize, true, nil
		}
	}
	return bot.DailyLotteryPrize{}, false, nil
}

func randomInt(max int) (int, error) {
	if max <= 0 {
		return 0, errors.New("invalid random range")
	}
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

func previousAttemptsLost(attempts []model.DailyLotteryAttempt) bool {
	if len(attempts) != dailyLotteryAttempts-1 {
		return false
	}
	for _, attempt := range attempts {
		if attempt.Result != "lost" {
			return false
		}
	}
	return true
}

func dailyLotteryConfigToBot(row model.DailyLotteryConfig) bot.DailyLotteryConfig {
	return bot.DailyLotteryConfig{ChatID: row.ChatID, Enabled: row.Enabled, DailyAttempts: dailyLotteryAttempts, CostPoints: row.CostPoints, PaidEnabled: row.PaidEnabled, GuaranteeOnLast: row.GuaranteeOnLast}
}

func (s *DailyLotteryService) ResetAttempts(ctx context.Context, chatID, userID int64) error {
	if chatID == 0 || userID == 0 {
		return errors.New("chat_id and user_id are required")
	}
	if s == nil || s.store == nil || s.store.DB == nil {
		return nil
	}
	date := time.Now().In(chinaLocation()).Format("2006-01-02")
	return s.store.DB.WithContext(ctx).Where("chat_id = ? AND user_id = ? AND draw_date = ?", chatID, userID, date).Delete(&model.DailyLotteryAttempt{}).Error
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
