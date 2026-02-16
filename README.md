# Project Setup

## Setup
Copy environment file and configure database:
cp .env.example .env
Edit `.env` for DB and related fields.

## Migration
go run cmd/migrate/main.go

## Run Server
go run cmd/server/main.go

## API Endpoints

**Login**  
POST `/auth/login`  
{ "email": "test@gmail.com" }

**Verify Token**  
GET `/auth/verify?token=UUID`

**Create Session and get callback URL with deep link.**  
POST `/session`  
{ "token": "UUID" }

**Submit Feedback**  
POST `/api/feedback`  
Headers: { "Authorization": "Bearer <JWT_TOKEN>" }  
Body: { "content": "text" }
