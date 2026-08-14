package tofu_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	childexec "github.com/projectious-work/ainfra/internal/exec"
	"github.com/projectious-work/ainfra/internal/tofu"
)

func TestAdapterPreservesNativeInputOrderAndIntent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, name := range []string{"backend-one.hcl", "backend-two.hcl", "one.tfvars", "two.tfvars"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("fixture"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "tofu"), 0o700); err != nil {
		t.Fatal(err)
	}
	var requests []childexec.Request
	adapter := tofu.Adapter{Run: func(_ context.Context, request childexec.Request) (childexec.Result, error) {
		requests = append(requests, request)
		return childexec.Result{Started: true}, nil
	}}
	if _, err := adapter.Init(context.Background(), root, "tofu", []string{"../backend-one.hcl", "../backend-two.hcl"}); err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.Plan(context.Background(), root, "tofu", "../plan.tfplan", []string{"../one.tfvars", "../two.tfvars"}, true); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(requests[0].Args, []string{"init", "-input=false", "-no-color", "-backend-config=../backend-one.hcl", "-backend-config=../backend-two.hcl"}) {
		t.Fatalf("init args: %#v", requests[0].Args)
	}
	if !reflect.DeepEqual(requests[1].Args, []string{"plan", "-input=false", "-no-color", "-out=../plan.tfplan", "-destroy", "-var-file=../one.tfvars", "-var-file=../two.tfvars"}) {
		t.Fatalf("plan args: %#v", requests[1].Args)
	}
}

func TestApplyUsesOnlyExactSavedPlan(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "tofu"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "plan.tfplan"), []byte("plan"), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter := tofu.Adapter{Run: func(_ context.Context, request childexec.Request) (childexec.Result, error) {
		want := []string{"apply", "-input=false", "-no-color", "../plan.tfplan"}
		if !reflect.DeepEqual(request.Args, want) {
			t.Fatalf("apply args=%#v", request.Args)
		}
		return childexec.Result{Started: true}, nil
	}}
	if _, err := adapter.Apply(context.Background(), root, "tofu", "../plan.tfplan"); err != nil {
		t.Fatal(err)
	}
}

func TestShowSummaryRetainsOnlyStructuralActionCounts(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "tofu"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "plan.tfplan"), []byte("sensitive"), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter := tofu.Adapter{Run: func(_ context.Context, request childexec.Request) (childexec.Result, error) {
		if strings.Join(request.Args, " ") != "show -json ../plan.tfplan" {
			t.Fatalf("show args: %#v", request.Args)
		}
		_, _ = request.IO.Stdout.Write([]byte(`{"resource_changes":[{"change":{"actions":["create"]}},{"change":{"actions":["delete","create"]}},{"change":{"actions":["no-op"]}}],"secret":"must-not-survive"}`))
		return childexec.Result{Started: true}, nil
	}}
	summary, _, err := adapter.ShowSummary(context.Background(), root, "tofu", "../plan.tfplan")
	if err != nil {
		t.Fatal(err)
	}
	if summary.Create != 1 || summary.Replace != 1 || summary.NoOp != 1 {
		t.Fatalf("summary: %#v", summary)
	}
}

func TestAdapterRejectsUnsafePathsAndFailedChildren(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "tofu"), 0o700); err != nil {
		t.Fatal(err)
	}
	adapter := tofu.Adapter{Run: func(context.Context, childexec.Request) (childexec.Result, error) {
		return childexec.Result{Started: true, ExitCode: 7}, nil
	}}
	if _, err := adapter.Init(context.Background(), root, "tofu", []string{"../escape"}); err == nil {
		t.Fatal("unsafe backend path accepted")
	}
	if _, err := adapter.Plan(context.Background(), root, "tofu", "plan.tfplan", nil, false); err == nil {
		t.Fatal("failed child accepted")
	}
}
