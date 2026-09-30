package worker

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"
)

func TestScheduledPostPermanentError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "deadline exceeded", err: context.DeadlineExceeded, want: false},
		{name: "network timeout", err: &url.Error{Op: "Post", URL: "https://api.telegram.org", Err: context.DeadlineExceeded}, want: false},
		{name: "telegram unavailable", err: errors.New("telegram: 502 Bad Gateway"), want: false},
		{name: "chat missing", err: errors.New("Error 400: Bad Request: chat not found"), want: true},
		{name: "bot kicked", err: errors.New("Error 403: Forbidden: bot was kicked from the supergroup chat"), want: true},
		{name: "insufficient rights", err: errors.New("Error 400: Bad Request: not enough rights to send text messages"), want: true},
		{name: "invalid task", err: fmt.Errorf("photo scheduled post requires media"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scheduledPostPermanentError(tt.err); got != tt.want {
				t.Fatalf("scheduledPostPermanentError(%v) = %t, want %t", tt.err, got, tt.want)
			}
		})
	}
}
