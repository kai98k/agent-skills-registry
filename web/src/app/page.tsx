import { getCategories, searchSkills } from "@/lib/api";
import { Hero } from "@/components/home/hero";
import { CategoryGrid } from "@/components/home/category-grid";
import { FeaturedSkills } from "@/components/home/featured-skills";

export default async function HomePage() {
  const [categories, trending, latest] = await Promise.all([
    getCategories().catch(() => ({ categories: [] })),
    searchSkills({ sort: "stars", per_page: 6 }).catch(() => ({
      total: 0,
      page: 1,
      per_page: 6,
      results: [],
    })),
    searchSkills({ sort: "newest", per_page: 6 }).catch(() => ({
      total: 0,
      page: 1,
      per_page: 6,
      results: [],
    })),
  ]);

  return (
    <div className="mx-auto max-w-7xl px-4">
      <Hero />

      {categories.categories.length > 0 && (
        <CategoryGrid categories={categories.categories} />
      )}

      <FeaturedSkills
        title="Trending Skills"
        skills={trending.results}
        viewAllHref="/search?sort=stars"
      />

      <FeaturedSkills
        title="Latest Skills"
        skills={latest.results}
        viewAllHref="/search?sort=newest"
      />
    </div>
  );
}
