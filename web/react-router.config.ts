import type { Config } from "@react-router/dev/config";

export default {
  // Server-side render by default, to enable SPA mode set this to `false`
  ssr: true,
  /**
   * Keeps loader/action code (and their imports) out of the browser bundle.
   * Without this, dev/HMR can leave a static import to `.server` modules in a shared chunk.
   */
  future: {
    v8_splitRouteModules: "enforce",
  },
} satisfies Config;
