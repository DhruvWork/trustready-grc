// Copyright (c) 2025 TrustReady
//
// Cloud scanner proxy: forwards authenticated console requests to the
// TrustReady Security API. The browser never sees the scanner service
// credential; the GRC session is the only auth the UI needs.
package cloudscanner_v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

const tokenTTL = 10 * time.Hour

type Handler struct {
	baseURL  string
	email    string
	password string
	client   *http.Client

	mu    sync.Mutex
	token string
	exp   time.Time
}

// NewHandler builds the proxy and wraps it in the console session
// middleware, so only a logged-in GRC user can reach the scanner.
func NewHandler(iamSvc *iam.Service, cookieConfig securecookie.Config) http.Handler {
	h := &Handler{
		baseURL:  strings.TrimRight(os.Getenv("TRUSTREADY_CLOUD_SCANNER_URL"), "/"),
		email:    os.Getenv("TRUSTREADY_CLOUD_SCANNER_EMAIL"),
		password: os.Getenv("TRUSTREADY_CLOUD_SCANNER_PASSWORD"),
		client:   &http.Client{Timeout: 5 * time.Minute},
	}

	return authn.NewSessionMiddleware(iamSvc, cookieConfig)(h)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if authn.SessionFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if h.baseURL == "" {
		writeError(w, http.StatusServiceUnavailable, "cloud scanner is not configured")
		return
	}

	token, err := h.bearer(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, "cloud scanner is unavailable")
		return
	}

	path := r.URL.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	target := h.baseURL + path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	var body io.Reader
	if r.Body != nil {
		body = r.Body
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, body)
	if err != nil {
		writeError(w, http.StatusBadGateway, "cloud scanner is unavailable")
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if ct := r.Header.Get("Content-Type"); ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	if accept := r.Header.Get("Accept"); accept != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, "cloud scanner is unavailable")
		return
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		w.Header().Set("Content-Disposition", cd)
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *Handler) bearer(ctx context.Context) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.token != "" && time.Now().Before(h.exp) {
		return h.token, nil
	}

	payload, err := json.Marshal(map[string]string{
		"email":    h.email,
		"password": h.password,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.baseURL+"/auth/login", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("scanner login returned %d", resp.StatusCode)
	}

	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Token == "" {
		return "", fmt.Errorf("scanner login returned no token")
	}

	h.token = out.Token
	h.exp = time.Now().Add(tokenTTL)
	return h.token, nil
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}
