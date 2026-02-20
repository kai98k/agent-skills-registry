from fastapi import APIRouter, Depends
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession

from app.dependencies import get_db, get_s3
from app.config import settings
from app.schemas import HealthResponse

router = APIRouter()


@router.get("/health", response_model=HealthResponse)
async def health_check(db: AsyncSession = Depends(get_db)):
    db_status = "disconnected"
    storage_status = "disconnected"

    try:
        await db.execute(text("SELECT 1"))
        db_status = "connected"
    except Exception:
        pass

    try:
        s3 = get_s3()
        s3.head_bucket(Bucket=settings.s3_bucket)
        storage_status = "connected"
    except Exception:
        pass

    status = "ok" if db_status == "connected" and storage_status == "connected" else "degraded"
    return HealthResponse(status=status, database=db_status, storage=storage_status)
