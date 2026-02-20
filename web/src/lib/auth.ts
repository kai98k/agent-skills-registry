import NextAuth from "next-auth";
import GitHub from "next-auth/providers/github";

// Auth callbacks run server-side — use internal Docker URL if available
const API_BASE = process.env.INTERNAL_API_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

export const { handlers, signIn, signOut, auth } = NextAuth({
  providers: [
    GitHub({
      clientId: process.env.GITHUB_CLIENT_ID!,
      clientSecret: process.env.GITHUB_CLIENT_SECRET!,
    }),
  ],
  callbacks: {
    async signIn({ account }) {
      if (account?.provider === "github" && account.access_token) {
        try {
          const res = await fetch(`${API_BASE}/v1/auth/github`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              github_access_token: account.access_token,
            }),
          });
          if (!res.ok) return false;
          const data = await res.json();
          // Store API token in the account for session callback
          (account as Record<string, unknown>).api_token = data.api_token;
          (account as Record<string, unknown>).api_username = data.username;
        } catch {
          return false;
        }
      }
      return true;
    },
    async jwt({ token, account }) {
      if (account) {
        token.api_token = (account as Record<string, unknown>).api_token;
        token.api_username = (account as Record<string, unknown>).api_username;
      }
      return token;
    },
    async session({ session, token }) {
      (session as Record<string, unknown>).api_token = token.api_token;
      (session as Record<string, unknown>).api_username = token.api_username;
      return session;
    },
  },
});
