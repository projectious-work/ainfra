---
title: Configuration
weight: 40
---

ainfra combines immutable, typed configuration layers without changing a
deployment or template. From lowest to highest precedence, the layers are:

1. compiled defaults;
2. system configuration;
3. user configuration;
4. deployment-local `ainfra.config.yaml`;
5. an explicit `--config` file or `AINFRA_CONFIG`; and
6. supported environment variables and command options.

On Linux, the system and user files are `/etc/ainfra/config.yaml` and
`$XDG_CONFIG_HOME/ainfra/config.yaml` (falling back to
`~/.config/ainfra/config.yaml`). On macOS they are
`/Library/Application Support/ainfra/config.yaml` and
`~/Library/Application Support/ainfra/config.yaml`.

Files are strict YAML documents using `apiVersion:
ainfra.projectious.work/v1` and `kind: Configuration`. Unknown fields,
unsupported versions, invalid enum values, and malformed environment values
are rejected. Optional absent files are reported as absent; an explicitly
selected missing file is an error.

Project discovery uses, in order, a positional target, `--project`,
`AINFRA_PROJECT`, or the nearest ancestor containing `ainfra.yaml`. An explicit
directory never searches descendants. A directory and its contained
`ainfra.yaml` identify the same deployment.

Repository-controlled project configuration may set only UI choices and
`logging.level`. It cannot select executables, storage paths, or log
destinations. Native OpenTofu and Ansible variable files stay opaque and are
never translated into ainfra configuration.

Use the environment doctor to inspect the effective values and provenance:

```sh
ainfra doctor environment --format json
ainfra doctor environment --project path/to/deployment
```

The report shows each winning layer and overridden layers without displaying
secret values.

`ainfra template lock` and `ainfra template update` use the same precedence
for trusted acquisition controls. An absolute `paths.cache` selects private
template storage, and `executables.git` selects the Git executable. Project
configuration cannot control either value; such attempts fail closed. Git is
resolved lazily, so local-only locking does not require it.
