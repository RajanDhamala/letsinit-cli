# npm option (active)
FROM node:22-alpine AS development

WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
CMD ["npm", "run", "dev"]
EXPOSE 5173


FROM node:22-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build


FROM nginx:alpine AS production

COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

# pnpm option (commented)
# To use pnpm, comment all npm stages above and uncomment these stages.
# FROM node:22-alpine AS development
#
# RUN corepack enable
# WORKDIR /app
# COPY package.json pnpm-lock.yaml ./
# RUN pnpm install --frozen-lockfile
# COPY . .
# CMD ["pnpm", "run", "dev"]
# EXPOSE 5173
#
# FROM node:22-alpine AS build
#
# RUN corepack enable
# WORKDIR /app
# COPY package.json pnpm-lock.yaml ./
# RUN pnpm install --frozen-lockfile
# COPY . .
# RUN pnpm run build
#
# FROM nginx:alpine AS production
#
# COPY --from=build /app/dist /usr/share/nginx/html
# COPY nginx.conf /etc/nginx/conf.d/default.conf
#
# EXPOSE 80
# CMD ["nginx", "-g", "daemon off;"]
