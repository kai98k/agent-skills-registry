from datetime import datetime, timezone
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, Query, Request, UploadFile, File, Form
from sqlalchemy import select, func, update
from sqlalchemy.ext.asyncio import AsyncSession

from app.config import settings
from app.dependencies import get_db, get_current_user, get_optional_user, get_s3
from app.models import Category, Skill, SkillVersion, Star, User
from app.schemas import (
    ErrorResponse,
    PublishResponse,
    SearchResponse,
    SearchResultItem,
    SkillResponse,
    SkillVersionDetail,
    SkillVersionSummary,
    SkillVersionsResponse,
    StarResponse,
)
from app.services.markdown import render_markdown
from app.services.parser import (
    ParseError,
    compute_checksum,
    extract_and_parse,
    extract_providers,
    validate_provider_constraints,
)
from app.services.storage import StorageService

router = APIRouter()


@router.post("/skills/publish", status_code=201, response_model=PublishResponse)
async def publish_skill(
    file: UploadFile = File(...),
    providers: Optional[str] = Form(default=None),
    category: Optional[str] = Form(default=None),
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
    s3=Depends(get_s3),
):
    # Read and validate file size
    file_bytes = await file.read()
    if len(file_bytes) > settings.max_bundle_size:
        raise HTTPException(status_code=413, detail="Bundle exceeds 50MB limit")

    # Parse the bundle
    try:
        parsed = extract_and_parse(file_bytes, max_decompressed=settings.max_decompressed_size)
    except ParseError as e:
        raise HTTPException(status_code=400, detail=str(e))

    # Validate author matches token user
    if parsed.author != user.username:
        raise HTTPException(
            status_code=400,
            detail=f"Author '{parsed.author}' does not match authenticated user '{user.username}'",
        )

    # Determine providers
    if providers:
        provider_list = sorted(set(p.strip() for p in providers.split(",") if p.strip()))
    else:
        provider_list = parsed.providers

    # Validate provider-specific constraints
    try:
        validate_provider_constraints(parsed.name, provider_list)
    except ParseError as e:
        raise HTTPException(status_code=400, detail=str(e))

    # Resolve category
    category_id = None
    if category:
        cat_result = await db.execute(select(Category).where(Category.name == category))
        cat = cat_result.scalar_one_or_none()
        if cat:
            category_id = cat.id

    # Check skill ownership
    result = await db.execute(select(Skill).where(Skill.name == parsed.name))
    skill = result.scalar_one_or_none()

    if skill is None:
        # Create new skill
        skill = Skill(name=parsed.name, owner_id=user.id, category_id=category_id)
        db.add(skill)
        await db.flush()
    elif skill.owner_id != user.id:
        raise HTTPException(
            status_code=403,
            detail=f"Skill '{parsed.name}' is owned by another user",
        )
    else:
        # Update category if provided
        if category_id:
            skill.category_id = category_id

    # Check for duplicate version
    result = await db.execute(
        select(SkillVersion).where(
            SkillVersion.skill_id == skill.id,
            SkillVersion.version == parsed.version,
        )
    )
    if result.scalar_one_or_none() is not None:
        raise HTTPException(
            status_code=409,
            detail=f"Version {parsed.version} already exists",
        )

    # Compute checksum
    checksum = compute_checksum(file_bytes)

    # Upload to S3
    storage = StorageService(s3)
    bundle_key = storage.upload_bundle(parsed.name, parsed.version, file_bytes)

    # Store provider info in metadata
    full_metadata = dict(parsed.metadata)
    full_metadata["_registry"] = {
        "providers": provider_list,
    }

    # Render SKILL.md body to HTML and cache
    readme_html = render_markdown(parsed.body) if parsed.body else None
    skill.readme_html = readme_html

    # Create version record
    version = SkillVersion(
        skill_id=skill.id,
        version=parsed.version,
        bundle_key=bundle_key,
        meta=full_metadata,
        checksum=checksum,
        size_bytes=len(file_bytes),
        providers=provider_list,
        readme_raw=parsed.body or None,
    )
    db.add(version)

    # Update skill timestamp
    skill.updated_at = datetime.now(timezone.utc)

    await db.commit()
    await db.refresh(version)

    return PublishResponse(
        name=parsed.name,
        version=parsed.version,
        checksum=f"sha256:{checksum}",
        published_at=version.published_at,
        providers=provider_list,
    )


