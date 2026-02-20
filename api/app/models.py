import uuid
from datetime import datetime

from sqlalchemy import (
    BigInteger,
    Column,
    DateTime,
    ForeignKey,
    JSON,
    String,
    Text,
    UniqueConstraint,
    func,
)
from sqlalchemy.orm import DeclarativeBase, relationship, mapped_column


class Base(DeclarativeBase):
    pass


class User(Base):
    __tablename__ = "users"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    username = Column(String(64), unique=True, nullable=False)
    api_token = Column(String(128), unique=True, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    skills = relationship("Skill", back_populates="owner")


class Skill(Base):
    __tablename__ = "skills"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    name = Column(String(128), unique=True, nullable=False)
    owner_id = Column(String(36), ForeignKey("users.id"), nullable=False)
    downloads = Column(BigInteger, default=0)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    owner = relationship("User", back_populates="skills")
    versions = relationship("SkillVersion", back_populates="skill", order_by="SkillVersion.published_at.desc()")


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
    published_at = Column(DateTime(timezone=True), server_default=func.now())

    skill = relationship("Skill", back_populates="versions")

    __table_args__ = (
        UniqueConstraint("skill_id", "version", name="uq_skill_version"),
    )
