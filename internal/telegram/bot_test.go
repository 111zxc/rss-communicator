package telegram

import "testing"

func TestParseRegistrationCommand(t *testing.T) {
	tests := []struct {
		in       string
		wantCmd  string
		wantCode string
		ok       bool
	}{
		{in: "/start", wantCmd: "start", wantCode: "", ok: true},
		{in: "/start abc123", wantCmd: "start", wantCode: "ABC123", ok: true},
		{in: "/register vip", wantCmd: "register", wantCode: "VIP", ok: true},
		{in: "hello", ok: false},
	}

	for _, tt := range tests {
		gotCmd, gotCode, ok := parseRegistrationCommand(tt.in)
		if ok != tt.ok || gotCmd != tt.wantCmd || gotCode != tt.wantCode {
			t.Fatalf("parseRegistrationCommand(%q) = (%q, %q, %v)", tt.in, gotCmd, gotCode, ok)
		}
	}
}

func TestParseConfirmCallback(t *testing.T) {
	code, ok := parseConfirmCallback("confirm|ABC123")
	if !ok || code != "ABC123" {
		t.Fatalf("unexpected callback parse result: %q %v", code, ok)
	}
}
