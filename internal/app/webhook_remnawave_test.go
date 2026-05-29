package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyRemnawaveSignature_OK(t *testing.T) {
	secret := "topsecret"
	body := []byte(`{"event":"user.expired","data":{"telegramId":42}}`)
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	sig := hex.EncodeToString(m.Sum(nil))

	if err := verifyRemnawaveSignature(sig, secret, body); err != nil {
		t.Fatalf("ожидался OK, получили: %v", err)
	}

	if err := verifyRemnawaveSignature("sha256="+sig, secret, body); err != nil {
		t.Fatalf("ожидался OK c префиксом, получили: %v", err)
	}
}

func TestVerifyRemnawaveSignature_Bad(t *testing.T) {
	secret := "topsecret"
	body := []byte(`{"event":"user.expired"}`)
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	sig := hex.EncodeToString(m.Sum(nil))

	tampered := []byte(`{"event":"user.expired","data":{"telegramId":99}}`)
	if err := verifyRemnawaveSignature(sig, secret, tampered); err == nil {
		t.Fatalf("ожидалась ошибка при tampered теле")
	}
}

func TestVerifyRemnawaveSignature_EmptySecret(t *testing.T) {
	if err := verifyRemnawaveSignature("", "", []byte("anything")); err != nil {
		t.Fatalf("при пустом секрете ошибок быть не должно: %v", err)
	}
}

func TestVerifyRemnawaveSignature_MissingHeader(t *testing.T) {
	if err := verifyRemnawaveSignature("", "secret", []byte("body")); err == nil {
		t.Fatalf("ожидалась ошибка из-за отсутствия заголовка")
	}
}
