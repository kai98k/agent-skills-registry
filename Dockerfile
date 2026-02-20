# ==========================================================
# AgentSkills Omnibus Image
# Single image with PostgreSQL + MinIO + FastAPI + Next.js
# Usage: docker run -p 3000:3000 -p 8000:8000 agentskills
# ==========================================================

# ---------- Stage 1: Build Next.js ----------
FROM node:22-slim AS web-builder
WORKDIR /build

COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install

COPY web/ .
ARG NEXT_PUBLIC_API_URL=http://127.0.0.1:8000
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
RUN npm run build


# ---------- Stage 2: Final omnibus image ----------
FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

# ---- System packages ----
RUN apt-get update && apt-get install -y --no-install-recommends \
        # PostgreSQL 15 (bookworm default)
        postgresql postgresql-client \
        # Python
        python3 python3-pip python3-venv \
        # Node.js runtime (for Next.js standalone server)
        nodejs \
        # Process manager
        supervisor \
        # Utilities
        curl ca-certificates bash \
    && rm -rf /var/lib/apt/lists/*

# ---- Install MinIO server + mc client ----
RUN ARCH=$(dpkg --print-architecture) && \
    if [ "$ARCH" = "arm64" ]; then MINIO_ARCH="arm64"; MC_ARCH="arm64"; \
    else MINIO_ARCH="amd64"; MC_ARCH="amd64"; fi && \
    curl -fsSL "https://dl.min.io/server/minio/release/linux-${MINIO_ARCH}/minio" -o /usr/local/bin/minio && \
    curl -fsSL "https://dl.min.io/client/mc/release/linux-${MC_ARCH}/mc" -o /usr/local/bin/mc && \
    chmod +x /usr/local/bin/minio /usr/local/bin/mc

# ---- Create minio system user ----
RUN groupadd -r minio-user && useradd -r -g minio-user minio-user

# ---- Working directory ----
RUN mkdir -p /opt/agentskills /var/lib/agentskills /var/log/agentskills

# ---- Install Python API ----
COPY api/requirements.txt /opt/agentskills/api/requirements.txt
RUN python3 -m pip install --no-cache-dir --break-system-packages \
        -r /opt/agentskills/api/requirements.txt
COPY api/ /opt/agentskills/api/

# ---- Copy built Next.js app ----
COPY --from=web-builder /build/public /opt/agentskills/web/public
COPY --from=web-builder /build/.next/standalone/ /opt/agentskills/web/
COPY --from=web-builder /build/.next/static /opt/agentskills/web/.next/static

# ---- Copy init.sql ----
COPY init.sql /opt/agentskills/init.sql

# ---- Copy supervisor + entrypoint configs ----
COPY docker/supervisord.conf /etc/supervisor/supervisord.conf
COPY docker/entrypoint.sh /opt/agentskills/entrypoint.sh
RUN chmod +x /opt/agentskills/entrypoint.sh

# ---- Environment defaults ----
ENV AGENTSKILLS_DATA_DIR=/var/lib/agentskills
ENV MINIO_ROOT_USER=minioadmin
ENV MINIO_ROOT_PASSWORD=minioadmin
ENV GITHUB_CLIENT_ID=""
ENV GITHUB_CLIENT_SECRET=""
ENV NEXTAUTH_URL=http://localhost:3000
ENV NEXTAUTH_SECRET=change-me-in-production

# ---- Expose ports ----
# 3000 = Web UI, 8000 = API, 9000 = MinIO API, 9001 = MinIO Console
EXPOSE 3000 8000 9000 9001

# ---- Persistent data ----
VOLUME ["/var/lib/agentskills"]

ENTRYPOINT ["/opt/agentskills/entrypoint.sh"]
