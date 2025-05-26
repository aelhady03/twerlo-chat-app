# 📨 Twerlo Chat App

A real-time chatting service built with Go, featuring user authentication, direct messaging, broadcasting, and media uploads.

## 🧰 Tech Stack

- **Backend**: Go 1.24
- **Database**: PostgreSQL 17
- **Authentication**: JWT tokens
- **Real-time**: WebSockets
- **Frontend**: Vanilla HTML/CSS/JavaScript
- **Containerization**: Docker & Docker Compose

## 🚀 Quick Start

1. Clone the repository:

```bash
git clone https://github.com/aelhady03/twerlo-chat-app.git
cd twerlo-chat-app
```

2. Start with Docker Compose:

```bash
docker compose up --build -d
```

3. Access the application:

- Web Interface: http://localhost:8080

## ✅ Implemented Features

### Core Features

- **User Authentication**: JWT-based registration and login
- **Direct Messaging**: Send messages between users
- **Broadcast Messaging**: Send messages to multiple selected users
- **Message History**: Retrieve chat history with timestamps
- **Media Upload**: Upload and share images, videos, and files
- **Real-time Communication**: WebSocket-based instant messaging

### UI Features

- **Modern Web Interface**: Responsive design with clean UI
- **User Management**: View all users and online status
- **Broadcast UI**: Select specific users for broadcast messages
- **File Sharing**: Drag-and-drop file uploads with preview
- **Message Display**: Proper styling for different message types

### Technical Features

- **Clean Architecture**: Modular codebase with separation of concerns
- **Database Migrations**: Automated schema management
- **File Storage**: Local filesystem with volume mounting
- **CORS Support**: Cross-origin resource sharing enabled
- **Error Handling**: Comprehensive error responses

## 🔌 Key API Endpoints

```http
# Authentication
POST /api/auth/register
POST /api/auth/login

# Messaging
POST /api/messages/send
POST /api/messages/broadcast
GET  /api/messages/history

# Media
POST /api/media/upload

# Users
GET  /api/users
GET  /api/users/online

# WebSocket
WS   /ws?token=<jwt-token>
```

## 📁 Project Structure

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── api/            # HTTP handlers and routes
│   ├── auth/           # JWT authentication
│   ├── database/       # Database connection and migrations
│   ├── models/         # Data models
│   ├── repository/     # Data access layer
│   ├── service/        # Business logic
│   └── websocket/      # Real-time messaging
├── static/             # Frontend assets
├── uploads/            # Media file storage
└── docker-compose.yml  # Development setup
```

## 🔮 Future Work

The following bonus features are planned for future implementation:

- **Go Concurrency**: Enhanced concurrent message processing
- **Large Volume Storage**: Optimized storage for high message volumes
- **Pagination**: Advanced pagination for message history
- **Rate Limiting**: API rate limiting and abuse prevention
- **Testing**: Comprehensive unit and integration tests
- **Message Delivery Status**: Read receipts and delivery confirmations

## 📝 License

This project is licensed under the MIT License.
