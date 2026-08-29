package bot

import (
	"strings"
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/dabowin/sola/internal/api"
)

func TestDailyLotteryHasAvailablePrize(t *testing.T) {
	status := DailyLotteryStatus{Enabled: true, Remaining: 3, Prizes: []DailyLotteryPrize{{Amount: 10, Weight: 100, Enabled: true, AvailableCode: 0}}}
	if dailyLotteryHasAvailablePrize(status) {
		t.Fatal("empty inventory should not enable draw")
	}
	status.Prizes[0].AvailableCode = 1
	if !dailyLotteryHasAvailablePrize(status) {
		t.Fatal("available inventory should enable draw")
	}
}

func TestDailyLotteryMarkupKeepsDrawButtonAfterWin(t *testing.T) {
	status := DailyLotteryStatus{Enabled: true, Remaining: 1, PaidEnabled: true, PaidCostPoints: 2, Prizes: []DailyLotteryPrize{{Amount: 1, Weight: 100, Enabled: true, AvailableCode: 3}}}
	markup := dailyLotteryMarkup(1001, status)
	if markup == nil || markup.ReplyMarkup == nil {
		t.Fatal("expected lottery markup")
	}
	keyboard := markup.ReplyMarkup.(gotgbot.InlineKeyboardMarkup).InlineKeyboard
	if len(keyboard) == 0 || keyboard[0][0].Text != "🎲 免费抽一次" {
		t.Fatalf("expected free draw button, got %+v", keyboard)
	}
	status.Remaining = 0
	markup = dailyLotteryMarkup(1001, status)
	keyboard = markup.ReplyMarkup.(gotgbot.InlineKeyboardMarkup).InlineKeyboard
	if len(keyboard) == 0 || keyboard[0][0].Text != "💎 2 积分再抽一次" {
		t.Fatalf("expected paid draw button, got %+v", keyboard)
	}
}

func TestDailyLotteryUserResultHidesPrizeDetails(t *testing.T) {
	text := formatDailyLotteryResult(DailyLotteryDrawResult{Result: "won", Amount: 100, Code: "SECRET", RedeemURL: "https://example.com/redeem", Remaining: 2})
	if !strings.Contains(text, "兑换码已发放") || !strings.Contains(text, "兑换地址：https://example.com/redeem") {
		t.Fatalf("result text missing user-facing guidance: %s", text)
	}
	if strings.Contains(text, "100") || strings.Contains(text, "额度") || strings.Contains(text, "SECRET") {
		t.Fatalf("result text exposes prize details or code: %s", text)
	}
}

func TestLotteryJoinButtonVisibilityByJoinType(t *testing.T) {
	cases := []struct {
		name     string
		joinType string
		want     bool
	}{
		{name: "button lottery", joinType: "button", want: true},
		{name: "keyword lottery", joinType: "keyword", want: false},
		{name: "button and keyword lottery", joinType: "both", want: true},
		{name: "legacy lottery", joinType: "", want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := lotteryHasJoinButton(tc.joinType); got != tc.want {
				t.Fatalf("lotteryHasJoinButton(%q) = %v, want %v", tc.joinType, got, tc.want)
			}
		})
	}
}

func TestLotteryAnnouncementTextSeparatesButtonAndKeyword(t *testing.T) {
	buttonText := lotteryAnnouncementText(api.Lottery{
		ID:          1,
		Title:       "button draw",
		Prize:       "coupon",
		WinnerCount: 1,
		JoinType:    "button",
	})
	if !strings.Contains(buttonText, "按钮抽奖活动 #1") {
		t.Fatalf("button announcement missing button label: %s", buttonText)
	}
	if !strings.Contains(buttonText, "点击下方按钮参与抽奖。") {
		t.Fatalf("button announcement missing button instruction: %s", buttonText)
	}
	if strings.Contains(buttonText, "口令：") || strings.Contains(buttonText, "发送口令") {
		t.Fatalf("button announcement should not mention keyword: %s", buttonText)
	}

	keywordText := lotteryAnnouncementText(api.Lottery{
		ID:          2,
		Title:       "keyword draw",
		Prize:       "coupon",
		WinnerCount: 1,
		JoinType:    "keyword",
		JoinKeyword: "888",
	})
	if !strings.Contains(keywordText, "口令抽奖活动 #2") {
		t.Fatalf("keyword announcement missing keyword label: %s", keywordText)
	}
	if !strings.Contains(keywordText, "口令：888") || !strings.Contains(keywordText, "发送口令「888」参与抽奖。") {
		t.Fatalf("keyword announcement missing keyword instruction: %s", keywordText)
	}
	if strings.Contains(keywordText, "点击下方按钮参与抽奖。") {
		t.Fatalf("keyword announcement should not use button-only instruction: %s", keywordText)
	}
}
