export interface SkillVersionDetail {
  version: string;
  description: string;
  checksum: string;
  size_bytes: number;
  published_at: string;
  providers: string[];
  metadata: Record<string, unknown>;
}

export interface SkillResponse {
  name: string;
  owner: string;
  owner_avatar_url: string | null;
  downloads: number;
  stars_count: number;
  starred_by_me: boolean;
  category: string | null;
  readme_html: string | null;
  created_at: string;
  latest_version: SkillVersionDetail | null;
}

export interface SkillVersionSummary {
  version: string;
  checksum: string;
  size_bytes: number;
  published_at: string;
  providers: string[];
}

export interface SkillVersionsResponse {
  name: string;
  versions: SkillVersionSummary[];
}

export interface SearchResultItem {
  name: string;
  description: string;
  owner: string;
  owner_avatar_url: string | null;
  downloads: number;
  stars_count: number;
  latest_version: string;
  category: string | null;
  updated_at: string;
  tags: string[];
  providers: string[];
}

export interface SearchResponse {
  total: number;
  page: number;
  per_page: number;
  results: SearchResultItem[];
}

export interface CategoryItem {
  name: string;
  label: string;
  icon: string | null;
  skill_count: number;
}

export interface CategoriesResponse {
  categories: CategoryItem[];
}

export interface UserSkillItem {
  name: string;
  description: string;
  downloads: number;
  stars_count: number;
  latest_version: string;
  updated_at: string;
}

export interface UserResponse {
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  bio: string | null;
  created_at: string;
  skills: UserSkillItem[];
  total_downloads: number;
  total_stars: number;
}

export interface StarResponse {
  starred: boolean;
  stars_count: number;
}
