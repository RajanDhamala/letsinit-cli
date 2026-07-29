FROM node:22-alpine
WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY --chown=node:node . .

RUN chown -R node:node /app

USER node

EXPOSE 8000
CMD ["npm", "start"]


# To use pnpm, comment the npm block above and uncomment this block.
#
# FROM node:22-alpine
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
# RUN chown -R node:node /app
#
# USER node
#
# EXPOSE 8000
# CMD ["pnpm", "start"]
