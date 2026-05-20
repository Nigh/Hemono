package main

import (
	"regexp"
	"testing"
)

func TestGenerateInvitationCode_Format(t *testing.T) {
	code := generateInvitationCode()
	matched, err := regexp.MatchString(`^[A-Z]{3}-\d{6}$`, code)
	if err != nil {
		t.Fatalf("regex error: %v", err)
	}
	if !matched {
		t.Errorf("invitation code %q does not match expected format XXX-NNNNNN", code)
	}
}

func TestGenerateInvitationCode_Length(t *testing.T) {
	code := generateInvitationCode()
	if len(code) != 10 {
		t.Errorf("expected code length 10, got %d (%q)", len(code), code)
	}
}

func TestGenerateInvitationCode_DashPosition(t *testing.T) {
	code := generateInvitationCode()
	if code[3] != '-' {
		t.Errorf("expected '-' at position 3, got %q in %q", code[3], code)
	}
}

func TestGenerateInvitationCode_LetterPart(t *testing.T) {
	code := generateInvitationCode()
	for i := 0; i < 3; i++ {
		if code[i] < 'A' || code[i] > 'Z' {
			t.Errorf("expected uppercase letter at position %d, got %q in %q", i, code[i], code)
		}
	}
}

func TestGenerateInvitationCode_NumberPart(t *testing.T) {
	code := generateInvitationCode()
	for i := 4; i < 10; i++ {
		if code[i] < '0' || code[i] > '9' {
			t.Errorf("expected digit at position %d, got %q in %q", i, code[i], code)
		}
	}
}

func TestGenerateInvitationCode_Uniqueness(t *testing.T) {
	codes := make(map[string]bool)
	iterations := 100
	for i := 0; i < iterations; i++ {
		code := generateInvitationCode()
		if codes[code] {
			t.Errorf("duplicate code generated: %q (iteration %d)", code, i)
		}
		codes[code] = true
	}
}