@router.get("/skills/{name}", response_model=SkillResponse)
async def get_skill(
    name: str,
    db: AsyncSession = Depends(get_db),
    current_user: Optional[User] = Depends(get_optional_user),
):
    result = await db.execute(
        select(Skill).where(Skill.name == name)
    )
    skill = result.scalar_one_or_none()
    if skill is None:
        raise HTTPException(status_code=404, detail=f"Skill '{name}' not found")

    # Get owner
    owner_result = await db.execute(select(User).where(User.id == skill.owner_id))
    owner = owner_result.scalar_one()

    # Get category name
    category_name = None
    if skill.category_id:
        cat_result = await db.execute(select(Category).where(Category.id == skill.category_id))
        cat = cat_result.scalar_one_or_none()
        if cat:
            category_name = cat.name

    # Check if starred by current user
    starred_by_me = False
    if current_user:
        star_result = await db.execute(
            select(Star).where(Star.user_id == current_user.id, Star.skill_id == skill.id)
        )
        starred_by_me = star_result.scalar_one_or_none() is not None

    # Get latest version
    version_result = await db.execute(
        select(SkillVersion)
        .where(SkillVersion.skill_id == skill.id)
        .order_by(SkillVersion.published_at.desc())
        .limit(1)
    )
    latest = version_result.scalar_one_or_none()

    latest_detail = None
    if latest:
        latest_detail = SkillVersionDetail(
            version=latest.version,
            description=latest.meta.get("description", ""),
            checksum=f"sha256:{latest.checksum}",
            size_bytes=latest.size_bytes,
            published_at=latest.published_at,
            providers=latest.providers or ["generic"],
            metadata=latest.meta,
        )

    return SkillResponse(
        name=skill.name,
        owner=owner.username,
        owner_avatar_url=owner.avatar_url,
        downloads=skill.downloads,
        stars_count=skill.stars_count or 0,
        starred_by_me=starred_by_me,
        category=category_name,
        readme_html=skill.readme_html,
        created_at=skill.created_at,
        latest_version=latest_detail,
    )


@router.get("/skills/{name}/versions", response_model=SkillVersionsResponse)
async def list_versions(name: str, db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(Skill).where(Skill.name == name))
    skill = result.scalar_one_or_none()
    if skill is None:
        raise HTTPException(status_code=404, detail=f"Skill '{name}' not found")

    versions_result = await db.execute(
        select(SkillVersion)
        .where(SkillVersion.skill_id == skill.id)
        .order_by(SkillVersion.published_at.desc())
    )
    versions = versions_result.scalars().all()

    return SkillVersionsResponse(
        name=skill.name,
        versions=[
            SkillVersionSummary(
                version=v.version,
                checksum=f"sha256:{v.checksum}",
                size_bytes=v.size_bytes,
                published_at=v.published_at,
                providers=v.providers or ["generic"],
            )
            for v in versions
        ],
    )


@router.get("/skills/{name}/versions/{version}/download")
async def download_version(
    name: str,
    version: str,
    db: AsyncSession = Depends(get_db),
    s3=Depends(get_s3),
):
    result = await db.execute(select(Skill).where(Skill.name == name))
    skill = result.scalar_one_or_none()
    if skill is None:
        raise HTTPException(status_code=404, detail=f"Skill '{name}' not found")

    version_result = await db.execute(
        select(SkillVersion).where(
            SkillVersion.skill_id == skill.id,
            SkillVersion.version == version,
        )
    )
    sv = version_result.scalar_one_or_none()
    if sv is None:
        raise HTTPException(status_code=404, detail=f"Version '{version}' not found")

    # Increment download count
    await db.execute(
        update(Skill).where(Skill.id == skill.id).values(downloads=Skill.downloads + 1)
    )
    await db.commit()

    # Download from S3
    storage = StorageService(s3)
    try:
        data = storage.download_bundle(sv.bundle_key)
    except Exception:
        raise HTTPException(status_code=500, detail="Failed to retrieve bundle from storage")

    from fastapi.responses import Response

    return Response(
        content=data,
        media_type="application/gzip",
        headers={
            "Content-Disposition": f'attachment; filename="{name}-{version}.tar.gz"',
            "X-Checksum-SHA256": sv.checksum,
        },
    )


