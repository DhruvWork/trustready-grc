// Copyright (c) 2025 TrustReady
//
// Cloud scanner proxy: forwards authenticated console requests to the
// TrustReady Security API. The browser never sees the scanner service
// credential; the GRC session is the only auth the UI needs.
//
// Each GRC user is mapped to their own scanner identity (and therefore their
// own scanner organization), so one person's scans are never visible to
// another. The mapping is requested from the scanner with a service key that
// only this server holds; the browser cannot ask for someone else's token.
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

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

const tokenCacheTTL = 50 * time.Minute

type cachedToken struct {
	value string
	exp   time.Time
}

type organizationResolver func(context.Context, gid.GID) (gid.GID, error)

type Handler struct {
	baseURL    string
	serviceKey string
	client     *http.Client
	resolveOrg organizationResolver

	mu     sync.Mutex
	tokens map[string]cachedToken
}

// NewHandler builds the proxy and wraps it in the console session
// middleware, so only a logged-in GRC user can reach the scanner.
func NewHandler(iamSvc *iam.Service, cookieConfig securecookie.Config) http.Handler {
	h := &Handler{
		baseURL:    strings.TrimRight(os.Getenv("TRUSTREADY_CLOUD_SCANNER_URL"), "/"),
		serviceKey: os.Getenv("TRUSTREADY_CLOUD_SCANNER_SERVICE_KEY"),
		client:     &http.Client{Timeout: 5 * time.Minute},
		resolveOrg: func(ctx context.Context, membershipID gid.GID) (gid.GID, error) {
			membership, err := iamSvc.GetMembership(ctx, membershipID)
			if err != nil {
				return gid.Nil, err
			}

			return membership.OrganizationID, nil
		},
		tokens: make(map[string]cachedToken),
	}

	return authn.NewSessionMiddleware(iamSvc, cookieConfig)(h)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session := authn.SessionFromContext(r.Context())
	identity := authn.IdentityFromContext(r.Context())
	if session == nil || identity == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// The scanner identity is keyed by email. An unverified (or just changed)
	// address must never be allowed to inherit an existing scanner identity.
	if !identity.EmailAddressVerified {
		writeError(w, http.StatusForbidden, "email verification required")
		return
	}

	if h.baseURL == "" {
		writeError(w, http.StatusServiceUnavailable, "cloud scanner is not configured")
		return
	}

	if h.serviceKey == "" {
		writeError(w, http.StatusServiceUnavailable, "cloud scanner service key is not configured")
		return
	}

	organizationID := ""
	if session.MembershipID != nil {
		resolved, err := h.resolveOrg(r.Context(), *session.MembershipID)
		if err != nil {
			writeError(w, http.StatusForbidden, "organization membership required")
			return
		}
		organizationID = resolved.String()
	}

	email := strings.ToLower(strings.TrimSpace(identity.EmailAddress.String()))
	if email == "" {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	token, err := h.bearer(r.Context(), email, identity.FullName, organizationID)
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

func (h *Handler) bearer(ctx context.Context, email, displayName, organizationID string) (string, error) {
	cacheKey := email + "\x00" + organizationID

	h.mu.Lock()
	cached, ok := h.tokens[cacheKey]
	h.mu.Unlock()
	if ok && time.Now().Before(cached.exp) {
		return cached.value, nil
	}

	payload, err := json.Marshal(map[string]string{
		"email":               email,
		"display_name":        displayName,
		"grc_organization_id": organizationID,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.baseURL+"/auth/service-token", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", h.serviceKey)

	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("scanner service-token returned %d", resp.StatusCode)
	}

	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Token == "" {
		return "", fmt.Errorf("scanner service-token returned no token")
	}

	h.mu.Lock()
	h.tokens[cacheKey] = cachedToken{value: out.Token, exp: time.Now().Add(tokenCacheTTL)}
	h.mu.Unlock()

	return out.Token, nil
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}
