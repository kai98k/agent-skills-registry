import { Metadata } from "next";
import { notFound } from "next/navigation";
import { getSkill } from "@/lib/api";
import { SkillDetail } from "@/components/skills/skill-detail";

interface SkillPageProps {
  params: Promise<{ name: string }>;
}

export async function generateMetadata({
  params,
}: SkillPageProps): Promise<Metadata> {
  const { name } = await params;
  try {
    const skill = await getSkill(name);
    return {
      title: `${skill.name} - AgentSkills`,
      description:
        skill.latest_version?.description || `AI Agent Skill: ${skill.name}`,
    };
  } catch {
    return { title: "Skill Not Found - AgentSkills" };
  }
}

export default async function SkillPage({ params }: SkillPageProps) {
  const { name } = await params;
  let skill;
  try {
    skill = await getSkill(name);
  } catch {
    notFound();
  }

  return <SkillDetail skill={skill} />;
}
