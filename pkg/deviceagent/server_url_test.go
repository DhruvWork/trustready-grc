// Copyright (c) 2026 TrustReady <hello@trustready.io>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package deviceagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeServerURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "adds https to bare hostnames", input: "eu.trustready.io", want: EUConsoleURL},
		{name: "trims trailing slash", input: "https://us.trustready.io/", want: USConsoleURL},
		{name: "rejects whitespace-only input", input: "   ", wantErr: true},
		{name: "rejects paths", input: "https://trustready.example.com/workspace", wantErr: true},
		{name: "accepts uppercase scheme", input: "HTTPS://eu.trustready.io/", want: EUConsoleURL},
		{name: "lowercases mixed-case hostname", input: "HTTPS://US.TrustReady.Com/", want: USConsoleURL},
		{name: "lowercases bare mixed-case hostname", input: "EU.trustready.io", want: EUConsoleURL},
		{name: "lowercases hostname with port", input: "http://LocalHost:3000", want: "http://localhost:3000"},
		{name: "rejects user credentials", input: "https://user@eu.trustready.io", wantErr: true},
		{name: "rejects query parameters", input: "https://eu.trustready.io?foo=bar", wantErr: true},
		{name: "rejects fragments", input: "https://eu.trustready.io#fragment", wantErr: true},
		{name: "rejects port-only host", input: "https://:443", wantErr: true},
		{name: "rejects bare port", input: ":443", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeServerURL(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConsoleEnrollURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "hosted US console",
			input: USConsoleURL,
			want:  USConsoleURL + "/enroll",
		},
		{
			name:  "local dev console",
			input: "http://localhost:3000",
			want:  "http://localhost:3000/enroll",
		},
		{
			name:    "rejects whitespace-only input",
			input:   "  ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ConsoleEnrollURL(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsTrustReadyServers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "US console", input: USConsoleURL, want: true},
		{name: "EU console", input: EUConsoleURL, want: true},
		{name: "mixed-case US host", input: "HTTPS://US.TrustReady.Com", want: true},
		{name: "bare EU host", input: "eu.trustready.io", want: true},
		{name: "self-hosted", input: "https://trustready.example.com", want: false},
		{name: "http self-hosted", input: "http://localhost:3000", want: false},
		{name: "lookalike host", input: "https://us.trustready.io.evil.example", want: false},
		{name: "empty", input: "", want: false},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, IsTrustReadyServers(tt.input))
			},
		)
	}
}

func TestRequiresEnrollmentConfirm(t *testing.T) {
	t.Parallel()

	assert.False(t, RequiresEnrollmentConfirm(TrustTrustReadyCloud))
	assert.True(t, RequiresEnrollmentConfirm(TrustUnverified))
	assert.True(t, RequiresEnrollmentConfirm(TrustInsecure))
	assert.True(t, RequiresEnrollmentConfirm(TrustUnknown))
}

func TestEnrollmentConfirmMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		serverURL string
		trust     EnrollmentTrust
		want      string
	}{
		{
			name:      "unverified host",
			serverURL: "https://trustready.example.com",
			trust:     TrustUnverified,
			want:      "This device will report to https://trustready.example.com.\n\nThis is not a TrustReady server.",
		},
		{
			name:      "cleartext http",
			serverURL: "http://us.trustready.io",
			trust:     TrustInsecure,
			want:      "This device will report to http://us.trustready.io.\n\nThis connection is not encrypted (http).",
		},
		{
			name:      "probe failed",
			serverURL: USConsoleURL,
			trust:     TrustUnknown,
			want:      "This device will report to https://us.trustready.io.\n\nCould not verify TLS for this server.",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, EnrollmentConfirmMessage(tt.serverURL, tt.trust))
			},
		)
	}
}
