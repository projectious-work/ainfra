---
title: Templates
weight: 25
description: Bundled infrastructure templates and where to find them.
---

ainfra embeds its built-in templates in the release binary. Their reviewed
source remains under `templates/`; each directory contains an
`ainfra-template.yaml` manifest and a template-specific README.

| Template | Source | Purpose | Status |
|---|---|---|---|
| `hetzner-kubernetes-baseline` | [`templates/hetzner-kubernetes-baseline`](https://github.com/projectious-work/ainfra/tree/main/templates/hetzner-kubernetes-baseline) | Private-networked, hardened Debian 13 hosts prepared for a later Kubernetes installation | Initial validated template |

The binary does not currently use an external registry or download templates
at runtime. Clone the repository only when you want to inspect or adapt source.

## Inspect locally

```sh
find templates -mindepth 2 -maxdepth 2 -name ainfra-template.yaml -print
ainfra validate hetzner-kubernetes-baseline \
  --input templates/hetzner-kubernetes-baseline/inputs/example.input.yaml
```

The identifier supplied to the CLI is the directory name, not a filesystem
path:

```sh
ainfra plan hetzner-kubernetes-baseline \
  --input /path/to/environment.input.yaml
```

For selection and extension guidance, see the
[template strategy]({{< relref "/docs/concepts/template-strategy" >}}) and
[template authoring guide]({{< relref
"/docs/guides/authoring-templates" >}}).
