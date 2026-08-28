package service

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strings"

	"github.com/dabowin/sola/internal/bot"
	"github.com/dabowin/sola/internal/model"
	"github.com/dabowin/sola/internal/store"
	"gorm.io/gorm"
)

type WelcomeButtonService struct{ store *store.Store }

func NewWelcomeButtonService(st *store.Store) *WelcomeButtonService {
	return &WelcomeButtonService{store: st}
}

func (s *WelcomeButtonService) List(ctx context.Context, chatID int64) ([]bot.WelcomeButton, error) {
	if s == nil || s.store == nil || s.store.DB == nil {
		return []bot.WelcomeButton{}, nil
	}
	var rows []model.WelcomeButton
	if err := s.store.DB.WithContext(ctx).Where("chat_id = ?", chatID).Order("sort_order ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]bot.WelcomeButton, 0, len(rows))
	for _, row := range rows {
		items = append(items, welcomeButtonToBot(row))
	}
	return items, nil
}

func (s *WelcomeButtonService) Replace(ctx context.Context, chatID int64, inputs []bot.WelcomeButtonInput) ([]bot.WelcomeButton, error) {
	if chatID == 0 {
		return nil, errors.New("chat_id is required")
	}
	seen := make(map[string]struct{}, len(inputs))
	dailyLotteryCount := 0
	enabledCount := 0
	for i, item := range inputs {
		label := strings.TrimSpace(item.Label)
		if label == "" || len([]rune(label)) > 64 {
			return nil, errors.New("按钮名称不能为空且不能超过 64 个字符")
		}
		if _, ok := seen[label]; ok {
			return nil, errors.New("按钮名称不能重复")
		}
		seen[label] = struct{}{}
		if item.Enabled {
			enabledCount++
			if enabledCount > 20 {
				return nil, errors.New("启用的欢迎按钮最多 20 个")
			}
		}
		typ := strings.TrimSpace(item.ActionType)
		value := strings.TrimSpace(item.ActionValue)
		switch typ {
		case model.WelcomeButtonDailyLottery:
			if item.Enabled {
				dailyLotteryCount++
			}
			if dailyLotteryCount > 1 {
				return nil, errors.New("每个群组最多启用一个每日抽奖按钮")
			}
			value = ""
		case model.WelcomeButtonLink:
			u, err := url.Parse(value)
			if err != nil || u.Scheme != "https" || u.Host == "" {
				return nil, errors.New("链接按钮必须使用有效的 https 地址")
			}
		default:
			return nil, errors.New("无效的欢迎按钮类型")
		}
		if item.SortOrder < 0 {
			inputs[i].SortOrder = i
		}
		inputs[i].Label, inputs[i].ActionType, inputs[i].ActionValue = label, typ, value
	}
	if s == nil || s.store == nil || s.store.DB == nil {
		return inputsToWelcomeButtons(chatID, inputs), nil
	}
	err := s.store.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("chat_id = ?", chatID).Delete(&model.WelcomeButton{}).Error; err != nil {
			return err
		}
		for _, item := range inputs {
			row := model.WelcomeButton{ChatID: chatID, Label: item.Label, ActionType: item.ActionType, ActionValue: item.ActionValue, Enabled: item.Enabled, SortOrder: item.SortOrder}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.List(ctx, chatID)
}

func welcomeButtonToBot(row model.WelcomeButton) bot.WelcomeButton {
	return bot.WelcomeButton{ID: row.ID, ChatID: row.ChatID, Label: row.Label, ActionType: row.ActionType, ActionValue: row.ActionValue, Enabled: row.Enabled, SortOrder: row.SortOrder}
}

func inputsToWelcomeButtons(chatID int64, inputs []bot.WelcomeButtonInput) []bot.WelcomeButton {
	items := make([]bot.WelcomeButton, 0, len(inputs))
	for i, item := range inputs {
		order := item.SortOrder
		if order < 0 {
			order = i
		}
		items = append(items, bot.WelcomeButton{ChatID: chatID, Label: item.Label, ActionType: item.ActionType, ActionValue: item.ActionValue, Enabled: item.Enabled, SortOrder: order})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].SortOrder < items[j].SortOrder })
	return items
}
