import pytest
from httpx import AsyncClient
from unittest.mock import AsyncMock, patch, MagicMock

import httpx


def _mock_github_response(status_code=200, json_data=None):
    """Create a mock httpx Response."""
    if json_data is None:
        json_data = {
            "id": 12345,
            "login": "ghuser",
            "name": "GitHub User",
            "avatar_url": "https://avatars.githubusercontent.com/u/12345",
        }

    response = MagicMock(spec=httpx.Response)
    response.status_code = status_code
    response.json.return_value = json_data
    return response


@pytest.mark.asyncio
class TestGitHubAuth:
    @patch("app.routes.auth.httpx.AsyncClient")
    async def test_auth_new_user(self, mock_client_cls, client: AsyncClient):
        mock_ctx = AsyncMock()
        mock_ctx.get.return_value = _mock_github_response()
        mock_client_cls.return_value.__aenter__ = AsyncMock(return_value=mock_ctx)
        mock_client_cls.return_value.__aexit__ = AsyncMock(return_value=False)

        response = await client.post(
            "/v1/auth/github",
            json={"github_access_token": "gho_test123"},
        )
        assert response.status_code == 200
        data = response.json()
        assert data["username"] == "ghuser"
        assert data["display_name"] == "GitHub User"
        assert data["api_token"].startswith("ask-")

    @patch("app.routes.auth.httpx.AsyncClient")
    async def test_auth_invalid_token(self, mock_client_cls, client: AsyncClient):
        mock_ctx = AsyncMock()
        mock_ctx.get.return_value = _mock_github_response(status_code=401)
        mock_client_cls.return_value.__aenter__ = AsyncMock(return_value=mock_ctx)
        mock_client_cls.return_value.__aexit__ = AsyncMock(return_value=False)

        response = await client.post(
            "/v1/auth/github",
            json={"github_access_token": "bad-token"},
        )
        assert response.status_code == 401

    @patch("app.routes.auth.httpx.AsyncClient")
    async def test_auth_existing_user_by_github_id(self, mock_client_cls, client: AsyncClient):
        mock_ctx = AsyncMock()
        mock_ctx.get.return_value = _mock_github_response()
        mock_client_cls.return_value.__aenter__ = AsyncMock(return_value=mock_ctx)
        mock_client_cls.return_value.__aexit__ = AsyncMock(return_value=False)

        # First auth creates user
        response1 = await client.post(
            "/v1/auth/github",
            json={"github_access_token": "gho_test123"},
        )
        assert response1.status_code == 200
        token1 = response1.json()["api_token"]

        # Second auth with same GitHub ID returns same user
        response2 = await client.post(
            "/v1/auth/github",
            json={"github_access_token": "gho_test456"},
        )
        assert response2.status_code == 200
        token2 = response2.json()["api_token"]
        assert token1 == token2

    @patch("app.routes.auth.httpx.AsyncClient")
    async def test_auth_links_existing_username(self, mock_client_cls, client: AsyncClient):
        """GitHub auth should link to existing user with matching username."""
        mock_ctx = AsyncMock()
        # Use 'dev' username which already exists in test DB
        mock_ctx.get.return_value = _mock_github_response(json_data={
            "id": 99999,
            "login": "dev",
            "name": "Dev User",
            "avatar_url": "https://avatars.githubusercontent.com/u/99999",
        })
        mock_client_cls.return_value.__aenter__ = AsyncMock(return_value=mock_ctx)
        mock_client_cls.return_value.__aexit__ = AsyncMock(return_value=False)

        response = await client.post(
            "/v1/auth/github",
            json={"github_access_token": "gho_dev_token"},
        )
        assert response.status_code == 200
        data = response.json()
        assert data["username"] == "dev"
        # Should keep existing API token
        assert data["api_token"] == "dev-token-12345"

    @patch("app.routes.auth.httpx.AsyncClient")
    async def test_auth_missing_github_id(self, mock_client_cls, client: AsyncClient):
        mock_ctx = AsyncMock()
        mock_ctx.get.return_value = _mock_github_response(json_data={
            "login": "noone",
            "name": "No ID",
        })
        mock_client_cls.return_value.__aenter__ = AsyncMock(return_value=mock_ctx)
        mock_client_cls.return_value.__aexit__ = AsyncMock(return_value=False)

        response = await client.post(
            "/v1/auth/github",
            json={"github_access_token": "gho_bad"},
        )
        assert response.status_code == 400

    async def test_auth_missing_body(self, client: AsyncClient):
        response = await client.post("/v1/auth/github", json={})
        assert response.status_code == 422


@pytest.mark.asyncio
class TestCategories:
    async def test_list_categories_empty(self, client: AsyncClient):
        response = await client.get("/v1/categories")
        assert response.status_code == 200
        data = response.json()
        assert "categories" in data
        assert isinstance(data["categories"], list)


@pytest.mark.asyncio
class TestUserProfile:
    async def test_get_user_profile(self, client: AsyncClient):
        response = await client.get("/v1/users/dev")
        assert response.status_code == 200
        data = response.json()
        assert data["username"] == "dev"
        assert "skills" in data
        assert "total_downloads" in data
        assert "total_stars" in data

    async def test_get_user_profile_with_skills(self, client: AsyncClient):
        # Publish a skill first
        from tests.conftest import VALID_SKILL_MD, make_bundle_bytes

        bundle = make_bundle_bytes(VALID_SKILL_MD)
        await client.post(
            "/v1/skills/publish",
            headers={"Authorization": "Bearer dev-token-12345"},
            files={"file": ("test-skill.tar.gz", bundle, "application/gzip")},
        )

        response = await client.get("/v1/users/dev")
        assert response.status_code == 200
        data = response.json()
        assert len(data["skills"]) >= 1
        assert data["skills"][0]["name"] == "test-skill"

    async def test_get_user_not_found(self, client: AsyncClient):
        response = await client.get("/v1/users/nonexistent")
        assert response.status_code == 404
