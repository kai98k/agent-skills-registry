import secrets

import httpx
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.dependencies import get_db
from app.models import User
from app.schemas import GitHubAuthRequest, GitHubAuthResponse

router = APIRouter()


@router.post("/auth/github", response_model=GitHubAuthResponse)
async def github_auth(
    body: GitHubAuthRequest,
    db: AsyncSession = Depends(get_db),
):
    """Exchange a GitHub access token for an AgentSkills API token."""
    # Fetch GitHub user info
    async with httpx.AsyncClient() as client:
        resp = await client.get(
            "https://api.github.com/user",
            headers={
                "Authorization": f"Bearer {body.github_access_token}",
                "Accept": "application/json",
            },
        )
    if resp.status_code != 200:
        raise HTTPException(status_code=401, detail="Invalid GitHub access token")

    gh_data = resp.json()
    github_id = gh_data.get("id")
    username = gh_data.get("login", "")
    display_name = gh_data.get("name") or username
    avatar_url = gh_data.get("avatar_url", "")

    if not github_id:
        raise HTTPException(status_code=400, detail="Could not retrieve GitHub user ID")

    # Look up existing user by github_id
    result = await db.execute(select(User).where(User.github_id == github_id))
    user = result.scalar_one_or_none()

    if user is None:
        # Check if username is taken (e.g. by a CLI-only user)
        result = await db.execute(select(User).where(User.username == username))
        existing = result.scalar_one_or_none()
        if existing:
            # Link GitHub to existing user
            existing.github_id = github_id
            existing.display_name = display_name
            existing.avatar_url = avatar_url
            user = existing
        else:
            # Create new user
            api_token = f"ask-{secrets.token_hex(24)}"
            user = User(
                username=username,
                api_token=api_token,
                display_name=display_name,
                avatar_url=avatar_url,
                github_id=github_id,
            )
            db.add(user)
    else:
        # Update profile fields
        user.display_name = display_name
        user.avatar_url = avatar_url

    await db.commit()
    await db.refresh(user)

    return GitHubAuthResponse(
        username=user.username,
        display_name=user.display_name,
        avatar_url=user.avatar_url,
        api_token=user.api_token,
    )
