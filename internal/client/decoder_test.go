package client

import (
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	decoder := decoder{}

	tests := []struct {
		name  string
		input []byte
		msg   *Message
		valid bool
	}{
		{
			name:  "valid subscribe",
			input: []byte("<from nokka> @"),
			msg: &Message{
				Account: "nokka",
				Cmd:     TypeSubscribe,
			},
			valid: true,
		},
		{
			name:  "invalid subscribe",
			input: []byte("<from nokka> sub"),
			valid: false,
		},
		{
			name:  "valid subscribe - truncate message",
			input: []byte("<from nokka> @ truncated"),
			msg: &Message{
				Account: "nokka",
				Cmd:     TypeSubscribe,
			},
			valid: true,
		},
		{
			name:  "valid publish",
			input: []byte("<from nokka> # hello there"),
			msg: &Message{
				Account: "nokka",
				Cmd:     TypePublish,
				Message: "[nokka] hello there",
			},
			valid: true,
		},
		{
			name:  "valid publish with special characters",
			input: []byte("<from nokka> # hello there!yo;_; _> test/&ader>...//derp bu@ | && > (() t ;>"),
			msg: &Message{
				Account: "nokka",
				Cmd:     TypePublish,
				Message: "[nokka] hello there!yo;_; _> test/&ader>...//derp bu@ | && > (() t ;>",
			},
			valid: true,
		},
		{
			name:  "valid publish remove IP address",
			input: []byte("<from nokka> # hello there 118.99.81.204"),
			msg: &Message{
				Account: "nokka",
				Cmd:     TypePublish,
				Message: "[nokka] hello there ",
			},
			valid: true,
		},
		{
			name:  "valid publish remove IP address in the middle",
			input: []byte("<from nokka> # hello there 82.254.181.210 how are you?"),
			msg: &Message{
				Account: "nokka",
				Cmd:     TypePublish,
				Message: "[nokka] hello there  how are you?",
			},
			valid: true,
		},
		{
			name:  "invalid publish",
			input: []byte("<from nokka> pub"),
			valid: false,
		},
		{
			name:  "invalid publish - random",
			input: []byte("<from nokka> random message that won't get through"),
			valid: false,
		},
		{
			name:  "valid ban",
			input: []byte("<from nokka> ~ nokka_bo 25"),
			msg: &Message{
				Account: "nokka",
				Cmd:     TypeBan,
				Message: "cheatingaccount 5",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, valid := decoder.Decode(tt.input)

			if tt.valid != valid {
				t.Fatalf("expected valid = %v; got = %v", tt.valid, valid)
			}

			if !reflect.DeepEqual(tt.msg, msg) {
				t.Fatalf("expected: %v, got: %v", tt.msg, msg)
			}
		})
	}
}
