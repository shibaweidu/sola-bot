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

func TestDailyLotteryDrawAndDailyLimit(t *testing.T) {
	ctx := context.Background()
	st := newServiceTestStore(t)
	createPointTables(t, st.DB)
	createPointCenterTables(t, st.DB)
	createDailyLotteryTables(t, st.DB)
	svc := NewDailyLotteryService(st)

	if _, err := svc.UpdateConfig(ctx, bot.DailyLotteryConfig{ChatID: 1001, Enabled: true, CostPoints: 2, GuaranteeOnLast: true}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	code := model.DailyLotteryCode{ID: uuid.New(), ChatID: 1001, Code: "DAILY-10", Amount: 10, Status: "available", CreatedAt: now, UpdatedAt: now}
	if err := st.DB.Create(&code).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReplacePrizes(ctx, 1001, []bot.DailyLotteryPrize{{Amount: 10, Weight: 1000, Enabled: true}}); err != nil {
		t.Fatal(err)
	}
	if _, err := NewPointsService(st).AdjustUserPoints(ctx, 1001, 2001, 10, "seed"); err != nil {
		t.Fatal(err)
	}
	first, err := svc.Draw(ctx, 1001, 2001)
	if err != nil || first.Result != "won" || first.Code != "DAILY-10" || first.Remaining != 2 {
		t.Fatalf("first draw = %+v, err=%v", first, err)
	}
	if _, err := svc.Draw(ctx, 1001, 2001); err == nil {
		t.Fatal("expected empty inventory rejection without consuming attempt")
	}
	status, err := svc.Status(ctx, 1001, 2001)
	if err != nil || status.UsedAttempts != 1 || status.Remaining != 2 {
		t.Fatalf("status = %+v, err=%v", status, err)
	}
	date := time.Now().In(chinaLocation()).Format("2006-01-02")
	for i := 1; i <= 3; i++ {
		if err := st.DB.Create(&model.DailyLotteryAttempt{ChatID: 1001, UserID: 2003, DrawDate: date, AttemptNo: i, Result: "lost"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.Draw(ctx, 1001, 2003); err == nil {
		t.Fatal("expected daily draw limit rejection")
	}
}

func TestDailyLotteryLastAttemptGuarantee(t *testing.T) {
	ctx := context.Background()
	st := newServiceTestStore(t)
	createPointTables(t, st.DB)
	createPointCenterTables(t, st.DB)
	createDailyLotteryTables(t, st.DB)
	svc := NewDailyLotteryService(st)
	if _, err := svc.UpdateConfig(ctx, bot.DailyLotteryConfig{ChatID: 1001, Enabled: true, GuaranteeOnLast: true}); err != nil {
		t.Fatal(err)
	}
	date := time.Now().In(chinaLocation()).Format("2006-01-02")
	for i := 1; i <= 2; i++ {
		if err := st.DB.Create(&model.DailyLotteryAttempt{ChatID: 1001, UserID: 2002, DrawDate: date, AttemptNo: i, Result: "lost"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	if err := st.DB.Create(&model.DailyLotteryCode{ID: uuid.New(), ChatID: 1001, Code: "DAILY-50", Amount: 50, Status: "available", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReplacePrizes(ctx, 1001, []bot.DailyLotteryPrize{{Amount: 50, Weight: 1, Enabled: true}}); err != nil {
		t.Fatal(err)
	}
	result, err := svc.Draw(ctx, 1001, 2002)
	if err != nil || result.Result != "won" || !result.Guaranteed || result.Amount != 50 {
		t.Fatalf("guaranteed draw = %+v, err=%v", result, err)
	}
}

func TestDailyLotteryPaidDrawAndReset(t *testing.T) {
	ctx := context.Background()
	st := newServiceTestStore(t)
	createPointTables(t, st.DB)
	createPointCenterTables(t, st.DB)
	createDailyLotteryTables(t, st.DB)
	svc := NewDailyLotteryService(st)
	if _, err := svc.UpdateConfig(ctx, bot.DailyLotteryConfig{ChatID: 1001, Enabled: true, PaidEnabled: true, CostPoints: 2}); err != nil {
		t.Fatal(err)
	}
	date := time.Now().In(chinaLocation()).Format("2006-01-02")
	for i := 1; i <= 3; i++ {
		if err := st.DB.Create(&model.DailyLotteryAttempt{ChatID: 1001, UserID: 2004, DrawDate: date, AttemptNo: i, Result: "lost"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	if err := st.DB.Create(&model.DailyLotteryCode{ID: uuid.New(), ChatID: 1001, Code: "PAID-10", Amount: 10, Status: "available", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReplacePrizes(ctx, 1001, []bot.DailyLotteryPrize{{Amount: 10, Weight: 1000, Enabled: true}}); err != nil {
		t.Fatal(err)
	}
	if _, err := NewPointsService(st).AdjustUserPoints(ctx, 1001, 2004, 5, "seed"); err != nil {
		t.Fatal(err)
	}
	result, err := svc.Draw(ctx, 1001, 2004)
	if err != nil || result.Result != "won" || result.CostPoints != 2 || result.Remaining != 0 {
		t.Fatalf("paid draw = %+v, err=%v", result, err)
	}
	point, err := NewPointsService(st).GetUserPoint(ctx, 1001, 2004)
	if err != nil || point.TotalPoints != 3 {
		t.Fatalf("paid points = %+v, err=%v", point, err)
	}
	if err := svc.ResetAttempts(ctx, 1001, 2004); err != nil {
		t.Fatal(err)
	}
	status, err := svc.Status(ctx, 1001, 2004)
	if err != nil || status.UsedAttempts != 0 || status.Remaining != 3 {
		t.Fatalf("reset status = %+v, err=%v", status, err)
	}
}

func TestDailyLotterySyncsNewInventoryAmounts(t *testing.T) {
	ctx := context.Background()
	st := newServiceTestStore(t)
	createPointTables(t, st.DB)
	createPointCenterTables(t, st.DB)
	createDailyLotteryTables(t, st.DB)
	svc := NewDailyLotteryService(st)
	now := time.Now()
	if err := st.DB.Create(&model.DailyLotteryCode{ID: uuid.New(), ChatID: 1001, Code: "NEW-1", Amount: 1, Status: "available", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	prizes, err := svc.ListPrizes(ctx, 1001)
	if err != nil || len(prizes) != 1 || prizes[0].Amount != 1 || prizes[0].AvailableCode != 1 {
		t.Fatalf("synced prizes = %+v, err=%v", prizes, err)
	}
	// A depleted/configured amount remains editable and does not block saving
	// another amount that currently has inventory.
	if _, err := svc.ReplacePrizes(ctx, 1001, []bot.DailyLotteryPrize{{Amount: 10, Weight: 200, Enabled: true}, {Amount: 1, Weight: 800, Enabled: true}}); err != nil {
		t.Fatalf("replace prizes with depleted amount: %v", err)
	}
}

func createDailyLotteryTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	execSQL(t, db,
		`CREATE TABLE daily_lottery_configs (chat_id integer PRIMARY KEY, enabled boolean NOT NULL DEFAULT false, daily_attempts integer NOT NULL DEFAULT 3, cost_points integer NOT NULL DEFAULT 0, paid_enabled boolean NOT NULL DEFAULT false, guarantee_on_last boolean NOT NULL DEFAULT true, created_at datetime, updated_at datetime)`,
		`CREATE TABLE daily_lottery_prizes (id integer PRIMARY KEY AUTOINCREMENT, chat_id integer NOT NULL, amount integer NOT NULL, weight integer NOT NULL DEFAULT 1, enabled boolean NOT NULL DEFAULT true, created_at datetime, updated_at datetime)`,
		`CREATE UNIQUE INDEX idx_daily_lottery_prize_chat_amount ON daily_lottery_prizes(chat_id, amount)`,
		`CREATE TABLE daily_lottery_attempts (id integer PRIMARY KEY AUTOINCREMENT, chat_id integer NOT NULL, user_id integer NOT NULL, draw_date text NOT NULL, attempt_no integer NOT NULL, result text NOT NULL, amount integer NOT NULL DEFAULT 0, code_id text, cost_points integer NOT NULL DEFAULT 0, created_at datetime)`,
		`CREATE UNIQUE INDEX idx_daily_lottery_attempt ON daily_lottery_attempts(chat_id, user_id, draw_date, attempt_no)`,
		`CREATE TABLE daily_lottery_codes (id text PRIMARY KEY, chat_id integer NOT NULL, code text NOT NULL, amount integer NOT NULL, batch_name text NOT NULL DEFAULT '', redeem_url text NOT NULL DEFAULT '', expires_at datetime, status text NOT NULL DEFAULT 'available', assigned_user integer NOT NULL DEFAULT 0, assigned_at datetime, created_at datetime, updated_at datetime)`,
		`CREATE UNIQUE INDEX idx_daily_lottery_codes_chat_code ON daily_lottery_codes(chat_id, code)`,
	)
}
