import Link from "next/link";
import { SkillCard } from "@/components/skills/skill-card";
import type { SearchResultItem } from "@/types";

interface FeaturedSkillsProps {
  title: string;
  skills: SearchResultItem[];
  viewAllHref: string;
}

export function FeaturedSkills({
  title,
  skills,
  viewAllHref,
}: FeaturedSkillsProps) {
  if (skills.length === 0) return null;

  return (
    <section className="py-8">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold">{title}</h2>
        <Link
          href={viewAllHref}
          className="text-sm text-[var(--primary)] hover:underline"
        >
          See all &rarr;
        </Link>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {skills.map((skill) => (
          <SkillCard key={skill.name} skill={skill} />
        ))}
      </div>
    </section>
  );
}
