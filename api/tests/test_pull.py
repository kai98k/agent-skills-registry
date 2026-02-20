import pytest
from httpx import AsyncClient

from tests.conftest import VALID_SKILL_MD, make_bundle_bytes


@pytest.mark.asyncio
class TestPull:
    async def _publish(self, client: AsyncClient, skill_md: str = VALID_SKILL_MD):
        bundle = make_bundle_bytes(skill_md)
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code == 201
        return response.json()

    async def test_download_latest(self, client: AsyncClient):
        await self._publish(client)

        response = await client.get("/v1/skills/test-skill/versions/1.0.0/download")
        assert response.status_code == 200
        assert response.headers["content-type"] == "application/gzip"
        assert "X-Checksum-SHA256" in response.headers
        assert len(response.content) > 0

    async def test_download_specific_version(self, client: AsyncClient):
        await self._publish(client)

        # Publish v2
        v2_md = VALID_SKILL_MD.replace('version: "1.0.0"', 'version: "2.0.0"')
        await self._publish(client, v2_md)

        # Download v1 specifically
        response = await client.get("/v1/skills/test-skill/versions/1.0.0/download")
        assert response.status_code == 200

        # Download v2
        response = await client.get("/v1/skills/test-skill/versions/2.0.0/download")
        assert response.status_code == 200

    async def test_download_nonexistent_skill(self, client: AsyncClient):
        response = await client.get("/v1/skills/nonexistent/versions/1.0.0/download")
        assert response.status_code == 404

    async def test_download_nonexistent_version(self, client: AsyncClient):
        await self._publish(client)
        response = await client.get("/v1/skills/test-skill/versions/9.9.9/download")
        assert response.status_code == 404

    async def test_download_increments_count(self, client: AsyncClient):
        await self._publish(client)

        # Check initial downloads
        info = await client.get("/v1/skills/test-skill")
        initial_downloads = info.json()["downloads"]

        # Download
        await client.get("/v1/skills/test-skill/versions/1.0.0/download")

        # Check downloads incremented
        info = await client.get("/v1/skills/test-skill")
        assert info.json()["downloads"] == initial_downloads + 1

    async def test_content_disposition_header(self, client: AsyncClient):
        await self._publish(client)
        response = await client.get("/v1/skills/test-skill/versions/1.0.0/download")
        assert "test-skill-1.0.0.tar.gz" in response.headers["content-disposition"]
