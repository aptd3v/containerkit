#!/usr/bin/env python3
"""Check the four upstream module pins.

go.mod is the source of truth. pkg/kit/doc.go and the blob URLs in
tools/gencatalog/main.go must name the same versions. The historical pin
line in CHANGELOG.md is not read.

Usage:
  scripts/pins.py --check       fail when the three copies disagree
  scripts/pins.py --freshness   compare go.mod with the latest tags
  scripts/pins.py --self-test   run the parser fixtures

--freshness prints "current" or "behind" on the first line and a markdown
report after it. Exit 0 means the pins match the latest tags. Exit 1 means
at least one is behind. Exit 2 is a usage or command failure.
"""

import re
import subprocess
import sys
from pathlib import Path

root = Path(__file__).resolve().parent.parent

MODULES = (
    "github.com/moby/moby/api",
    "github.com/moby/moby/client",
    "github.com/compose-spec/compose-go/v2",
    "github.com/docker/compose/v5",
)

GO_MOD = Path("go.mod")
DOC = Path("pkg/kit/doc.go")
CATALOG = Path("tools/gencatalog/main.go")

REQUIRE_RE = re.compile(
    r"(?m)^[ \t]*(?:require[ \t]+)?(\S+)[ \t]+(v\S+)(?:[ \t]+//.*)?$"
)
DOC_RE = re.compile(r"(?m)^//\s+-\s+(\S+)\s+(v\S+)")
LIST_RE = re.compile(r"^(\S+)\s+(\S+)(?:\s+\[([^\]]+)\])?")

BLOB_RE = {
    "github.com/moby/moby/api": re.compile(
        r"https://github.com/moby/moby/blob/api/(v[^/\"']+)/"
    ),
    "github.com/moby/moby/client": re.compile(
        r"https://github.com/moby/moby/blob/client/(v[^/\"']+)/"
    ),
    "github.com/compose-spec/compose-go/v2": re.compile(
        r"https://github.com/compose-spec/compose-go/blob/(v[^/\"']+)/"
    ),
    "github.com/docker/compose/v5": re.compile(
        r"https://github.com/docker/compose/blob/(v[^/\"']+)/"
    ),
}


def die(msg, code=1):
    print(msg, file=sys.stderr)
    raise SystemExit(code)


def read_text(rel):
    path = root / rel
    if not path.is_file():
        die(f"missing {rel}", 2)
    return path.read_text()


def versions_in(text, pattern):
    found = {}
    for module, version in pattern.findall(text):
        if module not in MODULES:
            continue
        found.setdefault(module, set()).add(version)
    return found


def one_version(found, module, label, errors):
    versions = found.get(module, set())
    if len(versions) != 1:
        shown = ", ".join(sorted(versions)) or "none"
        errors.append(f"{label} names {module} as {shown}")
        return None
    return next(iter(versions))


def blob_versions(text):
    found = {}
    for module, pattern in BLOB_RE.items():
        found[module] = set(pattern.findall(text))
    return found


def disagreements(go_text, doc_text, catalog_text):
    errors = []
    selected = versions_in(go_text, REQUIRE_RE)
    documented = versions_in(doc_text, DOC_RE)
    catalogued = blob_versions(catalog_text)
    pins = {}
    for module in MODULES:
        version = one_version(selected, module, "go.mod", errors)
        if version is None:
            continue
        pins[module] = version
        doc_version = one_version(documented, module, "pkg/kit/doc.go", errors)
        if doc_version is not None and doc_version != version:
            errors.append(
                f"pkg/kit/doc.go has {module} {doc_version}; go.mod has {version}"
            )
        blob = one_version(catalogued, module, "tools/gencatalog/main.go", errors)
        if blob is not None and blob != version:
            errors.append(
                f"tools/gencatalog/main.go has {module} {blob}; go.mod has {version}"
            )
    return errors, pins


def check():
    errors, pins = disagreements(read_text(GO_MOD), read_text(DOC), read_text(CATALOG))
    if errors:
        die("\n".join(errors))
    for module in MODULES:
        print(f"{module} {pins[module]}")


