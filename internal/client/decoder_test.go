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
			input: []byte("<from nokka> sub hc"),
			msg: &Message{
				ChatID:  "hc",
				Account: "nokka",
				Cmd:     "sub",
			},
			valid: true,
		},
		{
			name:  "invalid subscribe - chat id",
			input: []byte("<from nokka> sub derp"),
			valid: false,
		},
		{
			name:  "invalid subscribe - cmd",
			input: []byte("<from nokka> suu hc"),
			valid: false,
		},
		{
			name:  "valid subscribe - truncate message",
			input: []byte("<from nokka> sub hc truncated"),
			msg: &Message{
				ChatID:  "hc",
				Account: "nokka",
				Cmd:     "sub",
			},
			valid: true,
		},
		{
			name:  "valid publish",
			input: []byte("<from nokka> pub hc hello there"),
			msg: &Message{
				ChatID:  "hc",
				Account: "nokka",
				Cmd:     "pub",
				Message: "[nokka] hello there",
			},
			valid: true,
		},
		{
			name:  "invalid publish - chat id",
			input: []byte("<from nokka> pub random hello there"),
			valid: false,
		},
		{
			name:  "invalid publish - cmd",
			input: []byte("<from nokka> ppp hc hello there"),
			valid: false,
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
