/**
 * Shared skill types (safe for client bundles).
 * All HTTP calls to the skills endpoints live in `skills-http.server.ts` (Node only).
 */
export type Skill = {
  id: string;
  name: string;
  description: string;
  content: string;
  created_at: string;
  updated_at: string;
};

export type SkillCreate = {
  name: string;
  description: string;
  content: string;
};

export type SkillPatch = {
  description?: string;
  content?: string;
};
