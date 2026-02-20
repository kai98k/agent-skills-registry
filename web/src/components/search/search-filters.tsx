"use client";

import { useRouter, useSearchParams } from "next/navigation";

const SORT_OPTIONS = [
  { value: "updated", label: "Recently Updated" },
  { value: "newest", label: "Newest" },
  { value: "stars", label: "Most Stars" },
  { value: "downloads", label: "Most Downloads" },
];

export function SearchFilters({
  categories,
}: {
  categories: { name: string; label: string }[];
}) {
  const router = useRouter();
  const searchParams = useSearchParams();

  function updateParam(key: string, value: string) {
    const params = new URLSearchParams(searchParams.toString());
    if (value) {
      params.set(key, value);
    } else {
      params.delete(key);
    }
    params.delete("page");
    router.push(`/search?${params.toString()}`);
  }

  return (
    <div className="flex flex-wrap gap-3">
      <select
        value={searchParams.get("category") || ""}
        onChange={(e) => updateParam("category", e.target.value)}
        className="px-3 py-1.5 rounded-md bg-[var(--muted)] border border-[var(--border)] text-sm"
      >
        <option value="">All Categories</option>
        {categories.map((cat) => (
          <option key={cat.name} value={cat.name}>
            {cat.label}
          </option>
        ))}
      </select>

      <select
        value={searchParams.get("sort") || "updated"}
        onChange={(e) => updateParam("sort", e.target.value)}
        className="px-3 py-1.5 rounded-md bg-[var(--muted)] border border-[var(--border)] text-sm"
      >
        {SORT_OPTIONS.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    </div>
  );
}
