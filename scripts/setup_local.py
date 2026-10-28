"""Create local-only credentials without replacing an existing environment."""

import os
import secrets
from pathlib import Path


def create_environment(target: Path) -> None:
    password = secrets.token_hex(24)
    storage_user = "museum-" + secrets.token_hex(8)
    storage_password = secrets.token_hex(24)
    values = {
        "POSTGRES_USER": "museum",
        "POSTGRES_DB": "museum",
        "POSTGRES_PORT": "55432",
        "POSTGRES_PASSWORD": password,
        "DATABASE_URL": f"postgres://museum:{password}@localhost:55432/museum?sslmode=disable",
        "DATABASE_URL_CONTAINER": f"postgres://museum:{password}@postgres:5432/museum?sslmode=disable",
        "MINIO_ROOT_USER": storage_user,
        "MINIO_ROOT_PASSWORD": storage_password,
        "API_PORT": "8080",
        "S3_ENDPOINT": "localhost:9000",
        "S3_ACCESS_KEY": storage_user,
        "S3_SECRET_KEY": storage_password,
        "S3_BUCKET": "recollections",
        "S3_USE_SSL": "false",
    }
    descriptor = os.open(target, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(descriptor, "w") as environment:
        environment.write("".join(f"{key}={value}\n" for key, value in values.items()))


if __name__ == "__main__":
    destination = Path(__file__).resolve().parents[1] / ".env"
    try:
        create_environment(destination)
    except FileExistsError:
        print(".env already exists; left unchanged.")
    else:
        print("Created private .env with random local credentials.")
