package bot

import (
	"strings"
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func TestForceSubscribeTargets(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "usernames and ids", raw: "@news\nhttps://t.me/alerts\n-1001234567890", want: []string{"@news", "@alerts", "-1001234567890"}},
		{name: "deduplicates", raw: "@news,@news; @news", want: []string{"@news"}},
		{name: "rejects invite links", raw: "https://t.me/+privateInvite", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := forceSubscribeTargets(tt.raw)
			if len(got) != len(tt.want) {
				t.Fatalf("targets = %#v, want %#v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("targets = %#v, want %#v", got, tt.want)
				}
			}
		})
	}
}

func TestForceSubscribeTargetURL(t *testing.T) {
	if got := forceSubscribeTargetURL("@example"); got != "https://t.me/example" {
		t.Fatalf("target URL = %q", got)
	}
	if got := forceSubscribeTargetURL("-1001234567890"); got != "" {
		t.Fatalf("numeric target URL = %q, want empty", got)
	}
}

func TestForceSubscribeLabelsAndMessage(t *testing.T) {
	cfg := ChatAdminConfig{
		ForceSubscribeChannelLabels: "@news | 公告频道\n-1001234567890|会员频道",
	}
	if got := forceSubscribeChannelLabel(cfg, "@news"); got != "公告频道" {
		t.Fatalf("label = %q, want 公告频道", got)
	}
	if got := forceSubscribeChannelLabel(cfg, "@other"); got != "@other" {
		t.Fatalf("unmapped label = %q, want @other", got)
	}
	message := formatForceSubscribeMessage("{name} 待订阅：{channels}", cfg, "小明", []string{"@news", "-1001234567890"}, "fallback")
	if !strings.Contains(message, "小明") || !strings.Contains(message, "公告频道、会员频道") {
		t.Fatalf("formatted message = %q", message)
	}
	if got := forceSubscribeMuteText(cfg, gotgbot.User{Id: 7, FirstName: "小明"}, []string{"@news"}); !strings.Contains(got, "公告频道") {
		t.Fatalf("mute text = %q", got)
	}
}

func TestIsNewChatMemberStatus(t *testing.T) {
	user := gotgbot.User{Id: 7}
	left := gotgbot.ChatMemberLeft{User: user}
	joined := gotgbot.ChatMemberMember{User: user}
	if !isNewChatMemberStatus(left, joined) {
		t.Fatal("left -> member should be treated as a join")
	}
	if isNewChatMemberStatus(joined, joined) {
		t.Fatal("member -> member should not be treated as a join")
	}
	restricted := gotgbot.ChatMemberRestricted{User: user, IsMember: true}
	if !isNewChatMemberStatus(left, restricted) || !chatMemberPresent(restricted) {
		t.Fatal("left -> restricted member should be treated as a join")
	}
}
