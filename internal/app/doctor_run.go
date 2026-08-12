package app

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/projectious-work/ainfra/internal/diagnostic"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/security"
)

// DoctorRunRequest identifies the deployment whose latest retained run is read.
type DoctorRunRequest = DoctorDeploymentRequest

// DoctorRun validates only the retained local run layout owned by a deployment.
func DoctorRun(
	request DoctorRunRequest,
	options DoctorEnvironmentOptions,
) (DoctorEnvironmentResponse, error) {
	if request.Target != "" && request.ProjectPath != "" {
		return DoctorEnvironmentResponse{}, errors.New("run TARGET conflicts with --project")
	}
	deployment, err := project.Load(project.ResolveOptions{
		WorkingDirectory: options.WorkingDirectory, ExplicitPath: request.Target,
		ProjectPath:     request.ProjectPath,
		EnvironmentPath: options.Environment["AINFRA_PROJECT"],
	})
	if err != nil {
		return DoctorEnvironmentResponse{}, classifyDoctorLoad("deployment", err)
	}
	effective, err := resolveDoctorConfiguration(
		DoctorEnvironmentRequest{
			ConfigPath: request.ConfigPath, Format: request.Format,
			OutputStyle: request.OutputStyle, Color: request.Color,
		}, options, deployment.Target.Root,
	)
	if err != nil {
		return DoctorEnvironmentResponse{}, err
	}
	finding := latestRunFinding(deployment.Target.Root)
	summary := output.DoctorSummary{}
	switch finding.Status {
	case "pass":
		summary.Pass = 1
	case "skip":
		summary.Skip = 1
	case "warning":
		summary.Warning = 1
	case "fail":
		summary.Fail = 1
	}
	return DoctorEnvironmentResponse{
		Result: output.Doctor{
			Scope: "run", Summary: summary,
			Findings: []diagnostic.Diagnostic{finding},
		},
		Format:      effective.Settings.UI.Format,
		OutputStyle: effective.Settings.UI.OutputStyle,
		Color:       effective.Settings.UI.Color,
	}, nil
}

func latestRunFinding(deploymentRoot string) diagnostic.Diagnostic {
	base := diagnostic.Diagnostic{
		Code: "AINFRA-E2501", Check: "run.latest-evidence", Scope: "run",
		Component: "run", Reconciliation: "not_available",
	}
	root, err := os.OpenRoot(deploymentRoot)
	if err != nil {
		base.Severity, base.Status = diagnostic.SeverityError, "fail"
		base.Message = "deployment root is unreadable"
		base.NextAction = "Correct deployment-directory permissions."
		return base
	}
	defer func() { _ = root.Close() }()
	runsDirectory, err := root.Open(".ainfra/runs")
	if errors.Is(err, os.ErrNotExist) {
		base.Severity, base.Status = diagnostic.SeverityInfo, "skip"
		base.Message = "no retained run is available"
		base.NextAction = "Run a lifecycle command before diagnosing retained run evidence."
		return base
	}
	defer func() { _ = runsDirectory.Close() }()
	entries, err := runsDirectory.ReadDir(-1)
	if err != nil {
		base.Severity, base.Status = diagnostic.SeverityError, "fail"
		base.Message = "retained run directory is unreadable"
		base.NextAction = "Correct local run-directory permissions."
		return base
	}
	if err != nil {
		base.Severity, base.Status = diagnostic.SeverityError, "fail"
		base.Message, base.NextAction = "retained run directory is unreadable", "Correct local run-directory permissions."
		return base
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && entry.Type()&os.ModeSymlink == 0 {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 0 {
		base.Severity, base.Status = diagnostic.SeverityInfo, "skip"
		base.Message, base.NextAction = "no retained run is available", "Run a lifecycle command before diagnosing retained run evidence."
		return base
	}
	sort.Strings(names)
	latestName := names[len(names)-1]
	for _, file := range []string{"run.json", "events.jsonl"} {
		path := ".ainfra/runs/" + latestName + "/" + file
		handle, openErr := root.Open(path)
		if openErr != nil {
			base.Severity, base.Status = diagnostic.SeverityError, "fail"
			base.Path = path
			base.Message = fmt.Sprintf("latest retained run is missing valid %s", file)
			base.NextAction = "Preserve the run directory and inspect the interrupted evidence manually."
			return base
		}
		information, statErr := handle.Stat()
		closeErr := handle.Close()
		if statErr != nil || closeErr != nil || !information.Mode().IsRegular() {
			base.Severity, base.Status = diagnostic.SeverityError, "fail"
			base.Path = path
			base.Message = fmt.Sprintf("latest retained run is missing valid %s", file)
			base.NextAction = "Preserve the run directory and inspect the interrupted evidence manually."
			return base
		}
	}
	base.Severity, base.Status = diagnostic.SeverityInfo, "pass"
	base.Path = ".ainfra/runs/" + latestName
	base.Message = "latest retained run has the required local evidence files"
	return base
}

func classifyDoctorLoad(subject string, err error) error {
	kind := "input"
	var refusal *security.Refusal
	if errors.As(err, &refusal) {
		kind = "security"
	}
	return &DoctorError{Kind: kind, Message: fmt.Sprintf("load %s contract: %s", subject, err)}
}
