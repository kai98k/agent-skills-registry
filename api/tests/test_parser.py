import pytest

from app.services.parser import (
    ParseError,
    extract_and_parse,
    extract_providers,
    parse_frontmatter,
    validate_provider_constraints,
    compute_checksum,
)
from tests.conftest import VALID_SKILL_MD, make_bundle_bytes


class TestParseFrontmatter:
    def test_valid_skill_md(self):
        parsed = parse_frontmatter(VALID_SKILL_MD)
        assert parsed.name == "test-skill"
        assert parsed.version == "1.0.0"
        assert parsed.description == "A test skill for validation"
        assert parsed.author == "dev"
        assert parsed.tags == ["test", "example"]
        assert parsed.license == "MIT"
        assert "Test Skill" in parsed.body

    def test_missing_name(self):
        content = """\
---
version: "1.0.0"
description: "test"
author: "dev"
---
body
"""
        with pytest.raises(ParseError, match="name"):
            parse_frontmatter(content)

    def test_missing_version(self):
        content = """\
---
name: "test-skill"
description: "test"
author: "dev"
---
body
"""
        with pytest.raises(ParseError, match="version"):
            parse_frontmatter(content)

    def test_invalid_semver(self):
        content = """\
---
name: "test-skill"
version: "1.0"
description: "test"
author: "dev"
---
body
"""
        with pytest.raises(ParseError, match="semver"):
            parse_frontmatter(content)

    def test_missing_description(self):
        content = """\
---
name: "test-skill"
version: "1.0.0"
author: "dev"
---
body
"""
        with pytest.raises(ParseError, match="description"):
            parse_frontmatter(content)

    def test_description_too_long(self):
        content = f"""\
---
name: "test-skill"
version: "1.0.0"
description: "{'x' * 257}"
author: "dev"
---
body
"""
        with pytest.raises(ParseError, match="256"):
            parse_frontmatter(content)

    def test_missing_author(self):
        content = """\
---
name: "test-skill"
version: "1.0.0"
description: "test"
---
body
"""
        with pytest.raises(ParseError, match="author"):
            parse_frontmatter(content)

    def test_name_uppercase(self):
        content = """\
---
name: "Test-Skill"
version: "1.0.0"
description: "test"
author: "dev"
---
body
"""
        with pytest.raises(ParseError, match="\\[a-z0-9"):
            parse_frontmatter(content)

    def test_name_consecutive_hyphens(self):
        content = """\
---
name: "test--skill"
version: "1.0.0"
description: "test"
author: "dev"
---
body
"""
        with pytest.raises(ParseError, match="consecutive"):
            parse_frontmatter(content)

    def test_name_too_short(self):
        content = """\
---
name: "ab"
version: "1.0.0"
description: "test"
author: "dev"
---
body
"""
        with pytest.raises(ParseError, match="3-64"):
            parse_frontmatter(content)

    def test_name_starts_with_hyphen(self):
        content = """\
---
name: "-test-skill"
version: "1.0.0"
description: "test"
author: "dev"
---
body
"""
        with pytest.raises(ParseError, match="hyphen"):
            parse_frontmatter(content)

    def test_too_many_tags(self):
        tags = "\n".join(f"  - tag{i}" for i in range(11))
        content = f"""\
---
name: "test-skill"
version: "1.0.0"
description: "test"
author: "dev"
tags:
{tags}
---
body
"""
        with pytest.raises(ParseError, match="10"):
            parse_frontmatter(content)

    def test_tag_invalid_chars(self):
        content = """\
---
name: "test-skill"
version: "1.0.0"
description: "test"
author: "dev"
tags:
  - Invalid_Tag
---
body
"""
        with pytest.raises(ParseError, match="Tag"):
            parse_frontmatter(content)

    def test_optional_fields_absent(self):
        content = """\
---
name: "test-skill"
version: "1.0.0"
description: "test"
author: "dev"
---
body
"""
        parsed = parse_frontmatter(content)
        assert parsed.tags == []
        assert parsed.license is None
        assert parsed.min_agent_version is None

    def test_compatibility_field(self):
        content = """\
---
name: "test-skill"
version: "1.0.0"
description: "test"
author: "dev"
compatibility: "Designed for Claude Code"
---
body
"""
        parsed = parse_frontmatter(content)
        assert parsed.compatibility == "Designed for Claude Code"


