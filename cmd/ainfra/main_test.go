package main

import "testing"

func TestApplyLoggingFlagsUsesClosedVerbosityPrecedence(t *testing.T) {
	t.Parallel()
	environment := map[string]string{}
	if err := applyLoggingFlags([]string{"status", "-vv", "--log-format", "json",
		"--log-file", "/tmp/ainfra.log", "--syslog"}, environment); err != nil {
		t.Fatal(err)
	}
	if environment["AINFRA_LOG_LEVEL"] != "debug" ||
		environment["AINFRA_LOG_FORMAT"] != "json" ||
		environment["AINFRA_LOG_FILE"] != "/tmp/ainfra.log" ||
		environment["AINFRA_LOG_SYSLOG"] != "true" {
		t.Fatalf("logging environment=%v", environment)
	}
	for _, arguments := range [][]string{{"-vvvv"}, {"-vv", "-vv"},
		{"-v", "--log-level", "trace"}, {"--log-file"}} {
		if err := applyLoggingFlags(arguments, map[string]string{}); err == nil {
			t.Fatalf("invalid logging flags accepted: %v", arguments)
		}
	}
}

func TestStructuredOutputDetection(t *testing.T) {
	t.Parallel()
	for _, arguments := range [][]string{{"--format=json", "status"},
		{"status", "--format", "yaml"}} {
		if !structuredOutput(arguments) {
			t.Errorf("structured output not detected: %v", arguments)
		}
	}
	if structuredOutput([]string{"status", "--format", "text"}) {
		t.Error("text output detected as structured")
	}
}
