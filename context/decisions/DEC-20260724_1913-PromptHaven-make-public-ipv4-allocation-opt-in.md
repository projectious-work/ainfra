---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260724_1913-PromptHaven-make-public-ipv4-allocation-opt-in
  created: '2026-07-24T19:13:57+00:00'
spec:
  title: Make public IPv4 allocation opt-in and keep management ingress private by
    default
  state: accepted
  decision: The standard deployment disables public IPv4. When IPv4 is disabled, the
    OpenTofu configuration must omit the public IPv4 allocation/provider resource
    entirely rather than allocate an address and leave it unused. Public IPv4 and
    IPv6 are explicit independent inputs. Public workload ingress may be enabled deliberately,
    but SSH and other management ingress remain private by default and any exception
    requires narrow reviewed CIDRs.
  context: Hetzner public IPv4 addresses usually add recurring cost. The baseline
    must avoid unnecessary allocation while retaining explicit support for workloads
    that need public addressing. Security requirements also prohibit public management
    ingress by default.
  rationale: Opt-in allocation reduces recurring cost and attack surface. Omitting
    the allocation makes the declared configuration, provider bill, and actual topology
    agree. Separating address allocation, workload ingress, and management ingress
    prevents enabling a public workload endpoint from silently exposing administration.
  alternatives:
  - option: Allocate public IPv4 on every node but block ingress
    rejected_because: It preserves avoidable cost and unnecessary public addressing
      even when the address is unused.
  - option: Forbid all public addressing
    rejected_because: Some workloads legitimately require public ingress, and an explicit
      reviewed configuration can support them without weakening management defaults.
  - option: Enable public IPv4 by default for convenience
    rejected_because: This increases both cost and exposure for the standard deployment.
  consequences: OpenTofu modules and tests must prove that disabled IPv4 creates no
    billable IPv4 resource or attachment. Outputs and inventory must handle absent
    IPv4 values. Documentation must distinguish public address allocation, workload
    ingress, and management access. IPv4 and IPv6 paths require separate positive
    and negative tests.
  decided_at: '2026-07-24T19:13:57+00:00'
---
