export function loader() {
  return new Response("ok\n", {
    headers: {
      "Content-Type": "text/plain; charset=utf-8",
    },
  });
}
