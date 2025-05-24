FROM node:18-alpine as builder

WORKDIR /app
COPY . .
RUN npm install && npm run build

FROM node:18-alpine

WORKDIR /app
COPY --from=builder /app/build ./build
COPY entrypoint.sh .

RUN npm install -g serve
RUN chmod +x entrypoint.sh

EXPOSE 3000
ENTRYPOINT ["./entrypoint.sh"]
