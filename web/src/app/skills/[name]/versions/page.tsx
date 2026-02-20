import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { getSkillVersions } from "@/lib/api";
import { VersionList } from "@/components/skills/version-list";

interface VersionsPageProps {
  params: Promise<{ name: string }>;
}

export default async function VersionsPage({ params }: VersionsPageProps) {
  const { name } = await params;
  let data;
  try {
    data = await getSkillVersions(name);
  } catch {
    notFound();
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      <Link
        href={`/skills/${name}`}
        className="inline-flex items-center gap-1 text-sm text-[var(--primary)] hover:underline mb-4"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        Back to {name}
      </Link>

      <h1 className="text-2xl font-bold mb-6">
        Versions of {data.name}
      </h1>

      <VersionList versions={data.versions} />
    </div>
  );
}
