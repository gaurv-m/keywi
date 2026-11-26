"""Isolated PTY sessions for developer checks and captures."""

import fcntl
import os
from pathlib import Path
import pty
import re
import select
import signal
import struct
import subprocess
import termios
import time


class TerminalSession:
    def __init__(self, binary, home, columns=100, rows=30, on_output=None):
        self.master, self.slave = pty.openpty()
        self.slave_path = os.ttyname(self.slave)
        self.original = termios.tcgetattr(self.slave)
        fcntl.ioctl(self.slave, termios.TIOCSWINSZ, struct.pack("HHHH", rows, columns, 0, 0))
        self.raw = bytearray()
        self.answered = set()
        self.on_output = on_output
        env = dict(os.environ, HOME=str(home), XDG_CONFIG_HOME=str(Path(home) / "config"),
                   XDG_DATA_HOME=str(Path(home) / "data"), TERM="xterm-256color", COLORTERM="truecolor")
        env.pop("NO_COLOR", None)
        self.process = subprocess.Popen([str(Path(binary).resolve())], stdin=self.slave,
                                        stdout=self.slave, stderr=self.slave, env=env,
                                        start_new_session=True)

    def pump(self, seconds=0.15):
        end = time.monotonic() + seconds
        while time.monotonic() < end:
            ready, _, _ = select.select([self.master], [], [], max(0, end-time.monotonic()))
            if not ready:
                continue
            try:
                chunk = os.read(self.master, 65536)
            except OSError:
                return
            if not chunk:
                return
            start = max(0, len(self.raw)-16)
            self.raw.extend(chunk)
            if self.on_output:
                self.on_output(chunk)
            probe = bytes(self.raw[start:])
            for query, reply in ((b"\x1b]11;?", b"\x1b]11;rgb:0000/0000/0000\x1b\\"),
                                 (b"\x1b[6n", b"\x1b[1;1R")):
                offset = probe.find(query)
                while offset >= 0:
                    token = (query, start+offset)
                    if token not in self.answered:
                        os.write(self.master, reply)
                        self.answered.add(token)
                    offset = probe.find(query, offset+len(query))

    def send(self, data, wait=0.15):
        os.write(self.master, data)
        self.pump(wait)

    def wait_for(self, text, since=0, timeout=6):
        end = time.monotonic()+timeout
        while time.monotonic() < end:
            self.pump(0.05)
            output = re.sub(rb"\x1b\[[0-?]*[ -/]*[@-~]", b"", bytes(self.raw[since:]))
            if text.encode() in output:
                return
            if self.process.poll() is not None:
                break
        raise AssertionError(f"Terminal did not show {text!r}; exit={self.process.poll()}")

    def close(self):
        try:
            if self.process.poll() is None:
                self.send(b"\x03", wait=0.1)
                try:
                    self.process.wait(timeout=4)
                except subprocess.TimeoutExpired:
                    os.killpg(self.process.pid, signal.SIGKILL)
                    self.process.wait(timeout=2)
            self.pump(0.05)
            try:
                restored = termios.tcgetattr(self.slave)
            except termios.error as error:
                if error.args[0] != 25:
                    raise
                fresh = os.open(self.slave_path, os.O_RDWR | os.O_NOCTTY)
                try:
                    restored = termios.tcgetattr(fresh)
                finally:
                    os.close(fresh)
            # Darwin may set PENDIN while restoring canonical input.
            expected = list(self.original)
            restored[3] &= ~getattr(termios, "PENDIN", 0)
            expected[3] &= ~getattr(termios, "PENDIN", 0)
            restored = restored == expected
            if self.process.returncode != 0 or not restored:
                raise AssertionError("Unclean exit or terminal settings were not restored")
            for sequence in (b"\x1b[?1049l", b"\x1b[?2004l", b"\x1b[?25h"):
                assert sequence in self.raw, "terminal display state was not restored"
        finally:
            try:
                termios.tcsetattr(self.slave, termios.TCSANOW, self.original)
            except termios.error:
                pass
            os.close(self.master)
            os.close(self.slave)
