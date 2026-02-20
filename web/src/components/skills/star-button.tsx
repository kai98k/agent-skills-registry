"use client";

import { Star } from "lucide-react";
import { useState } from "react";
import { starSkill, unstarSkill } from "@/lib/api";
import { formatNumber } from "@/lib/utils";

interface StarButtonProps {
  skillName: string;
  initialStarred: boolean;
  initialCount: number;
  token?: string;
}

export function StarButton({
  skillName,
  initialStarred,
  initialCount,
  token,
}: StarButtonProps) {
  const [starred, setStarred] = useState(initialStarred);
  const [count, setCount] = useState(initialCount);
  const [loading, setLoading] = useState(false);

  async function handleToggle() {
    if (!token) return;
    setLoading(true);
    try {
      const result = starred
        ? await unstarSkill(skillName, token)
        : await starSkill(skillName, token);
      setStarred(result.starred);
      setCount(result.stars_count);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }

  return (
    <button
      onClick={handleToggle}
      disabled={loading || !token}
      className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border text-sm font-medium transition-colors ${
        starred
          ? "border-yellow-500/50 bg-yellow-500/10 text-yellow-400"
          : "border-[var(--border)] bg-[var(--muted)] text-[var(--foreground)] hover:border-yellow-500/50"
      } disabled:opacity-50`}
    >
      <Star className={`h-4 w-4 ${starred ? "fill-yellow-400" : ""}`} />
      {starred ? "Starred" : "Star"}
      <span className="text-[var(--muted-foreground)]">
        {formatNumber(count)}
      </span>
    </button>
  );
}
