FROM node:lts-alpine

RUN apk add --no-cache git
# corepack is no longer bundled with Node 25+, so pnpm comes from npm
RUN npm install -g pnpm@latest

WORKDIR /app
