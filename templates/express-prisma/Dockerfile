FROM node:22-alpine AS production
WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY --chown=node:node . .

RUN npx prisma generate

RUN chown -R node:node /app

USER node

EXPOSE 8000
CMD ["npm", "start"]


# pnpm version (commented out — uncomment this block and
#
# FROM node:22-alpine AS production
#
# RUN corepack enable && corepack prepare pnpm@10.25.0 --activate
#
# WORKDIR /app
#
# COPY package.json pnpm-lock.yaml ./
# RUN pnpm install --frozen-lockfile
#
# COPY --chown=node:node . .
#
# RUN pnpm exec prisma generate
#
# RUN chown -R node:node /app
#
# USER node
#
# EXPOSE 8000
# CMD ["pnpm", "start"]
