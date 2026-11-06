package auth

import (
	"encoding/hex"
	"errors"
	"testing"
	"strings"
)

func TestPBKDF2SHA256KnownVector(t *testing.T) {
	got := pbkdf2SHA256([]byte("password"), []byte("salt"), 1, 32)
	want, _ := hex.DecodeString("120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b")
	if string(got) != string(want) {
		t.Fatalf("PBKDF2 vector mismatch: %x", got)
	}
}

func TestPasswordHashAndVerify(t *testing.T) {
	encoded, err := HashPassword("a-long-test-password")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := VerifyPassword(encoded, "a-long-test-password")
	if err != nil || !ok {
		t.Fatalf("valid password rejected: ok=%v err=%v", ok, err)
	}
	ok, err = VerifyPassword(encoded, "a-different-password")
	if err != nil || ok {
		t.Fatalf("wrong password accepted: ok=%v err=%v", ok, err)
	}
}

func TestPasswordValidationAndMalformedHash(t *testing.T) {
	if _, err := HashPassword("short"); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("short password error = %v", err)
	}
	if _, err := VerifyPassword("not-a-hash", "anything"); !errors.Is(err, ErrInvalidHash) {
		t.Fatalf("malformed hash error = %v", err)
	}
	if _, err := VerifyPassword("not-a-hash", strings.Repeat("x", maxPasswordBytes+1)); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("oversized password error = %v", err)
	}
}

func TestSessionTokenStoresOnlyDeterministicDigest(t *testing.T) {
	token, digest, err := NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	recomputed, err := HashSessionToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if digest != recomputed {
		t.Fatal("session digest did not match")
	}
	if _, err := HashSessionToken("not-a-session-token"); err == nil {
		t.Fatal("malformed token accepted")
	}
	other, otherDigest, err := NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	if token == other || digest == otherDigest {
		t.Fatal("session tokens were not unique")
	}
}
