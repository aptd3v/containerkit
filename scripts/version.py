#!/usr/bin/env python3
"""Sync and bump the containerkit semver.

VERSION is the source of truth. This script writes that file and the copies
that cannot read it:

  tsx/package.json
  tsx/package-lock.json   this package only; dependency versions stay put
  npm/@containerkit/*/package.json
  cmd/containerkit/main.go  fallback when a build does not pass -X main.version

tools/crossbuild reads VERSION unless CONTAINERKIT_VERSION is set.
GoReleaser still passes -X main.version. This script does not commit or tag.

Usage:
  scripts/version.py                  print the version; fail if a copy disagrees
  scripts/version.py --check          same check, quiet on success
  scripts/version.py show             print every copy
  scripts/version.py patch|minor|major
  scripts/version.py 1.2.3            set every copy; a leading v is dropped

patch, minor, and major follow npm semver: a prerelease promotes to its
release (1.2.3-rc.1 patch -> 1.2.3) instead of skipping ahead. A bump refuses
to run when copies already disagree. An explicit version overwrites every copy.
"""

import json
import re
import sys
from pathlib import Path

root = Path(__file__).resolve().parent.parent

SEMVER = re.compile(
    r"^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)"
    r"(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?"
    r"(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$"
)

VERSION_FILE = Path("VERSION")
PACKAGE_JSON = Path("tsx/package.json")
PACKAGE_LOCK = Path("tsx/package-lock.json")
MAIN_GO = Path("cmd/containerkit/main.go")

MAIN_RE = re.compile(r'(?m)^(\tversion\s*=\s*")([^"]*)(")')
BUMPS = ("patch", "minor", "major")
PLATFORMS = (
    "darwin-amd64",
    "darwin-arm64",
    "linux-amd64",
    "linux-arm64",
    "windows-amd64",
    "windows-arm64",
)


def platform_rels():
    return [f"npm/@containerkit/{name}/package.json" for name in PLATFORMS]


def die(msg, code=1):
    print(msg, file=sys.stderr)
    raise SystemExit(code)


def parse_version(raw):
    text = raw.strip()
    if text.startswith("v"):
        text = text[1:]
    if not SEMVER.fullmatch(text):
        die(f"not a semantic version: {raw}", 2)
    return text


def bump(version, kind):
    match = SEMVER.fullmatch(version)
    if match is None or kind not in BUMPS:
        die(f"cannot {kind} {version}", 2)
    major, minor, patch = (int(match.group(i)) for i in (1, 2, 3))
    pre = match.group(4)
    if kind == "major":
        if not (minor == 0 and patch == 0 and pre):
            major += 1
        minor = patch = 0
    elif kind == "minor":
        if not (patch == 0 and pre):
            minor += 1
        patch = 0
    elif not pre:
        patch += 1
    return f"{major}.{minor}.{patch}"


def read_text(rel):
    path = root / rel
    if not path.is_file():
        die(f"missing {rel}")
    return path.read_text()


def version_file():
    lines = read_text(VERSION_FILE).splitlines()
    if len(lines) != 1 or not lines[0].strip():
        die("VERSION must be a single semantic version")
    return parse_version(lines[0])


def package_name():
    data = json.loads(read_text(PACKAGE_JSON))
    name = data.get("name")
    if not isinstance(name, str) or not name:
        die(f"{PACKAGE_JSON} has no package name")
    return name


def self_version_re(name):
    return re.compile(
        r'("name"\s*:\s*"' + re.escape(name) + r'"\s*,\s*"version"\s*:\s*")([^"]*)(")'
    )


def go_version():
    matches = MAIN_RE.findall(read_text(MAIN_GO))
    if len(matches) != 1:
        die(f"{MAIN_GO} should contain exactly one version assignment")
    return matches[0][1]


def lock_versions():
    data = json.loads(read_text(PACKAGE_LOCK))
    root_ver = data.get("version")
    pkg = data.get("packages", {}).get("")
    if not isinstance(root_ver, str) or not isinstance(pkg, dict) or not isinstance(pkg.get("version"), str):
        die(f"{PACKAGE_LOCK} is missing this package's version")
    return root_ver, pkg["version"]


