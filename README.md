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


## License

See LICENSE file.

