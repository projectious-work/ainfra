package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/projectious-work/ainfra/internal/app"
)

func TestVerifyMCPAuthorizationRequiresExactIndependentBinding(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	request, grant := authorizationFixture(now)
	provider := authorizationProvider{grant: grant}
	result, err := app.VerifyMCPAuthorization(context.Background(), provider, request, now)
	if err != nil {
		t.Fatal(err)
	}
	if result.AuthorizationID != grant.AuthorizationID || result.Caller != grant.Caller ||
		result.Issuer != grant.Issuer || result.ExpiresAt != "2026-08-16T12:05:00Z" {
		t.Fatalf("unexpected authorization: %+v", result)
	}
}

func TestVerifyMCPAuthorizationAcceptsConfigureApplyBinding(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	request, grant := authorizationFixture(now)
	request.Operation = "configure"
	grant.Operation = "configure"
	if _, err := app.VerifyMCPAuthorization(context.Background(),
		authorizationProvider{grant: grant}, request, now); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyMCPAuthorizationFailsClosed(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	request, valid := authorizationFixture(now)
	tests := []struct {
		name   string
		mutate func(*app.MCPAuthorizationGrant)
	}{
		{"self approved", func(value *app.MCPAuthorizationGrant) { value.Issuer = value.Caller }},
		{"caller", func(value *app.MCPAuthorizationGrant) { value.Caller = "other-caller" }},
		{"root", func(value *app.MCPAuthorizationGrant) { value.ProjectRoot = "/other" }},
		{"operation", func(value *app.MCPAuthorizationGrant) { value.Operation = "deploy" }},
		{"plan ID", func(value *app.MCPAuthorizationGrant) { value.PlanID += "-other" }},
		{"plan digest", func(value *app.MCPAuthorizationGrant) { value.PlanDigest = "sha256:" + strings.Repeat("b", 64) }},
		{"intent", func(value *app.MCPAuthorizationGrant) { value.Intent = "destroy" }},
		{"future", func(value *app.MCPAuthorizationGrant) { value.ApprovedAt = now.Add(time.Second) }},
		{"expired", func(value *app.MCPAuthorizationGrant) { value.ExpiresAt = now }},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			grant := valid
			test.mutate(&grant)
			if _, err := app.VerifyMCPAuthorization(context.Background(),
				authorizationProvider{grant: grant}, request, now); err == nil {
				t.Fatal("mismatched authorization succeeded")
			}
		})
	}
	if _, err := app.VerifyMCPAuthorization(context.Background(), nil, request, now); err == nil {
		t.Fatal("missing provider succeeded")
	}
	if _, err := app.VerifyMCPAuthorization(context.Background(),
		authorizationProvider{grant: valid, err: errors.New("provider-secret")},
		request, now); err == nil || strings.Contains(err.Error(), "provider-secret") {
		t.Fatalf("provider failure was not sanitized: %v", err)
	}
	request.Approval = ""
	if _, err := app.VerifyMCPAuthorization(context.Background(), providerWith(valid), request, now); err == nil {
		t.Fatal("missing approval succeeded")
	}
}

func TestVerifyMCPAuthorizationRejectsInvalidExpectationBeforeProvider(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	request, grant := authorizationFixture(now)
	provider := &countingAuthorizationProvider{grant: grant}
	request.Intent = "destroy"
	if _, err := app.VerifyMCPAuthorization(context.Background(), provider, request, now); err == nil {
		t.Fatal("apply operation with destroy intent succeeded")
	}
	if provider.calls != 0 {
		t.Fatalf("provider called for invalid expectation: %d", provider.calls)
	}
}

func TestVerifyMCPAuthorizationPropagatesCancellationWithoutProviderDetail(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	request, grant := authorizationFixture(now)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := app.VerifyMCPAuthorization(ctx,
		authorizationProvider{grant: grant, err: errors.New("approval-secret")}, request, now)
	if !errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "approval-secret") {
		t.Fatalf("cancellation error = %v", err)
	}
}

func TestVerifyMCPAuthorizationRejectsCancellationWhenProviderIgnoresIt(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	request, grant := authorizationFixture(now)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := app.VerifyMCPAuthorization(ctx,
		authorizationProvider{grant: grant}, request, now); !errors.Is(err, context.Canceled) {
		t.Fatalf("ignored provider cancellation succeeded: %v", err)
	}
}

type authorizationProvider struct {
	grant app.MCPAuthorizationGrant
	err   error
}

func (provider authorizationProvider) Verify(context.Context, string) (app.MCPAuthorizationGrant, error) {
	return provider.grant, provider.err
}

type countingAuthorizationProvider struct {
	grant app.MCPAuthorizationGrant
	calls int
}

func (provider *countingAuthorizationProvider) Verify(context.Context,
	string,
) (app.MCPAuthorizationGrant, error) {
	provider.calls++
	return provider.grant, nil
}

func providerWith(grant app.MCPAuthorizationGrant) authorizationProvider {
	return authorizationProvider{grant: grant}
}

func authorizationFixture(now time.Time) (app.MCPMutationAuthorizationRequest,
	app.MCPAuthorizationGrant,
) {
	request := app.MCPMutationAuthorizationRequest{Approval: "opaque-approval",
		ProjectRoot: "/project", Operation: "apply",
		PlanID:     "20260816T120000Z-0123456789abcdef",
		PlanDigest: "sha256:" + strings.Repeat("a", 64), Intent: "apply", Caller: "agent-1"}
	grant := app.MCPAuthorizationGrant{AuthorizationID: "approval-1", Issuer: "operator-1",
		Caller: request.Caller, ProjectRoot: request.ProjectRoot, Operation: request.Operation,
		PlanID: request.PlanID, PlanDigest: request.PlanDigest, Intent: request.Intent,
		ApprovedAt: now.Add(-time.Minute), ExpiresAt: now.Add(5 * time.Minute)}
	return request, grant
}
