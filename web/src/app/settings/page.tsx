"use client";

import { Copy, Check } from "lucide-react";
import { useState } from "react";

export default function SettingsPage() {
  const [copied, setCopied] = useState(false);
  const token = ""; // TODO: get from session

  async function handleCopy() {
    if (!token) return;
    await navigator.clipboard.writeText(token);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <h1 className="text-2xl font-bold mb-6">Settings</h1>

      <section className="border border-[var(--border)] rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-2">API Token</h2>
        <p className="text-sm text-[var(--muted-foreground)] mb-4">
          Use this token with the AgentSkills CLI. Run{" "}
          <code className="bg-[var(--muted)] px-1.5 py-0.5 rounded text-xs">
            agentskills login
          </code>{" "}
          and paste your token when prompted.
        </p>

        {token ? (
          <div className="flex items-center gap-2">
            <code className="flex-1 bg-[var(--muted)] border border-[var(--border)] rounded-md px-3 py-2 text-sm font-mono truncate">
              {token}
            </code>
            <button
              onClick={handleCopy}
              className="shrink-0 px-3 py-2 rounded-md border border-[var(--border)] hover:bg-[var(--muted)] text-sm"
            >
              {copied ? (
                <Check className="h-4 w-4 text-green-400" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </button>
          </div>
        ) : (
          <p className="text-sm text-[var(--muted-foreground)]">
            Please sign in to view your API token.
          </p>
        )}
      </section>
    </div>
  );
}
