import hashlib
import os
import re
import tarfile
import tempfile
from dataclasses import dataclass, field

import frontmatter
import semver
import yaml


class ParseError(Exception):
    pass


@dataclass
class ParsedSkill:
    name: str
    version: str
    description: str
    author: str
    tags: list[str] = field(default_factory=list)
    license: str | None = None
    min_agent_version: str | None = None
    compatibility: str | None = None
    body: str = ""
    metadata: dict = field(default_factory=dict)
    providers: list[str] = field(default_factory=lambda: ["generic"])
    bundle_files: list[str] = field(default_factory=list)


# Known providers and their detection indicators within bundles
PROVIDER_BUNDLE_INDICATORS: dict[str, list[str]] = {
    "claude": [".claude/", "CLAUDE.md"],
    "gemini": [".gemini/", "GEMINI.md"],
    "codex": [".codex/", "AGENTS.md"],
    "copilot": [".github/copilot-instructions.md", ".github/skills/", ".github/agents/"],
    "cursor": [".cursor/", ".cursorrules"],
    "windsurf": [".windsurf/", ".windsurfrules"],
    "antigravity": [".antigravity/"],
}

# Keywords to detect from the compatibility field
PROVIDER_COMPAT_KEYWORDS: dict[str, list[str]] = {
    "claude": ["claude"],
    "gemini": ["gemini"],
    "codex": ["codex", "openai"],
    "copilot": ["copilot"],
    "cursor": ["cursor"],
    "windsurf": ["windsurf", "codeium"],
    "antigravity": ["antigravity"],
}


def extract_providers(metadata: dict, bundle_files: list[str]) -> list[str]:
    """Determine provider compatibility from frontmatter and bundle file listing."""
    providers: set[str] = set()

    # Check compatibility field
    compat = metadata.get("compatibility", "")
    if isinstance(compat, str):
        compat_lower = compat.lower()
        for provider, keywords in PROVIDER_COMPAT_KEYWORDS.items():
            if any(kw in compat_lower for kw in keywords):
                providers.add(provider)

    # Check bundle file paths
    for filepath in bundle_files:
        normalized = filepath
        while normalized.startswith("./"):
            normalized = normalized[2:]
        for provider, indicators in PROVIDER_BUNDLE_INDICATORS.items():
            for indicator in indicators:
                if indicator.endswith("/"):
                    if normalized.startswith(indicator) or normalized == indicator.rstrip("/"):
                        providers.add(provider)
                else:
                    if normalized == indicator:
                        providers.add(provider)

    return sorted(providers) if providers else ["generic"]


def validate_provider_constraints(name: str, providers: list[str]) -> None:
    """Apply provider-specific naming constraints."""
    if "claude" in providers:
        lowered = name.lower()
        if "anthropic" in lowered or "claude" in lowered:
            raise ParseError(
                f"Skill name '{name}' cannot contain 'anthropic' or 'claude' "
                f"for Claude-compatible skills"
            )


def validate_name(name: str) -> None:
    if not name or not isinstance(name, str):
        raise ParseError("Field 'name' is required")
    if len(name) < 3 or len(name) > 64:
        raise ParseError(f"Field 'name' must be 3-64 characters, got {len(name)}")
    if not re.match(r"^[a-z0-9\-]+$", name):
        raise ParseError("Field 'name' must match [a-z0-9\\-]")
    if "--" in name:
        raise ParseError("Field 'name' must not contain consecutive hyphens '--'")
    if name.startswith("-") or name.endswith("-"):
        raise ParseError("Field 'name' must not start or end with a hyphen")


def validate_version(version: str) -> None:
    if not version or not isinstance(version, str):
        raise ParseError("Field 'version' is required")
    try:
        semver.Version.parse(version)
    except ValueError:
        raise ParseError(f"Field 'version' must be valid semver, got '{version}'")


