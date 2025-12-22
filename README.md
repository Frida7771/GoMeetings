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

- Swagger UI: `http://localhost:8080/swagger/index.html`


### How it works

- **Offer side**: Creates a data channel, generates an offer, and sends messages every 5 seconds
- **Answer side**: Receives the data channel, creates an answer, and displays received messages
- **SDP Encoding**: Uses base64-encoded JSON for SDP exchange via helper functions


### Room Management (Requires Authentication)

- `GET /auth/room/list` - Paginated room list + join state
- `POST /auth/room/create` - Create room、设定 `join_code`（或自动生成）、可选短邀请码
- `PUT /auth/room/edit` - Update room metadata（含 join_code / short_code）
- `DELETE /auth/room/delete` - Delete room (owner only)
- `POST /auth/room/join` - Join room by `identity + join_code + display_name`
- `POST /auth/room/leave` - Leave room
- `GET /auth/room/members` - 查询参会者名单（含 display_name、加入时间）
- `POST /auth/room/share/start` - 发起屏幕共享（单房间单路、需房间成员）
- `POST /auth/room/share/stop` - 停止屏幕共享（共享者或房主可调用）
- `GET /auth/room/share/status` - 查询当前屏幕共享状态



## License

See LICENSE file.

