package bot

import "testing"

func TestParseScheduledPostActionDeepLink(t *testing.T) {
	action, chatID, ok := parseScheduledPostActionDeepLink("pa_i_-100123")
	if !ok || action != "invite_rewards" || chatID != -100123 {
		t.Fatalf("parse result = %q, %d, %v", action, chatID, ok)
	}
	for _, raw := range []string{"", "pa_x_-100123", "pa_s_bad", "pa_s_0"} {
		if _, _, ok := parseScheduledPostActionDeepLink(raw); ok {
			t.Fatalf("%q should be rejected", raw)
		}
	}
}
