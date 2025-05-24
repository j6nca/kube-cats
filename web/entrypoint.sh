#!/bin/sh

# Inject runtime environment variable into env.js
cat <<EOF > /app/build/env.js
window.env = {
  API_URL: "${API_URL}"
};
EOF

# Start the lightweight web server
serve -s build -l 3000