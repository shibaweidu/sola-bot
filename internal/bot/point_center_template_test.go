package bot

import (
	"strings"
	"testing"
)

func TestFormatInvitePageTemplate(t *testing.T) {
	data := invitePageTemplateData{
		InviteLink: "https://t.me/testbot?start=ref_abc", Group: "测试群",
		InviterReward: 100, InviteeReward: 20, ExchangeMinimum: 10,
		ExchangeRate: 1, InviteJoinURL: "https://t.me/test_group",
	}
	text := formatInvitePageTemplate("{group}|{invite_link}|{inviter_reward}|{invitee_reward}|{exchange_minimum}|{exchange_rate}|{invite_join_url}", data)
	want := "测试群|https://t.me/testbot?start=ref_abc|100|20|10|1|https://t.me/test_group"
	if text != want {
		t.Fatalf("formatted template = %q, want %q", text, want)
	}

	defaultText := formatInvitePageTemplate("", data)
	for _, value := range []string{"https://t.me/testbot?start=ref_abc", "测试群", "100", "20", "10"} {
		if !strings.Contains(defaultText, value) {
			t.Fatalf("default template does not contain %q: %q", value, defaultText)
		}
	}
}

func TestFormatInvitePageTemplateDoesNotAppendHardcodedSteps(t *testing.T) {
	text := formatInvitePageTemplate("活动说明\n{group}", invitePageTemplateData{Group: "测试群"})
	if text != "活动说明\n测试群" {
		t.Fatalf("custom template was changed: %q", text)
	}
}
