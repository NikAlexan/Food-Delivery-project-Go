FROM node:24-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
ARG API_GATEWAY_URL=http://api-gateway
ENV API_GATEWAY_URL=$API_GATEWAY_URL
RUN npm run build

FROM node:24-alpine
WORKDIR /app
COPY --from=builder /app/.output ./.output
ENV NODE_ENV=production
ENV PORT=3000
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
