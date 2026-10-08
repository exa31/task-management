# task-management

> Fullstack project scaffolded via **Create Stack CLI**

## 🏗️ Architecture Overview

### 🔹 Frontend: `React 19 + TypeScript`
- **Location:** `./frontend`
- **Stack:** React / Vite
- **Description:** SPA Client (Vite, TypeScript, Tailwind, ESLint)
- **Default Port:** `http://localhost:5173`

### 🔹 Backend: `Golang (Fiber v2)`
- **Location:** `./backend`
- **Stack:** Go / Fiber
- **Description:** Clean Architecture, PostgreSQL, Redis, RabbitMQ, JWT, Realtime (WS)
- **Default Port:** `http://localhost:8080`

## 🚀 Quick Start Guide

### 🛠️ Setup Frontend (`frontend`)

```bash
cd frontend
npm install
npm run dev
```

### 🛠️ Setup Backend (`backend`)

```bash
cd backend
go mod download
go run main.go
```

