#!/usr/bin/env python3
"""Tests for scripts/version.py. They copy the repo's version sites and never write them."""

import importlib.util
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent

spec = importlib.util.spec_from_file_location("containerkitversion", REPO / "scripts" / "version.py")
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

SITES = (
    "VERSION",
    "tsx/package.json",
    "tsx/package-lock.json",
    "cmd/containerkit/main.go",
    *mod.platform_rels(),
)


def snapshot():
    return {rel: (REPO / rel).read_text() for rel in SITES}


class VersionScriptTests(unittest.TestCase):
    def setUp(self):
        self._repo = snapshot()
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.addCleanup(self.assert_repo_unchanged)
        root = Path(self.tmp.name)
        for rel, text in self._repo.items():
            path = root / rel
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(text)
        mod.root = root

    def assert_repo_unchanged(self):
        mod.root = REPO
        for rel, text in self._repo.items():
            self.assertEqual((REPO / rel).read_text(), text, rel)

    def test_bump_math(self):
        self.assertEqual(mod.bump("0.8.1", "patch"), "0.8.2")
        self.assertEqual(mod.bump("0.8.1", "minor"), "0.9.0")
        self.assertEqual(mod.bump("0.8.1", "major"), "1.0.0")
        self.assertEqual(mod.bump("1.2.3-rc.1", "patch"), "1.2.3")
        self.assertEqual(mod.bump("1.2.0-rc.1", "minor"), "1.2.0")
        self.assertEqual(mod.bump("1.0.0-rc.1", "major"), "1.0.0")
        self.assertEqual(mod.bump("1.2.3-rc.1+build", "minor"), "1.3.0")

    def test_patch_keeps_dependency_versions(self):
        before = json.loads((mod.root / "tsx/package-lock.json").read_text())
        nxt = mod.bump(mod.version_file(), "patch")
        mod.apply(nxt)
        lock = json.loads((mod.root / "tsx/package-lock.json").read_text())
        self.assertEqual(lock["version"], nxt)
        self.assertEqual(lock["packages"][""]["version"], nxt)
        for key, pkg in before["packages"].items():
            if key == "":
                continue
            self.assertEqual(lock["packages"][key], pkg)
        pkg = json.loads((mod.root / "tsx/package.json").read_text())
        self.assertEqual(pkg["version"], nxt)
        self.assertEqual(pkg["devDependencies"], before["packages"][""]["devDependencies"])
        self.assertIn(f'\tversion = "{nxt}"\n', (mod.root / "cmd/containerkit/main.go").read_text())

    def test_patch_refuses_drift(self):
        go = mod.root / "cmd/containerkit/main.go"
        go.write_text(go.read_text().replace("0.8.1", "0.8.0", 1))
        version = (mod.root / "VERSION").read_text()
        with self.assertRaises(SystemExit) as raised:
            mod.main(["patch"])
        self.assertEqual(raised.exception.code, 1)
        self.assertEqual((mod.root / "VERSION").read_text(), version)
        self.assertIn("0.8.0", go.read_text())

    def test_explicit_version_repairs_drift(self):
        go = mod.root / "cmd/containerkit/main.go"
        go.write_text(go.read_text().replace("0.8.1", "0.8.0", 1))
        mod.main(["v1.4.0"])
        for _, got in mod.sites():
            self.assertEqual(got, "1.4.0")

    def test_failed_edit_writes_nothing(self):
        go = mod.root / "cmd/containerkit/main.go"
        original = go.read_text()
        go.write_text(original.replace('\tversion = "0.8.1"\n', "\t// version moved\n"))
        before = snapshot_tree(mod.root)
        with self.assertRaises(SystemExit) as raised:
            mod.main(["9.9.9"])
        self.assertEqual(raised.exception.code, 1)
        self.assertEqual(snapshot_tree(mod.root), before)

    def test_bad_version_is_usage(self):
        with self.assertRaises(SystemExit) as raised:
            mod.main(["1.2"])
        self.assertEqual(raised.exception.code, 2)
        self.assertEqual((mod.root / "VERSION").read_text(), "0.8.1\n")

    def test_show_and_check(self):
        with self.assertRaises(SystemExit) as raised:
            mod.main(["--check", "show"])
        self.assertEqual(raised.exception.code, 2)
        mod.main(["--check"])
        mod.main(["show"])


def snapshot_tree(root):
    return {rel: (root / rel).read_text() for rel in SITES}


class RepoVersionTests(unittest.TestCase):
    def test_check_current_tree(self):
        proc = subprocess.run(
            [sys.executable, str(REPO / "scripts" / "version.py"), "--check"],
            cwd=REPO,
            capture_output=True,
            text=True,
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(proc.stdout, "")

    def test_print_current_version(self):
        proc = subprocess.run(
            [sys.executable, str(REPO / "scripts" / "version.py")],
            cwd=REPO,
            capture_output=True,
            text=True,
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(proc.stdout.strip(), (REPO / "VERSION").read_text().strip())


if __name__ == "__main__":
    unittest.main()
