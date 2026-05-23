export async function loader({ request }: { request: Request }) {
  const { downloadBackup } = await import("~/lib/recipes-http.server");
  const upstream = await downloadBackup(request);
  const fallback = `attachment; filename="recipes-backup-${new Date()
    .toISOString()
    .slice(0, 10)}.zip"`;
  return new Response(upstream.body, {
    status: upstream.status,
    headers: {
      "Content-Type": upstream.headers.get("Content-Type") ?? "application/zip",
      "Content-Disposition":
        upstream.headers.get("Content-Disposition") ?? fallback,
    },
  });
}
