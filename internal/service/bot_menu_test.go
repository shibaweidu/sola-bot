package service

import (
	"context"
	"testing"

	"github.com/dabowin/sola/internal/model"
)

func TestBotMenuDefaults(t *testing.T) {
	service := NewBotMenuService(nil, "12345:test")
	member, err := service.List(context.Background(), model.BotMenuRoleMember)
	if err != nil {
		t.Fatal(err)
	}
	if len(member) != 7 || member[0].ActionKey != "invite_rewards" || member[6].ActionKey != "rank" {
		t.Fatalf("unexpected member defaults: %#v", member)
	}
	admin, err := service.List(context.Background(), model.BotMenuRoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if len(admin) != 11 || admin[7].ActionKey != "private_console" {
		t.Fatalf("unexpected admin defaults: %#v", admin)
	}
}

func TestValidateBotMenuItems(t *testing.T) {
	base := model.BotMenuItem{Role: model.BotMenuRoleMember, Label: "帮助", ActionType: model.BotMenuActionBuiltin, ActionKey: "help", Enabled: true}
	if err := validateBotMenuItems(model.BotMenuRoleMember, []model.BotMenuItem{base, base}); err == nil {
		t.Fatal("expected duplicate label/position error")
	}
	badLink := base
	badLink.Label = "链接"
	badLink.ActionType = model.BotMenuActionLink
	badLink.ActionKey = ""
	badLink.ActionValue = "http://example.com"
	if err := validateBotMenuItems(model.BotMenuRoleMember, []model.BotMenuItem{badLink}); err == nil {
		t.Fatal("expected invalid URL error")
	}
	tooMany := make([]model.BotMenuItem, 21)
	for i := range tooMany {
		tooMany[i] = model.BotMenuItem{Role: model.BotMenuRoleMember, Label: string(rune('a' + i)), ActionType: model.BotMenuActionBuiltin, ActionKey: "help", RowIndex: i / 4, ColumnIndex: i % 4, Enabled: true}
	}
	if err := validateBotMenuItems(model.BotMenuRoleMember, tooMany); err == nil {
		t.Fatal("expected enabled item limit error")
	}
	custom := model.BotMenuItem{Role: model.BotMenuRoleMember, Label: "活动", ActionType: model.BotMenuActionCustom, MessageText: "当前积分：{points}", ActionValue: "https://example.com", Enabled: true}
	if err := validateBotMenuItems(model.BotMenuRoleMember, []model.BotMenuItem{custom}); err != nil {
		t.Fatalf("custom menu should be valid: %v", err)
	}
	custom.MessageText = ""
	if err := validateBotMenuItems(model.BotMenuRoleMember, []model.BotMenuItem{custom}); err == nil {
		t.Fatal("expected custom message required error")
	}
	custom.MessageText = "活动"
	custom.ActionValue = "http://example.com"
	if err := validateBotMenuItems(model.BotMenuRoleMember, []model.BotMenuItem{custom}); err == nil {
		t.Fatal("expected custom URL validation error")
	}
	linksJSON, err := model.EncodeBotMenuInlineLinks([]model.BotMenuInlineLink{
		{Label: "官网", URL: "https://example.com", RowIndex: 0, ColumnIndex: 0},
		{Label: "购买", URL: "https://example.com/buy", RowIndex: 1, ColumnIndex: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	custom.ActionValue = ""
	custom.LinkMode = model.BotMenuLinkModeButtons
	custom.LinkButtonsJSON = linksJSON
	if err := validateBotMenuItems(model.BotMenuRoleMember, []model.BotMenuItem{custom}); err != nil {
		t.Fatalf("inline links should be valid: %v", err)
	}
	custom.LinkButtonsJSON = `[{"label":"官网","url":"http://example.com","row_index":0,"column_index":0}]`
	if err := validateBotMenuItems(model.BotMenuRoleMember, []model.BotMenuItem{custom}); err == nil {
		t.Fatal("expected inline link URL validation error")
	}
}
