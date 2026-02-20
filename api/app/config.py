from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    # Database
    database_url: str = "postgresql+asyncpg://dev:devpass@localhost:5432/agentskills"

    # MinIO / S3
    s3_endpoint: str = "http://localhost:9000"
    s3_access_key: str = "minioadmin"
    s3_secret_key: str = "minioadmin"
    s3_bucket: str = "skills"
    s3_region: str = "us-east-1"

    # App
    max_bundle_size: int = 50 * 1024 * 1024  # 50MB
    max_decompressed_size: int = 200 * 1024 * 1024  # 200MB
    api_prefix: str = "/v1"

    model_config = {"env_file": ".env", "extra": "ignore"}


settings = Settings()
