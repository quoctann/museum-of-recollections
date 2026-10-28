"""Pinned Go tooling and disposable PostgreSQL integration environment."""

import argparse
import os
from pathlib import Path
import secrets
import shutil
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parents[1]
API = ROOT / "apps" / "api"
GOOSE = "github.com/pressly/goose/v3/cmd/goose@v3.24.1"
SQLC = "github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0"
GOOSE_TAGS = "no_clickhouse,no_libsql,no_mssql,no_mysql,no_sqlite3,no_vertica,no_ydb"


def migrate(action):
    if action not in {"up", "down", "status"}:
        raise ValueError("Only up, down (one migration), and status are supported")
    if not os.environ.get("DATABASE_URL"):
        raise ValueError("DATABASE_URL must be explicitly configured")
    env = {**os.environ, "GOOSE_DRIVER": "postgres",
           "GOOSE_DBSTRING": os.environ["DATABASE_URL"],
           "GOOSE_MIGRATION_DIR": str(API / "db" / "migrations")}
    subprocess.run(["go", "run", f"-tags={GOOSE_TAGS}", GOOSE, action],
                   cwd=API, env=env, check=True)


def generated_files(directory):
    return {str(path.relative_to(directory)): path.read_bytes()
            for path in directory.rglob("*") if path.is_file()}


def generate(check=False):
    if not check:
        subprocess.run(["go", "run", SQLC, "generate"], cwd=API, check=True)
        return
    # Compare bytes rather than git diff: also works before the first commit.
    with tempfile.TemporaryDirectory(prefix="museum-sqlc-") as temporary:
        target = Path(temporary)
        shutil.copy(API / "sqlc.yaml", target)
        for name in ("migrations", "queries"):
            shutil.copytree(API / "db" / name, target / "db" / name)
        subprocess.run(["go", "run", SQLC, "generate", "-f",
                        str(target / "sqlc.yaml")], cwd=API, check=True)
        if generated_files(target / "db" / "generated") != generated_files(
            API / "db" / "generated"
        ):
            raise ValueError("sqlc output is stale; run make sqlc")


def integration():
    project = "museum-test-" + secrets.token_hex(6)
    base_env = {key: value for key, value in os.environ.items()
                if key not in {"DATABASE_URL", "TEST_DATABASE_URL"}}
    env = {**base_env, "TEST_POSTGRES_PASSWORD": secrets.token_urlsafe(32)}
    compose = ["docker", "compose", "--env-file", os.devnull,
               "-f", str(ROOT / "compose.test.yaml"),
               "-p", project]
    failed = False
    try:
        subprocess.run([*compose, "up", "-d", "--wait"], cwd=ROOT,
                       env=env, check=True)
        result = subprocess.run([*compose, "port", "postgres", "5432"],
                                env=env, check=True, capture_output=True, text=True)
        address = result.stdout.strip()
        if not address.startswith("127.0.0.1:") or not address[10:].isdigit():
            raise ValueError("Test PostgreSQL must use a loopback port")
        database_url = (
            f"postgres://museum_test:{env['TEST_POSTGRES_PASSWORD']}"
            f"@{address}/postgres?sslmode=disable"
        )
        env = {**env, "TEST_DATABASE_URL": database_url}
        # Exercise the pinned migration tool as well as adapter integration tests.
        migration_env = {**env, "DATABASE_URL": env["TEST_DATABASE_URL"]}
        for action in ("up", "up", "down", "up"):
            subprocess.run(["python3", str(Path(__file__).resolve()), action],
                           env=migration_env, check=True)
        subprocess.run(["go", "test", "-race", "-count=1", "-coverprofile="
                        + str(API / "coverage-integration.out"), "./..."],
                       cwd=API, env=env, check=True)
    except BaseException:
        failed = True
        raise
    finally:
        # Only this randomly named project's disposable resources are removed.
        try:
            subprocess.run([*compose, "down", "--volumes"], cwd=ROOT,
                           env=env, check=True)
        except (OSError, subprocess.CalledProcessError):
            if not failed:
                raise
            print(f"Cleanup also failed; inspect Docker project {project}.",
                  file=sys.stderr)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("up", "down", "status", "sqlc",
                                            "sqlc-check", "integration"))
    command = parser.parse_args().command
    if command in {"up", "down", "status"}:
        migrate(command)
    elif command in {"sqlc", "sqlc-check"}:
        generate(check=command == "sqlc-check")
    else:
        integration()


if __name__ == "__main__":
    try:
        main()
    except (ValueError, subprocess.CalledProcessError) as error:
        # Do not print process environments or database connection strings.
        raise SystemExit(str(error)) from None
