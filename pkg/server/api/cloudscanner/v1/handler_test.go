// Copyright (c) 2025 TrustReady

package cloudscanner_v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/server/api/authn"
)

func TestProxyUsesTheSignedInUsersScannerIdentity(t *testing.T) {
	var serviceTokenEmail string
	var serviceTokenKey string
	var proxiedAuthorization string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/service-token":
			serviceTokenKey = r.Header.Get("X-Api-Key")
			var body struct {
				Email string `json:"email"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			serviceTokenEmail = body.Email
			_, _ = io.WriteString(w, `{"token":"per-user-token"}`)
		case "/audits":
			proxiedAuthorization = r.Header.Get("Authorization")
			_, _ = io.WriteString(w, `{"audits":[],"count":0}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	membershipID := gid.New(gid.NilTenant, 1)
	organizationID := gid.New(gid.NilTenant, 2)
	h := &Handler{
		baseURL:    upstream.URL,
		serviceKey: "service-key",
		client:     upstream.Client(),
		resolveOrg: func(context.Context, gid.GID) (gid.GID, error) {
			return organizationID, nil
		},
		tokens: make(map[string]cachedToken),
	}

	session := &coredata.Session{MembershipID: &membershipID}
	identity := &coredata.Identity{
		EmailAddress: mail.Addr("User@Example.com"),
		FullName:     "User",
	}
	ctx := authn.ContextWithSession(context.Background(), session)
	ctx = authn.ContextWithIdentity(ctx, identity)

	req := httptest.NewRequest(http.MethodGet, "/audits", nil).WithContext(ctx)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if serviceTokenKey != "service-key" {
		t.Fatalf("expected service key to be sent, got %q", serviceTokenKey)
	}
	if serviceTokenEmail != "user@example.com" {
		t.Fatalf("expected the signed-in user's email, got %q", serviceTokenEmail)
	}
	if proxiedAuthorization != "Bearer per-user-token" {
		t.Fatalf("expected per-user token, got %q", proxiedAuthorization)
	}
}

func TestProxyRejectsAnonymousRequests(t *testing.T) {
	h := &Handler{
		baseURL:    "https://scanner.invalid",
		serviceKey: "service-key",
		client:     http.DefaultClient,
		tokens:     make(map[string]cachedToken),
	}

	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/audits", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestProxyFailsClosedWithoutServiceKey(t *testing.T) {
	membershipID := gid.New(gid.NilTenant, 1)
	session := &coredata.Session{MembershipID: &membershipID}
	identity := &coredata.Identity{EmailAddress: mail.Addr("user@example.com")}
	ctx := authn.ContextWithSession(context.Background(), session)
	ctx = authn.ContextWithIdentity(ctx, identity)

	h := &Handler{
		baseURL: "https://scanner.invalid",
		client:  http.DefaultClient,
		tokens:  make(map[string]cachedToken),
	}

	req := httptest.NewRequest(http.MethodGet, "/audits", nil).WithContext(ctx)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "service key") {
		t.Fatalf("expected service key error, got %s", recorder.Body.String())
	}
}