class TestExtractProviders:
    def test_from_compatibility_claude(self):
        result = extract_providers({"compatibility": "Designed for Claude Code"}, [])
        assert result == ["claude"]

    def test_from_compatibility_gemini(self):
        result = extract_providers({"compatibility": "Designed for Gemini CLI"}, [])
        assert result == ["gemini"]

    def test_from_compatibility_codex(self):
        result = extract_providers({"compatibility": "Works with OpenAI Codex"}, [])
        assert result == ["codex"]

    def test_from_compatibility_copilot(self):
        result = extract_providers({"compatibility": "GitHub Copilot compatible"}, [])
        assert result == ["copilot"]

    def test_from_compatibility_cursor(self):
        result = extract_providers({"compatibility": "Designed for Cursor IDE"}, [])
        assert result == ["cursor"]

    def test_from_compatibility_windsurf(self):
        result = extract_providers({"compatibility": "Windsurf editor"}, [])
        assert result == ["windsurf"]

    def test_from_compatibility_antigravity(self):
        result = extract_providers({"compatibility": "Antigravity agent"}, [])
        assert result == ["antigravity"]

    def test_from_bundle_claude_dir(self):
        result = extract_providers({}, [".claude/settings.json"])
        assert result == ["claude"]

    def test_from_bundle_claude_md(self):
        result = extract_providers({}, ["CLAUDE.md"])
        assert result == ["claude"]

    def test_from_bundle_gemini_dir(self):
        result = extract_providers({}, [".gemini/config"])
        assert result == ["gemini"]

    def test_from_bundle_codex_dir(self):
        result = extract_providers({}, [".codex/config.toml"])
        assert result == ["codex"]

    def test_from_bundle_copilot_instructions(self):
        result = extract_providers({}, [".github/copilot-instructions.md"])
        assert result == ["copilot"]

    def test_from_bundle_copilot_skills_dir(self):
        result = extract_providers({}, [".github/skills/my-skill/SKILL.md"])
        assert result == ["copilot"]

    def test_from_bundle_cursor_dir(self):
        result = extract_providers({}, [".cursor/rules/test.mdc"])
        assert result == ["cursor"]

    def test_from_bundle_cursorrules(self):
        result = extract_providers({}, [".cursorrules"])
        assert result == ["cursor"]

    def test_from_bundle_windsurf_dir(self):
        result = extract_providers({}, [".windsurf/rules/general.md"])
        assert result == ["windsurf"]

    def test_from_bundle_windsurfrules(self):
        result = extract_providers({}, [".windsurfrules"])
        assert result == ["windsurf"]

    def test_from_bundle_antigravity_dir(self):
        result = extract_providers({}, [".antigravity/rules.md"])
        assert result == ["antigravity"]

    def test_from_bundle_agents_md(self):
        result = extract_providers({}, ["AGENTS.md"])
        assert result == ["codex"]

    def test_none_detected_returns_generic(self):
        result = extract_providers({}, ["SKILL.md", "scripts/run.sh"])
        assert result == ["generic"]

    def test_multiple_providers(self):
        result = extract_providers(
            {"compatibility": "Works with Claude and Gemini"},
            [".cursor/rules/test.mdc"],
        )
        assert "claude" in result
        assert "gemini" in result
        assert "cursor" in result

    def test_no_compatibility_field(self):
        result = extract_providers({}, [])
        assert result == ["generic"]


class TestValidateProviderConstraints:
    def test_claude_name_with_anthropic(self):
        with pytest.raises(ParseError, match="anthropic"):
            validate_provider_constraints("my-anthropic-skill", ["claude"])

    def test_claude_name_with_claude(self):
        with pytest.raises(ParseError, match="claude"):
            validate_provider_constraints("claude-helper", ["claude"])

    def test_claude_valid_name(self):
        validate_provider_constraints("code-review", ["claude"])

    def test_generic_allows_claude_in_name(self):
        validate_provider_constraints("claude-helper", ["generic"])

    def test_non_claude_allows_anthropic_in_name(self):
        validate_provider_constraints("my-anthropic-tool", ["gemini"])


class TestExtractAndParse:
    def test_valid_bundle(self):
        bundle = make_bundle_bytes(VALID_SKILL_MD)
        parsed = extract_and_parse(bundle)
        assert parsed.name == "test-skill"
        assert parsed.version == "1.0.0"

    def test_nested_bundle(self):
        bundle = make_bundle_bytes(VALID_SKILL_MD, nested=True)
        parsed = extract_and_parse(bundle)
        assert parsed.name == "test-skill"

    def test_no_skill_md(self):
        bundle = make_bundle_bytes("", extra_files={"README.md": "hello"})
        # Override: create bundle without SKILL.md
        import io
        import tarfile

        buf = io.BytesIO()
        with tarfile.open(fileobj=buf, mode="w:gz") as tar:
            data = b"readme"
            info = tarfile.TarInfo(name="README.md")
            info.size = len(data)
            tar.addfile(info, io.BytesIO(data))
        bundle = buf.getvalue()

        with pytest.raises(ParseError, match="No SKILL.md"):
            extract_and_parse(bundle)

    def test_invalid_tar(self):
        with pytest.raises(ParseError, match="Invalid .tar.gz"):
            extract_and_parse(b"not a tar file")

    def test_provider_detection_in_bundle(self):
        bundle = make_bundle_bytes(
            VALID_SKILL_MD,
            extra_files={".claude/settings.json": "{}"},
        )
        parsed = extract_and_parse(bundle)
        assert "claude" in parsed.providers

    def test_provider_detection_from_compatibility(self):
        content = """\
---
name: "test-skill"
version: "1.0.0"
description: "test"
author: "dev"
compatibility: "Designed for Gemini CLI"
---
body
"""
        bundle = make_bundle_bytes(content)
        parsed = extract_and_parse(bundle)
        assert "gemini" in parsed.providers


class TestComputeChecksum:
    def test_checksum(self):
        data = b"hello world"
        result = compute_checksum(data)
        assert len(result) == 64  # SHA-256 hex digest
        assert result == "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

    def test_different_data_different_checksum(self):
        assert compute_checksum(b"a") != compute_checksum(b"b")
