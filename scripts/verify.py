#!/usr/bin/env python3
import json
import sqlite3
import subprocess
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]

required = [
    "README.md",
    "MASTER_SPEC.md",
    "AGENT_PROMPT.md",
    "AGENT_IMPLEMENTATION_GUIDE.md",
    "contracts/domain.ts",
    "contracts/template-runtime.d.ts",
    "contracts/scheduler-runtime.d.ts",
    "contracts/openapi.yaml",
    "migrations/001_init.sql",
    "docs/01-terminology-and-invariants.md",
    "docs/29-acceptance-matrix.md",
]

for rel in required:
    p = ROOT / rel
    if not p.exists() or p.stat().st_size == 0:
        raise SystemExit(f"missing required file: {rel}")

for p in ROOT.rglob("*.json"):
    with p.open("r", encoding="utf-8") as f:
        json.load(f)

with (ROOT / "contracts/openapi.yaml").open("r", encoding="utf-8") as f:
    api = yaml.safe_load(f)
if api.get("openapi") != "3.1.0":
    raise SystemExit("openapi.yaml must be OpenAPI 3.1.0")

con = sqlite3.connect(":memory:")
con.executescript((ROOT / "migrations/001_init.sql").read_text(encoding="utf-8"))
con.close()

subprocess.run(["go", "test", "./..."], cwd=ROOT / "reference/go", check=True)

# Verify package manifest when present.
manifest = ROOT / "MANIFEST.sha256"
if manifest.exists():
    import hashlib
    for raw in manifest.read_text(encoding="utf-8").splitlines():
        if not raw.strip():
            continue
        expected, rel = raw.split("  ", 1)
        data = (ROOT / rel).read_bytes()
        actual = hashlib.sha256(data).hexdigest()
        if actual != expected:
            raise SystemExit(f"checksum mismatch: {rel}")

print("verify: OK")