@router.post("/skills/{name}/star", response_model=StarResponse)
async def star_skill(
    name: str,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    """Star a skill."""
    result = await db.execute(select(Skill).where(Skill.name == name))
    skill = result.scalar_one_or_none()
    if skill is None:
        raise HTTPException(status_code=404, detail=f"Skill '{name}' not found")

    # Check if already starred
    existing = await db.execute(
        select(Star).where(Star.user_id == user.id, Star.skill_id == skill.id)
    )
    if existing.scalar_one_or_none() is not None:
        raise HTTPException(status_code=409, detail="Already starred")

    star = Star(user_id=user.id, skill_id=skill.id)
    db.add(star)

    # Increment stars_count
    skill.stars_count = (skill.stars_count or 0) + 1

    await db.commit()

    return StarResponse(starred=True, stars_count=skill.stars_count)


@router.delete("/skills/{name}/star", response_model=StarResponse)
async def unstar_skill(
    name: str,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    """Unstar a skill."""
    result = await db.execute(select(Skill).where(Skill.name == name))
    skill = result.scalar_one_or_none()
    if skill is None:
        raise HTTPException(status_code=404, detail=f"Skill '{name}' not found")

    existing = await db.execute(
        select(Star).where(Star.user_id == user.id, Star.skill_id == skill.id)
    )
    star = existing.scalar_one_or_none()
    if star is None:
        raise HTTPException(status_code=404, detail="Not starred")

    await db.delete(star)

    # Decrement stars_count
    skill.stars_count = max((skill.stars_count or 0) - 1, 0)

    await db.commit()

    return StarResponse(starred=False, stars_count=skill.stars_count)


@router.get("/skills", response_model=SearchResponse)
async def search_skills(
    q: Optional[str] = Query(default=None),
    tag: Optional[str] = Query(default=None),
    category: Optional[str] = Query(default=None),
    provider: Optional[str] = Query(default=None),
    sort: Optional[str] = Query(default=None),
    page: int = Query(default=1, ge=1),
    per_page: int = Query(default=20, ge=1, le=100),
    db: AsyncSession = Depends(get_db),
):
    # Build base query
    query = select(Skill)

    if q:
        pattern = f"%{q}%"
        query = query.where(Skill.name.ilike(pattern))

    # Filter by category
    if category:
        cat_result = await db.execute(select(Category).where(Category.name == category))
        cat = cat_result.scalar_one_or_none()
        if cat:
            query = query.where(Skill.category_id == cat.id)

    # Determine sort order
    sort_map = {
        "downloads": Skill.downloads.desc(),
        "stars": Skill.stars_count.desc(),
        "newest": Skill.created_at.desc(),
        "updated": Skill.updated_at.desc(),
    }
    order_clause = sort_map.get(sort, Skill.updated_at.desc())

    # Count total
    count_query = select(func.count()).select_from(query.subquery())
    total_result = await db.execute(count_query)
    total = total_result.scalar()

    # Paginate
    offset = (page - 1) * per_page
    skills_result = await db.execute(
        query.order_by(order_clause).offset(offset).limit(per_page)
    )
    skills = skills_result.scalars().all()

    # Build results
    results = []
    for skill in skills:
        # Get owner
        owner_result = await db.execute(select(User).where(User.id == skill.owner_id))
        owner = owner_result.scalar_one()

        # Get category name
        category_name = None
        if skill.category_id:
            cat_r = await db.execute(select(Category).where(Category.id == skill.category_id))
            cat_obj = cat_r.scalar_one_or_none()
            if cat_obj:
                category_name = cat_obj.name

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

        # Filter by tag if specified
        if tag:
            tags_in_meta = latest.meta.get("tags", [])
            if tag not in tags_in_meta:
                continue

        # Filter by provider if specified
        if provider:
            skill_providers = latest.providers or ["generic"]
            if provider not in skill_providers:
                continue

        results.append(
            SearchResultItem(
                name=skill.name,
                description=latest.meta.get("description", ""),
                owner=owner.username,
                owner_avatar_url=owner.avatar_url,
                downloads=skill.downloads,
                stars_count=skill.stars_count or 0,
                latest_version=latest.version,
                category=category_name,
                updated_at=skill.updated_at,
                tags=latest.meta.get("tags", []),
                providers=latest.providers or ["generic"],
            )
        )

    return SearchResponse(
        total=len(results),
        page=page,
        per_page=per_page,
        results=results,
    )
