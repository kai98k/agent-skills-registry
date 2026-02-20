"use client";

import Image from "next/image";
import Link from "next/link";
import { LogOut, Settings, User } from "lucide-react";
import { useState } from "react";

interface UserMenuProps {
  username: string;
  avatarUrl?: string | null;
}

export function UserMenu({ username, avatarUrl }: UserMenuProps) {
  const [open, setOpen] = useState(false);

  return (
    <div className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-2"
      >
        {avatarUrl ? (
          <Image
            src={avatarUrl}
            alt={username}
            width={28}
            height={28}
            className="rounded-full"
          />
        ) : (
          <div className="h-7 w-7 rounded-full bg-[var(--muted)] flex items-center justify-center">
            <User className="h-4 w-4" />
          </div>
        )}
      </button>

      {open && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setOpen(false)}
          />
          <div className="absolute right-0 top-full mt-2 w-48 rounded-lg border border-[var(--border)] bg-[var(--card)] shadow-lg z-50 py-1">
            <div className="px-3 py-2 border-b border-[var(--border)]">
              <p className="text-sm font-medium">{username}</p>
            </div>
            <Link
              href={`/user/${username}`}
              className="flex items-center gap-2 px-3 py-2 text-sm hover:bg-[var(--muted)]"
              onClick={() => setOpen(false)}
            >
              <User className="h-4 w-4" /> Profile
            </Link>
            <Link
              href="/settings"
              className="flex items-center gap-2 px-3 py-2 text-sm hover:bg-[var(--muted)]"
              onClick={() => setOpen(false)}
            >
              <Settings className="h-4 w-4" /> Settings
            </Link>
            <a
              href="/api/auth/signout"
              className="flex items-center gap-2 px-3 py-2 text-sm hover:bg-[var(--muted)] text-[var(--destructive)]"
            >
              <LogOut className="h-4 w-4" /> Sign out
            </a>
          </div>
        </>
      )}
    </div>
  );
}
