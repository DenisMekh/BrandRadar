FROM oven/bun:1 AS builder

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

COPY . .
RUN bun run build

# Production
FROM oven/bun:1-slim

WORKDIR /app

COPY --from=builder /app/dist ./dist
COPY server.js .

ENV NODE_ENV=production
ENV PORT=8081

EXPOSE 8081

CMD ["bun", "run", "server.js"]
