import pytest
from httpx import AsyncClient

from tests.conftest import VALID_SKILL_MD, make_bundle_bytes


@pytest.mark.asyncio
class TestStars:
    async def _publish(self, client: AsyncClient):
        bundle = make_bundle_bytes(VALID_SKILL_MD)
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
        )
        assert response.status_code == 201

    async def test_star_skill(self, client: AsyncClient):
        await self._publish(client)
        response = await client.post(
            "/v1/skills/test-skill/star",
            headers={"Authorization": "Bearer dev-token-12345"},
        )
        assert response.status_code == 200
        data = response.json()
        assert data["starred"] is True
        assert data["stars_count"] == 1

    async def test_star_no_auth(self, client: AsyncClient):
        await self._publish(client)
        response = await client.post("/v1/skills/test-skill/star")
        assert response.status_code in (401, 422)

    async def test_star_nonexistent_skill(self, client: AsyncClient):
        response = await client.post(
            "/v1/skills/nonexistent/star",
            headers={"Authorization": "Bearer dev-token-12345"},
        )
        assert response.status_code == 404

    async def test_star_duplicate(self, client: AsyncClient):
        await self._publish(client)
        await client.post(
            "/v1/skills/test-skill/star",
            headers={"Authorization": "Bearer dev-token-12345"},
        )
        response = await client.post(
            "/v1/skills/test-skill/star",
            headers={"Authorization": "Bearer dev-token-12345"},
        )
        assert response.status_code == 409

    async def test_unstar_skill(self, client: AsyncClient):
        await self._publish(client)
        # Star first
        await client.post(
            "/v1/skills/test-skill/star",
            headers={"Authorization": "Bearer dev-token-12345"},
        )
        # Then unstar
        response = await client.delete(
            "/v1/skills/test-skill/star",
            headers={"Authorization": "Bearer dev-token-12345"},
        )
        assert response.status_code == 200
        data = response.json()
        assert data["starred"] is False
        assert data["stars_count"] == 0

    async def test_unstar_not_starred(self, client: AsyncClient):
        await self._publish(client)
        response = await client.delete(
            "/v1/skills/test-skill/star",
            headers={"Authorization": "Bearer dev-token-12345"},
        )
        assert response.status_code == 404

    async def test_unstar_no_auth(self, client: AsyncClient):
        await self._publish(client)
        response = await client.delete("/v1/skills/test-skill/star")
        assert response.status_code in (401, 422)

    async def test_starred_by_me_in_skill_detail(self, client: AsyncClient):
        await self._publish(client)

        # Before starring - no auth
        response = await client.get("/v1/skills/test-skill")
        assert response.json()["starred_by_me"] is False

        # Star it
        await client.post(
            "/v1/skills/test-skill/star",
            headers={"Authorization": "Bearer dev-token-12345"},
        )

        # After starring - with auth
        response = await client.get(
            "/v1/skills/test-skill",
            headers={"Authorization": "Bearer dev-token-12345"},
        )
        data = response.json()
        assert data["starred_by_me"] is True
        assert data["stars_count"] == 1

    async def test_stars_count_persists(self, client: AsyncClient):
        await self._publish(client)

        # Star with dev user
        await client.post(
            "/v1/skills/test-skill/star",
            headers={"Authorization": "Bearer dev-token-12345"},
        )

        # Star with other user
        await client.post(
            "/v1/skills/test-skill/star",
            headers={"Authorization": "Bearer other-token-99999"},
        )

        # Check count
        response = await client.get("/v1/skills/test-skill")
        assert response.json()["stars_count"] == 2

    async def test_unstar_nonexistent_skill(self, client: AsyncClient):
        response = await client.delete(
            "/v1/skills/nonexistent/star",
            headers={"Authorization": "Bearer dev-token-12345"},
        )
        assert response.status_code == 404