def sites():
    lock_root, lock_pkg = lock_versions()
    pkg = json.loads(read_text(PACKAGE_JSON))
    ver = pkg.get("version")
    if not isinstance(ver, str):
        die(f"{PACKAGE_JSON} has no string version")
    found = [
        (str(VERSION_FILE), version_file()),
        (str(PACKAGE_JSON), ver),
        (f"{PACKAGE_LOCK}#/version", lock_root),
        (f"{PACKAGE_LOCK}#/packages//version", lock_pkg),
        (str(MAIN_GO), go_version()),
    ]
    for rel in platform_rels():
        data = json.loads(read_text(rel))
        got = data.get("version")
        if not isinstance(got, str):
            die(f"{rel} has no string version")
        found.append((str(rel), got))
    return found


def disagreements():
    found = sites()
    want = found[0][1]
    bad = [f"  {name}: {got}" for name, got in found if got != want]
    return want, bad, found


def report_drift(want, bad):
    print(f"version sites disagree with VERSION ({want}):", file=sys.stderr)
    print("\n".join(bad), file=sys.stderr)


def check(quiet):
    want, bad, _ = disagreements()
    if bad:
        report_drift(want, bad)
        raise SystemExit(1)
    if not quiet:
        print(want)


def show():
    want, bad, found = disagreements()
    width = max(len(name) for name, _ in found)
    for name, got in found:
        mark = "" if got == want else "  drifted"
        print(f"{name:<{width}}  {got}{mark}")
    if bad:
        raise SystemExit(1)


def replace_self_version(text, name, version):
    return self_version_re(name).subn(lambda m: m.group(1) + version + m.group(3), text)


def assert_package_version(data, version, name, lock=False):
    if data.get("name") != name or data.get("version") != version:
        die(f"package version was not updated (name={data.get('name')!r})")
    if not lock:
        return
    for key, pkg in data.get("packages", {}).items():
        if not isinstance(pkg, dict):
            continue
        if key == "" or pkg.get("name") == name:
            if pkg.get("version") != version:
                die(f"{PACKAGE_LOCK} package {key!r} version was not updated")


def render(version):
    name = package_name()
    files = {VERSION_FILE: version + "\n"}

    pkg_text, n = replace_self_version(read_text(PACKAGE_JSON), name, version)
    if n != 1:
        die(f"{PACKAGE_JSON} should contain exactly one name/version pair, found {n}")
    assert_package_version(json.loads(pkg_text), version, name)
    files[PACKAGE_JSON] = pkg_text

    lock_text, n = replace_self_version(read_text(PACKAGE_LOCK), name, version)
    if n < 2:
        die(f"{PACKAGE_LOCK} should contain the root and packages[''] versions, found {n}")
    assert_package_version(json.loads(lock_text), version, name, lock=True)
    files[PACKAGE_LOCK] = lock_text

    go_text, n = MAIN_RE.subn(lambda m: m.group(1) + version + m.group(3), read_text(MAIN_GO), count=1)
    if n != 1:
        die(f"could not update {MAIN_GO}")
    files[MAIN_GO] = go_text

    for rel in platform_rels():
        platform = json.loads(read_text(rel))
        platform_name = platform.get("name")
        text, n = replace_self_version(read_text(rel), platform_name, version)
        if n != 1:
            die(f"{rel} should contain exactly one name/version pair, found {n}")
        files[Path(rel)] = text
    return files


def apply(version):
    files = render(version)
    before = version_file() if (root / VERSION_FILE).is_file() else None
    for rel, text in files.items():
        path = root / rel
        if not path.is_file() or path.read_text() != text:
            path.write_text(text)
    check(quiet=True)
    if before == version:
        print(version)
    else:
        print(f"{before or '(none)'} -> {version}")


def main(argv=None):
    args = list(sys.argv[1:] if argv is None else argv)
    if args[:1] in (["-h"], ["--help"]):
        if len(args) > 1:
            die("usage: scripts/version.py [--check|show|patch|minor|major|version]", 2)
        print(__doc__.strip())
        return
    if not args:
        check(quiet=False)
        return
    if args == ["--check"]:
        check(quiet=True)
        return
    if args == ["show"]:
        show()
        return
    if len(args) != 1:
        die("usage: scripts/version.py [--check|show|patch|minor|major|version]", 2)
    if args[0] in BUMPS:
        want, bad, _ = disagreements()
        if bad:
            report_drift(want, bad)
            die(f"refusing to {args[0]} until the copies agree")
        apply(bump(want, args[0]))
        return
    apply(parse_version(args[0]))


if __name__ == "__main__":
    main()
