#!/usr/bin/env python3
"""Exercise real native release binaries with offline downloads and a private HOME."""

import argparse
import hashlib
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import tarfile

from terminal_session import TerminalSession

ROOT = Path(__file__).resolve().parents[2]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("release", nargs="?", type=Path, help="directory containing archives and checksums.txt")
    parser.add_argument("--binary", type=Path, help="package a native binary for the check")
    args = parser.parse_args()
    if bool(args.release) == bool(args.binary):
        parser.error("provide a release directory or --binary")
    with tempfile.TemporaryDirectory(prefix="keywi-native-install-") as directory:
        home = Path(directory)
        if args.binary:
            release = home / "release"
            release.mkdir()
            platform = {"Darwin": "darwin", "Linux": "linux"}[os.uname().sysname]
            arch = {"arm64": "arm64", "aarch64": "arm64", "x86_64": "amd64", "amd64": "amd64"}[os.uname().machine]
            archive = release / f"keywi_{platform}_{arch}.tar.gz"
            with tarfile.open(archive, "w:gz") as bundle:
                bundle.add(args.binary.resolve(), arcname="keywi")
            digest = hashlib.sha256(archive.read_bytes()).hexdigest()
            (release / "checksums.txt").write_text(f"{digest}  {archive.name}\n")
        else:
            release = args.release.resolve()
        tools = home / "tools"
        tools.mkdir()
        curl = tools / "curl"
        curl.write_text('''#!/usr/bin/env python3
import os, pathlib, shutil, sys
args=sys.argv[1:]
url=next(a for a in args if a.startswith('https://'))
assert url.startswith('https://github.com/gaurv-m/keywi/releases/')
shutil.copyfile(pathlib.Path(os.environ['KEYWI_TEST_RELEASE'])/url.rsplit('/',1)[1], args[args.index('-o')+1])
''')
        curl.chmod(0o755)
        env = dict(os.environ, HOME=str(home), ZDOTDIR=str(home), SHELL="/bin/bash",
                   KEYWI_TEST_RELEASE=str(release), PATH=str(tools)+":"+os.environ["PATH"])
        for variable in ("KEYWI_INSTALL_DIR", "KEYWI_VERSION", "KEYWI_REPO"):
            env.pop(variable, None)
        binary = home / ".local/bin/keywi"
        for action in ("install", "upgrade"):
            subprocess.run(["sh", str(ROOT / "install.sh"), action], env=env, check=True)
            assert binary.is_file()
        startup = 'export PATH="$HOME/.local/bin:$PATH"\n'
        assert (home / ".bashrc").read_text().count(startup) == 1
        (home / ".zshrc").write_text(startup)
        (home / ".bashrc").write_text(startup)
        for shell in ("zsh", "bash"):
            executable = shutil.which(shell)
            if not executable:
                print(f"Skipped {shell}: not installed")
                continue
            command = 'cd /tmp && command -v keywi && keywi --version'
            flags = ["-d", "-i", "-c"] if shell == "zsh" else ["--noprofile", "--rcfile", str(home / ".bashrc"), "-i", "-c"]
            result = subprocess.run([executable, *flags, command], env=env, text=True, capture_output=True, check=True)
            assert str(binary) in result.stdout and "keywi " in result.stdout
            print(f"Fresh {shell} found keywi from /tmp")
        terminal = TerminalSession(binary, home)
        try:
            terminal.wait_for("Casual")
            terminal.send(b"?")
            terminal.wait_for("Made by G")
        finally:
            terminal.close()
        data = home / "Library/Application Support/keywi" if os.uname().sysname == "Darwin" else home / "data/keywi"
        assert data.is_dir(), "profile missing"
        subprocess.run(["sh", str(ROOT / "install.sh"), "uninstall"], env=env, check=True)
        assert not binary.exists() and data.is_dir()
    print("Native install, upgrade, shell PATH, terminal launch, and uninstall passed")


if __name__ == "__main__":
    main()
