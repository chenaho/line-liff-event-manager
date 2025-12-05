# LINE LIFF Event Manager

A lightweight event management system integrated with LINE LIFF.

## Tech Stack
- **Frontend**: Vue 3, Pinia, TailwindCSS, LIFF SDK
- **Backend**: Go (Gin), Firestore
- **Infrastructure**: Docker Compose, Nginx

## Setup

### Prerequisites
- Docker & Docker Compose
- Go 1.23+
- Node.js 18+

### Local Development

1. **Backend Setup**
   ```bash
   cd backend
   go mod download
   # Create .env file with FIREBASE_CREDENTIALS path if needed
   go run cmd/main.go
   ```
   **How to get `firebase-key.json`:**
   1. Go to [Firebase Console](https://console.firebase.google.com/) > Select your project.
   2. Click the **Gear icon** (Project Settings) > **Service accounts** tab.
   3. Click **Generate new private key** > **Generate key**.
   4. Save the file as `firebase-key.json` in this project's root directory.

2. **Frontend Setup**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

3. **Docker Compose (Full Stack)**
   ```bash
   docker-compose up --build
   ```
   Access the app at http://localhost

## Project Structure
- `backend/`: Go API server
- `frontend/`: Vue 3 SPA
- `nginx/`: Nginx configuration
