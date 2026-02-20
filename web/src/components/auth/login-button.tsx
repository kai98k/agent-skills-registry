"use client";

import { Github } from "lucide-react";

export function LoginButton() {
  return (
    <a
      href="/api/auth/signin/github"
      className="inline-flex items-center gap-2 px-4 py-2 rounded-md bg-[var(--foreground)] text-[var(--background)] font-medium text-sm hover:opacity-90"
    >
      <Github className="h-4 w-4" />
      Sign in with GitHub
    </a>
  );
}
