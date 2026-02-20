import asyncio
import io
import tarfile
import uuid

import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient
from sqlalchemy import event
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from app.config import settings
from app.dependencies import get_db, get_s3
from app.main import app
from app.models import Base, User


# Use SQLite for tests
TEST_DATABASE_URL = "sqlite+aiosqlite:///:memory:"

test_engine = create_async_engine(TEST_DATABASE_URL, echo=False)
TestSessionFactory = async_sessionmaker(test_engine, expire_on_commit=False)


@pytest.fixture(scope="session")
def event_loop():
    loop = asyncio.new_event_loop()
    yield loop
    loop.close()


@pytest_asyncio.fixture
async def test_db():
    async with test_engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

    async with TestSessionFactory() as session:
        # Seed test user
        dev_user = User(
            id="00000000-0000-0000-0000-000000000001",
            username="dev",
            api_token="dev-token-12345",
        )
        other_user = User(
            id="00000000-0000-0000-0000-000000000002",
            username="other",
            api_token="other-token-99999",
        )
        session.add(dev_user)
        session.add(other_user)
        await session.commit()

    async with TestSessionFactory() as session:
        yield session

    async with test_engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)


class MockS3:
    """In-memory S3 mock for tests."""

    def __init__(self):
        self._objects: dict[str, bytes] = {}

    def put_object(self, Bucket: str, Key: str, Body: bytes, **kwargs):
        self._objects[f"{Bucket}/{Key}"] = Body

    def get_object(self, Bucket: str, Key: str):
        full_key = f"{Bucket}/{Key}"
        if full_key not in self._objects:
            raise Exception(f"NoSuchKey: {Key}")
        body = self._objects[full_key]
        return {"Body": io.BytesIO(body), "ContentLength": len(body)}

    def head_bucket(self, Bucket: str):
        return {}

    def head_object(self, Bucket: str, Key: str):
        full_key = f"{Bucket}/{Key}"
        if full_key not in self._objects:
            raise Exception(f"NoSuchKey: {Key}")
        return {"ContentLength": len(self._objects[full_key])}


@pytest.fixture
def mock_s3():
    return MockS3()


@pytest_asyncio.fixture
async def client(test_db: AsyncSession, mock_s3: MockS3):
    async def override_get_db():
        yield test_db

    def override_get_s3():
        return mock_s3

    app.dependency_overrides[get_db] = override_get_db
    app.dependency_overrides[get_s3] = override_get_s3

    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test",
    ) as ac:
        yield ac

    app.dependency_overrides.clear()


def make_bundle_bytes(
    skill_md_content: str,
    extra_files: dict[str, str] | None = None,
    nested: bool = False,
) -> bytes:
    """Create a .tar.gz bundle in memory with given SKILL.md content."""
    buf = io.BytesIO()
    with tarfile.open(fileobj=buf, mode="w:gz") as tar:
        prefix = "my-skill/" if nested else ""

        # Add SKILL.md
        skill_data = skill_md_content.encode("utf-8")
        info = tarfile.TarInfo(name=f"{prefix}SKILL.md")
        info.size = len(skill_data)
        tar.addfile(info, io.BytesIO(skill_data))

        # Add extra files if provided
        if extra_files:
            for name, content in extra_files.items():
                data = content.encode("utf-8")
                info = tarfile.TarInfo(name=f"{prefix}{name}")
                info.size = len(data)
                tar.addfile(info, io.BytesIO(data))

    return buf.getvalue()


VALID_SKILL_MD = """\
---
name: "test-skill"
version: "1.0.0"
description: "A test skill for validation"
author: "dev"
tags:
  - test
  - example
license: "MIT"
---

# Test Skill

This is a test skill.
"""
