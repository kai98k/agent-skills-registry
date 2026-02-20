import io

import pytest
import pytest_asyncio
from httpx import AsyncClient

from tests.conftest import VALID_SKILL_MD, make_bundle_bytes


@pytest.mark.asyncio
class TestPublish:
    async def test_publish_success(self, client: AsyncClient):
        bundle = make_bundle_bytes(VALID_SKILL_MD)
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code == 201
        data = response.json()
        assert data["name"] == "test-skill"
        assert data["version"] == "1.0.0"
        assert data["checksum"].startswith("sha256:")
        assert "published_at" in data
        assert "providers" in data

    async def test_publish_no_auth(self, client: AsyncClient):
        bundle = make_bundle_bytes(VALID_SKILL_MD)
        response = await client.post(
            "/v1/skills/publish",
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code in (401, 422)

    async def test_publish_invalid_token(self, client: AsyncClient):
        bundle = make_bundle_bytes(VALID_SKILL_MD)
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer invalid-token"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code == 401

    async def test_publish_duplicate_version(self, client: AsyncClient):
        bundle = make_bundle_bytes(VALID_SKILL_MD)
        # First publish
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code == 201

        # Second publish with same version
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code == 409

    async def test_publish_name_taken_by_other(self, client: AsyncClient):
        bundle = make_bundle_bytes(VALID_SKILL_MD)
        # First user publishes
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code == 201

        # Other user tries same name, different version
        other_skill_md = VALID_SKILL_MD.replace("1.0.0", "2.0.0").replace('author: "dev"', 'author: "other"')
        bundle2 = make_bundle_bytes(other_skill_md)
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer other-token-99999"},
            files={"file": ("test-skill.tar.gz", bundle2, "application/gzip")},
        )
        assert response.status_code == 403

    async def test_publish_author_mismatch(self, client: AsyncClient):
        wrong_author = VALID_SKILL_MD.replace('author: "dev"', 'author: "someone-else"')
        bundle = make_bundle_bytes(wrong_author)
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code == 400
        assert "author" in response.json()["detail"].lower()

    async def test_publish_invalid_tar(self, client: AsyncClient):
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("bad.tar.gz", b"not a tar", "application/gzip")},
        )
        assert response.status_code == 400

    async def test_publish_with_explicit_providers(self, client: AsyncClient):
        bundle = make_bundle_bytes(VALID_SKILL_MD)
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
            data={"providers": "claude,gemini"},
        )
        assert response.status_code == 201
        data = response.json()
        assert "claude" in data["providers"]
        assert "gemini" in data["providers"]

    async def test_publish_auto_detect_provider_from_bundle(self, client: AsyncClient):
        bundle = make_bundle_bytes(
            VALID_SKILL_MD,
            extra_files={".claude/settings.json": "{}"},
        )
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code == 201
        assert "claude" in response.json()["providers"]

    async def test_publish_claude_forbidden_name(self, client: AsyncClient):
        bad_name_md = VALID_SKILL_MD.replace('name: "test-skill"', 'name: "claude-helper"')
        bundle = make_bundle_bytes(
            bad_name_md,
            extra_files={".claude/settings.json": "{}"},
        )
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("bad.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code == 400
        assert "claude" in response.json()["detail"].lower()

    async def test_publish_missing_skill_md(self, client: AsyncClient):
        import tarfile

        buf = io.BytesIO()
        with tarfile.open(fileobj=buf, mode="w:gz") as tar:
            data = b"readme"
            info = tarfile.TarInfo(name="README.md")
            info.size = len(data)
            tar.addfile(info, io.BytesIO(data))
        bundle = buf.getvalue()

        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code == 400
        assert "SKILL.md" in response.json()["detail"]
