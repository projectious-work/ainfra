## Basis and enforcement

ainfra follows [Effective Go](https://go.dev/doc/effective_go),
[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), and the
[Go documentation conventions](https://go.dev/doc/comment). This chapter fixes
project choices where general Go guidance permits several reasonable styles.

`gofmt`, `goimports`, compiler errors, vet, static analysis, and tests enforce
mechanical rules. Review concentrates on names, ownership, boundaries,
correctness, and clarity that tools cannot decide.

- **AINFRA-CODE-001:** committed Go source MUST pass `gofmt` and `goimports`.
- **AINFRA-CODE-002:** lint suppression MUST be narrow and include a reason;
  repository-wide convenience suppressions are prohibited.
- **AINFRA-CODE-003:** generated source MUST identify its generator, be
  reproducible, and never be edited manually.

## Packages and files

Package names are short, lower-case, singular words without underscores or
generic buckets such as `util`, `common`, `helpers`, or `misc`. A package owns
one responsibility explainable in one sentence. Importers SHOULD read without
stutter: `diagnostic.Report`, not `diagnostic.DiagnosticReport`.

Production packages remain under `internal/` until a real external consumer
justifies a separately versioned API. Files are named for a use case or concept,
such as `plan.go`, `reconcile.go`, or `git_source.go`; `manager.go` and
`utils.go` are review signals.

Imports are arranged by `goimports`. Dot imports are prohibited. Blank imports
require an adjacent rationale. Aliases are used only for a real collision or
misleading imported name.

Package-level mutable state is prohibited. Constants and immutable tables are
acceptable. Production `init` functions require design review; construction
and registration SHOULD be visible in the composition root.

## Names

Identifiers use `MixedCaps`, never underscores. Initialisms retain consistent
capitalization: `ID`, `URL`, `HTTP`, `SSH`, `JSON`, and `API`, producing
`runID`, `templateURL`, and `parseJSON`.

- Exported names are precise and stable package contracts.
- Local name length is proportional to scope; one-letter names are limited to
  conventional tiny scopes such as indexes and receivers.
- Receiver names are short and consistent; `this` and `self` are prohibited.
- Boolean names state positive predicates such as `enabled` or `hasPlan`.
- Getters omit `Get` when a natural name is available.
- Single-operation interfaces normally use a natural `-er` name.
- Domain terminology uses one documented spelling across code and schemas.

Names describe domain intent rather than mechanics. `Manager`, `Processor`,
`Handler`, `Data`, and `Info` require a more precise responsibility unless they
are established protocol terms.

## Functions and control flow

Functions do one coherent job and make ownership and side effects visible.
There is no arbitrary line limit, but mixed abstraction levels, unrelated
branches, or a name requiring “and” are extraction signals.

Use guard clauses and early returns to keep the successful path readable. Avoid
deep nesting, clever boolean expressions, and hidden mutation. Use a switch for
a closed decision when it is clearer than mutually exclusive `if` chains.

Group related parameters into a typed options value when several travel
together or booleans become ambiguous. Context, when required, is the first
parameter and is never stored in a long-lived struct.

Pointers express mutation, identity, optionality, or impractical copy cost—not
habit. Copy slices, maps, bytes, and option values at ownership boundaries when
later mutation could violate an invariant.

## Types, interfaces, and APIs

Prefer concrete types and useful zero values. Constructors establish invariants
or dependencies, not merely allocate structs. Fields remain unexported unless
direct mutation is a deliberate stable contract.

Interfaces are defined by the consumer and contain only its required
operations. Prefer accepting a narrow interface and returning a concrete type.
Do not create an interface solely to mock a type; introduce it at a real domain,
process, filesystem, time, or terminal boundary.

Use typed constants for closed states and operations. Safety-relevant switches
have an explicit default refusal or a test that catches new values. Persistence
conversions are explicit and versioned.

Generics are used only when they remove genuine type-safe duplication without
hiding domain meaning. Reflection and `unsafe` require a recorded design and
security rationale. CGO remains prohibited unless a reviewed dependency makes
it unavoidable.

## Errors and diagnostics

Return errors; do not panic for input, filesystem, child-process, network, or
recoverable operational failures. Panic is reserved for impossible internal
invariants and composition-time programmer errors. Packages MUST NOT call
`os.Exit`; only `cmd/ainfra` maps the final result to an exit code.

Errors add concise context and preserve causes with `%w` when callers may use
`errors.Is` or `errors.As`. Error strings start lower-case and omit terminal
punctuation. Do not log and return the same error at every layer: the owning
boundary adds context and the command boundary chooses presentation.

Sentinel and typed errors exist only for stable caller logic. Human wording is
not an API. User failures become typed diagnostics with a stable code, safe
evidence, and next action. Redaction occurs before errors, logs, diagnostics,
or renderers receive sensitive values.

Ignored errors require a reason. Cleanup errors are joined or reported when
they affect correctness; a deferred close MUST NOT hide a failed durable write.

## Documentation and comments

Every package has a package comment explaining its responsibility and boundary.
Every exported identifier has a complete-sentence doc comment describing
observable behavior, important errors, ownership, concurrency safety, and side
effects.

Comments explain why, invariants, security constraints, protocols, or non-
obvious tradeoffs. They do not translate obvious code into prose. Workarounds
link an issue or upstream source. TODOs include an issue and do not serve as
indefinite design storage.

Complex packages SHOULD include a short example. Public behavior is documented
in user docs and the active phase note as implemented. A code change that
invalidates a comment or example updates it in the same change.

- **AINFRA-CODE-010:** exported contracts and non-obvious invariants MUST be
  documented before review approval.
- **AINFRA-CODE-011:** comments MUST NOT claim behavior unenforced by code or
  tests.

## Dependencies and construction

Pass dependencies explicitly through constructors or use-case methods. Service
locators, hidden global registries, ambient mutable configuration, and package
singletons are prohibited. The composition root selects concrete adapters.

Prefer the standard library when it remains clear and testable. A third-party
dependency must provide maintained behavior risky or wasteful to reproduce.
Wrap a dependency only to establish an ainfra boundary, not automatically.

Parse configuration once into an immutable typed value. Domain packages do not
read environment variables or global flags. Filesystem, process, clock,
randomness, and terminal capabilities enter through explicit narrow boundaries.

## Concurrency and cancellation

Sequential code is the default. Concurrency is introduced only for independent
work with a measurable benefit. Every goroutine has an owner, cancellation,
completion, and error path. Fire-and-forget goroutines are prohibited.

Prefer ownership transfer through channels over shared mutation. When shared
state is simpler, protect it with the narrowest lock and document the invariant.
Do not copy mutex-containing values. Producers own channel closure.

Cancellation propagates through `context.Context` and process signals. No
blocking channel, process wait, or IO loop may hide cancellation indefinitely.
Concurrency changes require race, cancellation, partial-output, and shutdown
tests.

## Logging and output

Production packages do not print with `fmt.Print*`, `log.Print*`, or ad hoc
terminal escapes. They return typed results, diagnostics, and events. Only
renderers write command output; only configured sinks write operational logs.

Log messages are concise present-tense facts with structured fields. A field
name has one type and meaning across the codebase. Verbosity checks belong in
logging rather than scattered domain conditionals.

## Clean-change rules

- Separate refactoring from behavior changes when it improves reviewability.
- Delete dead code; do not comment it out or retain speculative abstractions.
- Do not generalize for roadmap phases that remain only potential.
- Prefer straightforward duplication over a premature abstraction.
- Preserve behavior unless specs, migration notes, tests, and changelog change.
- Leave the touched package simpler without expanding unrelated scope.

These are design signals, not incentives for mechanical micro-functions or
excessive interfaces. Readability and correct boundaries are the objective.
