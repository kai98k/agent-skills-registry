import uuid
from datetime import datetime

from sqlalchemy import (
    BigInteger,
    Column,
    DateTime,
    ForeignKey,
    Integer,
    JSON,
    String,
    Text,
    UniqueConstraint,
    func,
)
from sqlalchemy.orm import DeclarativeBase, relationship


class Base(DeclarativeBase):
    pass


class User(Base):
    __tablename__ = "users"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    username = Column(String(64), unique=True, nullable=False)
    api_token = Column(String(128), unique=True, nullable=False)
    display_name = Column(String(128), nullable=True)
    avatar_url = Column(Text, nullable=True)
    github_id = Column(BigInteger, unique=True, nullable=True)
    bio = Column(String(256), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    skills = relationship("Skill", back_populates="owner")
    stars = relationship("Star", back_populates="user")


class Category(Base):
    __tablename__ = "categories"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    name = Column(String(64), unique=True, nullable=False)
    label = Column(String(128), nullable=False)
    description = Column(String(256), nullable=True)
    icon = Column(String(64), nullable=True)
    sort_order = Column(Integer, default=0)

    skills = relationship("Skill", back_populates="category")


class Skill(Base):
    __tablename__ = "skills"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    name = Column(String(128), unique=True, nullable=False)
    owner_id = Column(String(36), ForeignKey("users.id"), nullable=False)
    category_id = Column(String(36), ForeignKey("categories.id"), nullable=True)
    downloads = Column(BigInteger, default=0)
    stars_count = Column(BigInteger, default=0)
    readme_html = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    owner = relationship("User", back_populates="skills")
    category = relationship("Category", back_populates="skills")
    versions = relationship("SkillVersion", back_populates="skill", order_by="SkillVersion.published_at.desc()")
    stars = relationship("Star", back_populates="skill")


class SkillVersion(Base):
    __tablename__ = "skill_versions"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    skill_id = Column(String(36), ForeignKey("skills.id", ondelete="CASCADE"), nullable=False)
    version = Column(String(32), nullable=False)
    bundle_key = Column(Text, nullable=False)
    meta = Column("metadata", JSON, nullable=False)
    checksum = Column(String(64), nullable=False)
    size_bytes = Column(BigInteger, nullable=False)
    providers = Column(JSON, default=list)
    readme_raw = Column(Text, nullable=True)
    published_at = Column(DateTime(timezone=True), server_default=func.now())

    skill = relationship("Skill", back_populates="versions")

    __table_args__ = (
        UniqueConstraint("skill_id", "version", name="uq_skill_version"),
    )


class Star(Base):
    __tablename__ = "stars"

    user_id = Column(String(36), ForeignKey("users.id", ondelete="CASCADE"), primary_key=True)
    skill_id = Column(String(36), ForeignKey("skills.id", ondelete="CASCADE"), primary_key=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    user = relationship("User", back_populates="stars")
    skill = relationship("Skill", back_populates="stars")
