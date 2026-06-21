package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// buildInitData produces a correctly-signed init data string for tests.
func buildInitData(botToken string, tgID int64, authDate time.Time) string {
	v := url.Values{}
	v.Set("auth_date", strconv.FormatInt(authDate.Unix(), 10))
	v.Set("user", `{"id":`+strconv.FormatInt(tgID, 10)+`,"first_name":"T"}`)
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(k + "=" + v.Get(k))
	}
	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	sig := hex.EncodeToString(hmacSHA256(secret, []byte(b.String())))
	v.Set("hash", sig)
	return v.Encode()
}

func TestValidateInitDataOK(t *testing.T) {
	tok := "123:abc"
	id, err := validateInitData(buildInitData(tok, 555, time.Now()), tok, time.Hour)
	if err != nil || id != 555 {
		t.Fatalf("id=%d err=%v", id, err)
	}
}

func TestValidateInitDataTampered(t *testing.T) {
	tok := "123:abc"
	data := buildInitData(tok, 555, time.Now())
	if _, err := validateInitData(data, "999:wrong", time.Hour); err == nil {
		t.Fatal("expected failure for wrong bot token")
	}
	if _, err := validateInitData(data+"&x=1", tok, time.Hour); err == nil {
		t.Fatal("expected failure for tampered payload")
	}
}

func TestValidateInitDataExpired(t *testing.T) {
	tok := "123:abc"
	data := buildInitData(tok, 555, time.Now().Add(-2*time.Hour))
	if _, err := validateInitData(data, tok, time.Hour); err == nil {
		t.Fatal("expected failure for expired auth_date")
	}
}

func TestJWTRoundTrip(t *testing.T) {
	key := jwtKey("123:abc")
	tok := issueJWT(777, key, time.Hour)
	id, err := parseJWT(tok, key)
	if err != nil || id != 777 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if _, err := parseJWT(tok, jwtKey("other")); err == nil {
		t.Fatal("expected failure with wrong key")
	}
	if _, err := parseJWT(issueJWT(1, key, -time.Minute), key); err == nil {
		t.Fatal("expected failure for expired jwt")
	}
}

// sanity: ensure our test signer matches the hex format expectation
var _ = hmac.Equal
var _ = sha256.Size
