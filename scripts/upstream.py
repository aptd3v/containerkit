#!/usr/bin/env python3
"""Show what changed upstream since the selected module pins.

Reads the versions in go.mod and the latest tags from the module proxy.
When a pin is behind, prints the GitHub release notes for each tag in
between, then the compare commits and the file diffstat.

  compose-go   github.com/compose-spec/compose-go/v2
  compose      github.com/docker/compose/v5, the Compose SDK
  engine       github.com/moby/moby/api and github.com/moby/moby/client

Usage:
  scripts/upstream.py
  scripts/upstream.py compose-go
  scripts/upstream.py compose
  scripts/upstream.py engine
  scripts/upstream.py compose-go --base v2.16.2
  scripts/upstream.py --self-test

With no names, every group is reported. A current pin is one line.
--base compares that tag to the latest tag of one module.
Exit 0 after a report. Exit 2 on usage or a failed lookup.
"""

import json
import os
import subprocess
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

root = Path(__file__).resolve().parent.parent

SPECS = (
    {
        "name": "compose-go",
        "group": "compose-go",
        "title": "compose-go",
        "module": "github.com/compose-spec/compose-go/v2",
        "repo": "compose-spec/compose-go",
        "tag": lambda version: version,
    },
    {
        "name": "compose",
        "group": "compose",
        "title": "Compose SDK",
        "module": "github.com/docker/compose/v5",
        "repo": "docker/compose",
        "tag": lambda version: version,
    },
    {
        "name": "api",
        "group": "engine",
        "title": "Engine SDK",
        "module": "github.com/moby/moby/api",
        "repo": "moby/moby",
        "tag": lambda version: "api/" + version,
    },
    {
        "name": "client",
        "group": "engine",
        "title": "Engine SDK",
        "module": "github.com/moby/moby/client",
        "repo": "moby/moby",
        "tag": lambda version: "client/" + version,
    },
)

ALIASES = {
    "compose-go": "compose-go",
    "composego": "compose-go",
    "compose": "compose",
    "compose-sdk": "compose",
    "composesdk": "compose",
    "engine": "engine",
    "engine-sdk": "engine",
    "enginesdk": "engine",
    "api": "api",
    "client": "client",
}


def die(msg, code=2):
    print(msg, file=sys.stderr)
    raise SystemExit(code)


def version_key(version):
    body = version[1:] if version.startswith("v") else version
    body = body.split("+", 1)[0].split("-", 1)[0]
    try:
        return tuple(int(part) for part in body.split("."))
    except ValueError:
        die(f"cannot compare version {version}")


def go_json(args):
    try:
        out = subprocess.run(
            ["go", *args],
            cwd=root,
            check=True,
            capture_output=True,
            text=True,
        )
    except subprocess.CalledProcessError as err:
        detail = (err.stderr or err.stdout or str(err)).strip()
        die(detail or "go failed")
    decoder = json.JSONDecoder()
    text = out.stdout
    index = 0
    items = []
    while index < len(text):
        while index < len(text) and text[index].isspace():
            index += 1
        if index >= len(text):
            break
        item, index = decoder.raw_decode(text, index)
        items.append(item)
    return items


def module_span(module):
    found = go_json(["list", "-m", "-u", "-json", module])
    if len(found) != 1:
        die(f"go list returned {len(found)} records for {module}")
    item = found[0]
    selected = item["Version"]
    latest = item.get("Update", {}).get("Version", selected)
    return selected, latest


def versions_through(module, base, head):
    try:
        out = subprocess.run(
            ["go", "list", "-m", "-versions", module],
            cwd=root,
            check=True,
            capture_output=True,
            text=True,
        )
    except subprocess.CalledProcessError as err:
        detail = (err.stderr or err.stdout or str(err)).strip()
        die(detail or f"go list -m -versions {module} failed")
    listed = out.stdout.split()[1:]
    lo = version_key(base)
    hi = version_key(head)
    chosen = []
    for version in listed:
        if "-" in version:
            continue
        key = version_key(version)
        if lo < key <= hi:
            chosen.append(version)
    if head not in chosen:
        chosen.append(head)
    return chosen


def github(path):
    url = "https://api.github.com" + path
    headers = {
        "Accept": "application/vnd.github+json",
        "User-Agent": "containerkit-upstream",
        "X-GitHub-Api-Version": "2022-11-28",
    }
    token = os.environ.get("GITHUB_TOKEN") or os.environ.get("GH_TOKEN")
    if token:
        headers["Authorization"] = "Bearer " + token
    request = urllib.request.Request(url, headers=headers)
    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            return json.load(response)
    except urllib.error.HTTPError as err:
        if err.code == 404:
            return None
        body = err.read().decode("utf-8", "replace")[:400].strip()
        die(f"GitHub {path}: {err.code} {body}")
    except urllib.error.URLError as err:
        die(f"GitHub {path}: {err.reason}")


def quote_tag(tag):
    return urllib.parse.quote(tag, safe="")


def release_notes(repo, tag):
    payload = github(f"/repos/{repo}/releases/tags/{quote_tag(tag)}")
    if payload is None:
        return None
    body = (payload.get("body") or "").strip()
    return body or None


def compare(repo, base_tag, head_tag):
    path = f"/repos/{repo}/compare/{quote_tag(base_tag)}...{quote_tag(head_tag)}"
    payload = github(path)
    if payload is None:
        die(f"no compare {base_tag}...{head_tag} in {repo}")
    return payload


