import Link from "next/link";

export function Footer() {
  return (
    <footer className="border-t border-[var(--border)] py-8 mt-16">
      <div className="mx-auto max-w-7xl px-4 flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-[var(--muted-foreground)]">
        <p>&copy; {new Date().getFullYear()} AgentSkills. Open source under MIT.</p>
        <nav className="flex items-center gap-6">
          <Link href="/search" className="hover:text-[var(--foreground)]">
            Explore
          </Link>
          <a
            href="https://github.com/anthropics/agent-skills-registry"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-[var(--foreground)]"
          >
            GitHub
          </a>
        </nav>
      </div>
    </footer>
  );
}
