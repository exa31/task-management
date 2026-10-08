# ==========================================
# Multi-stage Dockerfile for React Vite Frontend
# ==========================================

# Stage 1: Build static assets
FROM node:22-alpine AS builder

WORKDIR /build

# Copy dependency manifests
COPY package.json package-lock.json ./

# Install dependencies cleanly
RUN npm ci

# Copy frontend source code
COPY . .

# Build production bundle
ARG VITE_API_URL=""
ARG VITE_APP_API_URL=""
ARG VITE_WS_URL=""
ARG VITE_GOOGLE_CLIENT_ID=""
ARG VITE_APPLE_CLIENT_ID=""
ARG VITE_ENABLE_MOCKS="false"
ARG VITE_APP_VERSION=""

ENV VITE_API_URL=$VITE_API_URL
ENV VITE_APP_API_URL=$VITE_APP_API_URL
ENV VITE_WS_URL=$VITE_WS_URL
ENV VITE_GOOGLE_CLIENT_ID=$VITE_GOOGLE_CLIENT_ID
ENV VITE_APPLE_CLIENT_ID=$VITE_APPLE_CLIENT_ID
ENV VITE_ENABLE_MOCKS=$VITE_ENABLE_MOCKS
ENV VITE_APP_VERSION=$VITE_APP_VERSION

RUN npm run build

# Stage 2: Serve with lightweight Nginx
FROM nginx:alpine AS runner

# Remove default nginx website
RUN rm -rf /usr/share/nginx/html/*

# Copy built assets from builder
COPY --from=builder /build/dist /usr/share/nginx/html

# Copy custom Nginx configuration for SPA routing
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
