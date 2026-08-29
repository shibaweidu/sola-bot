package service

import (
	"context"
	"testing"
	"time"

	"github.com/dabowin/sola/internal/bot"
	"github.com/dabowin/sola/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestDailyLotteryInventoryImportAndSummary(t *testing.T) {
	st := newServiceTestStore(t)
	createDailyLotteryInventoryTables(t, st.DB)
	svc := NewDailyLotteryService(st)
	ctx := context.Background()
	result, err := svc.ImportCodes(ctx, bot.DailyLotteryCodeImportRequest{ChatID: 1001, Amount: 10, BatchName: "首批", Codes: []string{"A", "B", "A"}})
	if err != nil || result.Imported != 2 || result.Skipped != 1 {
		t.Fatalf("first import = %+v, err=%v", result, err)
	}
	result, err = svc.ImportCodes(ctx, bot.DailyLotteryCodeImportRequest{ChatID: 1001, Amount: 10, Codes: []string{"A", "C"}})
	if err != nil || result.Imported != 1 || result.Skipped != 1 {
		t.Fatalf("second import = %+v, err=%v", result, err)
	}
	if err := st.DB.Create(&model.DailyLotteryCode{ChatID: 1001, Code: "D", Amount: 50, Status: "assigned"}).Error; err != nil {
		t.Fatal(err)
	}
	summaries, err := svc.SummarizeCodes(ctx, 1001)
	if err != nil || len(summaries) != 2 {
		t.Fatalf("summaries = %+v, err=%v", summaries, err)
	}
	if summaries[0].Amount != 10 || summaries[0].Available != 3 || summaries[0].Total != 3 {
		t.Fatalf("10 summary = %+v", summaries[0])
	}
	if summaries[1].Amount != 50 || summaries[1].Assigned != 1 {
		t.Fatalf("50 summary = %+v", summaries[1])
	}
	items, err := svc.ListCodes(ctx, 1001, func() *int { value := 10; return &value }(), "available", 20)
	if err != nil || len(items) != 3 {
		t.Fatalf("filtered items = %+v, err=%v", items, err)
	}
}

func TestDailyLotteryInventoryBatchOperations(t *testing.T) {
	st := newServiceTestStore(t)
	createDailyLotteryInventoryTables(t, st.DB)
	now := time.Now()
	available := model.DailyLotteryCode{ID: uuid.New(), ChatID: 1001, Code: "DAILY-A", Amount: 10, Status: "available", CreatedAt: now, UpdatedAt: now}
	assigned := model.DailyLotteryCode{ID: uuid.New(), ChatID: 1001, Code: "DAILY-B", Amount: 10, Status: "assigned", CreatedAt: now, UpdatedAt: now}
	if err := st.DB.Create(&available).Error; err != nil {
		t.Fatal(err)
	}
	if err := st.DB.Create(&assigned).Error; err != nil {
		t.Fatal(err)
	}
	svc := NewDailyLotteryService(st)
	url := "https://new.example/redeem"
	result, err := svc.BatchUpdateCodes(context.Background(), bot.DailyLotteryCodeBatchUpdateRequest{ChatID: 1001, IDs: []string{available.ID.String(), assigned.ID.String()}, RedeemURL: &url})
	if err != nil || result.Updated != 1 || result.Skipped != 1 {
		t.Fatalf("batch update = %+v, err=%v", result, err)
	}
	result, err = svc.BatchDeleteCodes(context.Background(), bot.DailyLotteryCodeBatchDeleteRequest{ChatID: 1001, IDs: []string{available.ID.String(), assigned.ID.String()}})
	if err != nil || result.Deleted != 1 || result.Skipped != 1 {
		t.Fatalf("batch delete = %+v, err=%v", result, err)
	}
}

func createDailyLotteryInventoryTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	execSQL(t, db,
		`CREATE TABLE daily_lottery_codes (id text PRIMARY KEY, chat_id integer NOT NULL, code text NOT NULL, amount integer NOT NULL, batch_name text NOT NULL DEFAULT '', redeem_url text NOT NULL DEFAULT '', expires_at datetime, status text NOT NULL DEFAULT 'available', assigned_user integer NOT NULL DEFAULT 0, assigned_at datetime, created_at datetime, updated_at datetime)`,
		`CREATE UNIQUE INDEX idx_daily_lottery_codes_chat_code ON daily_lottery_codes(chat_id, code)`,
	)
}
