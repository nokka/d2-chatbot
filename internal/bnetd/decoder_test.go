package bnetd

import (
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	decoder := decoder{}

	tests := []struct {
		name   string
		input  string
		change *StatusChange
		valid  bool
	}{
		{
			name:  "user login",
			input: "Aug 28 08:25:22 [info ] _client_loginreq2: [28] \"nokka\" logged in (correct password)",
			change: &StatusChange{
				Account: "nokka",
				Online:  true,
			},
			valid: true,
		},
		{
			name:  "user logout",
			input: "Aug 28 09:01:48 [info ] conn_destroy: [28] \"nokka\" logged out",
			change: &StatusChange{
				Account: "nokka",
				Online:  false,
			},
			valid: true,
		},
		{
			name:   "unrelated parse",
			input:  "Aug 28 09:01:48 [debug] sd_tcpinput: [28] read returned -1 (closing connection)",
			change: nil,
			valid:  false,
		},
		{
			name:  "bot login",
			input: "Aug 28 08:15:05 [info ] handle_telnet_packet: [28] \"Discord\" bot logged in (correct password)",
			change: &StatusChange{
				Account: "discord",
				Online:  true,
			},
			valid: true,
		},
		{
			name:  "bot logout",
			input: "Aug 28 09:15:35 [info ] conn_destroy: [28] \"discord\" logged out",
			change: &StatusChange{
				Account: "discord",
				Online:  false,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change, valid := decoder.Decode(tt.input)

			if tt.valid != valid {
				t.Fatalf("expected valid = %v; got = %v", tt.valid, valid)
			}

			if !reflect.DeepEqual(tt.change, change) {
				t.Fatalf("expected: %v, got: %v", tt.change, change)
			}
		})
	}
}
