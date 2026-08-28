package service

import (
	"context"
	"testing"

	"github.com/dabowin/sola/internal/bot"
	"gorm.io/gorm"
)

func TestWelcomeButtonValidation(t *testing.T) {
	service := &WelcomeButtonService{}
	ctx := context.Background()
	_, err := service.Replace(ctx, 1001, []bot.WelcomeButtonInput{
		{Label: "抽奖", ActionType: "daily_lottery", Enabled: true},
		{Label: "活动", ActionType: "link", ActionValue: "http://example.com", Enabled: true},
	})
	if err == nil {
		t.Fatal("expected https validation error")
	}
	_, err = service.Replace(ctx, 1001, []bot.WelcomeButtonInput{
		{Label: "抽奖1", ActionType: "daily_lottery", Enabled: true},
		{Label: "抽奖2", ActionType: "daily_lottery", Enabled: true},
	})
	if err == nil {
		t.Fatal("expected duplicate daily lottery error")
	}
}

func TestWelcomeButtonReplace(t *testing.T) {
	st := newServiceTestStore(t)
	createWelcomeButtonTables(t, st.DB)
	service := NewWelcomeButtonService(st)
	items, err := service.Replace(context.Background(), 1001, []bot.WelcomeButtonInput{{Label: "抽奖", ActionType: "daily_lottery", Enabled: true, SortOrder: 0}, {Label: "小铺", ActionType: "link", ActionValue: "https://example.com", Enabled: true, SortOrder: 1}})
	if err != nil || len(items) != 2 || items[0].ActionType != "daily_lottery" {
		t.Fatalf("items = %+v, err=%v", items, err)
	}
}

func createWelcomeButtonTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	execSQL(t, db, `CREATE TABLE welcome_buttons (id integer PRIMARY KEY AUTOINCREMENT, chat_id integer NOT NULL, label text NOT NULL, action_type text NOT NULL, action_value text NOT NULL DEFAULT '', enabled boolean NOT NULL DEFAULT true, sort_order integer NOT NULL DEFAULT 0, created_at datetime, updated_at datetime)`, `CREATE UNIQUE INDEX idx_welcome_buttons_chat_label ON welcome_buttons(chat_id, label)`)
}
