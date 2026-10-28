"""Safety boundaries for database tooling; no Docker needed."""

import importlib.util
import os
import subprocess
import tempfile
import unittest
from unittest.mock import patch
from pathlib import Path


class APIToolsTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        path = Path(__file__).resolve().parents[1] / "scripts" / "api_tools.py"
        spec = importlib.util.spec_from_file_location("api_tools", path)
        cls.tools = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(cls.tools)

    def test_migration_requires_explicit_database(self):
        with patch.dict(os.environ, {}, clear=True):
            with self.assertRaises(ValueError):
                self.tools.migrate("up")

    def test_migration_passes_credentials_only_in_environment(self):
        url = "postgres://example.invalid/test"
        with patch.dict(os.environ, {"DATABASE_URL": url}), patch.object(
            self.tools.subprocess, "run"
        ) as run:
            self.tools.migrate("status")
        args, kwargs = run.call_args
        self.assertNotIn(url, " ".join(args[0]))
        self.assertEqual(kwargs["env"]["GOOSE_DBSTRING"], url)
        self.assertEqual(kwargs["env"]["GOOSE_DRIVER"], "postgres")
        self.assertIn("@v3.24.1", " ".join(args[0]))

    def test_migration_rejects_arbitrary_command(self):
        with self.assertRaises(ValueError):
            self.tools.migrate("reset")

    def test_integration_always_cleans_own_project_after_failure(self):
        calls = []

        def run(command, **kwargs):
            calls.append((command, kwargs))
            if "port" in command:
                return subprocess.CompletedProcess(command, 0, "127.0.0.1:54321\n")
            if command[:2] == ["go", "test"]:
                raise subprocess.CalledProcessError(1, command)
            return subprocess.CompletedProcess(command, 0)

        with patch.object(self.tools.subprocess, "run", side_effect=run):
            with self.assertRaises(subprocess.CalledProcessError):
                self.tools.integration()
        self.assertEqual(calls[-1][0][-2:], ["down", "--volumes"])
        test = next(call for call in calls if call[0][:2] == ["go", "test"])
        self.assertIn("127.0.0.1:54321", test[1]["env"]["TEST_DATABASE_URL"])
        self.assertNotIn("DATABASE_URL", test[1]["env"])
        self.assertIn("museum-test-", calls[0][0][calls[0][0].index("-p") + 1])
        self.assertEqual(calls[0][0][2:4], ["--env-file", os.devnull])

    def test_cleanup_failure_preserves_primary_failure(self):
        def run(command, **kwargs):
            raise subprocess.CalledProcessError(2 if "down" in command else 1, command)

        with patch.object(self.tools.subprocess, "run", side_effect=run), patch(
            "sys.stderr"
        ):
            with self.assertRaises(subprocess.CalledProcessError) as caught:
                self.tools.integration()
        self.assertEqual(caught.exception.returncode, 1)

    def test_drift_check_compares_without_modifying_originals(self):
        cases = ({"query.go": b"original"}, {"query.go": b"changed"},
                 {}, {"query.go": b"original", "extra.go": b"extra"})
        for files in cases:
            with self.subTest(files=files), tempfile.TemporaryDirectory() as directory:
                api = Path(directory)
                (api / "sqlc.yaml").write_text("version: 2")
                for name in ("migrations", "queries", "generated"):
                    (api / "db" / name).mkdir(parents=True)
                original = api / "db" / "generated" / "query.go"
                original.write_bytes(b"original")

                def run(command, **kwargs):
                    target = Path(command[-1]).parent / "db" / "generated"
                    target.mkdir(parents=True)
                    for name, content in files.items():
                        (target / name).write_bytes(content)

                with patch.object(self.tools, "API", api), patch.object(
                    self.tools.subprocess, "run", side_effect=run
                ):
                    if files == cases[0]:
                        self.tools.generate(check=True)
                    else:
                        with self.assertRaises(ValueError):
                            self.tools.generate(check=True)
                self.assertEqual(original.read_bytes(), b"original")
                self.assertEqual(list(original.parent.iterdir()), [original])

    def test_generator_failure_is_not_reported_as_success(self):
        with tempfile.TemporaryDirectory() as directory:
            api = Path(directory)
            (api / "sqlc.yaml").touch()
            for name in ("migrations", "queries"):
                (api / "db" / name).mkdir(parents=True)
            with patch.object(self.tools, "API", api), patch.object(
                self.tools.subprocess, "run",
                side_effect=subprocess.CalledProcessError(1, ["go"])
            ):
                with self.assertRaises(subprocess.CalledProcessError):
                    self.tools.generate(check=True)

    def test_integration_migration_cycle_and_environment_isolation(self):
        calls = []

        def run(command, **kwargs):
            calls.append((command, kwargs))
            return subprocess.CompletedProcess(command, 0, "127.0.0.1:54321\n")

        with patch.dict(os.environ, {"DATABASE_URL": "do-not-use",
                                     "TEST_DATABASE_URL": "also-do-not-use"}), patch.object(
            self.tools.subprocess, "run", side_effect=run
        ):
            self.tools.integration()
        migrations = [call for call in calls if call[0][0] == "python3"]
        self.assertEqual([call[0][-1] for call in migrations], ["up", "up", "down", "up"])
        for _, kwargs in migrations:
            self.assertEqual(kwargs["env"]["DATABASE_URL"],
                             kwargs["env"]["TEST_DATABASE_URL"])
            self.assertNotIn("do-not-use", kwargs["env"]["DATABASE_URL"])

    def test_non_loopback_test_database_is_rejected_and_cleaned(self):
        calls = []

        def run(command, **kwargs):
            calls.append(command)
            return subprocess.CompletedProcess(command, 0, "0.0.0.0:54321\n")

        with patch.object(self.tools.subprocess, "run", side_effect=run):
            with self.assertRaises(ValueError):
                self.tools.integration()
        self.assertEqual(calls[-1][-2:], ["down", "--volumes"])
        self.assertFalse(any(command[:2] == ["go", "test"] for command in calls))


if __name__ == "__main__":
    unittest.main()
