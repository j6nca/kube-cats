FROM node:18-alpine AS builder

WORKDIR /app

# Copy and install dependencies
COPY package.json package-lock.json ./
RUN npm ci

# Copy source code
COPY . .
RUN npm run build

# Output stage
FROM alpine:3.18 AS output

WORKDIR /output
COPY --from=builder /app/build .