FROM node:24.13.0-bookworm-slim AS build
WORKDIR /workspace
COPY apps/web/package.json ./
RUN npm install --no-audit --no-fund
COPY apps/web/ ./
RUN npm run typecheck && npm run test && npm run build

FROM node:24.13.0-bookworm-slim
ENV NODE_ENV=production NITRO_HOST=0.0.0.0 NITRO_PORT=3000
RUN useradd --create-home --uid 10001 app
WORKDIR /app
COPY --from=build --chown=app:app /workspace/.output ./.output
USER app
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
