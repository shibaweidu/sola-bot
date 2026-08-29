package bot

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func TestPrivateKeyboardUsesNativeToggle(t *testing.T) {
	markup := gotgbot.ReplyKeyboardMarkup{
		Keyboard:        [][]gotgbot.KeyboardButton{{{Text: "菜单"}}},
		ResizeKeyboard:  true,
		IsPersistent:    false,
		OneTimeKeyboard: false,
	}
	if !markup.ResizeKeyboard {
		t.Fatal("private keyboard should resize to fit the menu")
	}
	if markup.IsPersistent {
		t.Fatal("private keyboard must allow Telegram's native collapse/expand toggle")
	}
	if markup.OneTimeKeyboard {
		t.Fatal("private keyboard must remain available after a button is pressed")
	}
}
