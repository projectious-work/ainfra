package app

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/projectious-work/ainfra/internal/security"
)

const (
	mcpAuthorizationDomain = "ainfra-mcp-authorization-v1\n"
	maxMCPTrustStoreBytes  = 1 << 20
)

type mcpTrustStoreDocument struct {
	SchemaVersion int              `json:"schemaVersion"`
	Issuers       []mcpTrustIssuer `json:"issuers"`
}

type mcpTrustIssuer struct {
	ID        string `json:"id"`
	PublicKey string `json:"publicKey"`
}

type mcpApprovalEnvelope struct {
	SchemaVersion int             `json:"schemaVersion"`
	Grant         json.RawMessage `json:"grant"`
	Signature     string          `json:"signature"`
}

type mcpApprovalGrant struct {
	AuthorizationID string `json:"authorizationId"`
	Issuer          string `json:"issuer"`
	Caller          string `json:"caller"`
	ProjectRoot     string `json:"projectRoot"`
	Operation       string `json:"operation"`
	PlanID          string `json:"planId"`
	PlanDigest      string `json:"planDigest"`
	Intent          string `json:"intent"`
	ApprovedAt      string `json:"approvedAt"`
	ExpiresAt       string `json:"expiresAt"`
}

type ed25519MCPAuthorizationProvider struct {
	issuers map[string]ed25519.PublicKey
}

// LoadMCPAuthorizationTrust loads a fixed Ed25519 issuer trust store. The
// startup path must be absolute and name a real regular file, never a symlink.
func LoadMCPAuthorizationTrust(path string) (MCPAuthorizationProvider, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, errors.New("MCP authorization trust path must be absolute and clean")
	}
	if err := security.RequireRegular(path); err != nil {
		return nil, fmt.Errorf("validate MCP authorization trust store: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect MCP authorization trust store: %w", err)
	}
	if info.Mode().Perm()&0o022 != 0 {
		return nil, errors.New("MCP authorization trust store is writable by another user")
	}
	contents, err := os.ReadFile(path) // #nosec G304 -- validated startup-only operator path.
	if err != nil {
		return nil, fmt.Errorf("read MCP authorization trust store: %w", err)
	}
	if len(contents) == 0 || len(contents) > maxMCPTrustStoreBytes {
		return nil, errors.New("MCP authorization trust store has invalid size")
	}
	var document mcpTrustStoreDocument
	if err := decodeStrictMCPJSON(contents, &document); err != nil {
		return nil, errors.New("MCP authorization trust store is invalid")
	}
	if document.SchemaVersion != 1 || len(document.Issuers) == 0 ||
		len(document.Issuers) > 128 {
		return nil, errors.New("MCP authorization trust store is invalid")
	}
	issuers := make(map[string]ed25519.PublicKey, len(document.Issuers))
	for _, issuer := range document.Issuers {
		key, decodeErr := base64.StdEncoding.Strict().DecodeString(issuer.PublicKey)
		if !validAuthorizationID(issuer.ID) || decodeErr != nil ||
			len(key) != ed25519.PublicKeySize {
			return nil, errors.New("MCP authorization trust store is invalid")
		}
		if _, duplicate := issuers[issuer.ID]; duplicate {
			return nil, errors.New("MCP authorization trust store contains duplicate issuer")
		}
		issuers[issuer.ID] = ed25519.PublicKey(append([]byte(nil), key...))
	}
	return ed25519MCPAuthorizationProvider{issuers: issuers}, nil
}

func (provider ed25519MCPAuthorizationProvider) Verify(ctx context.Context,
	approval string,
) (MCPAuthorizationGrant, error) {
	if err := ctx.Err(); err != nil {
		return MCPAuthorizationGrant{}, err
	}
	var envelope mcpApprovalEnvelope
	if decodeStrictMCPJSON([]byte(approval), &envelope) != nil || envelope.SchemaVersion != 1 {
		return MCPAuthorizationGrant{}, errors.New("invalid Ed25519 MCP approval envelope")
	}
	var grant mcpApprovalGrant
	if decodeStrictMCPJSON(envelope.Grant, &grant) != nil {
		return MCPAuthorizationGrant{}, errors.New("invalid Ed25519 MCP approval grant")
	}
	key, trusted := provider.issuers[grant.Issuer]
	signature, err := base64.StdEncoding.Strict().DecodeString(envelope.Signature)
	if !trusted || err != nil || len(signature) != ed25519.SignatureSize {
		return MCPAuthorizationGrant{}, errors.New("untrusted Ed25519 MCP approval envelope")
	}
	if !ed25519.Verify(key, append([]byte(mcpAuthorizationDomain), envelope.Grant...), signature) {
		return MCPAuthorizationGrant{}, errors.New("invalid Ed25519 MCP approval signature")
	}
	approvedAt, approvedErr := time.Parse(time.RFC3339, grant.ApprovedAt)
	expiresAt, expiresErr := time.Parse(time.RFC3339, grant.ExpiresAt)
	if approvedErr != nil || expiresErr != nil {
		return MCPAuthorizationGrant{}, errors.New("invalid Ed25519 MCP approval time")
	}
	return MCPAuthorizationGrant{AuthorizationID: grant.AuthorizationID,
		Issuer: grant.Issuer, Caller: grant.Caller, ProjectRoot: grant.ProjectRoot,
		Operation: grant.Operation, PlanID: grant.PlanID, PlanDigest: grant.PlanDigest,
		Intent: grant.Intent, ApprovedAt: approvedAt, ExpiresAt: expiresAt}, nil
}

func decodeStrictMCPJSON(contents []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("document must contain exactly one JSON value")
	}
	return nil
}
