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
    owner_avatar_url: str | None = None
    downloads: int
    stars_count: int = 0
    starred_by_me: bool = False
    category: str | None = None
    readme_html: str | None = None
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
    owner_avatar_url: str | None = None
    downloads: int
    stars_count: int = 0
    latest_version: str
    category: str | None = None
    updated_at: datetime
    tags: list[str]
    providers: list[str] = ["generic"]


class SearchResponse(BaseModel):
    total: int
    page: int
    per_page: int
    results: list[SearchResultItem]


class StarResponse(BaseModel):
    starred: bool
    stars_count: int


class CategoryItem(BaseModel):
    name: str
    label: str
    icon: str | None = None
    skill_count: int = 0


class CategoriesResponse(BaseModel):
    categories: list[CategoryItem]


class GitHubAuthRequest(BaseModel):
    github_access_token: str


class GitHubAuthResponse(BaseModel):
    username: str
    display_name: str | None = None
    avatar_url: str | None = None
    api_token: str


class UserSkillItem(BaseModel):
    name: str
    description: str
    downloads: int
    stars_count: int = 0
    latest_version: str
    updated_at: datetime


class UserResponse(BaseModel):
    username: str
    display_name: str | None = None
    avatar_url: str | None = None
    bio: str | None = None
    created_at: datetime
    skills: list[UserSkillItem]
    total_downloads: int = 0
    total_stars: int = 0


class ErrorResponse(BaseModel):
    error: str
