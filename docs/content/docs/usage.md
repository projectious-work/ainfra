---
title: Usage
weight: 50
---

The Phase 1 command shell provides static help and version information without
project discovery, network access, or child-tool execution:

```sh
ainfra help
ainfra help version
ainfra version
ainfra --format json version
```

Each operation will include its expected inputs, outputs, safety checks, and
recovery path as its roadmap phase is implemented. Help already describes the
planned canonical command surface, but only `help` and `version` execute in
Phase 1.
