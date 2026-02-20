import { formatBytes, formatRelativeDate } from "@/lib/utils";
import type { SkillVersionSummary } from "@/types";

export function VersionList({ versions }: { versions: SkillVersionSummary[] }) {
  return (
    <div className="border border-[var(--border)] rounded-lg overflow-hidden">
      <table className="w-full text-sm">
        <thead>
          <tr className="bg-[var(--muted)]">
            <th className="text-left px-4 py-2 font-medium">Version</th>
            <th className="text-left px-4 py-2 font-medium">Size</th>
            <th className="text-left px-4 py-2 font-medium">Published</th>
            <th className="text-left px-4 py-2 font-medium">Providers</th>
          </tr>
        </thead>
        <tbody>
          {versions.map((v) => (
            <tr key={v.version} className="border-t border-[var(--border)]">
              <td className="px-4 py-2 font-mono">{v.version}</td>
              <td className="px-4 py-2 text-[var(--muted-foreground)]">
                {formatBytes(v.size_bytes)}
              </td>
              <td className="px-4 py-2 text-[var(--muted-foreground)]">
                {formatRelativeDate(v.published_at)}
              </td>
              <td className="px-4 py-2">
                <div className="flex gap-1 flex-wrap">
                  {v.providers.map((p) => (
                    <span
                      key={p}
                      className="px-1.5 py-0.5 text-xs rounded bg-[var(--muted)] text-[var(--muted-foreground)]"
                    >
                      {p}
                    </span>
                  ))}
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
