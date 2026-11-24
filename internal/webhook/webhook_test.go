package webhook

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serve(t *testing.T, method, secret, body string) (int, []byte) {
	t.Helper()
	var got []byte
	h := Handler("s3cret-value", func(b []byte) { got = b })
	req := httptest.NewRequest(method, "/webhook", strings.NewReader(body))
	if secret != "" {
		req.Header.Set(SecretHeader, secret)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code, got
}

func TestAcceptsUpdateWithCorrectSecret(t *testing.T) {
	code, got := serve(t, http.MethodPost, "s3cret-value", `{"update_id":1}`)
	if code != http.StatusOK || string(got) != `{"update_id":1}` {
		t.Fatalf("code=%d body=%q", code, got)
	}
}

func TestRejectsMissingOrWrongSecret(t *testing.T) {
	for _, s := range []string{"", "wrong", "s3cret-valu"} {
		code, got := serve(t, http.MethodPost, s, `{}`)
		if code != http.StatusUnauthorized || got != nil {
			t.Fatalf("secret %q: code=%d handled=%v", s, code, got != nil)
		}
	}
}

func TestRejectsNonPost(t *testing.T) {
	if code, _ := serve(t, http.MethodGet, "s3cret-value", ""); code != http.StatusMethodNotAllowed {
		t.Fatalf("code=%d", code)
	}
}

func TestEmptyConfiguredSecretRejectsEverything(t *testing.T) {
	h := Handler("", func([]byte) { t.Fatal("handler must not run") })
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestRejectsOversizedBody(t *testing.T) {
	code, got := serve(t, http.MethodPost, "s3cret-value", strings.Repeat("a", MaxBodyBytes+1))
	if code != http.StatusBadRequest || got != nil {
		t.Fatalf("code=%d", code)
	}
}
