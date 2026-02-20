import { Package, Download, Users } from "lucide-react";
import { formatNumber } from "@/lib/utils";

interface StatsBarProps {
  totalSkills: number;
  totalDownloads: number;
  totalCategories: number;
}

export function StatsBar({
  totalSkills,
  totalDownloads,
  totalCategories,
}: StatsBarProps) {
  return (
    <div className="flex items-center justify-center gap-8 py-6 text-sm text-[var(--muted-foreground)]">
      <div className="flex items-center gap-2">
        <Package className="h-4 w-4" />
        <span>
          <strong className="text-[var(--foreground)]">
            {formatNumber(totalSkills)}
          </strong>{" "}
          skills
        </span>
      </div>
      <div className="flex items-center gap-2">
        <Download className="h-4 w-4" />
        <span>
          <strong className="text-[var(--foreground)]">
            {formatNumber(totalDownloads)}
          </strong>{" "}
          downloads
        </span>
      </div>
      <div className="flex items-center gap-2">
        <Users className="h-4 w-4" />
        <span>
          <strong className="text-[var(--foreground)]">
            {formatNumber(totalCategories)}
          </strong>{" "}
          categories
        </span>
      </div>
    </div>
  );
}
