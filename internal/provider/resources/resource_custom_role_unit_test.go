package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTokensServer stands in for the IAM ListApiTokens/GetApiToken endpoints.
type fakeTokensServer struct {
	mu         sync.Mutex
	tokens     []iam.ApiToken
	listDelay  time.Duration
	listBroken bool
}

func newFakeTokensServer(t *testing.T, tokens []iam.ApiToken) (*httptest.Server, *fakeTokensServer) {
	t.Helper()
	fake := &fakeTokensServer{tokens: tokens}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fake.mu.Lock()
		delay, broken, allTokens := fake.listDelay, fake.listBroken, fake.tokens
		fake.mu.Unlock()

		segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		tokensIdx := -1
		for i, s := range segments {
			if s == "tokens" {
				tokensIdx = i
				break
			}
		}
		if tokensIdx == -1 {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if tokensIdx < len(segments)-1 {
			tokenId := segments[tokensIdx+1]
			for _, tok := range allTokens {
				if tok.Id == tokenId {
					_ = json.NewEncoder(w).Encode(tok)
					return
				}
			}
			http.NotFound(w, r)
			return
		}

		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-r.Context().Done():
				return
			}
		}
		if broken {
			_, _ = w.Write([]byte("{not valid json"))
			return
		}

		limit, offset := 1000, 0
		if v := r.URL.Query().Get("limit"); v != "" {
			limit, _ = strconv.Atoi(v)
		}
		if v := r.URL.Query().Get("offset"); v != "" {
			offset, _ = strconv.Atoi(v)
		}

		var page []iam.ApiToken
		if offset < len(allTokens) {
			page = allTokens[offset:min(offset+limit, len(allTokens))]
		}

		_ = json.NewEncoder(w).Encode(iam.ApiTokensPaginated{
			Tokens:     page,
			Limit:      limit,
			Offset:     offset,
			TotalCount: len(allTokens),
		})
	}))
	return srv, fake
}

func fakeDirectAccessToken(id, name string, roles ...string) iam.ApiToken {
	tok := iam.ApiToken{Id: id, Name: name, Kind: iam.ApiTokenKindDIRECTACCESS}
	if len(roles) > 0 {
		apiRoles := make([]iam.ApiTokenRole, 0, len(roles))
		for _, r := range roles {
			apiRoles = append(apiRoles, iam.ApiTokenRole{Role: r})
		}
		tok.Roles = &apiRoles
	}
	return tok
}

func TestUnit_FindDirectAccessTokensUsingRole_NoMatches(t *testing.T) {
	srv, _ := newFakeTokensServer(t, []iam.ApiToken{
		fakeDirectAccessToken("tok-1", "one", "OtherRole"),
		fakeDirectAccessToken("tok-2", "two"),
	})
	defer srv.Close()

	client, err := iam.NewIamClient(srv.URL, "token", "test")
	require.NoError(t, err)

	matches, err := findDirectAccessTokensUsingRoleWithLimits(context.Background(), client, "org", "TargetRole", 1000, 4)
	require.NoError(t, err)
	assert.Empty(t, matches)
}

func TestUnit_FindDirectAccessTokensUsingRole_MatchesAcrossPages(t *testing.T) {
	srv, _ := newFakeTokensServer(t, []iam.ApiToken{
		fakeDirectAccessToken("tok-0", "zero", "OtherRole"),
		fakeDirectAccessToken("tok-1", "one"),
		fakeDirectAccessToken("tok-2", "two", "AnotherRole"),
		fakeDirectAccessToken("tok-3", "three", "TargetRole"),
		fakeDirectAccessToken("tok-4", "four"),
	})
	defer srv.Close()

	client, err := iam.NewIamClient(srv.URL, "token", "test")
	require.NoError(t, err)

	matches, err := findDirectAccessTokensUsingRoleWithLimits(context.Background(), client, "org", "TargetRole", 2, 4)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, "tok-3", matches[0].Id)
}

func TestUnit_FindDirectAccessTokensUsingRole_ListErrorFailsOpen(t *testing.T) {
	srv, fake := newFakeTokensServer(t, []iam.ApiToken{fakeDirectAccessToken("tok-1", "one", "TargetRole")})
	defer srv.Close()
	fake.mu.Lock()
	fake.listBroken = true
	fake.mu.Unlock()

	client, err := iam.NewIamClient(srv.URL, "token", "test")
	require.NoError(t, err)

	_, err = findDirectAccessTokensUsingRoleWithLimits(context.Background(), client, "org", "TargetRole", 1000, 4)
	assert.Error(t, err)
}

func TestUnit_FindDirectAccessTokensUsingRole_RespectsTimeout(t *testing.T) {
	srv, fake := newFakeTokensServer(t, []iam.ApiToken{fakeDirectAccessToken("tok-1", "one", "TargetRole")})
	defer srv.Close()
	fake.mu.Lock()
	fake.listDelay = 500 * time.Millisecond
	fake.mu.Unlock()

	client, err := iam.NewIamClient(srv.URL, "token", "test")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = findDirectAccessTokensUsingRoleWithLimits(ctx, client, "org", "TargetRole", 1000, 4)
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Less(t, elapsed, 400*time.Millisecond, "lookup should abort on context timeout, not wait for the slow response")
}

func TestUnit_FormatTokenListForDiagnostic_TruncatesAtTen(t *testing.T) {
	var tokens []iam.ApiToken
	for i := range 12 {
		tokens = append(tokens, fakeDirectAccessToken(fmt.Sprintf("tok-%d", i), fmt.Sprintf("name-%d", i)))
	}

	result := formatTokenListForDiagnostic(tokens)

	assert.Contains(t, result, "name-0 (id: tok-0)")
	assert.Contains(t, result, "name-9 (id: tok-9)")
	assert.NotContains(t, result, "name-10")
	assert.Contains(t, result, "and 2 more")
}

func TestUnit_FormatTokenListForDiagnostic_NoTruncationUnderLimit(t *testing.T) {
	tokens := []iam.ApiToken{fakeDirectAccessToken("tok-1", "one"), fakeDirectAccessToken("tok-2", "two")}

	result := formatTokenListForDiagnostic(tokens)

	assert.Equal(t, "one (id: tok-1), two (id: tok-2)", result)
}
