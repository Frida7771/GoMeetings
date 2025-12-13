# GoMeetings

An online meeting system built with Go and Gin framework, featuring WebRTC support for real-time communication.

## Tech Stack

- **Framework**: Gin
- **ORM**: GORM
- **Database**: MySQL
- **Authentication**: JWT
- **Go Version**: 1.24.5

## Quick Start


### 1. Install Dependencies

```bash
go mod download
```

### 2. Configure Environment Variables

Create a `.env` file in the project root directory:

```env
DB_PASS=your_mysql_password
```

### 3. Create Database

```sql
CREATE DATABASE meeting CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 4. Run the Project

```bash
cd internal/server
go run main.go
```

The server runs on `:8080` by default.

## WebRTC Testing

The project includes WebRTC data channel test programs to demonstrate peer-to-peer communication.

### Testing Data Channels

1. **Start the Answer side** (receiver):
```bash
cd internal/test/data-channels/answer
go run main.go
```

2. **Start the Offer side** (sender) in another terminal:
```bash
cd internal/test/data-channels/offer
go run main.go
```

3. **Exchange SDP offers/answers**:
   - The offer program will print an encoded offer
   - Copy and paste it into the answer program
   - The answer program will print an encoded answer
   - Copy and paste it back into the offer program
   - Once connected, the data channel will open and messages will be transmitted

### How it works

- **Offer side**: Creates a data channel, generates an offer, and sends messages every 5 seconds
- **Answer side**: Receives the data channel, creates an answer, and displays received messages
- **SDP Encoding**: Uses base64-encoded JSON for SDP exchange via helper functions

## API Endpoints

### Public Endpoints

- `GET /ping` - Health check

### Authentication

- `POST /auth/user/login` - User login

### Meeting Management (Requires Authentication)

- `GET /auth/meeting/list` - Get meeting list (query: page, size, keyword)
- `POST /auth/meeting/create` - Create meeting
- `PUT /auth/meeting/edit` - Edit meeting
- `DELETE /auth/meeting/delete` - Delete meeting (query: identity)

## Project Structure

```
GoMeetings/
├── internal/
│   ├── models/          # Data models
│   ├── server/          # Server code
│   │   ├── router/      # Route configuration
│   │   └── service/     # Business logic
│   ├── middlewares/     # Middlewares
│   ├── helper/          # Helper functions (including WebRTC Encode/Decode)
│   ├── define/          # Constants
│   └── test/            # Test programs
│       └── data-channels/
│           ├── offer/   # WebRTC offer side test
│           └── answer/ # WebRTC answer side test
└── go.mod
```

## License

See LICENSE file.
