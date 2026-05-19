FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

RUN npm run build

FROM node:20-alpine

WORKDIR /app

ENV HOST=0.0.0.0
ENV PORT=3000

COPY --from=builder /app/.output /app/.output
COPY --from=builder /app/package.json /app/package.json

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
