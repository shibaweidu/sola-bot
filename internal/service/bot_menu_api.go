package service

import (
	"context"

	"github.com/dabowin/sola/internal/api"
	"github.com/dabowin/sola/internal/model"
)

type botMenuAPIService struct {
	service *BotMenuService
}

func (s *botMenuAPIService) List(ctx context.Context, role string) ([]api.BotMenuItem, error) {
	items, err := s.service.List(ctx, role)
	if err != nil {
		return nil, err
	}
	return botMenuItemsToAPI(items), nil
}

func (s *botMenuAPIService) Replace(ctx context.Context, role string, inputs []api.BotMenuItemInput) ([]api.BotMenuItem, error) {
	items := make([]model.BotMenuItem, 0, len(inputs))
	for _, input := range inputs {
		links := make([]model.BotMenuInlineLink, 0, len(input.LinkButtons))
		for _, link := range input.LinkButtons {
			links = append(links, model.BotMenuInlineLink{Label: link.Label, URL: link.URL, RowIndex: link.RowIndex, ColumnIndex: link.ColumnIndex})
		}
		linksJSON, err := model.EncodeBotMenuInlineLinks(links)
		if err != nil {
			return nil, err
		}
		items = append(items, model.BotMenuItem{
			Role:            role,
			Label:           input.Label,
			Icon:            input.Icon,
			ActionType:      input.ActionType,
			ActionKey:       input.ActionKey,
			ActionValue:     input.ActionValue,
			MessageText:     input.MessageText,
			LinkMode:        input.LinkMode,
			LinkButtonsJSON: linksJSON,
			RowIndex:        input.RowIndex,
			ColumnIndex:     input.ColumnIndex,
			Enabled:         input.Enabled,
		})
	}
	items, err := s.service.Replace(ctx, role, items)
	if err != nil {
		return nil, err
	}
	return botMenuItemsToAPI(items), nil
}

func (s *botMenuAPIService) Reset(ctx context.Context, role string) ([]api.BotMenuItem, error) {
	items, err := s.service.Reset(ctx, role)
	if err != nil {
		return nil, err
	}
	return botMenuItemsToAPI(items), nil
}

func botMenuItemsToAPI(items []model.BotMenuItem) []api.BotMenuItem {
	result := make([]api.BotMenuItem, 0, len(items))
	for _, item := range items {
		links, _ := model.ParseBotMenuInlineLinks(item.LinkButtonsJSON)
		apiLinks := make([]api.BotMenuInlineLink, 0, len(links))
		for _, link := range links {
			apiLinks = append(apiLinks, api.BotMenuInlineLink{Label: link.Label, URL: link.URL, RowIndex: link.RowIndex, ColumnIndex: link.ColumnIndex})
		}
		result = append(result, api.BotMenuItem{
			ID:            item.ID.String(),
			TelegramBotID: item.TelegramBotID,
			Role:          item.Role,
			Label:         item.Label,
			Icon:          item.Icon,
			ActionType:    item.ActionType,
			ActionKey:     item.ActionKey,
			ActionValue:   item.ActionValue,
			MessageText:   item.MessageText,
			LinkMode:      model.NormalizeBotMenuLinkMode(item.LinkMode),
			LinkButtons:   apiLinks,
			RowIndex:      item.RowIndex,
			ColumnIndex:   item.ColumnIndex,
			Enabled:       item.Enabled,
		})
	}
	return result
}
