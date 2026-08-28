package service

import (
	"context"

	"github.com/dabowin/sola/internal/api"
	"github.com/dabowin/sola/internal/bot"
)

type welcomeButtonAPIService struct{ service *WelcomeButtonService }

func (s *welcomeButtonAPIService) List(ctx context.Context, chatID int64) ([]api.WelcomeButton, error) {
	items, err := s.service.List(ctx, chatID)
	if err != nil {
		return nil, err
	}
	result := make([]api.WelcomeButton, 0, len(items))
	for _, item := range items {
		result = append(result, api.WelcomeButton{ID: item.ID, ChatID: item.ChatID, Label: item.Label, ActionType: item.ActionType, ActionValue: item.ActionValue, Enabled: item.Enabled, SortOrder: item.SortOrder})
	}
	return result, nil
}

func (s *welcomeButtonAPIService) Replace(ctx context.Context, chatID int64, inputs []api.WelcomeButtonInput) ([]api.WelcomeButton, error) {
	items := make([]bot.WelcomeButtonInput, 0, len(inputs))
	for _, item := range inputs {
		items = append(items, bot.WelcomeButtonInput{Label: item.Label, ActionType: item.ActionType, ActionValue: item.ActionValue, Enabled: item.Enabled, SortOrder: item.SortOrder})
	}
	result, err := s.service.Replace(ctx, chatID, items)
	if err != nil {
		return nil, err
	}
	itemsOut := make([]api.WelcomeButton, 0, len(result))
	for _, item := range result {
		itemsOut = append(itemsOut, api.WelcomeButton{ID: item.ID, ChatID: item.ChatID, Label: item.Label, ActionType: item.ActionType, ActionValue: item.ActionValue, Enabled: item.Enabled, SortOrder: item.SortOrder})
	}
	return itemsOut, nil
}
