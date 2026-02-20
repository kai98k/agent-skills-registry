import { Suspense } from "react";
import { getCategories, searchSkills } from "@/lib/api";
import { SearchBar } from "@/components/search/search-bar";
import { SearchFilters } from "@/components/search/search-filters";
import { SearchResults } from "@/components/search/search-results";

interface SearchPageProps {
  searchParams: Promise<{
    q?: string;
    category?: string;
    tag?: string;
    provider?: string;
    sort?: string;
    page?: string;
  }>;
}

export default async function SearchPage({ searchParams }: SearchPageProps) {
  const params = await searchParams;
  const page = parseInt(params.page || "1", 10);

  const [results, categories] = await Promise.all([
    searchSkills({
      q: params.q,
      category: params.category,
      tag: params.tag,
      provider: params.provider,
      sort: params.sort || "updated",
      page,
      per_page: 20,
    }),
    getCategories().catch(() => ({ categories: [] })),
  ]);

  // Build base URL for pagination
  const urlParams = new URLSearchParams();
  if (params.q) urlParams.set("q", params.q);
  if (params.category) urlParams.set("category", params.category);
  if (params.tag) urlParams.set("tag", params.tag);
  if (params.sort) urlParams.set("sort", params.sort);
  const baseUrl = `/search?${urlParams.toString()}`;

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-4">
          {params.q ? `Results for "${params.q}"` : "Explore Skills"}
        </h1>
        <p className="text-sm text-[var(--muted-foreground)] mb-4">
          {results.total} {results.total === 1 ? "result" : "results"}
        </p>
      </div>

      <div className="flex flex-col sm:flex-row gap-4 mb-6">
        <div className="flex-1">
          <Suspense>
            <SearchBar />
          </Suspense>
        </div>
        <Suspense>
          <SearchFilters categories={categories.categories} />
        </Suspense>
      </div>

      <SearchResults data={results} currentPage={page} baseUrl={baseUrl} />
    </div>
  );
}
