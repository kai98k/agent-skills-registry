import Link from "next/link";
import { Star, Download, Package } from "lucide-react";
import { formatNumber, formatRelativeDate } from "@/lib/utils";
import type { SearchResultItem } from "@/types";

export function SkillCard({ skill }: { skill: SearchResultItem }) {
  return (
    <Link
      href={`/skills/${skill.name}`}
      className="block p-4 rounded-lg border border-[var(--border)] bg-[var(--card)] hover:border-[var(--primary)]/50 transition-colors"
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2 min-w-0">
          <Package className="h-4 w-4 text-[var(--primary)] shrink-0" />
          <h3 className="font-semibold text-sm truncate">{skill.name}</h3>
          <span className="text-xs text-[var(--muted-foreground)]">
            v{skill.latest_version}
          </span>
        </div>
      </div>

      <p className="mt-1 text-xs text-[var(--muted-foreground)] line-clamp-2">
        by {skill.owner}
      </p>

      <p className="mt-2 text-sm text-[var(--muted-foreground)] line-clamp-2">
        {skill.description}
      </p>

      <div className="mt-3 flex items-center gap-4 text-xs text-[var(--muted-foreground)]">
        <span className="flex items-center gap-1">
          <Star className="h-3.5 w-3.5" />
          {formatNumber(skill.stars_count)}
        </span>
        <span className="flex items-center gap-1">
          <Download className="h-3.5 w-3.5" />
          {formatNumber(skill.downloads)}
        </span>
        <span>{formatRelativeDate(skill.updated_at)}</span>
      </div>

      {skill.tags.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1">
          {skill.tags.slice(0, 4).map((tag) => (
            <span
              key={tag}
              className="px-1.5 py-0.5 text-xs rounded bg-[var(--muted)] text-[var(--muted-foreground)]"
            >
              {tag}
            </span>
          ))}
        </div>
      )}
    </Link>
  );
}
