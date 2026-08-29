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
	page, err := s.ListCodesPage(ctx, chatID, amount, status, 1, limit)
	return page.Items, err
}

func (s *DailyLotteryService) ListCodesPage(ctx context.Context, chatID int64, amount *int, status string, page, pageSize int) (bot.DailyLotteryCodePage, error) {
	if s == nil || s.store == nil || s.store.DB == nil {
		return bot.DailyLotteryCodePage{Items: []bot.DailyLotteryCode{}, Page: 1, PageSize: 20}, nil
	}
	if chatID == 0 {
		return bot.DailyLotteryCodePage{}, errors.New("chat_id is required")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	query := s.store.DB.WithContext(ctx).Model(&model.DailyLotteryCode{}).Where("chat_id = ?", chatID)
	if amount != nil {
		if *amount <= 0 {
			return bot.DailyLotteryCodePage{}, errors.New("额度必须大于 0")
		}
		query = query.Where("amount = ?", *amount)
	}
	if status != "" {
		if status != "available" && status != "assigned" && status != "used" {
			return bot.DailyLotteryCodePage{}, errors.New("无效库存状态")
		}
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return bot.DailyLotteryCodePage{}, err
	}
	query = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize)
	var rows []model.DailyLotteryCode
	if err := query.Find(&rows).Error; err != nil {
		return bot.DailyLotteryCodePage{}, err
	}
	items := make([]bot.DailyLotteryCode, 0, len(rows))
	for _, row := range rows {
		items = append(items, dailyLotteryCodeToBot(row))
	}
	return bot.DailyLotteryCodePage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
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
		importedAmount := false
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
			importedAmount = true
		}
		if importedAmount {
			// Inventory import is the explicit event that may create a missing
			// prize. This keeps manual prize deletion persistent across reads.
			if err := ensureInventoryPrize(tx, req.ChatID, req.Amount); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return bot.DailyLotteryCodeImportResult{}, fmt.Errorf("导入抽奖兑换码失败: %w", err)
	}
	return result, nil
}

func (s *DailyLotteryService) BatchUpdateCodes(ctx context.Context, req bot.DailyLotteryCodeBatchUpdateRequest) (bot.InventoryBatchResult, error) {
	if req.ChatID == 0 {
		return bot.InventoryBatchResult{}, errors.New("chat_id is required")
	}
	ids, err := parseInventoryIDs(req.IDs)
	if err != nil {
		return bot.InventoryBatchResult{}, err
	}
	updates, err := inventoryUpdates(req.RedeemURL, req.BatchName, req.ExpiresAt)
	if err != nil {
		return bot.InventoryBatchResult{}, err
	}
	if len(updates) == 0 {
		return bot.InventoryBatchResult{}, errors.New("至少提供一个要修改的字段")
	}
	result := bot.InventoryBatchResult{Skipped: len(ids)}
	err = s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.DailyLotteryCode{}).Where("id IN ? AND chat_id = ? AND status = ?", ids, req.ChatID, "available").Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		result.Updated = int(res.RowsAffected)
		result.Skipped = len(ids) - result.Updated
		return nil
	})
	return result, err
}

func (s *DailyLotteryService) BatchDeleteCodes(ctx context.Context, req bot.DailyLotteryCodeBatchDeleteRequest) (bot.InventoryBatchResult, error) {
	if req.ChatID == 0 {
		return bot.InventoryBatchResult{}, errors.New("chat_id is required")
	}
	ids, err := parseInventoryIDs(req.IDs)
	if err != nil {
		return bot.InventoryBatchResult{}, err
	}
	result := bot.InventoryBatchResult{Skipped: len(ids)}
	err = s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("id IN ? AND chat_id = ? AND status = ?", ids, req.ChatID, "available").Delete(&model.DailyLotteryCode{})
		if res.Error != nil {
			return res.Error
		}
		result.Deleted = int(res.RowsAffected)
		result.Skipped = len(ids) - result.Deleted
		return nil
	})
	return result, err
}

func dailyLotteryCodeToBot(row model.DailyLotteryCode) bot.DailyLotteryCode {
	return bot.DailyLotteryCode{ID: row.ID.String(), ChatID: row.ChatID, Code: row.Code, Amount: row.Amount, BatchName: row.BatchName, RedeemURL: row.RedeemURL, ExpiresAt: row.ExpiresAt, Status: row.Status, AssignedUser: row.AssignedUser, AssignedAt: row.AssignedAt}
}
