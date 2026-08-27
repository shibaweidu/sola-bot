package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/dabowin/sola/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestPointCenterSignExchangeAndReferral(t *testing.T) {
	ctx := context.Background()
	st := newServiceTestStore(t)
	createPointTables(t, st.DB)
	createPointCenterTables(t, st.DB)
	svc := NewPointCenterService(st, "12345:test")

	cfg, err := svc.GetConfig(ctx, 1001)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ExchangeURL = "https://example.com/redeem"
	cfg.InviteJoinURL = "https://t.me/example_group"
	cfg.InviteJoinText = "请先加入官方群并完成验证。"
	cfg.InviteSuccessText = "成功加入 {group}，+{invitee_points}，当前 {points}"
	if _, err := svc.UpdateConfig(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := svc.GetConfig(ctx, 1001)
	if err != nil || loaded.InviteJoinURL != cfg.InviteJoinURL || loaded.InviteJoinText != cfg.InviteJoinText || loaded.InviteSuccessText != cfg.InviteSuccessText {
		t.Fatalf("invite config = %+v, err=%v", loaded, err)
	}

	sign, err := svc.Sign(ctx, 1001, 2001)
	if err != nil || !sign.Signed || sign.Reward != 1 {
		t.Fatalf("sign = %+v, err=%v", sign, err)
	}
	if _, err := svc.Sign(ctx, 1001, 2001); err == nil {
		t.Fatal("expected duplicate daily sign rejection")
	}
	if _, err := NewPointsService(st).AdjustUserPoints(ctx, 1001, 2001, 10, "seed"); err != nil {
		t.Fatal(err)
	}
	code := model.ExchangeCode{BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()}, Code: "TEN-CODE", Amount: 10, Status: "available"}
	if err := st.DB.Create(&code).Error; err != nil {
		t.Fatal(err)
	}
	exchange, err := svc.Exchange(ctx, 1001, 2001, 10)
	if err != nil {
		t.Fatal(err)
	}
	if exchange.Code != "TEN-CODE" || exchange.Amount != 10 || exchange.RedeemURL != cfg.ExchangeURL {
		t.Fatalf("exchange = %+v", exchange)
	}
	var updated model.ExchangeCode
	if err := st.DB.First(&updated, "id = ?", code.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.Status != "assigned" || updated.AssignedUser != 2001 {
		t.Fatalf("updated code = %+v", updated)
	}

	link, err := svc.EnsureReferralLink(ctx, 12345, 1001, 3001, "testbot")
	if err != nil || !strings.Contains(link.Link, "start=ref_") {
		t.Fatalf("link = %+v, err=%v", link, err)
	}
	startCode := link.Link[strings.LastIndex(link.Link, "ref_"):]
	start, err := svc.RegisterReferralStart(ctx, 12345, startCode, 3002)
	if err != nil || !start.Accepted {
		t.Fatalf("referral start = %+v, err=%v", start, err)
	}
	reward, err := svc.ActivateReferral(ctx, 12345, 1001, 3002)
	if err != nil || !reward.Rewarded || reward.InviterReward != 100 || reward.InviteeReward != 20 {
		t.Fatalf("referral reward = %+v, err=%v", reward, err)
	}
	var inviter model.UserPoint
	if err := st.DB.First(&inviter, "chat_id = ? AND user_id = ?", 1001, 3001).Error; err != nil || inviter.TotalPoints != 100 {
		t.Fatalf("inviter points = %+v, err=%v", inviter, err)
	}
	var invitee model.UserPoint
	if err := st.DB.First(&invitee, "chat_id = ? AND user_id = ?", 1001, 3002).Error; err != nil || invitee.TotalPoints != 20 {
		t.Fatalf("invitee points = %+v, err=%v", invitee, err)
	}
	if second, err := svc.ActivateReferral(ctx, 12345, 1001, 3002); err != nil || second.Rewarded {
		t.Fatalf("duplicate reward = %+v, err=%v", second, err)
	}
	if err := svc.ResetReferralForTesting(ctx, 12345, 1001, 3002); err != nil {
		t.Fatalf("reset referral = %v", err)
	}
	var resetReferral model.Referral
	if err := st.DB.First(&resetReferral, "bot_id = ? AND chat_id = ? AND invitee_id = ?", 12345, 1001, 3002).Error; err != nil || resetReferral.Status != "started" {
		t.Fatalf("reset referral row = %+v, err=%v", resetReferral, err)
	}
	if err := st.DB.First(&invitee, "chat_id = ? AND user_id = ?", 1001, 3002).Error; err != nil || invitee.TotalPoints != 0 {
		t.Fatalf("reset invitee points = %+v, err=%v", invitee, err)
	}
	if err := st.DB.First(&inviter, "chat_id = ? AND user_id = ?", 1001, 3001).Error; err != nil || inviter.TotalPoints != 0 {
		t.Fatalf("reset inviter points = %+v, err=%v", inviter, err)
	}
	var rewardLogs int64
	if err := st.DB.Model(&model.PointLog{}).Where("chat_id = ? AND reason LIKE ?", 1001, "invite_%:%").Count(&rewardLogs).Error; err != nil || rewardLogs != 0 {
		t.Fatalf("reset reward logs = %d, err=%v", rewardLogs, err)
	}
}

func createPointCenterTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	execSQL(t, db,
		`CREATE TABLE point_center_configs (chat_id integer PRIMARY KEY, invite_enabled boolean, inviter_reward integer, invitee_reward integer, sign_enabled boolean, sign_reward integer, exchange_enabled boolean, exchange_minimum integer, exchange_rate integer, exchange_url text, exchange_instructions text, purchase_url text, purchase_text text, shop_url text, shop_text text, invite_text text, invite_page_template text, invite_join_url text, invite_join_text text, invite_success_text text, points_text text, sign_text text, exchange_text text, rank_text text, updated_at datetime)`,
		`CREATE TABLE daily_sign_records (id integer PRIMARY KEY AUTOINCREMENT, chat_id integer, user_id integer, sign_date text, reward integer, created_at datetime, UNIQUE(chat_id, user_id, sign_date))`,
		`CREATE TABLE exchange_codes (id text PRIMARY KEY, code text UNIQUE, batch_name text, amount integer, redeem_url text, expires_at datetime, status text, assigned_user integer, assigned_chat integer, assigned_at datetime, created_at datetime, updated_at datetime, deleted_at datetime)`,
		`CREATE TABLE exchange_orders (id text PRIMARY KEY, user_id integer, chat_id integer, points_spent integer, amount integer, code_id text, status text, redeem_url text, created_at datetime, updated_at datetime, deleted_at datetime)`,
		`CREATE TABLE referral_codes (id text PRIMARY KEY, code text UNIQUE, bot_id integer, chat_id integer, inviter_id integer, use_count integer, is_active boolean, created_at datetime, updated_at datetime, deleted_at datetime)`,
		`CREATE TABLE referrals (id text PRIMARY KEY, code_id text, bot_id integer, chat_id integer, inviter_id integer, invitee_id integer, status text, started_at datetime, activated_at datetime, rewarded_at datetime, created_at datetime, updated_at datetime, deleted_at datetime, UNIQUE(bot_id, invitee_id))`,
	)
}
