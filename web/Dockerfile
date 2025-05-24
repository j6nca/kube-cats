FROM node:18-alpine as builder

ARG API_URL
ENV API_URL=http://kube-cats-api

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
