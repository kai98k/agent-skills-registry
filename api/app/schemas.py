from datetime import datetime

from pydantic import BaseModel


class HealthResponse(BaseModel):
    status: str
    database: str
    storage: str


class PublishResponse(BaseModel):
    name: str
    version: str
    checksum: str
    published_at: datetime
    providers: list[str] = ["generic"]


class SkillVersionDetail(BaseModel):
    version: str
    description: str
    checksum: str
    size_bytes: int
    published_at: datetime
    providers: list[str] = ["generic"]
    metadata: dict


class SkillResponse(BaseModel):
    name: str
    owner: str
    downloads: int
    created_at: datetime
    latest_version: SkillVersionDetail | None = None


class SkillVersionSummary(BaseModel):
    version: str
    checksum: str
    size_bytes: int
    published_at: datetime
    providers: list[str] = ["generic"]


class SkillVersionsResponse(BaseModel):
    name: str
    versions: list[SkillVersionSummary]


class SearchResultItem(BaseModel):
    name: str
    description: str
    owner: str
    downloads: int
    latest_version: str
    updated_at: datetime
    tags: list[str]
    providers: list[str] = ["generic"]


class SearchResponse(BaseModel):
    total: int
    page: int
    per_page: int
    results: list[SearchResultItem]


class ErrorResponse(BaseModel):
    error: str
