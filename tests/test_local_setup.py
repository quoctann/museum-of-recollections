"""Tests for safe local environment bootstrapping."""

import importlib.util
import tempfile
import unittest
from pathlib import Path


class LocalSetupTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        path = Path(__file__).resolve().parents[1] / "scripts" / "setup_local.py"
        spec = importlib.util.spec_from_file_location("setup_local", path)
        cls.module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(cls.module)

    def test_creates_private_environment_with_matching_credentials(self):
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory) / ".env"
            self.module.create_environment(target)
            values = dict(line.split("=", 1) for line in target.read_text().splitlines())
            self.assertEqual(target.stat().st_mode & 0o777, 0o600)
            self.assertGreaterEqual(len(values["POSTGRES_PASSWORD"]), 32)
            self.assertIn(values["POSTGRES_PASSWORD"], values["DATABASE_URL"])
            self.assertIn("@postgres:5432/", values["DATABASE_URL_CONTAINER"])
            self.assertIn(f"@localhost:{values['POSTGRES_PORT']}/", values["DATABASE_URL"])
            self.assertEqual(values["MINIO_ROOT_USER"], values["S3_ACCESS_KEY"])
            self.assertEqual(values["MINIO_ROOT_PASSWORD"], values["S3_SECRET_KEY"])
            self.assertEqual(values["S3_ENDPOINT"], "localhost:9000")

    def test_never_overwrites_existing_environment(self):
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory) / ".env"
            target.write_text("existing user configuration")
            with self.assertRaises(FileExistsError):
                self.module.create_environment(target)
            self.assertEqual(target.read_text(), "existing user configuration")

    def test_credentials_are_generated_independently(self):
        with tempfile.TemporaryDirectory() as directory:
            first, second = Path(directory) / "one", Path(directory) / "two"
            self.module.create_environment(first)
            self.module.create_environment(second)
            self.assertNotEqual(first.read_text(), second.read_text())

    def test_never_follows_existing_or_dangling_symlinks(self):
        with tempfile.TemporaryDirectory() as directory:
            original = Path(directory) / "original"
            original.write_text("unchanged")
            target = Path(directory) / ".env"
            target.symlink_to(original)
            with self.assertRaises(FileExistsError):
                self.module.create_environment(target)
            self.assertEqual(original.read_text(), "unchanged")
            original.unlink()
            with self.assertRaises(FileExistsError):
                self.module.create_environment(target)
            self.assertFalse(original.exists())


if __name__ == "__main__":
    unittest.main()
