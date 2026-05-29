# Build
docker compose build                       # build all images
docker compose build auth-svc              # build one
docker compose build --no-cache            # ignore layer cache

# Start
docker compose up -d                       # all services, background
docker compose up                          # foreground (see logs streamed)
docker compose up -d auth-svc              # one service + its deps
docker compose up -d --build               # build then start

# Stop
docker compose stop                        # stop without removing
docker compose down                        # stop AND remove containers (KEEP volumes)
docker compose down -v                     # stop + remove + WIPE volumes ⚠️ DATA LOSS

# Restart
docker compose restart                     # restart everything
docker compose restart auth-svc            # restart one