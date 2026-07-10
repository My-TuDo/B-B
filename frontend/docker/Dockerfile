FROM node:22-alpine AS builder

ARG NPM_REGISTRY=https://registry.npmmirror.com

WORKDIR /app

RUN corepack enable && corepack prepare pnpm@latest --activate

RUN if [ "$NPM_REGISTRY" != "" ]; then \
      pnpm config set registry "$NPM_REGISTRY"; \
    fi

COPY package.json pnpm-lock.yaml ./
RUN pnpm install

COPY . .
RUN pnpm build

FROM nginx:alpine

COPY --from=builder /app/.output/public /usr/share/nginx/html
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
