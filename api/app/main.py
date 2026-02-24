from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.config import settings
from app.routes import health, skills


@asynccontextmanager
async def lifespan(app: FastAPI):
    yield


app = FastAPI(title="AgentSkills Registry", lifespan=lifespan)

app.include_router(health.router, prefix=settings.api_prefix)
app.include_router(skills.router, prefix=settings.api_prefix)
