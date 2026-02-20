import io

from app.config import settings


class StorageService:
    def __init__(self, s3_client):
        self.s3 = s3_client
        self.bucket = settings.s3_bucket

    def upload_bundle(self, name: str, version: str, file_bytes: bytes) -> str:
        """Upload a .tar.gz bundle to S3/MinIO. Returns the object key."""
        key = f"{name}/{version}.tar.gz"
        self.s3.put_object(
            Bucket=self.bucket,
            Key=key,
            Body=file_bytes,
            ContentType="application/gzip",
        )
        return key

    def download_bundle(self, bundle_key: str) -> bytes:
        """Download a bundle from S3/MinIO. Returns raw bytes."""
        response = self.s3.get_object(Bucket=self.bucket, Key=bundle_key)
        return response["Body"].read()

    def check_health(self) -> bool:
        """Check if the S3/MinIO bucket is accessible."""
        try:
            self.s3.head_bucket(Bucket=self.bucket)
            return True
        except Exception:
            return False