def commit_lines(payload):
    lines = []
    for commit in payload.get("commits") or []:
        message = commit.get("commit", {}).get("message", "").splitlines()
        subject = message[0] if message else ""
        sha = (commit.get("sha") or "")[:7]
        lines.append(f"{sha} {subject}".rstrip())
    total = payload.get("total_commits", len(lines))
    return lines, total


def file_lines(payload):
    lines = []
    for item in payload.get("files") or []:
        status = {"added": "A", "removed": "D", "renamed": "R"}.get(item.get("status"), "M")
        name = item.get("filename", "")
        if item.get("previous_filename"):
            name = f"{item['previous_filename']} -> {name}"
        lines.append(f"{status} {name}  +{item.get('additions', 0)} -{item.get('deletions', 0)}")
    return lines


def report_module(spec, base, head):
    tag = spec["tag"]
    base_tag = tag(base)
    head_tag = tag(head)
    print(f"{spec['module']}")
    print(f"{base} -> {head}")
    print(f"https://github.com/{spec['repo']}/compare/{base_tag}...{head_tag}")
    print()
    for version in versions_through(spec["module"], base, head):
        notes = release_notes(spec["repo"], tag(version))
        print(version)
        if notes is None:
            print("No GitHub release notes for this tag.")
        else:
            print(notes)
        print()
    payload = compare(spec["repo"], base_tag, head_tag)
    commits, total = commit_lines(payload)
    if total == len(commits):
        print(f"commits ({total})")
    else:
        print(f"commits ({len(commits)} of {total})")
    if commits:
        print("\n".join(commits))
    else:
        print("(none)")
    print()
    files = file_lines(payload)
    print(f"files ({len(files)})")
    if files:
        print("\n".join(files))
    else:
        print("(none)")
    if len(files) >= 300:
        print("GitHub returned the first 300 files.")
    print()


def report_current(spec, version):
    print(f"  {spec['module']} {version} current")


def specs_for(names):
    if not names:
        return list(SPECS)
    chosen = []
    seen = set()
    for name in names:
        key = ALIASES.get(name.lower())
        if key is None:
            die(f"unknown module {name}; use compose-go, compose, or engine")
        for spec in SPECS:
            if spec["name"] == key or spec["group"] == key:
                if spec["name"] not in seen:
                    chosen.append(spec)
                    seen.add(spec["name"])
    return chosen


def run(names, base):
    specs = specs_for(names)
    if base is not None and len(specs) != 1:
        die("--base applies to one module: compose-go, compose, api, or client")
    title = None
    for spec in specs:
        if spec["title"] != title:
            if title is not None:
                print()
            title = spec["title"]
            print(title)
        selected, latest = module_span(spec["module"])
        start = base or selected
        if version_key(start) > version_key(latest):
            die(f"{spec['module']} {start} is newer than the latest tag {latest}")
        if start == latest:
            report_current(spec, latest)
            continue
        print()
        report_module(spec, start, latest)


def self_test():
    if version_key("v2.16.2") >= version_key("v2.17.0"):
        die("v2.16.2 should sort before v2.17.0")
    if version_key("v1.56.1") >= version_key("v1.57.0"):
        die("v1.56.1 should sort before v1.57.0")
    if version_key("v0.6.2") >= version_key("v0.6.3"):
        die("v0.6.2 should sort before v0.6.3")
    names = [spec["name"] for spec in specs_for(["engine"])]
    if names != ["api", "client"]:
        die(f"engine should select the API and the client, got {names}")
    if [spec["name"] for spec in specs_for(["compose-sdk"])] != ["compose"]:
        die("compose-sdk should select the Compose SDK")
    if [spec["name"] for spec in specs_for([])] != ["compose-go", "compose", "api", "client"]:
        die("the default selection should be compose-go, Compose, then Engine")
    commits, total = commit_lines(
        {
            "total_commits": 2,
            "commits": [
                {"sha": "abcdef123456", "commit": {"message": "first\n\nbody"}},
                {"sha": "1234567890ab", "commit": {"message": "second"}},
            ],
        }
    )
    if commits != ["abcdef1 first", "1234567 second"] or total != 2:
        die(f"commit lines: {commits} {total}")
    files = file_lines(
        {
            "files": [
                {"filename": "a.go", "status": "modified", "additions": 3, "deletions": 1},
                {"filename": "b.go", "status": "added", "additions": 4, "deletions": 0},
            ]
        }
    )
    if files != ["M a.go  +3 -1", "A b.go  +4 -0"]:
        die(f"file lines: {files}")
    print("ok")


def main(argv=None):
    args = list(sys.argv[1:] if argv is None else argv)
    if args in (["-h"], ["--help"]):
        print(__doc__.strip())
        return
    if args == ["--self-test"]:
        self_test()
        return
    base = None
    names = []
    index = 0
    while index < len(args):
        arg = args[index]
        if arg == "--base":
            if index + 1 >= len(args):
                die("usage: scripts/upstream.py [compose-go|compose|engine] [--base version]")
            base = args[index + 1]
            index += 2
            continue
        if arg.startswith("-"):
            die("usage: scripts/upstream.py [compose-go|compose|engine] [--base version]")
        names.append(arg)
        index += 1
    run(names, base)


if __name__ == "__main__":
    main()
