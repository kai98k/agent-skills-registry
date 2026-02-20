import { Tag, Calendar, Scale, HardDrive, Layers } from "lucide-react";
import { formatBytes, formatRelativeDate } from "@/lib/utils";
import { InstallCommand } from "./install-command";
import type { SkillResponse } from "@/types";

export function SkillSidebar({ skill }: { skill: SkillResponse }) {
  const lv = skill.latest_version;

  return (
    <aside className="space-y-6">
      {/* Install */}
      <div>
        <h3 className="text-sm font-semibold mb-2">Install</h3>
        <InstallCommand skillName={skill.name} />
      </div>

      {/* Version */}
      {lv && (
        <div>
          <h3 className="text-sm font-semibold mb-1 flex items-center gap-1.5">
            <Layers className="h-3.5 w-3.5" /> Version
          </h3>
          <p className="text-sm text-[var(--muted-foreground)]">
            {lv.version} (latest)
          </p>
          <p className="text-xs text-[var(--muted-foreground)]">
            Published {formatRelativeDate(lv.published_at)}
          </p>
        </div>
      )}

      {/* License */}
      {lv?.metadata?.license && (
        <div>
          <h3 className="text-sm font-semibold mb-1 flex items-center gap-1.5">
            <Scale className="h-3.5 w-3.5" /> License
          </h3>
          <p className="text-sm text-[var(--muted-foreground)]">
            {String(lv.metadata.license)}
          </p>
        </div>
      )}

      {/* Tags */}
      {lv?.metadata?.tags && Array.isArray(lv.metadata.tags) && lv.metadata.tags.length > 0 && (
        <div>
          <h3 className="text-sm font-semibold mb-2 flex items-center gap-1.5">
            <Tag className="h-3.5 w-3.5" /> Tags
          </h3>
          <div className="flex flex-wrap gap-1">
            {(lv.metadata.tags as string[]).map((tag) => (
              <span
                key={tag}
                className="px-2 py-0.5 text-xs rounded-full bg-[var(--muted)] text-[var(--muted-foreground)]"
              >
                {tag}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Category */}
      {skill.category && (
        <div>
          <h3 className="text-sm font-semibold mb-1">Category</h3>
          <p className="text-sm text-[var(--muted-foreground)]">{skill.category}</p>
        </div>
      )}

      {/* Size */}
      {lv && (
        <div>
          <h3 className="text-sm font-semibold mb-1 flex items-center gap-1.5">
            <HardDrive className="h-3.5 w-3.5" /> Size
          </h3>
          <p className="text-sm text-[var(--muted-foreground)]">
            {formatBytes(lv.size_bytes)}
          </p>
        </div>
      )}

      {/* Providers */}
      {lv && lv.providers.length > 0 && (
        <div>
          <h3 className="text-sm font-semibold mb-2">Providers</h3>
          <div className="flex flex-wrap gap-1">
            {lv.providers.map((p) => (
              <span
                key={p}
                className="px-2 py-0.5 text-xs rounded-full bg-[var(--primary)]/10 text-[var(--primary)]"
              >
                {p}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Published date */}
      <div>
        <h3 className="text-sm font-semibold mb-1 flex items-center gap-1.5">
          <Calendar className="h-3.5 w-3.5" /> Created
        </h3>
        <p className="text-sm text-[var(--muted-foreground)]">
          {formatRelativeDate(skill.created_at)}
        </p>
      </div>
    </aside>
  );
}
