from fastapi import APIRouter, Depends
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession

from app.dependencies import get_db
from app.models import Category, Skill
from app.schemas import CategoriesResponse, CategoryItem

router = APIRouter()


@router.get("/categories", response_model=CategoriesResponse)
async def list_categories(db: AsyncSession = Depends(get_db)):
    """List all categories with their skill counts."""
    # Get categories ordered by sort_order
    result = await db.execute(
        select(Category).order_by(Category.sort_order)
    )
    categories = result.scalars().all()

    items = []
    for cat in categories:
        # Count skills in this category
        count_result = await db.execute(
            select(func.count()).select_from(Skill).where(Skill.category_id == cat.id)
        )
        skill_count = count_result.scalar() or 0

        items.append(
            CategoryItem(
                name=cat.name,
                label=cat.label,
                icon=cat.icon,
                skill_count=skill_count,
            )
        )

    return CategoriesResponse(categories=items)
