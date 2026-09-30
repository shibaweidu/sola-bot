package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/dabowin/sola/internal/api"
)

func TestNormalizeScheduledPostInlineKeyboard(t *testing.T) {
	valid := `[[{"text":"签到","action":"sign"},{"text":"官网","url":"https://example.com"}]]`
	normalized, err := normalizeScheduledPostInlineKeyboard(valid)
	if err != nil || normalized != valid {
		t.Fatalf("normalize valid keyboard = %q, %v", normalized, err)
	}

	tests := []struct {
		name string
		raw  string
	}{
		{name: "malformed", raw: `{`},
		{name: "empty text", raw: `[[{"text":"","action":"sign"}]]`},
		{name: "duplicate text", raw: `[[{"text":"签到","action":"sign"},{"text":"签到","action":"points"}]]`},
		{name: "non https", raw: `[[{"text":"官网","url":"http://example.com"}]]`},
		{name: "unsupported action", raw: `[[{"text":"操作","action":"admin"}]]`},
		{name: "too many columns", raw: `[[{"text":"1","action":"sign"},{"text":"2","action":"points"},{"text":"3","action":"rank"},{"text":"4","action":"exchange"},{"text":"5","action":"invite_rewards"}]]`},
		{name: "too many buttons", raw: `[[{"text":"1","action":"sign"},{"text":"2","action":"points"},{"text":"3","action":"rank"},{"text":"4","action":"exchange"}],[{"text":"5","action":"sign"},{"text":"6","action":"points"},{"text":"7","action":"rank"},{"text":"8","action":"exchange"}],[{"text":"9","action":"sign"},{"text":"10","action":"points"},{"text":"11","action":"rank"}]]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeScheduledPostInlineKeyboard(tt.raw)
			if !errors.Is(err, api.ErrBadRequest) {
				t.Fatalf("error = %v, want ErrBadRequest", err)
			}
		})
	}

	if normalized, err := normalizeScheduledPostInlineKeyboard(strings.Repeat(" ", 3)); err != nil || normalized != "[]" {
		t.Fatalf("empty keyboard = %q, %v", normalized, err)
	}
}
