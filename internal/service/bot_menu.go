package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/dabowin/sola/internal/model"
	"github.com/dabowin/sola/internal/store"
)

var builtinBotMenuActions = map[string]struct{}{
	"points":          {},
	"sign":            {},
	"rank":            {},
	"lottery":         {},
	"help":            {},
	"info":            {},
	"private_console": {},
	"admin_center":    {},
	"admin_config":    {},
	"scheduled_posts": {},
	"hide_keyboard":   {},
	"invite_rewards":  {},
	"exchange":        {},
	"purchase":        {},
	"shop":            {},
}

type BotMenuService struct {
	store         *store.Store
	telegramBotID int64
}

func NewBotMenuService(st *store.Store, token string) *BotMenuService {
	return &BotMenuService{
		store:         st,
		telegramBotID: telegramBotIDFromToken(token),
	}
}

func telegramBotIDFromToken(token string) int64 {
	first, _, ok := strings.Cut(strings.TrimSpace(token), ":")
	if !ok {
		return 0
	}
	id, err := strconv.ParseInt(first, 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func (s *BotMenuService) List(ctx context.Context, role string) ([]model.BotMenuItem, error) {
	if !model.IsValidBotMenuRole(role) {
		return nil, fmt.Errorf("invalid bot menu role")
	}
	if s == nil || s.store == nil || s.store.DB == nil || s.telegramBotID == 0 {
		return defaultBotMenu(role), nil
	}

	var items []model.BotMenuItem
	err := s.store.DB.WithContext(ctx).
		Where("telegram_bot_id = ? AND role = ?", s.telegramBotID, role).
		Order("row_index ASC, column_index ASC, created_at ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		if _, err := s.Replace(ctx, role, defaultBotMenu(role)); err != nil {
			return nil, err
		}
		return s.List(ctx, role)
	}
	return items, nil
}

func (s *BotMenuService) Replace(ctx context.Context, role string, items []model.BotMenuItem) ([]model.BotMenuItem, error) {
	if !model.IsValidBotMenuRole(role) {
		return nil, fmt.Errorf("invalid bot menu role")
	}
	if err := validateBotMenuItems(role, items); err != nil {
		return nil, err
	}
	if s == nil || s.store == nil || s.store.DB == nil || s.telegramBotID == 0 {
		return items, nil
	}

	err := s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("telegram_bot_id = ? AND role = ?", s.telegramBotID, role).Delete(&model.BotMenuItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ID = model.BotMenuItem{}.ID
			items[i].TelegramBotID = s.telegramBotID
			items[i].Role = role
			items[i].Label = strings.TrimSpace(items[i].Label)
			items[i].Icon = strings.TrimSpace(items[i].Icon)
			items[i].ActionType = strings.TrimSpace(items[i].ActionType)
			items[i].ActionKey = strings.TrimSpace(items[i].ActionKey)
			items[i].ActionValue = strings.TrimSpace(items[i].ActionValue)
			items[i].MessageText = strings.TrimSpace(items[i].MessageText)
			items[i].LinkMode = model.NormalizeBotMenuLinkMode(items[i].LinkMode)
			links, err := model.ParseBotMenuInlineLinks(items[i].LinkButtonsJSON)
			if err != nil {
				return err
			}
			for linkIndex := range links {
				links[linkIndex].Label = strings.TrimSpace(links[linkIndex].Label)
				links[linkIndex].URL = strings.TrimSpace(links[linkIndex].URL)
			}
			items[i].LinkButtonsJSON, err = model.EncodeBotMenuInlineLinks(links)
			if err != nil {
				return err
			}
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.List(ctx, role)
}

func (s *BotMenuService) Reset(ctx context.Context, role string) ([]model.BotMenuItem, error) {
	return s.Replace(ctx, role, defaultBotMenu(role))
}

func validateBotMenuItems(role string, items []model.BotMenuItem) error {
	labels := make(map[string]struct{}, len(items))
	positions := make(map[string]struct{}, len(items))
	enabledCount := 0
	for _, item := range items {
		if item.Role != "" && item.Role != role {
			return fmt.Errorf("menu item role does not match request role")
		}
		label := strings.TrimSpace(item.RenderLabel())
		if label == "" {
			return fmt.Errorf("menu button label is required")
		}
		if len([]rune(label)) > 64 {
			return fmt.Errorf("menu button label is too long")
		}
		if item.RowIndex < 0 || item.ColumnIndex < 0 || item.ColumnIndex >= 4 {
			return fmt.Errorf("invalid menu button position")
		}
		position := fmt.Sprintf("%d:%d", item.RowIndex, item.ColumnIndex)
		if _, exists := positions[position]; exists {
			return fmt.Errorf("duplicate menu button position")
		}
		positions[position] = struct{}{}
		if _, exists := labels[label]; exists {
			return fmt.Errorf("duplicate menu button label")
		}
		labels[label] = struct{}{}
		if item.Enabled {
			enabledCount++
		}
		if !model.IsValidBotMenuAction(item.ActionType) {
			return fmt.Errorf("invalid menu action type")
		}
		if item.ActionType == model.BotMenuActionBuiltin {
			if _, ok := builtinBotMenuActions[item.ActionKey]; !ok {
				return fmt.Errorf("unsupported built-in menu action: %s", item.ActionKey)
			}
		} else if item.ActionType == model.BotMenuActionLink {
			if !model.IsValidBotMenuLink(item.ActionValue) {
				return fmt.Errorf("menu link must be a valid https URL")
			}
		} else {
			if strings.TrimSpace(item.MessageText) == "" {
				return fmt.Errorf("custom menu message is required")
			}
			if len([]rune(item.MessageText)) > 4000 {
				return fmt.Errorf("custom menu message is too long")
			}
			linkMode := model.NormalizeBotMenuLinkMode(item.LinkMode)
			if !model.IsValidBotMenuLinkMode(linkMode) {
				return fmt.Errorf("invalid custom menu link mode")
			}
			if linkMode == model.BotMenuLinkModeText {
				if strings.TrimSpace(item.ActionValue) != "" && !model.IsValidBotMenuLink(item.ActionValue) {
					return fmt.Errorf("custom menu link must be a valid https URL")
				}
				continue
			}
			links, err := model.ParseBotMenuInlineLinks(item.LinkButtonsJSON)
			if err != nil {
				return fmt.Errorf("invalid custom menu inline links")
			}
			if len(links) == 0 || len(links) > 10 {
				return fmt.Errorf("custom menu inline buttons must contain 1 to 10 links")
			}
			linkPositions := make(map[string]struct{}, len(links))
			for _, link := range links {
				if strings.TrimSpace(link.Label) == "" || len([]rune(strings.TrimSpace(link.Label))) > 64 {
					return fmt.Errorf("inline button label is required and must not exceed 64 characters")
				}
				if !model.IsValidBotMenuLink(link.URL) {
					return fmt.Errorf("inline button link must be a valid https URL")
				}
				if link.RowIndex < 0 || link.ColumnIndex < 0 || link.ColumnIndex >= 4 {
					return fmt.Errorf("invalid inline button position")
				}
				position := fmt.Sprintf("%d:%d", link.RowIndex, link.ColumnIndex)
				if _, exists := linkPositions[position]; exists {
					return fmt.Errorf("duplicate inline button position")
				}
				linkPositions[position] = struct{}{}
			}
		}
	}
	if enabledCount > 20 {
		return fmt.Errorf("a menu can contain at most 20 enabled buttons")
	}
	return nil
}

func defaultBotMenu(role string) []model.BotMenuItem {
	items := []model.BotMenuItem{
		{Role: role, Label: "🎁 邀请免费领额度", Icon: "🎁", ActionType: model.BotMenuActionBuiltin, ActionKey: "invite_rewards", RowIndex: 0, ColumnIndex: 0, Enabled: true},
		{Role: role, Label: "💎 我的积分", Icon: "💎", ActionType: model.BotMenuActionBuiltin, ActionKey: "points", RowIndex: 0, ColumnIndex: 1, Enabled: true},
		{Role: role, Label: "📅 每日签到", Icon: "📅", ActionType: model.BotMenuActionBuiltin, ActionKey: "sign", RowIndex: 1, ColumnIndex: 0, Enabled: true},
		{Role: role, Label: "🎫 积分兑换免费额度", Icon: "🎫", ActionType: model.BotMenuActionBuiltin, ActionKey: "exchange", RowIndex: 1, ColumnIndex: 1, Enabled: true},
		{Role: role, Label: "🛒 直接购买额度", Icon: "🛒", ActionType: model.BotMenuActionBuiltin, ActionKey: "purchase", RowIndex: 2, ColumnIndex: 0, Enabled: true},
		{Role: role, Label: "🏪 小铺地址", Icon: "🏪", ActionType: model.BotMenuActionBuiltin, ActionKey: "shop", RowIndex: 2, ColumnIndex: 1, Enabled: true},
		{Role: role, Label: "🏆 积分榜", Icon: "🏆", ActionType: model.BotMenuActionBuiltin, ActionKey: "rank", RowIndex: 3, ColumnIndex: 0, Enabled: true},
	}
	if role != model.BotMenuRoleAdmin {
		return items
	}
	items = append(items,
		model.BotMenuItem{Role: role, Label: "📋 运营工作台", Icon: "📋", ActionType: model.BotMenuActionBuiltin, ActionKey: "private_console", RowIndex: 4, ColumnIndex: 0, Enabled: true},
		model.BotMenuItem{Role: role, Label: "🛡 群管中心", Icon: "🛡", ActionType: model.BotMenuActionBuiltin, ActionKey: "admin_center", RowIndex: 4, ColumnIndex: 1, Enabled: true},
		model.BotMenuItem{Role: role, Label: "⚙️ 群组配置", Icon: "⚙️", ActionType: model.BotMenuActionBuiltin, ActionKey: "admin_config", RowIndex: 5, ColumnIndex: 0, Enabled: true},
		model.BotMenuItem{Role: role, Label: "📣 定时发帖", Icon: "📣", ActionType: model.BotMenuActionBuiltin, ActionKey: "scheduled_posts", RowIndex: 5, ColumnIndex: 1, Enabled: true},
	)
	return items
}
