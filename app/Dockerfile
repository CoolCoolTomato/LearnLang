FROM node:22-alpine AS base
WORKDIR /app

RUN corepack enable

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

FROM base AS dev
COPY . .

EXPOSE 5173
CMD ["pnpm", "dev", "--host", "0.0.0.0", "--port", "5173"]

FROM base AS build
COPY . .

ARG VITE_API_BASE_URL=
ARG VITE_BASENAME=
ARG VITE_API_PREFIX=/api/user
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_BASENAME=${VITE_BASENAME}
ENV VITE_API_PREFIX=${VITE_API_PREFIX}

RUN pnpm build

FROM nginx:1.27-alpine AS production
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
