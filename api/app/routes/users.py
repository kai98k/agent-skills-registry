from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession

from app.dependencies import get_db
from app.models import Skill, SkillVersion, User
from app.schemas import UserResponse, UserSkillItem

router = APIRouter()


@router.get("/users/{username}", response_model=UserResponse)
async def get_user_profile(username: str, db: AsyncSession = Depends(get_db)):
    """Get public user profile and their published skills."""
    result = await db.execute(select(User).where(User.username == username))
    user = result.scalar_one_or_none()
    if user is None:
        raise HTTPException(status_code=404, detail=f"User '{username}' not found")

    # Get user's skills
    skills_result = await db.execute(
        select(Skill)
        .where(Skill.owner_id == user.id)
        .order_by(Skill.updated_at.desc())
    )
    skills = skills_result.scalars().all()

    skill_items = []
    total_downloads = 0
    total_stars = 0

    for skill in skills:
        total_downloads += skill.downloads or 0
        total_stars += skill.stars_count or 0

        # Get latest version
        latest_result = await db.execute(
            select(SkillVersion)
            .where(SkillVersion.skill_id == skill.id)
            .order_by(SkillVersion.published_at.desc())
            .limit(1)
        )
        latest = latest_result.scalar_one_or_none()
        if latest is None:
            continue

        skill_items.append(
            UserSkillItem(
                name=skill.name,
                description=latest.meta.get("description", ""),
                downloads=skill.downloads or 0,
                stars_count=skill.stars_count or 0,
                latest_version=latest.version,
                updated_at=skill.updated_at,
            )
        )

    return UserResponse(
        username=user.username,
        display_name=user.display_name,
        avatar_url=user.avatar_url,
        bio=user.bio,
        created_at=user.created_at,
        skills=skill_items,
        total_downloads=total_downloads,
        total_stars=total_stars,
    )
