package app

import "testing"

func TestNotApplicableDeployStagesRecordEverySkippedStage(t *testing.T) {
	t.Parallel()
	stages := notApplicableDeployStages("inventory mode none")
	want := []string{"output", "inventory", "configure", "configure-check"}
	if len(stages) != len(want) {
		t.Fatalf("stages=%+v", stages)
	}
	for index, stage := range stages {
		if stage.Operation != want[index] || stage.Applicability != "not-applicable" ||
			stage.Status != "not-applicable" || stage.Reason == "" {
			t.Fatalf("stage[%d]=%+v", index, stage)
		}
	}
}
