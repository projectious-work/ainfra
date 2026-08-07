// Package command defines CLI syntax, input conversion, and renderer selection.
package command

// ExitCode is a stable process outcome defined by the CLI contract.
type ExitCode int

const (
	// ExitSuccess reports a successful requested operation.
	ExitSuccess ExitCode = iota
	// ExitOperationFailed reports an operation or child-engine failure.
	ExitOperationFailed
	// ExitInvalidInput reports an invalid command or input contract.
	ExitInvalidInput
	// ExitDependency reports a missing or incompatible dependency.
	ExitDependency
	// ExitSecurity reports a security or trust refusal.
	ExitSecurity
	// ExitStaleBinding reports a stale reviewed-plan binding.
	ExitStaleBinding
	// ExitInterrupted reports an ambiguous interrupted state needing recovery.
	ExitInterrupted
)
