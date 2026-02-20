import { Metadata } from "next";
import { notFound } from "next/navigation";
import Image from "next/image";
import Link from "next/link";
import { Calendar, Download, Star } from "lucide-react";
import { getUserProfile } from "@/lib/api";
import { formatNumber, formatRelativeDate } from "@/lib/utils";

interface UserPageProps {
  params: Promise<{ username: string }>;
}

export async function generateMetadata({
  params,
}: UserPageProps): Promise<Metadata> {
  const { username } = await params;
  return {
    title: `${username} - AgentSkills`,
    description: `View ${username}'s published AI agent skills on AgentSkills.`,
  };
}

export default async function UserPage({ params }: UserPageProps) {
  const { username } = await params;
  let user;
  try {
    user = await getUserProfile(username);
  } catch {
    notFound();
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      {/* Profile header */}
      <div className="flex items-start gap-4 mb-8">
        {user.avatar_url ? (
          <Image
            src={user.avatar_url}
            alt={user.username}
            width={80}
            height={80}
            className="rounded-full"
          />
        ) : (
          <div className="h-20 w-20 rounded-full bg-[var(--muted)] flex items-center justify-center text-2xl font-bold">
            {user.username[0].toUpperCase()}
          </div>
        )}
        <div>
          <h1 className="text-2xl font-bold">{user.username}</h1>
          {user.display_name && (
            <p className="text-[var(--muted-foreground)]">
              {user.display_name}
            </p>
          )}
          {user.bio && (
            <p className="text-sm text-[var(--muted-foreground)] mt-1">
              {user.bio}
            </p>
          )}
          <div className="flex items-center gap-4 mt-2 text-sm text-[var(--muted-foreground)]">
            <span className="flex items-center gap-1">
              <Calendar className="h-3.5 w-3.5" />
              Joined {formatRelativeDate(user.created_at)}
            </span>
            <span className="flex items-center gap-1">
              <Download className="h-3.5 w-3.5" />
              {formatNumber(user.total_downloads)} total downloads
            </span>
            <span className="flex items-center gap-1">
              <Star className="h-3.5 w-3.5" />
              {formatNumber(user.total_stars)} total stars
            </span>
          </div>
        </div>
      </div>

      {/* Skills list */}
      <h2 className="text-lg font-semibold mb-4">
        Published Skills ({user.skills.length})
      </h2>
      {user.skills.length > 0 ? (
        <div className="space-y-3">
          {user.skills.map((skill) => (
            <Link
              key={skill.name}
              href={`/skills/${skill.name}`}
              className="flex items-center justify-between p-4 rounded-lg border border-[var(--border)] bg-[var(--card)] hover:border-[var(--primary)]/50 transition-colors"
            >
              <div>
                <h3 className="font-semibold text-sm">{skill.name}</h3>
                <p className="text-xs text-[var(--muted-foreground)] mt-0.5">
                  {skill.description}
                </p>
              </div>
              <div className="flex items-center gap-4 text-xs text-[var(--muted-foreground)] shrink-0 ml-4">
                <span>v{skill.latest_version}</span>
                <span className="flex items-center gap-1">
                  <Star className="h-3 w-3" />
                  {formatNumber(skill.stars_count)}
                </span>
                <span className="flex items-center gap-1">
                  <Download className="h-3 w-3" />
                  {formatNumber(skill.downloads)}
                </span>
              </div>
            </Link>
          ))}
        </div>
      ) : (
        <p className="text-[var(--muted-foreground)] text-center py-8">
          No published skills yet.
        </p>
      )}
    </div>
  );
}
