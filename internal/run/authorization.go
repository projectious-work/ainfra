package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/projectious-work/ainfra/internal/security"
)

// AuthorizationBinding is the minimum immutable saved-plan identity needed
// before an independent authorization provider may be consulted.
type AuthorizationBinding struct {
	PlanID     string
	PlanDigest string
	Intent     string
}

// AuthorizationRecord is sanitized durable evidence for one MCP approval.
type AuthorizationRecord struct {
	SchemaVersion   int    `json:"schemaVersion"`
	AuthorizationID string `json:"authorizationId"`
	Operation       string `json:"operation"`
	PlanID          string `json:"planId"`
	PlanDigest      string `json:"planDigest"`
	Caller          string `json:"caller"`
	Issuer          string `json:"issuer"`
	ExpiresAt       string `json:"expiresAt"`
	RecordedAt      string `json:"recordedAt"`
}

// LoadAuthorizationBinding strictly reads a plan record without invoking an
// engine. Full plan reverification still occurs under the operation lock.
func LoadAuthorizationBinding(runsRoot, id, deploymentName, intent string) (
	AuthorizationBinding, error,
) {
	if !validID(id) || deploymentName == "" || intent != "apply" && intent != "destroy" {
		return AuthorizationBinding{}, errors.New("invalid authorization binding request")
	}
	root, err := security.ResolveContained(runsRoot, id)
	if err != nil {
		return AuthorizationBinding{}, fmt.Errorf("resolve authorization plan: %w", err)
	}
	var record PlanRecord
	if err := readStrictJSON(root, "plan-record.json", &record); err != nil {
		return AuthorizationBinding{}, fmt.Errorf("read authorization plan: %w", err)
	}
	if record.SchemaVersion != 1 || record.RunID != id || record.Intent != intent ||
		record.Deployment.Name != deploymentName || !validSHA256(record.Plan.Digest) {
		return AuthorizationBinding{}, errors.New("saved plan authorization binding is invalid")
	}
	return AuthorizationBinding{PlanID: id, PlanDigest: record.Plan.Digest,
		Intent: record.Intent}, nil
}

// RecordAuthorization exclusively retains sanitized approval evidence before
// execution. A second attempt cannot overwrite or reuse the authorization.
func RecordAuthorization(runsRoot string, record AuthorizationRecord) error {
	return RecordOperationAuthorization(runsRoot, "authorization.json", record)
}

// RecordOperationAuthorization exclusively retains sanitized approval
// evidence under an operation-specific fixed filename.
func RecordOperationAuthorization(runsRoot, name string, record AuthorizationRecord) error {
	valid := false
	switch name {
	case "authorization.json":
		valid = record.Operation == "apply" || record.Operation == "destroy"
	case "authorization-configure.json":
		valid = record.Operation == "configure"
	case "authorization-deploy.json":
		valid = record.Operation == "deploy"
	}
	if !valid {
		return errors.New("invalid authorization evidence name")
	}
	root, err := security.ResolveContained(runsRoot, record.PlanID)
	if err != nil {
		return fmt.Errorf("resolve authorization evidence: %w", err)
	}
	contents, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	file, err := security.CreatePrivateFile(root, name)
	if err != nil {
		return fmt.Errorf("create authorization evidence: %w", err)
	}
	_, writeErr := file.Write(append(contents, '\n'))
	return errors.Join(writeErr, file.Sync(), file.Close())
}

// NewAuthorizationRecord builds the closed evidence shape after application
// authorization validation has removed opaque approval material.
func NewAuthorizationRecord(authorizationID, operation, planID, planDigest,
	caller, issuer, expiresAt string, recordedAt time.Time,
) AuthorizationRecord {
	return AuthorizationRecord{SchemaVersion: 1, AuthorizationID: authorizationID,
		Operation: operation, PlanID: planID, PlanDigest: planDigest, Caller: caller,
		Issuer: issuer, ExpiresAt: expiresAt,
		RecordedAt: recordedAt.UTC().Format(time.RFC3339Nano)}
}

func validSHA256(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, character := range value[len("sha256:"):] {
		if character < '0' || character > '9' && character < 'a' || character > 'f' {
			return false
		}
	}
	return true
}
