FROM node:22-alpine AS builder

ARG NPM_REGISTRY=https://registry.npmmirror.com
ENV npm_config_registry=${NPM_REGISTRY}
ENV CI=true

WORKDIR /app

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --ignore-scripts

COPY . .
RUN pnpm build

FROM node:22-alpine

WORKDIR /app/.output/server

COPY --from=builder /app/.output /app/.output

ENV HOST=0.0.0.0
ENV PORT=3000

EXPOSE 3000

CMD ["node", "index.mjs"]
