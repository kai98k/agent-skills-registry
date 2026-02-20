import type {
  CategoriesResponse,
  SearchResponse,
  SkillResponse,
  SkillVersionsResponse,
  StarResponse,
  UserResponse,
} from "@/types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

// --- Server-side fetches (used in Server Components with ISR) ---

export async function getSkill(name: string): Promise<SkillResponse> {
  const res = await fetch(`${API_BASE}/v1/skills/${encodeURIComponent(name)}`, {
    next: { revalidate: 60 },
  });
  if (!res.ok) throw new Error(`Skill not found: ${name}`);
  return res.json();
}

export async function getSkillVersions(
  name: string
): Promise<SkillVersionsResponse> {
  const res = await fetch(
    `${API_BASE}/v1/skills/${encodeURIComponent(name)}/versions`,
    { next: { revalidate: 60 } }
  );
  if (!res.ok) throw new Error(`Skill not found: ${name}`);
  return res.json();
}

export async function searchSkills(params: {
  q?: string;
  category?: string;
  tag?: string;
  provider?: string;
  sort?: string;
  page?: number;
  per_page?: number;
}): Promise<SearchResponse> {
  const searchParams = new URLSearchParams();
  if (params.q) searchParams.set("q", params.q);
  if (params.category) searchParams.set("category", params.category);
  if (params.tag) searchParams.set("tag", params.tag);
  if (params.provider) searchParams.set("provider", params.provider);
  if (params.sort) searchParams.set("sort", params.sort);
  if (params.page) searchParams.set("page", String(params.page));
  if (params.per_page) searchParams.set("per_page", String(params.per_page));

  const res = await fetch(`${API_BASE}/v1/skills?${searchParams.toString()}`, {
    cache: "no-store",
  });
  if (!res.ok) throw new Error("Search failed");
  return res.json();
}

export async function getCategories(): Promise<CategoriesResponse> {
  const res = await fetch(`${API_BASE}/v1/categories`, {
    next: { revalidate: 300 },
  });
  if (!res.ok) throw new Error("Failed to fetch categories");
  return res.json();
}

export async function getUserProfile(
  username: string
): Promise<UserResponse> {
  const res = await fetch(
    `${API_BASE}/v1/users/${encodeURIComponent(username)}`,
    { next: { revalidate: 120 } }
  );
  if (!res.ok) throw new Error(`User not found: ${username}`);
  return res.json();
}

// --- Client-side fetches (used in Client Components for mutations) ---

export async function starSkill(
  name: string,
  token: string
): Promise<StarResponse> {
  const res = await fetch(
    `${API_BASE}/v1/skills/${encodeURIComponent(name)}/star`,
    {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    }
  );
  if (!res.ok) throw new Error("Failed to star skill");
  return res.json();
}

export async function unstarSkill(
  name: string,
  token: string
): Promise<StarResponse> {
  const res = await fetch(
    `${API_BASE}/v1/skills/${encodeURIComponent(name)}/star`,
    {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    }
  );
  if (!res.ok) throw new Error("Failed to unstar skill");
  return res.json();
}
