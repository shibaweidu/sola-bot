package bot

import (
	"strings"
	"testing"
)

func TestFormatReferralSuccessText(t *testing.T) {
	got := formatReferralSuccessText("{name}|{group}|{invitee_points}|{points}", "小明", "官方群", 20, 35)
	if got != "小明|官方群|20|35" {
		t.Fatalf("formatted referral success text = %q", got)
	}
}

func TestFormatReferralSuccessTextDefault(t *testing.T) {
	got := formatReferralSuccessText("", "小明", "官方群", 20, 35)
	if got == "" || !containsAll(got, "+20", "35", "积分可兑换额度") {
		t.Fatalf("default referral success text = %q", got)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
