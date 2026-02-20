import Link from "next/link";
import { SkillCard } from "@/components/skills/skill-card";
import type { SearchResponse } from "@/types";

interface SearchResultsProps {
  data: SearchResponse;
  currentPage: number;
  baseUrl: string;
}

export function SearchResults({
  data,
  currentPage,
  baseUrl,
}: SearchResultsProps) {
  const totalPages = Math.ceil(data.total / data.per_page);

  if (data.results.length === 0) {
    return (
      <div className="text-center py-16 text-[var(--muted-foreground)]">
        <p className="text-lg">No skills found</p>
        <p className="text-sm mt-2">Try adjusting your search or filters.</p>
      </div>
    );
  }

  return (
    <div>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {data.results.map((skill) => (
          <SkillCard key={skill.name} skill={skill} />
        ))}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-4 mt-8 text-sm">
          {currentPage > 1 && (
            <Link
              href={`${baseUrl}&page=${currentPage - 1}`}
              className="px-3 py-1.5 rounded-md border border-[var(--border)] hover:bg-[var(--muted)]"
            >
              &larr; Prev
            </Link>
          )}
          <span className="text-[var(--muted-foreground)]">
            Page {currentPage} of {totalPages}
          </span>
          {currentPage < totalPages && (
            <Link
              href={`${baseUrl}&page=${currentPage + 1}`}
              className="px-3 py-1.5 rounded-md border border-[var(--border)] hover:bg-[var(--muted)]"
            >
              Next &rarr;
            </Link>
          )}
        </div>
      )}
    </div>
  );
}
