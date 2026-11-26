#!/usr/bin/env python3
"""Offline integration tests for installer integrity, upgrade and removal rules."""
import hashlib
import io
import os
import pathlib
import subprocess
import tarfile
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]


class InstallerTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="keywi-installer-")
        self.addCleanup(self.temp.cleanup)
        self.root = pathlib.Path(self.temp.name)
        self.tools = self.root / "tools"
        self.tools.mkdir()
        self.release = self.root / "release"
        self.release.mkdir()
        self.env = dict(os.environ, KEYWI_REPO="test-owner/keywi", KEYWI_INSTALL_DIR=str(self.root / "bin"),
                        FAKE_RELEASE=str(self.release), FAKE_SYSTEM="Darwin", FAKE_ARCH="arm64")
        self.env["PATH"] = str(self.tools) + ":" + os.environ["PATH"]
        self.script("uname", '#!/bin/sh\nif [ "$1" = -s ]; then echo "$FAKE_SYSTEM"; else echo "$FAKE_ARCH"; fi\n')
        self.script("curl", '''#!/usr/bin/env python3
import os,pathlib,sys,shutil
args=sys.argv[1:]
assert args[args.index('--proto')+1]=='=https'
assert args[args.index('--proto-redir')+1]=='=https'
url=next(a for a in args if a.startswith('https://'))
assert url.startswith('https://github.com/test-owner/keywi/releases/')
source=pathlib.Path(os.environ['FAKE_RELEASE'])/url.rsplit('/',1)[1]
if not source.exists():sys.exit(22)
shutil.copyfile(source,args[args.index('-o')+1])
''')
        self.archive()

    def script(self, name, body):
        p = self.tools / name
        p.write_text(body)
        p.chmod(0o755)

    def archive(self, system="darwin", arch="arm64", symlink=False):
        name = f"keywi_{system}_{arch}.tar.gz"
        binary = b'#!/bin/sh\n# github.com/gaurv-m/keywi\necho "keywi 1.0.0 (test)"\n'
        with tarfile.open(self.release / name, "w:gz") as tar:
            info = tarfile.TarInfo("keywi")
            info.mode = 0o755
            if symlink:
                info.type = tarfile.SYMTYPE
                info.linkname = "/etc/passwd"
                tar.addfile(info)
            else:
                info.size = len(binary)
                tar.addfile(info, io.BytesIO(binary))
        digest = hashlib.sha256((self.release / name).read_bytes()).hexdigest()
        (self.release / "checksums.txt").write_text(f"{digest}  {name}\n")

    def run_install(self, *args):
        return subprocess.run(["sh", str(ROOT / "install.sh"), *args], env=self.env,
                              capture_output=True, text=True)

    def test_install_upgrade_uninstall(self):
        result = self.run_install()
        self.assertEqual(result.returncode, 0, result.stderr)
        target = self.root / "bin/keywi"
        self.assertTrue(target.exists())
        self.assertIn("export PATH", result.stdout)
        self.assertEqual(self.run_install("upgrade").returncode, 0)
        self.assertEqual(self.run_install("uninstall").returncode, 0)
        self.assertFalse(target.exists())

    def test_default_install_saves_shell_path_once(self):
        line = 'export PATH="$HOME/.local/bin:$PATH"'
        for shell, filename in (("bash", ".bashrc"), ("zsh", ".zshrc")):
            with self.subTest(shell=shell):
                home = self.root / shell
                home.mkdir()
                profile = home / filename
                profile.write_text("# existing settings\n")
                self.env.update(HOME=str(home), SHELL="/bin/"+shell, ZDOTDIR=str(home))
                self.env.pop("KEYWI_INSTALL_DIR", None)
                result = self.run_install()
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertIn("Open a new terminal", result.stdout)
                self.assertEqual(self.run_install("upgrade").returncode, 0)
                saved = profile.read_text()
                self.assertTrue(saved.startswith("# existing settings\n"))
                self.assertEqual(saved.count(line), 1)
                self.assertTrue((home / ".local/bin/keywi").exists())

    def test_shell_profile_symlink_is_preserved(self):
        home = self.root / "home"
        home.mkdir()
        target = home / "shared-settings"
        target.write_text("# managed settings\n")
        (home / ".zshrc").symlink_to(target)
        self.env.update(HOME=str(home), SHELL="/bin/zsh", ZDOTDIR=str(home))
        self.env.pop("KEYWI_INSTALL_DIR", None)
        result = self.run_install()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("Add to your shell profile", result.stdout)
        self.assertEqual(target.read_text(), "# managed settings\n")

    def test_supported_platforms(self):
        for system, machine, platform, arch in [("Darwin", "x86_64", "darwin", "amd64"),
                ("Linux", "aarch64", "linux", "arm64"), ("Linux", "x86_64", "linux", "amd64")]:
            with self.subTest(system=system, arch=arch):
                self.env.update(FAKE_SYSTEM=system, FAKE_ARCH=machine)
                self.archive(platform, arch)
                result = self.run_install()
                self.assertEqual(result.returncode, 0, result.stderr)

    def test_bad_checksum_preserves_existing_binary(self):
        self.assertEqual(self.run_install().returncode, 0)
        target = self.root / "bin/keywi"
        before = target.read_bytes()
        (self.release / "checksums.txt").write_text("0" * 64 + "  keywi_darwin_arm64.tar.gz\n")
        result = self.run_install("upgrade")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("checksum mismatch", result.stderr)
        self.assertEqual(target.read_bytes(), before)
        self.assertFalse(list(target.parent.glob(".keywi.*")))

    def test_download_failure(self):
        (self.release / "keywi_darwin_arm64.tar.gz").unlink()
        result = self.run_install()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("download failed", result.stderr)
        self.assertFalse((self.root / "bin/keywi").exists())

    def test_reject_symlink_archive(self):
        self.archive(symlink=True)
        self.assertNotEqual(self.run_install().returncode, 0)
        self.assertFalse((self.root / "bin/keywi").exists())

    def test_invalid_inputs(self):
        for repo in ("OWNER/keywi", "x/../../y", "x/y?z", "x/y/z", "x", "x/$(id)"):
            with self.subTest(repo=repo):
                self.env["KEYWI_REPO"] = repo
                self.assertNotEqual(self.run_install().returncode, 0)
        self.env["KEYWI_REPO"] = "test-owner/keywi"
        self.env["FAKE_SYSTEM"] = "Windows"
        self.assertNotEqual(self.run_install().returncode, 0)

    def test_duplicate_checksums_rejected(self):
        p = self.release / "checksums.txt"
        p.write_text(p.read_text() * 2)
        self.assertNotEqual(self.run_install().returncode, 0)

    def test_uninstall_does_not_delete_unrelated_file(self):
        target = self.root / "bin/keywi"
        target.parent.mkdir()
        target.write_text("unrelated")
        self.assertNotEqual(self.run_install("uninstall").returncode, 0)
        self.assertEqual(target.read_text(), "unrelated")


if __name__ == "__main__":
    unittest.main(verbosity=2)
