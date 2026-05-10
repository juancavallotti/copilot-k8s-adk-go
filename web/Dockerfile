FROM node:20-alpine AS development-dependencies-env
COPY . /app
WORKDIR /app
RUN npm ci

FROM node:20-alpine AS production-dependencies-env
COPY ./package.json package-lock.json /app/
WORKDIR /app
RUN npm ci --omit=dev

FROM node:20-alpine AS build-env
COPY . /app/
COPY --from=development-dependencies-env /app/node_modules /app/node_modules
WORKDIR /app
RUN npm run build

FROM node:20-alpine
ENV NODE_ENV=production
# HTTP port for react-router-serve (override in k8s with containerPort).
ENV PORT=3000
# Recipes backend base URL (no trailing slash). Examples:
#   http://recipes-api:4000   — pod → cluster Service (server-side only if browser cannot reach it)
#   https://api.example.com   — public API origin
#   (unset)                   — same-origin relative "/api" (use Ingress or a reverse proxy)
ENV RECIPES_API_BASE=
EXPOSE 3000
COPY ./package.json package-lock.json /app/
COPY --from=production-dependencies-env /app/node_modules /app/node_modules
COPY --from=build-env /app/build /app/build
WORKDIR /app
CMD ["npm", "run", "start"]