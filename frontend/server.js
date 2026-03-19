const port = process.env.PORT || 8081;

const server = Bun.serve({
  port: port,
  async fetch(req) {
    const url = new URL(req.url);
    let path = url.pathname;
    
    // Serve index.html for root or if no extension (SPA routing)
    if (path === "/" || !path.includes(".")) {
      path = "/index.html";
    }

    const file = Bun.file(`./dist${path}`);
    if (await file.exists()) {
      return new Response(file);
    }

    // Fallback to index.html for any unknown paths (client-side routing)
    return new Response(Bun.file("./dist/index.html"));
  },
});

console.log(`Server running at http://localhost:${server.port}`);
