package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/projectious-work/ainfra/internal/doctor"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/security"
	templatecontract "github.com/projectious-work/ainfra/internal/template"
)

// DoctorTemplateRequest identifies one already-resolved local template.
type DoctorTemplateRequest struct {
	Target      string
	ConfigPath  string
	Format      *string
	OutputStyle *string
	Color       *string
}

// DoctorTemplate validates a local template without acquiring or updating it.
func DoctorTemplate(
	request DoctorTemplateRequest,
	options DoctorEnvironmentOptions,
) (DoctorEnvironmentResponse, error) {
	target := request.Target
	if target == "" {
		target = options.WorkingDirectory
	} else if !filepath.IsAbs(target) {
		target = filepath.Join(options.WorkingDirectory, target)
	}
	contract, err := templatecontract.Load(target)
	if err != nil {
		kind := "input"
		var refusal *security.Refusal
		if errors.As(err, &refusal) {
			kind = "security"
		}
		return DoctorEnvironmentResponse{}, &DoctorError{
			Kind: kind, Message: fmt.Sprintf("load template contract: %s", err),
		}
	}
	effective, err := resolveDoctorConfiguration(
		DoctorEnvironmentRequest{
			ConfigPath: request.ConfigPath, Format: request.Format,
			OutputStyle: request.OutputStyle, Color: request.Color,
		},
		options, "",
	)
	if err != nil {
		return DoctorEnvironmentResponse{}, err
	}
	report := doctor.TemplateRegistry(doctor.TemplateInput{
		Name: contract.Name, Version: contract.Version, Root: contract.Root,
		HasAnsible: contract.Ansible != nil, Inventory: contract.Inventory,
	}).Run(context.Background(), doctor.ScopeTemplate, doctor.Input{}, doctor.Capabilities{})
	return DoctorEnvironmentResponse{
		Result: output.Doctor{
			Scope: "template",
			Summary: output.DoctorSummary{
				Pass: report.Summary.Pass, Skip: report.Summary.Skip,
				Warning: report.Summary.Warning, Fail: report.Summary.Fail,
			},
			Findings: report.Findings,
		},
		Format:      effective.Settings.UI.Format,
		OutputStyle: effective.Settings.UI.OutputStyle,
		Color:       effective.Settings.UI.Color,
	}, nil
}
