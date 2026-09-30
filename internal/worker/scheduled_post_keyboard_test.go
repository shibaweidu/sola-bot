package worker

import "testing"

func TestScheduledPostActionURL(t *testing.T) {
	tests := []struct {
		action string
		want   string
	}{
		{action: "sign", want: "https://t.me/KAOLAAIBOT?start=pa_s_-100123"},
		{action: "daily_lottery", want: "https://t.me/KAOLAAIBOT?start=dl_-100123"},
		{action: "invite_rewards", want: "https://t.me/KAOLAAIBOT?start=pa_i_-100123"},
		{action: "points", want: "https://t.me/KAOLAAIBOT?start=pa_p_-100123"},
		{action: "rank", want: "https://t.me/KAOLAAIBOT?start=pa_r_-100123"},
		{action: "exchange", want: "https://t.me/KAOLAAIBOT?start=pa_e_-100123"},
	}
	for _, tt := range tests {
		if got := scheduledPostActionURL("@KAOLAAIBOT", tt.action, -100123); got != tt.want {
			t.Errorf("action %s URL = %q, want %q", tt.action, got, tt.want)
		}
	}
	if got := scheduledPostActionURL("KAOLAAIBOT", "unknown", -100123); got != "" {
		t.Fatalf("unknown action URL = %q", got)
	}
}

func TestParseInlineKeyboardBuildsActionURL(t *testing.T) {
	markup, err := parseInlineKeyboard(`[[{"text":"签到","action":"sign"},{"text":"官网","url":"https://example.com"}]]`, "KAOLAAIBOT", -100123)
	if err != nil {
		t.Fatal(err)
	}
	if markup == nil || len(markup.InlineKeyboard) != 1 || len(markup.InlineKeyboard[0]) != 2 {
		t.Fatalf("markup = %#v", markup)
	}
	if got := markup.InlineKeyboard[0][0].Url; got != "https://t.me/KAOLAAIBOT?start=pa_s_-100123" {
		t.Fatalf("action URL = %q", got)
	}
}
