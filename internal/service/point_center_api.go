package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/dabowin/sola/internal/api"
	"github.com/dabowin/sola/internal/bot"
	"github.com/dabowin/sola/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointCenterAPIService struct{ service *PointCenterService }

func (s *pointCenterAPIService) GetConfig(ctx context.Context, chatID int64) (api.PointCenterConfig, error) {
	cfg, err := s.service.GetConfig(ctx, chatID)
	return api.PointCenterConfig{ChatID: cfg.ChatID, InviteEnabled: cfg.InviteEnabled, InviterReward: cfg.InviterReward, InviteeReward: cfg.InviteeReward, SignEnabled: cfg.SignEnabled, SignReward: cfg.SignReward, ExchangeEnabled: cfg.ExchangeEnabled, ExchangeMinimum: cfg.ExchangeMinimum, ExchangeRate: cfg.ExchangeRate, ExchangeURL: cfg.ExchangeURL, ExchangeInstructions: cfg.ExchangeInstructions, PurchaseURL: cfg.PurchaseURL, PurchaseText: cfg.PurchaseText, ShopURL: cfg.ShopURL, ShopText: cfg.ShopText, InviteText: cfg.InviteText, InvitePageTemplate: cfg.InvitePageTemplate, InviteJoinURL: cfg.InviteJoinURL, InviteJoinText: cfg.InviteJoinText, InviteSuccessText: cfg.InviteSuccessText, PointsText: cfg.PointsText, SignText: cfg.SignText, ExchangeText: cfg.ExchangeText, RankText: cfg.RankText}, err
}

func (s *pointCenterAPIService) UpdateConfig(ctx context.Context, input api.PointCenterConfig) (api.PointCenterConfig, error) {
	cfg, err := s.service.UpdateConfig(ctx, botPointCenterFromAPI(input))
	if err != nil {
		return api.PointCenterConfig{}, err
	}
	return (&pointCenterAPIService{service: s.service}).GetConfig(ctx, cfg.ChatID)
}

func (s *pointCenterAPIService) ResetReferralForTesting(ctx context.Context, chatID, userID int64) error {
	return s.service.ResetReferralForTesting(ctx, 0, chatID, userID)
}

func (s *pointCenterAPIService) ListExchangeCodes(ctx context.Context, query api.ExchangeCodeListQuery) ([]api.ExchangeCode, error) {
	if s.service == nil || s.service.store == nil || s.service.store.DB == nil {
		return []api.ExchangeCode{}, nil
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	db := s.service.store.DB.WithContext(ctx).Model(&model.ExchangeCode{}).Order("created_at DESC").Limit(limit)
	if strings.TrimSpace(query.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(query.Status))
	}
	if query.Amount != nil {
		db = db.Where("amount = ?", *query.Amount)
	}
	var rows []model.ExchangeCode
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]api.ExchangeCode, 0, len(rows))
	for _, row := range rows {
		out = append(out, exchangeCodeToAPI(row))
	}
	return out, nil
}

func (s *pointCenterAPIService) SummarizeExchangeCodes(ctx context.Context) ([]api.ExchangeCodeSummary, error) {
	if s.service == nil || s.service.store == nil || s.service.store.DB == nil {
		return []api.ExchangeCodeSummary{}, nil
	}
	var rows []model.ExchangeCode
	if err := s.service.store.DB.WithContext(ctx).Model(&model.ExchangeCode{}).Select("amount, status").Find(&rows).Error; err != nil {
		return nil, err
	}
	byAmount := make(map[int]*api.ExchangeCodeSummary)
	for _, row := range rows {
		item := byAmount[row.Amount]
		if item == nil {
			item = &api.ExchangeCodeSummary{Amount: row.Amount}
			byAmount[row.Amount] = item
		}
		item.Total++
		switch strings.ToLower(strings.TrimSpace(row.Status)) {
		case "available":
			item.Available++
		case "used":
			item.Used++
		case "assigned":
			item.Assigned++
		}
	}
	result := make([]api.ExchangeCodeSummary, 0, len(byAmount))
	for _, item := range byAmount {
		result = append(result, *item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Amount < result[j].Amount })
	return result, nil
}

func (s *pointCenterAPIService) ImportExchangeCodes(ctx context.Context, req api.ExchangeCodeImportRequest) (api.ExchangeCodeImportResult, error) {
	if s.service == nil || s.service.store == nil || s.service.store.DB == nil {
		return api.ExchangeCodeImportResult{}, errors.New("database is not configured")
	}
	if req.Amount <= 0 || len(req.Codes) == 0 {
		return api.ExchangeCodeImportResult{}, errors.New("兑换码和额度不能为空")
	}
	if req.RedeemURL != "" && !validHTTPSURL(req.RedeemURL) {
		return api.ExchangeCodeImportResult{}, errors.New("兑换地址必须使用 https")
	}
	seen := make(map[string]struct{}, len(req.Codes))
	result := api.ExchangeCodeImportResult{}
	err := s.service.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, raw := range req.Codes {
			code := strings.TrimSpace(raw)
			if code == "" {
				result.Skipped++
				continue
			}
			if _, ok := seen[code]; ok {
				result.Skipped++
				continue
			}
			seen[code] = struct{}{}
			var count int64
			if err := tx.Model(&model.ExchangeCode{}).Where("code = ?", code).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				result.Skipped++
				continue
			}
			now := time.Now()
			row := model.ExchangeCode{BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now}, Code: code, BatchName: strings.TrimSpace(req.BatchName), Amount: req.Amount, RedeemURL: strings.TrimSpace(req.RedeemURL), ExpiresAt: req.ExpiresAt, Status: "available"}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			result.Imported++
		}
		return nil
	})
	return result, err
}

func botPointCenterFromAPI(input api.PointCenterConfig) bot.PointCenterConfig {
	return bot.PointCenterConfig{ChatID: input.ChatID, InviteEnabled: input.InviteEnabled, InviterReward: input.InviterReward, InviteeReward: input.InviteeReward, SignEnabled: input.SignEnabled, SignReward: input.SignReward, ExchangeEnabled: input.ExchangeEnabled, ExchangeMinimum: input.ExchangeMinimum, ExchangeRate: input.ExchangeRate, ExchangeURL: input.ExchangeURL, ExchangeInstructions: input.ExchangeInstructions, PurchaseURL: input.PurchaseURL, PurchaseText: input.PurchaseText, ShopURL: input.ShopURL, ShopText: input.ShopText, InviteText: input.InviteText, InvitePageTemplate: input.InvitePageTemplate, InviteJoinURL: input.InviteJoinURL, InviteJoinText: input.InviteJoinText, InviteSuccessText: input.InviteSuccessText, PointsText: input.PointsText, SignText: input.SignText, ExchangeText: input.ExchangeText, RankText: input.RankText}
}

func exchangeCodeToAPI(row model.ExchangeCode) api.ExchangeCode {
	return api.ExchangeCode{ID: row.ID.String(), Code: row.Code, BatchName: row.BatchName, Amount: row.Amount, RedeemURL: row.RedeemURL, ExpiresAt: row.ExpiresAt, Status: row.Status, AssignedUser: row.AssignedUser, AssignedChat: row.AssignedChat, AssignedAt: row.AssignedAt}
}