def validate_description(description: str) -> None:
    if not description or not isinstance(description, str):
        raise ParseError("Field 'description' is required")
    if len(description) < 1 or len(description) > 256:
        raise ParseError(f"Field 'description' must be 1-256 characters, got {len(description)}")


def validate_author(author: str) -> None:
    if not author or not isinstance(author, str):
        raise ParseError("Field 'author' is required")


def validate_tags(tags: list) -> None:
    if not isinstance(tags, list):
        raise ParseError("Field 'tags' must be a list")
    if len(tags) > 10:
        raise ParseError(f"Field 'tags' allows max 10 items, got {len(tags)}")
    for tag in tags:
        if not isinstance(tag, str):
            raise ParseError(f"Each tag must be a string, got {type(tag).__name__}")
        if not re.match(r"^[a-z0-9\-]{1,32}$", tag):
            raise ParseError(f"Tag '{tag}' must match [a-z0-9\\-]{{1,32}}")


def parse_frontmatter(content: str) -> ParsedSkill:
    """Parse SKILL.md content and validate frontmatter."""
    post = frontmatter.loads(content)
    meta = dict(post.metadata)

    # Required fields
    name = meta.get("name", "")
    validate_name(name)

    version = meta.get("version", "")
    validate_version(version)

    description = meta.get("description", "")
    validate_description(description)

    author = meta.get("author", "")
    validate_author(author)

    # Optional fields
    tags = meta.get("tags", [])
    if tags:
        validate_tags(tags)

    license_val = meta.get("license")
    min_agent_version = meta.get("min_agent_version")
    compatibility = meta.get("compatibility")

    return ParsedSkill(
        name=name,
        version=version,
        description=description,
        author=author,
        tags=tags or [],
        license=license_val,
        min_agent_version=min_agent_version,
        compatibility=compatibility,
        body=post.content,
        metadata=meta,
    )


def _safe_extract_path(member_name: str, dest_dir: str) -> str:
    """Validate that an extracted path stays within dest_dir."""
    target = os.path.realpath(os.path.join(dest_dir, member_name))
    if not target.startswith(os.path.realpath(dest_dir)):
        raise ParseError(f"Path traversal detected: {member_name}")
    return target


def extract_and_parse(file_bytes: bytes, max_decompressed: int = 200 * 1024 * 1024) -> ParsedSkill:
    """Extract a .tar.gz bundle, find SKILL.md, parse and validate it."""
    tmpdir = tempfile.mkdtemp()
    try:
        try:
            tar = tarfile.open(fileobj=__import__("io").BytesIO(file_bytes), mode="r:gz")
        except tarfile.TarError:
            raise ParseError("Invalid .tar.gz file")

        # Validate paths and sizes before extraction
        total_size = 0
        bundle_files = []
        for member in tar.getmembers():
            _safe_extract_path(member.name, tmpdir)
            if member.isfile():
                total_size += member.size
                if total_size > max_decompressed:
                    raise ParseError(f"Decompressed size exceeds {max_decompressed} bytes limit")
            bundle_files.append(member.name)

        # Extract all
        tar.extractall(path=tmpdir, filter="data")
        tar.close()

        # Find SKILL.md (root or one level deep)
        skill_md_path = None
        for root, _dirs, files in os.walk(tmpdir):
            depth = root.replace(tmpdir, "").count(os.sep)
            if depth > 1:
                continue
            if "SKILL.md" in files:
                skill_md_path = os.path.join(root, "SKILL.md")
                break

        if skill_md_path is None:
            raise ParseError("No SKILL.md found in bundle")

        with open(skill_md_path, "r", encoding="utf-8") as f:
            content = f.read()

        parsed = parse_frontmatter(content)
        parsed.bundle_files = bundle_files

        # Determine providers
        parsed.providers = extract_providers(parsed.metadata, bundle_files)

        return parsed

    finally:
        import shutil
        shutil.rmtree(tmpdir, ignore_errors=True)


def compute_checksum(data: bytes) -> str:
    """Compute SHA-256 hex digest of data."""
    return hashlib.sha256(data).hexdigest()
