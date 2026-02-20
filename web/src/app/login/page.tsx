import { LoginButton } from "@/components/auth/login-button";

export default function LoginPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] px-4">
      <h1 className="text-2xl font-bold mb-2">Sign in to AgentSkills</h1>
      <p className="text-[var(--muted-foreground)] mb-8 text-center max-w-md">
        Sign in with your GitHub account to star skills, manage your published
        skills, and access your API token.
      </p>
      <LoginButton />
    </div>
  );
}
