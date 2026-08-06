## Output, logging, and evidence are separate

ainfra has three distinct information channels:

- **command output** communicates the requested result to a human or program;
- **operational logs** explain CLI decisions and child-process activity; and
- **run evidence** is the durable lifecycle record described elsewhere.

Command output goes to stdout. Logs and diagnostics go to configured sinks,
with stderr as the default sink. Run evidence goes only to the run store. A log
file or syslog stream MUST NOT be treated as authoritative run evidence, and
rotating a log MUST NOT remove run evidence.

## Console output

The text renderer SHOULD provide a polished terminal experience comparable to
Python Rich while remaining a native Go implementation. Rendering starts from
semantic result objects, not strings containing terminal escape sequences. The
renderer MAY use panels, tables, trees, status glyphs, spinners, progress bars,
and syntax-aware diffs when they materially improve comprehension.

Global output controls are:

| Option | Values | Purpose |
|---|---|---|
| `--format` | `text`, `json` | Select human-oriented text or the versioned machine-result envelope. |
| `--output-style` | `auto`, `rich`, `plain` | Select decorated terminal rendering or stable undecorated text. |
| `--color` | `auto`, `always`, `never` | Control ANSI color independently of output style. |

`auto` selects rich output only when stdout is an interactive terminal with
adequate capabilities. Pipes, redirected output, dumb terminals, and unknown
terminal capabilities select plain output. `NO_COLOR` and `AINFRA_COLOR=never`
disable color; an explicit `--color` flag has normal flag precedence.

Rich output MUST degrade safely when Unicode, color, cursor movement, or width
detection is unavailable. Color MUST never be the only carrier of status.
Status symbols MUST have adjacent text, tables MUST remain understandable in
plain output, and narrow terminals MUST wrap rather than truncate identifiers,
paths, or required next actions. Progress rendering MUST not leave stale lines
after success, failure, or cancellation.

JSON output disables panels, progress animation, color, and terminal control
sequences. Plain text MUST be stable enough for human comparison, but scripts
MUST use JSON rather than parse prose. Prompts go to the controlling terminal
where possible and MUST never contaminate JSON stdout.

- **AINFRA-OUTPUT-001:** all commands MUST render from shared typed semantic
  results so rich, plain, and JSON modes communicate equivalent outcomes.
- **AINFRA-OUTPUT-002:** non-TTY stdout MUST default to plain output and MUST
  contain no ANSI control sequences.
- **AINFRA-OUTPUT-003:** `--format json` MUST produce exactly one versioned
  result stream on stdout and no human decoration.
- **AINFRA-OUTPUT-004:** secrets and sensitive plan/state content MUST be
  redacted before reaching any renderer, not by renderer-specific cleanup.
- **AINFRA-OUTPUT-005:** rich and plain renderers MUST have golden tests for
  common widths, no-color mode, Unicode fallback, failures, and cancellation.
- **AINFRA-OUTPUT-006:** verbosity changes logs, not the schema or meaning of
  command results.

## Logging and verbosity

Logging uses structured events internally. The default level is `warn` so
normal command output remains concise.

| Option | Effective level | Intended content |
|---|---|---|
| no `-v` | `warn` | Warnings, refusals, and errors requiring attention. |
| `-v` | `info` | Lifecycle stages, selected sources, and child-tool start/finish events. |
| `-vv` | `debug` | Resolution decisions, check applicability, bindings, and sanitized child commands. |
| `-vvv` | `trace` | Fine-grained local control flow and sanitized IO metadata for diagnosis. |

`--log-level error|warn|info|debug|trace` is the explicit equivalent and is
mutually exclusive with `-v`. More than three `v` flags is invalid. Increasing
verbosity MUST NOT enable child-tool debug flags automatically, retain raw
secrets, bypass redaction, or add sensitive values to run evidence.

Every log event contains a UTC RFC 3339 timestamp, level, component, message,
command, and correlation identifiers when available. Stable diagnostic code,
deployment, template digest, run ID, and child executable MAY be included when
non-sensitive. Events MUST NOT include native variable contents, environment
values, credentials, raw plans, state, or unsanitized child output.

