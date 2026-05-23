import { type RouteConfig, index, layout, route } from "@react-router/dev/routes";

export default [
  route("healthz", "routes/healthz.ts"),
  route("recipes/backup", "routes/recipes.backup.ts"),
  layout("routes/app-layout.tsx", [
    index("routes/_index.tsx"),
    route("create", "routes/create.tsx"),
    route("recipe/:id/edit", "routes/recipe-edit.tsx"),
    route("recipe/:id", "routes/recipe.tsx"),
    route("traces", "routes/traces._index.tsx"),
    route("traces/:event_id", "routes/traces.$event_id.tsx"),
    route("skills", "routes/skills._index.tsx"),
    route("skills/:id", "routes/skills.$id.tsx"),
  ]),
] satisfies RouteConfig;
