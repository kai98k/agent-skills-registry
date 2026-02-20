"use client";

import Link from "next/link";
import { Package, Search } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

export function Header() {
  const router = useRouter();
  const [query, setQuery] = useState("");

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (query.trim()) {
      router.push(`/search?q=${encodeURIComponent(query.trim())}`);
    }
  }

  return (
    <header className="sticky top-0 z-50 border-b border-[var(--border)] bg-[var(--background)]/95 backdrop-blur">
      <div className="mx-auto max-w-7xl flex items-center justify-between px-4 h-14">
        <Link href="/" className="flex items-center gap-2 font-semibold text-lg">
          <Package className="h-5 w-5 text-[var(--primary)]" />
          AgentSkills
        </Link>

        <form onSubmit={handleSearch} className="hidden sm:flex items-center flex-1 max-w-md mx-8">
          <div className="relative w-full">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-[var(--muted-foreground)]" />
            <input
              type="text"
              placeholder="Search skills..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-1.5 rounded-md bg-[var(--muted)] border border-[var(--border)] text-sm focus:outline-none focus:ring-2 focus:ring-[var(--primary)] placeholder:text-[var(--muted-foreground)]"
            />
          </div>
        </form>

        <nav className="flex items-center gap-4 text-sm">
          <Link href="/search" className="text-[var(--muted-foreground)] hover:text-[var(--foreground)]">
            Explore
          </Link>
          <Link
            href="/login"
            className="px-3 py-1.5 rounded-md bg-[var(--primary)] text-[var(--primary-foreground)] hover:opacity-90 text-sm font-medium"
          >
            Sign in
          </Link>
        </nav>
      </div>
    </header>
  );
}