## Log destinations

The default destination is human-readable stderr. Configuration MAY select
one or more destinations:

- `stderr`: text or JSON, without rotation;
- `file`: append-only text or JSON Lines with size/age rotation, retained-file
  count, and optional compression; or
- `syslog`: the local operating-system syslog endpoint with configurable
  facility and tag.

Remote syslog transport is outside v1 because it introduces credentials,
network failure, and sensitive-data transport policy. A user can deploy a
system log forwarder independently.

`--log-format text|json` sets the default sink format, `--log-file PATH` adds
or replaces the rotating file sink, and `--syslog` enables the local syslog
sink. Detailed rotation and syslog options live in configuration files so the
global flag surface remains small.

Log directories and files are created owner-only where supported. File sinks
MUST reject symlinks and special files, rotate without following path changes,
and use safe rename/create behavior. Default rotation is 10 MiB, five retained
files, seven days, and compression enabled. Explicitly configured sink
initialization failure is a command error; ainfra MUST NOT silently discard a
requested audit destination.

- **AINFRA-LOG-001:** all sinks consume the same already-redacted structured
  event; sink implementations MUST NOT receive secret-bearing alternatives.
- **AINFRA-LOG-002:** file rotation MUST have deterministic size, count, age,
  and compression behavior with boundary tests.
- **AINFRA-LOG-003:** sink failures during execution MUST be surfaced on every
  remaining healthy sink and in the final result.
- **AINFRA-LOG-004:** log format is independent from command output format.
- **AINFRA-LOG-005:** concurrent child streams MUST preserve event records and
  correlation even when their display order interleaves.

## Configuration files

CLI configuration controls presentation, logging, local storage paths, and
executable discovery. It MUST NOT contain deployment variables, template
selection, backend settings, credentials, approvals, or infrastructure policy.
Those remain in native engine inputs, deployment contracts, explicit flags, or
external credential mechanisms.

ainfra reads these optional files from lowest to highest precedence:

| Layer | Linux | macOS |
|---|---|---|
| System | `/etc/ainfra/config.yaml` | `/Library/Application Support/ainfra/config.yaml` |
| User | `${XDG_CONFIG_HOME:-$HOME/.config}/ainfra/config.yaml` | `$HOME/Library/Application Support/ainfra/config.yaml` |
| Project | `<deployment-root>/.ainfra/config.yaml` | `<deployment-root>/.ainfra/config.yaml` |
| Explicit | `--config PATH` or `AINFRA_CONFIG` | `--config PATH` or `AINFRA_CONFIG` |

The explicit file is an additional highest-precedence file layer; it does not
silently disable lower layers. A missing auto-discovered file is ignored. A
missing, unreadable, or invalid explicitly selected file is an error. Unknown
keys and unsupported `apiVersion` values are errors in every layer.

Relative paths in system, user, and explicit files resolve from that file's
directory. Relative paths in project configuration resolve from the deployment
root. Environment path overrides MUST be absolute. Paths still undergo normal
containment, file-type, ownership, and permission checks.

## Environment variables

ainfra supports this closed set for its own configuration:

| Variable | Meaning |
|---|---|
| `AINFRA_CONFIG` | Explicit additional configuration file; overridden by `--config`. |
| `AINFRA_PROJECT` | Deployment root or `ainfra.yaml`; overridden by `--project`. |
| `AINFRA_FORMAT` | Command output format: `text` or `json`. |
| `AINFRA_OUTPUT_STYLE` | Console style: `auto`, `rich`, or `plain`. |
| `AINFRA_COLOR` | Color mode: `auto`, `always`, or `never`. |
| `NO_COLOR` | Standard presence-based request to disable color. |
| `AINFRA_NON_INTERACTIVE` | Boolean non-interactive mode. |
| `AINFRA_LOG_LEVEL` | `error`, `warn`, `info`, `debug`, or `trace`. |
| `AINFRA_LOG_FORMAT` | Default sink format: `text` or `json`. |
| `AINFRA_LOG_FILE` | Add an owner-only rotating file destination at an absolute path. |
| `AINFRA_LOG_SYSLOG` | Boolean enabling the local syslog destination. |
| `AINFRA_CACHE_DIR` | Absolute local template cache directory. |
| `AINFRA_RUN_DIR` | Absolute local run-evidence directory. |
| `AINFRA_TOFU_PATH` | Absolute OpenTofu executable path. |
| `AINFRA_ANSIBLE_RUNNER_PATH` | Absolute Ansible Runner executable path. |
| `AINFRA_GIT_PATH` | Absolute Git executable path. |
| `AINFRA_SSH_PATH` | Absolute SSH executable path. |

