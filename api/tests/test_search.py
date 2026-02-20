import pytest
from httpx import AsyncClient

from tests.conftest import VALID_SKILL_MD, make_bundle_bytes


@pytest.mark.asyncio
class TestSearch:
    async def _publish(self, client: AsyncClient, name: str = "test-skill", **overrides):
        md = VALID_SKILL_MD
        if name != "test-skill":
            md = md.replace('name: "test-skill"', f'name: "{name}"')
        for key, val in overrides.items():
            if key == "version":
                md = md.replace('version: "1.0.0"', f'version: "{val}"')
            elif key == "description":
                md = md.replace('description: "A test skill for validation"', f'description: "{val}"')
        bundle = make_bundle_bytes(md)
        response = await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": (f"{name}.tar.gz", bundle, "application/gzip")},
            data=overrides.get("form_data", {}),
        )
        assert response.status_code == 201
        return response.json()

    async def test_get_skill_info(self, client: AsyncClient):
        await self._publish(client)
        response = await client.get("/v1/skills/test-skill")
        assert response.status_code == 200
        data = response.json()
        assert data["name"] == "test-skill"
        assert data["owner"] == "dev"
        assert data["latest_version"]["version"] == "1.0.0"
        assert "providers" in data["latest_version"]

    async def test_get_skill_not_found(self, client: AsyncClient):
        response = await client.get("/v1/skills/nonexistent")
        assert response.status_code == 404

    async def test_list_versions(self, client: AsyncClient):
        await self._publish(client)
        v2_md = VALID_SKILL_MD.replace('version: "1.0.0"', 'version: "2.0.0"')
        bundle2 = make_bundle_bytes(v2_md)
        await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle2, "application/gzip")},
        )

        response = await client.get("/v1/skills/test-skill/versions")
        assert response.status_code == 200
        data = response.json()
        assert data["name"] == "test-skill"
        assert len(data["versions"]) == 2
        version_strings = {v["version"] for v in data["versions"]}
        assert "1.0.0" in version_strings
        assert "2.0.0" in version_strings
        assert "providers" in data["versions"][0]

    async def test_list_versions_not_found(self, client: AsyncClient):
        response = await client.get("/v1/skills/nonexistent/versions")
        assert response.status_code == 404

    async def test_search_by_keyword(self, client: AsyncClient):
        await self._publish(client)
        response = await client.get("/v1/skills?q=test")
        assert response.status_code == 200
        data = response.json()
        assert data["total"] >= 1
        assert any(r["name"] == "test-skill" for r in data["results"])

    async def test_search_no_results(self, client: AsyncClient):
        response = await client.get("/v1/skills?q=zzzznonexistent")
        assert response.status_code == 200
        data = response.json()
        assert data["total"] == 0
        assert data["results"] == []

    async def test_search_results_include_providers(self, client: AsyncClient):
        bundle = make_bundle_bytes(VALID_SKILL_MD)
        await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
            data={"providers": "claude"},
        )
        response = await client.get("/v1/skills?q=test")
        assert response.status_code == 200
        data = response.json()
        assert len(data["results"]) >= 1
        assert "providers" in data["results"][0]
        assert "claude" in data["results"][0]["providers"]

    async def test_search_filter_by_provider(self, client: AsyncClient):
        # Publish with claude provider
        bundle = make_bundle_bytes(VALID_SKILL_MD)
        await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
            data={"providers": "claude"},
        )

        # Publish another skill with gemini provider
        gemini_md = VALID_SKILL_MD.replace('name: "test-skill"', 'name: "gemini-skill"')
        bundle2 = make_bundle_bytes(gemini_md)
        await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("gemini-skill.tar.gz", bundle2, "application/gzip")},
            data={"providers": "gemini"},
        )

        # Search with provider filter
        response = await client.get("/v1/skills?provider=claude")
        data = response.json()
        for result in data["results"]:
            assert "claude" in result["providers"]

    async def test_search_pagination(self, client: AsyncClient):
        await self._publish(client)
        response = await client.get("/v1/skills?page=1&per_page=1")
        assert response.status_code == 200
        data = response.json()
        assert data["page"] == 1
        assert data["per_page"] == 1

    async def test_search_all_skills(self, client: AsyncClient):
        await self._publish(client)
        response = await client.get("/v1/skills")
        assert response.status_code == 200
        data = response.json()
        assert data["total"] >= 1
