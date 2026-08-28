package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dabowin/sola/internal/bot"
	"github.com/dabowin/sola/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *DailyLotteryService) ListCodes(ctx context.Context, chatID int64, amount *int, status string, limit int) ([]bot.DailyLotteryCode, error) {
	if s == nil || s.store == nil || s.store.DB == nil {
		return []bot.DailyLotteryCode{}, nil
	}
	if chatID == 0 {
		return nil, errors.New("chat_id is required")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := s.store.DB.WithContext(ctx).Where("chat_id = ?", chatID).Order("created_at DESC").Limit(limit)
	if amount != nil {
		if *amount <= 0 {
			return nil, errors.New("额度必须大于 0")
		}
		query = query.Where("amount = ?", *amount)
	}
	if status != "" {
		if status != "available" && status != "assigned" && status != "used" {
			return nil, errors.New("无效库存状态")
		}
		query = query.Where("status = ?", status)
	}
	var rows []model.DailyLotteryCode
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]bot.DailyLotteryCode, 0, len(rows))
	for _, row := range rows {
		items = append(items, dailyLotteryCodeToBot(row))
	}
	return items, nil
}

func (s *DailyLotteryService) SummarizeCodes(ctx context.Context, chatID int64) ([]bot.DailyLotteryCodeSummary, error) {
	if s == nil || s.store == nil || s.store.DB == nil {
		return []bot.DailyLotteryCodeSummary{}, nil
	}
	if chatID == 0 {
		return nil, errors.New("chat_id is required")
	}
	var rows []struct {
		Amount int
		Status string
		Count  int
	}
	if err := s.store.DB.WithContext(ctx).Model(&model.DailyLotteryCode{}).
		Select("amount, status, COUNT(*) AS count").Where("chat_id = ?", chatID).
		Group("amount, status").Order("amount ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	byAmount := map[int]*bot.DailyLotteryCodeSummary{}
	for _, row := range rows {
		item := byAmount[row.Amount]
		if item == nil {
			item = &bot.DailyLotteryCodeSummary{Amount: row.Amount}
			byAmount[row.Amount] = item
		}
		switch row.Status {
		case "available":
			item.Available = row.Count
		case "assigned":
			item.Assigned = row.Count
		case "used":
			item.Used = row.Count
		}
		item.Total += row.Count
	}
	result := make([]bot.DailyLotteryCodeSummary, 0, len(byAmount))
	for _, item := range byAmount {
		result = append(result, *item)
	}
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Amount < result[i].Amount {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result, nil
}

func (s *DailyLotteryService) ImportCodes(ctx context.Context, req bot.DailyLotteryCodeImportRequest) (bot.DailyLotteryCodeImportResult, error) {
	if s == nil || s.store == nil || s.store.DB == nil {
		return bot.DailyLotteryCodeImportResult{}, errors.New("抽奖库存服务尚未接入数据库")
	}
	if req.ChatID == 0 || req.Amount <= 0 {
		return bot.DailyLotteryCodeImportResult{}, errors.New("群组和额度不能为空，额度必须大于 0")
	}
	if strings.TrimSpace(req.RedeemURL) != "" {
		u, err := url.Parse(strings.TrimSpace(req.RedeemURL))
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return bot.DailyLotteryCodeImportResult{}, errors.New("兑换地址必须使用有效的 https 地址")
		}
	}
	codes := make([]string, 0, len(req.Codes))
	seen := make(map[string]struct{}, len(req.Codes))
	duplicateCount := 0
	for _, raw := range req.Codes {
		code := strings.TrimSpace(raw)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			duplicateCount++
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return bot.DailyLotteryCodeImportResult{}, errors.New("至少填写一个兑换码")
	}
	result := bot.DailyLotteryCodeImportResult{Skipped: duplicateCount}
	err := s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, code := range codes {
			var count int64
			if err := tx.Model(&model.DailyLotteryCode{}).Where("chat_id = ? AND code = ?", req.ChatID, code).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				result.Skipped++
				continue
			}
			now := time.Now()
			row := model.DailyLotteryCode{ID: uuid.New(), ChatID: req.ChatID, Code: code, Amount: req.Amount, BatchName: strings.TrimSpace(req.BatchName), RedeemURL: strings.TrimSpace(req.RedeemURL), ExpiresAt: req.ExpiresAt, Status: "available", CreatedAt: now, UpdatedAt: now}
			createResult := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
			if createResult.Error != nil {
				return createResult.Error
			}
			if createResult.RowsAffected == 0 {
				result.Skipped++
				continue
			}
			result.Imported++
		}
		return nil
	})
	if err != nil {
		return bot.DailyLotteryCodeImportResult{}, fmt.Errorf("导入抽奖兑换码失败: %w", err)
	}
	return result, nil
}

func dailyLotteryCodeToBot(row model.DailyLotteryCode) bot.DailyLotteryCode {
	return bot.DailyLotteryCode{ID: row.ID.String(), ChatID: row.ChatID, Code: row.Code, Amount: row.Amount, BatchName: row.BatchName, RedeemURL: row.RedeemURL, ExpiresAt: row.ExpiresAt, Status: row.Status, AssignedUser: row.AssignedUser, AssignedAt: row.AssignedAt}
}