Booleans accept only `true` or `false`, case-insensitively. Empty values are
invalid except that presence of `NO_COLOR`, with any value, disables color.
ainfra MUST NOT support environment variables containing secrets. Provider,
OpenTofu, Ansible, Git, and SSH variables are child-tool inputs governed by the
subprocess environment policy; they are not ainfra configuration and MUST NOT
be copied into logs or configuration snapshots.

## Precedence and merging

Effective configuration is resolved in this order, from lowest to highest:

1. compiled defaults;
2. system file;
3. user file;
4. project file;
5. explicit file from `AINFRA_CONFIG` or `--config`;
6. supported `AINFRA_*` environment variables and `NO_COLOR`; and
7. command-line flags.

`--config` overrides which explicit file `AINFRA_CONFIG` names. `--project`
similarly overrides `AINFRA_PROJECT`. Maps merge by key, scalars replace lower
values, and lists replace rather than append. This makes log destination order
and removal deterministic. Conflicting mutually exclusive flags are errors.

`ainfra doctor environment -v` shows the loaded file paths, origin of each
effective non-sensitive value, and ignored absent layers. It MUST redact
sensitive-looking path components and MUST never print environment values from
outside the supported set.

Configuration that affects executable selection, cache materialization, run
location, or execution behavior is recorded and bound where required by the
plan protocol. Presentation and logging choices are not plan bindings.

Because project configuration can arrive with an untrusted repository, the
project layer MAY set only `ui` fields and `logging.level`. Executable paths,
storage paths, log destinations, and syslog settings from a project file are
rejected. Those settings require a system, user, explicitly selected file,
supported environment variable, or flag.

- **AINFRA-CONFIG-001:** effective configuration MUST be represented as a
  typed immutable value before command execution.
- **AINFRA-CONFIG-002:** configuration parsing and merge behavior MUST be
  identical across commands and covered by precedence fixtures.
- **AINFRA-CONFIG-003:** configuration files MUST reject inline secrets and
  fields outside the published schema.
- **AINFRA-CONFIG-004:** project-controlled configuration MUST NOT weaken
  security, redaction, plan review, containment, or confirmation rules.
- **AINFRA-CONFIG-005:** diagnostic output MUST identify a value's source
  layer without disclosing sensitive values.
- **AINFRA-CONFIG-006:**
  [`../../schemas/v1/config.schema.json`](../../schemas/v1/config.schema.json)
  defines the structural file contract, and
  [`../../examples/v1/config.yaml`](../../examples/v1/config.yaml) is the
  maintained example.
- **AINFRA-CONFIG-007:** a project configuration file MUST NOT select an
  executable, redirect storage or logs, or enable a new output destination.

## Relationship to `ainfra configure`

`ainfra configure --run RUN_ID` is a deployment lifecycle command: it invokes
Ansible against generated inventory using the declared native variable files.
It does not create, edit, reconcile, or display CLI configuration files.

Like every command, `configure` consumes the already resolved CLI configuration
for logging, presentation, executable discovery, cache, and run locations. CLI
configuration MUST NOT select its playbook, inventory, host variables, or
Ansible extra variables. The noun “configuration” in this chapter refers to
ainfra process settings; the `configure` verb refers only to configuring
deployment hosts through Ansible.