def parse_updates(text):
    updates = []
    wanted = set(MODULES)
    for line in text.splitlines():
        match = LIST_RE.match(line.strip())
        if match is None:
            continue
        module, selected, latest = match.group(1), match.group(2), match.group(3)
        if module not in wanted or latest is None:
            continue
        updates.append((module, selected, latest))
    return updates


def report(updates):
    lines = [
        "These modules in `go.mod` are behind the latest tag:",
        "",
        "| Module | Selected | Latest |",
        "| --- | --- | --- |",
    ]
    for module, selected, latest in updates:
        lines.append(f"| `{module}` | `{selected}` | `{latest}` |")
    lines.extend(
        [
            "",
            "`go list -m -u` reported this.",
            "The pull-request check requires `go.mod`, `pkg/kit/doc.go`, and `tools/gencatalog/main.go` to name the same versions.",
        ]
    )
    return "\n".join(lines) + "\n"


def list_upstream():
    cmd = ["go", "list", "-m", "-u", *MODULES]
    try:
        out = subprocess.run(cmd, cwd=root, check=True, capture_output=True, text=True)
    except subprocess.CalledProcessError as err:
        detail = (err.stderr or err.stdout or str(err)).strip()
        die(detail or "go list -m -u failed", 2)
    return out.stdout


def freshness():
    updates = parse_updates(list_upstream())
    if updates:
        print("behind")
        sys.stdout.write(report(updates))
        raise SystemExit(1)
    print("current")
    print("The four module pins match the latest tags.")


def self_test():
    go_text = "\n".join(f"\t{module} v1.2.3" for module in MODULES) + "\n"
    doc_text = "\n".join(f"//   - {module} v1.2.3" for module in MODULES) + "\n"
    catalog_text = "\n".join(
        [
            "https://github.com/moby/moby/blob/api/v1.2.3/api/types/types.go",
            "https://github.com/moby/moby/blob/client/v1.2.3/client/client.go",
            "https://github.com/compose-spec/compose-go/blob/v1.2.3/types/types.go",
            "https://github.com/docker/compose/blob/v1.2.3/pkg/api/api.go",
        ]
    )
    errors, pins = disagreements(go_text, doc_text, catalog_text)
    if errors or len(pins) != 4:
        die("matching copies should agree: " + "; ".join(errors), 2)

    drifted = doc_text.replace(
        "github.com/moby/moby/api v1.2.3", "github.com/moby/moby/api v9.9.9", 1
    )
    errors, _ = disagreements(go_text, drifted, catalog_text)
    if not any("v9.9.9" in err and "v1.2.3" in err for err in errors):
        die("a doc.go drift should name both versions", 2)

    doubled = catalog_text + "\nhttps://github.com/moby/moby/blob/api/v8.8.8/api/x.go\n"
    errors, _ = disagreements(go_text, doc_text, doubled)
    if not any("v1.2.3" in err and "v8.8.8" in err for err in errors):
        die("two catalog versions should be reported", 2)

    changelog = "Pins verified: github.com/moby/moby/api v0.0.1\n"
    errors, _ = disagreements(go_text, doc_text, catalog_text + changelog)
    if errors:
        die("text outside the blob URLs should not count", 2)

    sample = "\n".join(
        [
            "github.com/moby/moby/api v1.57.0",
            "github.com/docker/compose/v5 v5.7.1 [v5.7.2]",
            "github.com/moby/moby/client v0.6.3",
            "github.com/compose-spec/compose-go/v2 v2.16.2",
            "golang.org/x/net v0.1.0 [v0.2.0]",
        ]
    )
    updates = parse_updates(sample)
    if updates != [("github.com/docker/compose/v5", "v5.7.1", "v5.7.2")]:
        die(f"unexpected updates: {updates}", 2)
    if parse_updates("github.com/moby/moby/api v1.57.0\n"):
        die("a current module should not be an update", 2)
    print("ok")


def main(argv=None):
    args = list(sys.argv[1:] if argv is None else argv)
    if args in (["-h"], ["--help"]):
        print(__doc__.strip())
        return
    if args == ["--check"]:
        check()
        return
    if args == ["--freshness"]:
        freshness()
        return
    if args == ["--self-test"]:
        self_test()
        return
    die(
        "usage: scripts/pins.py [--check|--freshness|--self-test]",
        2,
    )


if __name__ == "__main__":
    main()
