"use client";

import { Copy, Check } from "lucide-react";
import { useState } from "react";

export function InstallCommand({ skillName }: { skillName: string }) {
  const [copied, setCopied] = useState(false);
  const command = `agentskills pull ${skillName}`;

  async function handleCopy() {
    await navigator.clipboard.writeText(command);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="flex items-center gap-2 rounded-md bg-[var(--muted)] border border-[var(--border)] px-3 py-2">
      <code className="text-sm flex-1 truncate font-mono">{command}</code>
      <button
        onClick={handleCopy}
        className="shrink-0 text-[var(--muted-foreground)] hover:text-[var(--foreground)]"
      >
        {copied ? <Check className="h-4 w-4 text-green-400" /> : <Copy className="h-4 w-4" />}
      </button>
    </div>
  );
}
