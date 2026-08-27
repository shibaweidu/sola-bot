package bot

import (
	"strings"
	"testing"
)

func TestRenderPrivateMenuText(t *testing.T) {
	text := renderPrivateMenuText(
		"你好 {name}\n群组：{group}\n积分：{points}/{today_invite}/{successful_invites}/{exchanged_amount}",
		"小明", "测试群", PointCenterDashboard{CurrentPoints: 20, TodayInvitePoints: 3, SuccessfulInvites: 5, ExchangedAmount: 10},
	)
	for _, part := range []string{"小明", "测试群", "20", "3", "5", "10"} {
		if !strings.Contains(text, part) {
			t.Fatalf("rendered text %q does not contain %q", text, part)
		}
	}
}

func TestAppendOptionalLink(t *testing.T) {
	if got := appendOptionalLink("说明", "链接", ""); got != "说明" {
		t.Fatalf("empty URL result = %q", got)
	}
	if got := appendOptionalLink("说明", "链接", "https://example.com"); !strings.HasSuffix(got, "链接：https://example.com") {
		t.Fatalf("URL result = %q", got)
	}
}

func TestCustomMenuInlineMarkup(t *testing.T) {
	markup, err := customMenuInlineMarkup(`[{"label":"第二个","url":"https://example.com/2","row_index":1,"column_index":0},{"label":"第一个","url":"https://example.com/1","row_index":0,"column_index":0},{"label":"并排","url":"https://example.com/side","row_index":0,"column_index":1}]`)
	if err != nil {
		t.Fatal(err)
	}
	if len(markup.InlineKeyboard) != 2 || len(markup.InlineKeyboard[0]) != 2 || markup.InlineKeyboard[0][0].Text != "第一个" || markup.InlineKeyboard[1][0].Url != "https://example.com/2" {
		t.Fatalf("unexpected inline markup: %#v", markup.InlineKeyboard)
	}
}

func TestFormatExchangeResultOptionalURL(t *testing.T) {
	withoutURL := formatExchangeResult(ExchangeResult{
		Amount:       10,
		Points:       10,
		Code:         "ABC-123",
		Instructions: "打开兑换页面使用兑换码。",
	})
	if strings.Contains(withoutURL, "ABC-123") {
		t.Fatalf("exchange details should not contain the redemption code: %q", withoutURL)
	}
	if strings.Contains(withoutURL, "兑换地址：") {
		t.Fatalf("empty redeem URL should not render an address line: %q", withoutURL)
	}

	withURL := formatExchangeResult(ExchangeResult{
		Amount:       10,
		Points:       10,
		Code:         "ABC-123",
		Instructions: "打开兑换页面使用兑换码。",
		RedeemURL:    "https://example.com/redeem",
	})
	if !strings.HasSuffix(withURL, "兑换地址：https://example.com/redeem") {
		t.Fatalf("redeem URL should be appended to the message: %q", withURL)
	}
	code := formatExchangeCode(ExchangeResult{Code: "ABC-123"})
	if !strings.Contains(code, "ABC-123") || strings.Contains(code, "兑换额度") {
		t.Fatalf("exchange code message should be short and copyable: %q", code)
	}
}
