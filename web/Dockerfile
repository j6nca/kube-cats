FROM node:22-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:22-alpine

WORKDIR /app
RUN npm install -g serve
COPY --from=builder /app/build ./build
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# build/config holds only the generated runtime config and is the one path
# written at start-up. Owned by the runtime uid so the entrypoint still works
# when no volume is mounted; mount a writable volume here under
# readOnlyRootFilesystem.
RUN mkdir -p /app/build/config && chown -R 1000:1000 /app/build/config

# Run as non-root by default so the image does not depend on the platform
# imposing a securityContext to be safe.
USER 1000

EXPOSE 3000

# The entrypoint regenerates the runtime config from the environment on every
# start, so configuration is a restart rather than a rebuild.
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["serve", "-s", "build", "-l", "3000"]
