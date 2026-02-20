import Link from "next/link";
import Image from "next/image";
import { Star, Download, Package } from "lucide-react";
import { formatNumber } from "@/lib/utils";
import { SkillSidebar } from "./skill-sidebar";
import type { SkillResponse } from "@/types";

export function SkillDetail({ skill }: { skill: SkillResponse }) {
  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      {/* Header */}
      <div className="flex items-start justify-between gap-4 mb-8">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Package className="h-5 w-5 text-[var(--primary)]" />
            <h1 className="text-2xl font-bold">{skill.name}</h1>
          </div>
          <div className="flex items-center gap-2 text-sm text-[var(--muted-foreground)]">
            <span>by</span>
            <Link
              href={`/user/${skill.owner}`}
              className="flex items-center gap-1.5 hover:text-[var(--foreground)]"
            >
              {skill.owner_avatar_url && (
                <Image
                  src={skill.owner_avatar_url}
                  alt={skill.owner}
                  width={20}
                  height={20}
                  className="rounded-full"
                />
              )}
              {skill.owner}
            </Link>
          </div>
          <div className="flex items-center gap-4 mt-2 text-sm text-[var(--muted-foreground)]">
            <span className="flex items-center gap-1">
              <Star className="h-4 w-4" />
              {formatNumber(skill.stars_count)}
            </span>
            <span className="flex items-center gap-1">
              <Download className="h-4 w-4" />
              {formatNumber(skill.downloads)}
            </span>
          </div>
        </div>
      </div>

      {/* Content grid */}
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_300px] gap-8">
        {/* Main content - Markdown */}
        <div>
          {skill.readme_html ? (
            <div
              className="prose"
              dangerouslySetInnerHTML={{ __html: skill.readme_html }}
            />
          ) : (
            <p className="text-[var(--muted-foreground)]">
              No README content available.
            </p>
          )}

          {/* Versions link */}
          <div className="mt-8">
            <Link
              href={`/skills/${skill.name}/versions`}
              className="text-sm text-[var(--primary)] hover:underline"
            >
              View all versions &rarr;
            </Link>
          </div>
        </div>

        {/* Sidebar */}
        <SkillSidebar skill={skill} />
      </div>
    </div>
  );
}
