from contextlib import asynccontextmanager

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

from app.config import settings
from app.routes import health, skills


@asynccontextmanager
async def lifespan(app: FastAPI):
    yield


app = FastAPI(title="AgentSkills Registry", version="0.1.0", lifespan=lifespan)

app.include_router(health.router, prefix=settings.api_prefix)
app.include_router(skills.router, prefix=settings.api_prefix)


@app.exception_handler(Exception)
async def generic_exception_handler(request: Request, exc: Exception):
    return JSONResponse(status_code=500, content={"error": "Internal server error"})
