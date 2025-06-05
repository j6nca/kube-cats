FROM node:18-alpine as builder

WORKDIR /app
COPY . .
RUN npm install && npm run build

FROM node:18-alpine

WORKDIR /app
COPY --from=builder /app/build ./build

RUN npm install -g serve

CMD ["serve", "-s", "build"]
EXPOSE 3000
