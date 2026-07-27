---
title: Templates
weight: 25
description: Bundled infrastructure templates and where to find them.
---

ainfra discovers bundled templates as direct children of `templates/` in the
source repository. Each directory contains an `ainfra-template.yaml` manifest
and a template-specific README.

| Template | Source | Purpose | Status |
|---|---|---|---|
| `hetzner-kubernetes-baseline` | [`templates/hetzner-kubernetes-baseline`](https://github.com/projectious-work/ainfra/tree/main/templates/hetzner-kubernetes-baseline) | Private-networked, hardened Debian 13 hosts prepared for a later Kubernetes installation | Initial validated template |

The repository does not currently use an external registry and does not
download templates at runtime. Clone the repository to inspect or adapt a
template.

## Inspect locally

```sh
find templates -mindepth 2 -maxdepth 2 -name ainfra-template.yaml -print
uv run ainfra validate hetzner-kubernetes-baseline \
  --input templates/hetzner-kubernetes-baseline/inputs/example.input.yaml
```

The identifier supplied to the CLI is the directory name, not a filesystem
path:

```sh
uv run ainfra plan hetzner-kubernetes-baseline \
  --input /path/to/environment.input.yaml \
  --plan-id reviewed-plan-001
```

For selection and extension guidance, see the
[template strategy]({{< relref "/docs/concepts/template-strategy" >}}) and
[template authoring guide]({{< relref
"/docs/guides/authoring-templates" >}}).
