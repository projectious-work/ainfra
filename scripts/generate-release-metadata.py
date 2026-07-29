#!/usr/bin/env python3
"""Generate deterministic SPDX and per-archive provenance release assets."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import subprocess
from datetime import UTC, datetime
from pathlib import Path


def sha256(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write_json(path: Path, document: object) -> None:
    path.write_text(
        json.dumps(document, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, required=True)
    parser.add_argument("--version", required=True)
    parser.add_argument("--target", action="append", default=[])
    args = parser.parse_args()
    root = args.root.resolve()
    dist = root / "dist"
    tooling_commit = subprocess.check_output(
        ["git", "-C", root, "rev-parse", "HEAD"], text=True
    ).strip()
    commit = os.environ.get("AINFRA_RELEASE_SOURCE_COMMIT", tooling_commit)
    metadata = json.loads(
        subprocess.check_output(
            [
                "cargo",
                "metadata",
                "--locked",
                "--format-version",
                "1",
                "--manifest-path",
                str(root / "Cargo.toml"),
            ],
            text=True,
        )
    )
    epoch = int(os.environ.get("SOURCE_DATE_EPOCH", "0"))
    created = datetime.fromtimestamp(epoch, UTC).strftime("%Y-%m-%dT%H:%M:%SZ")
    packages = []
    for package in sorted(
        metadata["packages"], key=lambda item: (item["name"], item["version"])
    ):
        packages.append(
            {
                "SPDXID": (
                    f"SPDXRef-Package-{package['name']}-{package['version']}"
                ),
                "copyrightText": "NOASSERTION",
                "downloadLocation": "NOASSERTION",
                "filesAnalyzed": False,
                "licenseConcluded": "NOASSERTION",
                "licenseDeclared": package.get("license") or "NOASSERTION",
                "name": package["name"],
                "versionInfo": package["version"],
            }
        )
    sbom = {
        "SPDXID": "SPDXRef-DOCUMENT",
        "creationInfo": {
            "created": created,
            "creators": ["Tool: ainfra-generate-release-metadata"],
        },
        "dataLicense": "CC0-1.0",
        "documentNamespace": (
            "https://github.com/projectious-work/ainfra/"
            f"releases/v{args.version}/spdx/{commit}"
        ),
        "name": f"ainfra-v{args.version}",
        "packages": packages,
        "spdxVersion": "SPDX-2.3",
    }
    write_json(dist / f"ainfra-v{args.version}.spdx.json", sbom)

    lock_digest = sha256(root / "Cargo.lock")
    for target in args.target:
        archive = dist / f"ainfra-v{args.version}-{target}.tar.gz"
        if not archive.is_file():
            raise SystemExit(f"missing release archive: {archive}")
        provenance = {
            "_type": "https://in-toto.io/Statement/v1",
            "predicate": {
                "buildDefinition": {
                    "buildType": (
                        "https://projectious-work.github.io/ainfra/"
                        "release-build/v1"
                    ),
                    "externalParameters": {
                        "target": target,
                        "version": args.version,
                    },
                    "resolvedDependencies": [
                        {
                            "digest": {"gitCommit": commit},
                            "uri": (
                                "git+https://github.com/projectious-work/ainfra"
                            ),
                        },
                        {
                            "digest": {"sha256": lock_digest},
                            "uri": "file:Cargo.lock",
                        },
                    ],
                },
                "runDetails": {
                    "builder": {
                        "id": (
                            "https://projectious-work.github.io/ainfra/"
                            "local-release"
                        )
                    },
                    "metadata": {
                        "invocationId": f"v{args.version}-{commit}",
                        "toolingCommit": tooling_commit,
                    },
                },
            },
            "predicateType": "https://slsa.dev/provenance/v1",
            "subject": [
                {
                    "digest": {"sha256": sha256(archive)},
                    "name": archive.name,
                }
            ],
        }
        write_json(
            dist / f"ainfra-v{args.version}-{target}.provenance.json",
            provenance,
        )


if __name__ == "__main__":
    main()
