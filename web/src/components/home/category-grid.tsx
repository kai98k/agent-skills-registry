import Link from "next/link";
import type { CategoryItem } from "@/types";

const CATEGORY_ICONS: Record<string, string> = {
  development: "\uD83D\uDCBB",
  productivity: "\u26A1",
  "ai-ml": "\uD83E\uDDE0",
  devops: "\uD83D\uDD27",
  "data-analysis": "\uD83D\uDCCA",
  security: "\uD83D\uDD12",
  documentation: "\uD83D\uDCDD",
  testing: "\uD83E\uDDEA",
  design: "\uD83C\uDFA8",
  other: "\uD83D\uDCE6",
};

export function CategoryGrid({ categories }: { categories: CategoryItem[] }) {
  return (
    <section className="py-12">
      <h2 className="text-xl font-semibold mb-6">Categories</h2>
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3">
        {categories.map((cat) => (
          <Link
            key={cat.name}
            href={`/categories/${cat.name}`}
            className="flex flex-col items-center gap-2 p-4 rounded-lg border border-[var(--border)] bg-[var(--card)] hover:border-[var(--primary)]/50 transition-colors"
          >
            <span className="text-2xl">
              {cat.icon || CATEGORY_ICONS[cat.name] || "\uD83D\uDCE6"}
            </span>
            <span className="text-sm font-medium">{cat.label}</span>
            <span className="text-xs text-[var(--muted-foreground)]">
              {cat.skill_count} skills
            </span>
          </Link>
        ))}
      </div>
    </section>
  );
}
