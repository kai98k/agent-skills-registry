"""SKILL.md Markdown → safe HTML rendering service."""

import bleach
import markdown
from markdown.extensions.codehilite import CodeHiliteExtension
from markdown.extensions.fenced_code import FencedCodeExtension
from markdown.extensions.tables import TableExtension


ALLOWED_TAGS = [
    "h1", "h2", "h3", "h4", "h5", "h6",
    "p", "br", "hr",
    "strong", "em", "del", "code", "pre",
    "a", "img",
    "ul", "ol", "li",
    "blockquote",
    "table", "thead", "tbody", "tr", "th", "td",
    "div", "span",
]

ALLOWED_ATTRIBUTES = {
    "a": ["href", "title", "rel"],
    "img": ["src", "alt", "title", "width", "height"],
    "code": ["class"],
    "div": ["class"],
    "span": ["class"],
    "pre": ["class"],
    "td": ["align"],
    "th": ["align"],
}


def render_markdown(raw_markdown: str) -> str:
    """Convert Markdown to sanitized HTML.

    Uses python-markdown with fenced code, syntax highlighting, and tables.
    Output is sanitized with bleach to prevent XSS.
    """
    html = markdown.markdown(
        raw_markdown,
        extensions=[
            FencedCodeExtension(),
            CodeHiliteExtension(css_class="highlight", guess_lang=False),
            TableExtension(),
            "md_in_html",
        ],
    )

    clean_html = bleach.clean(
        html,
        tags=ALLOWED_TAGS,
        attributes=ALLOWED_ATTRIBUTES,
        strip=True,
    )

    return clean_html
