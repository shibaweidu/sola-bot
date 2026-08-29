package service

import (
	"context"
	"fmt"

	"github.com/dabowin/sola/internal/api"
	"github.com/dabowin/sola/internal/bot"
)

type dailyLotteryAPIService struct{ service *DailyLotteryService }

func (s *dailyLotteryAPIService) GetConfig(ctx context.Context, chatID int64) (api.DailyLotteryConfig, error) {
	cfg, err := s.service.GetConfig(ctx, chatID)
	if err != nil {
		return api.DailyLotteryConfig{}, err
	}
	return api.DailyLotteryConfig{ChatID: cfg.ChatID, Enabled: cfg.Enabled, DailyAttempts: cfg.DailyAttempts, CostPoints: cfg.CostPoints, PaidEnabled: cfg.PaidEnabled, GuaranteeOnLast: cfg.GuaranteeOnLast}, nil
}

func (s *dailyLotteryAPIService) ListCodes(ctx context.Context, chatID int64, amount *int, status string, limit int) ([]api.DailyLotteryCode, error) {
	items, err := s.service.ListCodes(ctx, chatID, amount, status, limit)
	if err != nil {
		return nil, err
	}
	result := make([]api.DailyLotteryCode, 0, len(items))
	for _, item := range items {
		result = append(result, api.DailyLotteryCode{ID: item.ID, ChatID: item.ChatID, Code: item.Code, Amount: item.Amount, BatchName: item.BatchName, RedeemURL: item.RedeemURL, ExpiresAt: item.ExpiresAt, Status: item.Status, AssignedUser: item.AssignedUser, AssignedAt: item.AssignedAt})
	}
	return result, nil
}

func (s *dailyLotteryAPIService) ListCodesPage(ctx context.Context, chatID int64, amount *int, status string, page, pageSize int) (api.DailyLotteryCodePage, error) {
	result, err := s.service.ListCodesPage(ctx, chatID, amount, status, page, pageSize)
	if err != nil {
		return api.DailyLotteryCodePage{}, err
	}
	items := make([]api.DailyLotteryCode, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, api.DailyLotteryCode{ID: item.ID, ChatID: item.ChatID, Code: item.Code, Amount: item.Amount, BatchName: item.BatchName, RedeemURL: item.RedeemURL, ExpiresAt: item.ExpiresAt, Status: item.Status, AssignedUser: item.AssignedUser, AssignedAt: item.AssignedAt})
	}
	return api.DailyLotteryCodePage{Items: items, Total: result.Total, Page: result.Page, PageSize: result.PageSize}, nil
}

func (s *dailyLotteryAPIService) SummarizeCodes(ctx context.Context, chatID int64) ([]api.DailyLotteryCodeSummary, error) {
	items, err := s.service.SummarizeCodes(ctx, chatID)
	if err != nil {
		return nil, err
	}
	result := make([]api.DailyLotteryCodeSummary, 0, len(items))
	for _, item := range items {
		result = append(result, api.DailyLotteryCodeSummary{Amount: item.Amount, Available: item.Available, Assigned: item.Assigned, Used: item.Used, Total: item.Total})
	}
	return result, nil
}

func (s *dailyLotteryAPIService) ImportCodes(ctx context.Context, req api.DailyLotteryCodeImportRequest) (api.DailyLotteryCodeImportResult, error) {
	result, err := s.service.ImportCodes(ctx, bot.DailyLotteryCodeImportRequest{ChatID: req.ChatID, Amount: req.Amount, BatchName: req.BatchName, RedeemURL: req.RedeemURL, ExpiresAt: req.ExpiresAt, Codes: req.Codes})
	return api.DailyLotteryCodeImportResult{Imported: result.Imported, Skipped: result.Skipped}, err
}

func (s *dailyLotteryAPIService) BatchUpdateCodes(ctx context.Context, req api.DailyLotteryCodeBatchUpdateRequest) (api.InventoryBatchResult, error) {
	result, err := s.service.BatchUpdateCodes(ctx, bot.DailyLotteryCodeBatchUpdateRequest{ChatID: req.ChatID, IDs: req.IDs, RedeemURL: req.RedeemURL, BatchName: req.BatchName, ExpiresAt: req.ExpiresAt})
	return api.InventoryBatchResult{Updated: result.Updated, Deleted: result.Deleted, Skipped: result.Skipped}, err
}

func (s *dailyLotteryAPIService) BatchDeleteCodes(ctx context.Context, req api.DailyLotteryCodeBatchDeleteRequest) (api.InventoryBatchResult, error) {
	result, err := s.service.BatchDeleteCodes(ctx, bot.DailyLotteryCodeBatchDeleteRequest{ChatID: req.ChatID, IDs: req.IDs})
	return api.InventoryBatchResult{Updated: result.Updated, Deleted: result.Deleted, Skipped: result.Skipped}, err
}

func (s *dailyLotteryAPIService) UpdateConfig(ctx context.Context, req api.DailyLotteryUpdateRequest) (api.DailyLotteryConfig, error) {
	if req.PaidEnabled && req.CostPoints <= 0 {
		return api.DailyLotteryConfig{}, fmt.Errorf("开启积分抽奖后，每次消耗积分必须大于 0")
	}
	if _, err := s.ReplacePrizes(ctx, int64(req.ChatID), req.Prizes); err != nil {
		return api.DailyLotteryConfig{}, err
	}
	cfg, err := s.service.UpdateConfig(ctx, bot.DailyLotteryConfig{ChatID: int64(req.ChatID), Enabled: req.Enabled, CostPoints: req.CostPoints, PaidEnabled: req.PaidEnabled, GuaranteeOnLast: req.GuaranteeOnLast})
	if err != nil {
		return api.DailyLotteryConfig{}, err
	}
	return api.DailyLotteryConfig{ChatID: cfg.ChatID, Enabled: cfg.Enabled, DailyAttempts: cfg.DailyAttempts, CostPoints: cfg.CostPoints, PaidEnabled: cfg.PaidEnabled, GuaranteeOnLast: cfg.GuaranteeOnLast}, nil
}

func (s *dailyLotteryAPIService) ResetAttempts(ctx context.Context, chatID, userID int64) error {
	return s.service.ResetAttempts(ctx, chatID, userID)
}

func (s *dailyLotteryAPIService) ListPrizes(ctx context.Context, chatID int64) ([]api.DailyLotteryPrize, error) {
	items, err := s.service.ListPrizes(ctx, chatID)
	if err != nil {
		return nil, err
	}
	return dailyLotteryPrizesToAPI(items), nil
}

func (s *dailyLotteryAPIService) ReplacePrizes(ctx context.Context, chatID int64, inputs []api.DailyLotteryPrizeInput) ([]api.DailyLotteryPrize, error) {
	items := make([]bot.DailyLotteryPrize, 0, len(inputs))
	for _, item := range inputs {
		items = append(items, bot.DailyLotteryPrize{ChatID: chatID, Amount: item.Amount, Weight: item.Weight, Enabled: item.Enabled})
	}
	result, err := s.service.ReplacePrizes(ctx, chatID, items)
	if err != nil {
		return nil, err
	}
	return dailyLotteryPrizesToAPI(result), nil
}

func dailyLotteryPrizesToAPI(items []bot.DailyLotteryPrize) []api.DailyLotteryPrize {
	result := make([]api.DailyLotteryPrize, 0, len(items))
	for _, item := range items {
		result = append(result, api.DailyLotteryPrize{ID: item.ID, ChatID: item.ChatID, Amount: item.Amount, Weight: item.Weight, Enabled: item.Enabled, AvailableCode: item.AvailableCode})
	}
	return result
}
