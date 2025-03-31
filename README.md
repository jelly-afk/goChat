# goChat

A real-time chat application built with Go and WebSockets.

## Features

- User authentication with JWT
- Real-time messaging using WebSockets
- Public and private chat rooms
- Message persistence in PostgreSQL

## Setup

1. **Prerequisites**
   - Go 1.16+
   - PostgreSQL

2. **Installation**
   ```bash
   # Clone and enter directory
   git clone https://github.com/yourusername/chat-app.git
   cd chat-app

   # Install dependencies
   go mod download

   # Create database
 
   ```

3. **Run the app**
   ```bash
   go run ./cmd/web
   ```

## API Endpoints

### Public
- `POST /user/register` - Register
- `POST /user/login` - Login

### Protected
- `POST /chat/create` - Create chat
- `POST /chat/message` - Send message
- `GET /chat/messages/:chat_id` - Get messages
- `POST /chat/join` - Join chat
- `GET /ws` - WebSocket connection

## Environment Variables

- `PORT` - Server port (default: 4000)
- `DB_DSN` - Database connection string
- `JWT_SECRET` - JWT secret key 