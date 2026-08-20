---
title: Tutorials
weight: 80
---

## Validate a template in a clean room

Copy the reference template, make only the required local edits, then run:

```sh
ainfra doctor template . --format json
./tests/validate.sh
```

The first command validates the portable authoring contract. The test script
runs the compatible native tool checks and fixture assertions. Do not create
cloud resources as part of this tutorial: live lifecycle work requires explicit
approval at the point it is performed.

See the [template authoring guide](/docs/templates/) for the required layout
and the [AI guide](/docs/template-authoring-ai/) for an execution checklist.
