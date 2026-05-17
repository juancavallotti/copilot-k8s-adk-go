import { type RouteConfig, index, layout, route } from "@react-router/dev/routes";

export default [
  route("healthz", "routes/healthz.ts"),
  layout("routes/app-layout.tsx", [
    index("routes/_index.tsx"),
    route("create", "routes/create.tsx"),
    route("recipe/:id/edit", "routes/recipe-edit.tsx"),
    route("recipe/:id", "routes/recipe.tsx"),
  ]),
] satisfies RouteConfig;
