package tofu_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
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
	if _, err := adapter.Init(context.Background(), root, "tofu", []string{"backend-one.hcl", "backend-two.hcl"}); err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.Plan(context.Background(), root, "tofu", "plan.tfplan", []string{"one.tfvars", "two.tfvars"}, true); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(requests[0].Args, []string{"init", "-input=false", "-no-color", "-backend-config=backend-one.hcl", "-backend-config=backend-two.hcl"}) {
		t.Fatalf("init args: %#v", requests[0].Args)
	}
	if !reflect.DeepEqual(requests[1].Args, []string{"plan", "-input=false", "-no-color", "-out=plan.tfplan", "-destroy", "-var-file=one.tfvars", "-var-file=two.tfvars"}) {
		t.Fatalf("plan args: %#v", requests[1].Args)
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
