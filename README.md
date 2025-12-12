# GoMeetings

An online meeting system built with Go and Gin framework.

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
│   ├── helper/          # Helper functions
│   └── define/          # Constants
└── go.mod
```

## License

See LICENSE file.
