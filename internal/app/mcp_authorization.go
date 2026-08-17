package app

import (
	"context"
	"errors"
	"path/filepath"
	"regexp"
	"time"
)

const maxMCPApprovalBytes = 64 << 10

var sha256DigestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// MCPMutationAuthorizationRequest is the exact binding a mutation handler
// expects an independent approval provider to verify.
type MCPMutationAuthorizationRequest struct {
	Approval    string
	ProjectRoot string
	Operation   string
	PlanID      string
	PlanDigest  string
	Intent      string
	Caller      string
}

// MCPAuthorizationGrant is a provider-verified approval assertion. The
// application compares every field rather than trusting partial decisions.
type MCPAuthorizationGrant struct {
	AuthorizationID string
	Issuer          string
	Caller          string
	ProjectRoot     string
	Operation       string
	PlanID          string
	PlanDigest      string
	Intent          string
	ApprovedAt      time.Time
	ExpiresAt       time.Time
}

// MCPAuthorizationProvider independently verifies opaque approval material.
// Implementations own signature, identity, and trust-store policy.
type MCPAuthorizationProvider interface {
	Verify(context.Context, string) (MCPAuthorizationGrant, error)
}

// MCPAuthorization is the sanitized evidence retained by a mutation adapter.
type MCPAuthorization struct {
	AuthorizationID string `json:"authorizationId"`
	Caller          string `json:"caller"`
	Issuer          string `json:"issuer"`
	ExpiresAt       string `json:"expiresAt"`
}

// VerifyMCPAuthorization fails closed unless an independent provider returns
// a complete, current, exact, non-self-issued grant.
func VerifyMCPAuthorization(ctx context.Context, provider MCPAuthorizationProvider,
	request MCPMutationAuthorizationRequest, now time.Time,
) (MCPAuthorization, error) {
	if provider == nil {
		return MCPAuthorization{}, errors.New("independent MCP authorization is unavailable")
	}
	if len(request.Approval) == 0 || len(request.Approval) > maxMCPApprovalBytes {
		return MCPAuthorization{}, errors.New("MCP approval material has invalid size")
	}
	if !validMCPAuthorizationRequest(request) {
		return MCPAuthorization{}, errors.New("MCP authorization expectation is invalid")
	}
	if err := ctx.Err(); err != nil {
		return MCPAuthorization{}, err
	}
	grant, err := provider.Verify(ctx, request.Approval)
	if err != nil {
		if ctx.Err() != nil {
			return MCPAuthorization{}, ctx.Err()
		}
		return MCPAuthorization{}, errors.New("independent MCP authorization was not verified")
	}
	if err := ctx.Err(); err != nil {
		return MCPAuthorization{}, err
	}
	if !validAuthorizationID(grant.AuthorizationID) ||
		!validAuthorizationID(grant.Issuer) || grant.Issuer == request.Caller ||
		grant.Caller != request.Caller || grant.ProjectRoot != request.ProjectRoot ||
		grant.Operation != request.Operation || grant.PlanID != request.PlanID ||
		grant.PlanDigest != request.PlanDigest || grant.Intent != request.Intent {
		return MCPAuthorization{}, errors.New("MCP authorization binding mismatch")
	}
	current := now.UTC()
	if grant.ApprovedAt.IsZero() || grant.ExpiresAt.IsZero() ||
		grant.ApprovedAt.UTC().After(current) || !grant.ExpiresAt.UTC().After(current) ||
		!grant.ExpiresAt.UTC().After(grant.ApprovedAt.UTC()) {
		return MCPAuthorization{}, errors.New("MCP authorization is not currently valid")
	}
	return MCPAuthorization{AuthorizationID: grant.AuthorizationID,
		Caller: grant.Caller, Issuer: grant.Issuer,
		ExpiresAt: grant.ExpiresAt.UTC().Format(time.RFC3339)}, nil
}

func validMCPAuthorizationRequest(request MCPMutationAuthorizationRequest) bool {
	if request.ProjectRoot == "" || !filepath.IsAbs(request.ProjectRoot) ||
		filepath.Clean(request.ProjectRoot) != request.ProjectRoot ||
		!validAuthorizationID(request.PlanID) ||
		!sha256DigestPattern.MatchString(request.PlanDigest) ||
		!validAuthorizationID(request.Caller) {
		return false
	}
	switch request.Operation {
	case "apply", "configure", "deploy":
		return request.Intent == "apply"
	case "destroy":
		return request.Intent == "destroy"
	case "reconcile":
		return request.Intent == "reconcile"
	case "template-lock", "template-update":
		return request.Intent == request.Operation
	default:
		return false
	}
}

func validAuthorizationID(value string) bool {
	if len(value) < 1 || len(value) > 128 || value == "." || value == ".." {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' ||
			character == '.' || character == '_' || character == '-' || character == ':' {
			continue
		}
		return false
	}
	return true
}
