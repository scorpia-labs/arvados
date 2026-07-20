# Copyright (C) The Arvados Authors. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
import re
import subprocess


class TestSanity:
    """Test that the `arv-cluster-activity` script is installed and can be
    run as a standalone CLI app, that it understands the "-h" and "--version"
    CLI arguments, and that it exits with status 2 upon invalid CLI arguments.
    """
    script = "arv-cluster-activity"

    def run(self, *args, **kwargs) -> subprocess.CompletedProcess:
        return subprocess.run(
            [self.script, *args],
            capture_output=True, text=True, check=False, **kwargs
        )

    def test_help(self):
        completed_process = self.run("-h")
        assert completed_process.returncode == 0
        assert f"usage: {self.script}" in completed_process.stdout
        assert not completed_process.stderr

    def test_version(self):
        completed_process = self.run("--version")
        assert completed_process.returncode == 0
        assert re.match(
            (
                rf"^{re.escape(self.script)}"
                r" [0-9]+\.[0-9]+\.[0-9]+(\.dev[0-9]+)?$\n"
            ),
            completed_process.stdout
        )
        assert not completed_process.stderr

    def test_invalid_argument(self):
        completed_process = self.run("--x-invalid-argument")
        assert completed_process.returncode == 2
        assert not completed_process.stdout
        assert f"usage: {self.script}" in completed_process.stderr
