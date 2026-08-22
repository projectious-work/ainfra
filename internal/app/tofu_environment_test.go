package app

import (
	"slices"
	"testing"
)

func TestBuildTofuEnvironmentAllowsHetznerCredentialOnly(t *testing.T) {
	t.Parallel()
	environment, err := buildTofuEnvironment([]string{
		"HOME=/tmp/operator",
		"HCLOUD_TOKEN=provider-secret",
		"UNRELATED_SECRET=must-not-pass",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"HCLOUD_TOKEN=provider-secret",
		"HOME=/tmp/operator",
		"TF_IN_AUTOMATION=1",
	}
	if !slices.Equal(environment.Values(), want) {
		t.Fatalf("environment = %#v, want %#v", environment.Values(), want)
	}
}
