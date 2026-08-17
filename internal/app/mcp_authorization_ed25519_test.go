package app

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEd25519MCPAuthorizationProviderVerifiesCanonicalGrant(t *testing.T) {
	t.Parallel()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	trustPath := filepath.Join(t.TempDir(), "trust.json")
	writeMCPAuthorizationJSON(t, trustPath, mcpTrustStoreDocument{SchemaVersion: 1,
		Issuers: []mcpTrustIssuer{{ID: "operator-1",
			PublicKey: base64.StdEncoding.EncodeToString(publicKey)}}})
	provider, err := LoadMCPAuthorizationTrust(trustPath)
	if err != nil {
		t.Fatal(err)
	}
	grant := mcpApprovalGrant{AuthorizationID: "approval-1", Issuer: "operator-1",
		Caller: "agent-1", ProjectRoot: "/project", Operation: "apply",
		PlanID:     "20260816T120000Z-0123456789abcdef",
		PlanDigest: "sha256:" + strings.Repeat("a", 64), Intent: "apply",
		ApprovedAt: "2026-08-16T12:00:00Z", ExpiresAt: "2026-08-16T12:05:00Z"}
	approval := signedMCPApproval(t, privateKey, grant)
	verified, err := provider.Verify(context.Background(), approval)
	if err != nil {
		t.Fatal(err)
	}
	if verified.AuthorizationID != grant.AuthorizationID || verified.Issuer != grant.Issuer ||
		verified.PlanDigest != grant.PlanDigest ||
		!verified.ExpiresAt.Equal(time.Date(2026, 8, 16, 12, 5, 0, 0, time.UTC)) {
		t.Fatalf("verified grant: %+v", verified)
	}

	var envelope mcpApprovalEnvelope
	if err := json.Unmarshal([]byte(approval), &envelope); err != nil {
		t.Fatal(err)
	}
	var tamperedGrant mcpApprovalGrant
	if err := json.Unmarshal(envelope.Grant, &tamperedGrant); err != nil {
		t.Fatal(err)
	}
	tamperedGrant.Caller = "attacker"
	envelope.Grant, err = json.Marshal(tamperedGrant)
	if err != nil {
		t.Fatal(err)
	}
	tampered, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Verify(context.Background(), string(tampered)); err == nil {
		t.Fatal("tampered approval succeeded")
	}
	tamperedGrant.Caller = grant.Caller
	tamperedGrant.Issuer = "unknown"
	envelope.Grant, err = json.Marshal(tamperedGrant)
	if err != nil {
		t.Fatal(err)
	}
	unknown, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Verify(context.Background(), string(unknown)); err == nil {
		t.Fatal("unknown issuer succeeded")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := provider.Verify(ctx, approval); err == nil {
		t.Fatal("cancelled verification succeeded")
	}
}

func TestLoadMCPAuthorizationTrustFailsClosed(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	relative := filepath.Join("relative", "trust.json")
	if _, err := LoadMCPAuthorizationTrust(relative); err == nil {
		t.Fatal("relative trust path succeeded")
	}
	trustPath := filepath.Join(root, "trust.json")
	writeMCPAuthorizationJSON(t, trustPath, mcpTrustStoreDocument{SchemaVersion: 1,
		Issuers: []mcpTrustIssuer{{ID: "operator-1", PublicKey: "invalid"}}})
	if _, err := LoadMCPAuthorizationTrust(trustPath); err == nil {
		t.Fatal("invalid trust key succeeded")
	}
	if err := os.Chmod(trustPath, 0o620); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadMCPAuthorizationTrust(trustPath); err == nil {
		t.Fatal("group-writable trust store succeeded")
	}
	if err := os.Chmod(trustPath, 0o600); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(root, "trust-link.json")
	if err := os.Symlink(trustPath, symlink); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadMCPAuthorizationTrust(symlink); err == nil {
		t.Fatal("symlink trust path succeeded")
	}
}

func signedMCPApproval(t *testing.T, privateKey ed25519.PrivateKey,
	grant mcpApprovalGrant,
) string {
	t.Helper()
	payload, err := json.Marshal(grant)
	if err != nil {
		t.Fatal(err)
	}
	signature := ed25519.Sign(privateKey,
		append([]byte(mcpAuthorizationDomain), payload...))
	contents, err := json.Marshal(mcpApprovalEnvelope{SchemaVersion: 1, Grant: payload,
		Signature: base64.StdEncoding.EncodeToString(signature)})
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

func writeMCPAuthorizationJSON(t *testing.T, path string, value any) {
	t.Helper()
	contents, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}
}
