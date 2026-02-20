"use client";

import { Search } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

export function Hero() {
  const router = useRouter();
  const [query, setQuery] = useState("");

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (query.trim()) {
      router.push(`/search?q=${encodeURIComponent(query.trim())}`);
    }
  }

  return (
    <section className="py-20 text-center">
      <h1 className="text-4xl sm:text-5xl font-bold tracking-tight">
        Discover AI Agent Skills
      </h1>
      <p className="mt-4 text-lg text-[var(--muted-foreground)] max-w-xl mx-auto">
        The open registry for agent capabilities. Publish and install skill
        bundles for Claude, Gemini, Codex, Copilot, and more.
      </p>

      <form onSubmit={handleSearch} className="mt-8 max-w-lg mx-auto">
        <div className="relative">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-[var(--muted-foreground)]" />
          <input
            type="text"
            placeholder="Search skills..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="w-full pl-12 pr-4 py-3 rounded-lg bg-[var(--muted)] border border-[var(--border)] text-base focus:outline-none focus:ring-2 focus:ring-[var(--primary)] placeholder:text-[var(--muted-foreground)]"
          />
        </div>
      </form>
    </section>
  );
}
