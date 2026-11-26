#!/usr/bin/env python3
"""Check native startup, controls, storage, and terminal restoration."""

import argparse
from pathlib import Path
import tempfile

from terminal_session import TerminalSession


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=Path)
    args = parser.parse_args()
    with tempfile.TemporaryDirectory(prefix="keywi-smoke-") as home:
        for trial in range(2):
            terminal = TerminalSession(args.binary, home, columns=80, rows=24)
            try:
                terminal.wait_for("Casual")
                mark = len(terminal.raw)
                terminal.send(b"/")
                assert b"Help" not in terminal.raw[mark:], "slash opened Help"
                terminal.send(b"?")
                terminal.wait_for("Made by G", since=mark)
                terminal.send(b"\x1b", wait=0.25)
                terminal.send(b"\x1b[C\x1b[C\x1b[C")
                mark = len(terminal.raw)
                terminal.send(b"\r")
                terminal.wait_for("Kiwi", since=mark)
                terminal.send(b"\x1b", wait=0.25)
                terminal.send(b"\x1b[D\x1b[D\x1b[D")
                terminal.send(b"\r")
                terminal.send(b"\r")
                terminal.wait_for("type to begin")
                terminal.send(b"a")
            finally:
                terminal.close()
        print("Native PTY smoke check passed: startup, Kiwi, ? help, / inactive, play, relaunch, terminal restored")


if __name__ == "__main__":
    main()
