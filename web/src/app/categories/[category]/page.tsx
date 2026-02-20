import { searchSkills } from "@/lib/api";
import { SkillCard } from "@/components/skills/skill-card";

interface CategoryPageProps {
  params: Promise<{ category: string }>;
}

export default async function CategoryPage({ params }: CategoryPageProps) {
  const { category } = await params;
  const results = await searchSkills({ category, sort: "stars", per_page: 50 });

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <h1 className="text-2xl font-bold mb-2 capitalize">
        {category.replace(/-/g, " ")}
      </h1>
      <p className="text-sm text-[var(--muted-foreground)] mb-6">
        {results.total} {results.total === 1 ? "skill" : "skills"} in this
        category
      </p>

      {results.results.length > 0 ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {results.results.map((skill) => (
            <SkillCard key={skill.name} skill={skill} />
          ))}
        </div>
      ) : (
        <p className="text-[var(--muted-foreground)] text-center py-16">
          No skills in this category yet.
        </p>
      )}
    </div>
  );
}
