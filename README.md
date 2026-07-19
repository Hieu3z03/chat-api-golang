# Go Gin Clean Starter

Boilerplate Go API theo mô hình Controller - Service - Repository, chạy trực tiếp trên máy và không phụ thuộc Docker.

## Thành phần giữ lại

- Gin HTTP server
- Clean architecture theo module
- PostgreSQL qua GORM
- MongoDB qua MongoDB Go Driver
- Dependency injection với `samber/do`
- Migration, seeder và module generator
- JWT authentication và CRUD user cơ bản

Tích hợp SMTP, xác thực email, quên mật khẩu qua email và toàn bộ cấu hình Docker đã được loại bỏ. Trường `email` vẫn là định danh dùng để đăng ký và đăng nhập.

## Yêu cầu

- Go theo phiên bản trong `go.mod`
- PostgreSQL đang chạy local; có thể tạo và quản lý database bằng pgAdmin
- MongoDB đang chạy local hoặc có MongoDB URI truy cập được

## Cấu hình

1. Tạo database PostgreSQL, ví dụ `go_gin_clean`, bằng pgAdmin.
2. Sao chép `.env.example` thành `.env`.
3. Thay thông tin kết nối và secret:

```env
APP_NAME=go-gin-clean
GOLANG_PORT=8888
JWT_SECRET=change_me

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=change_me
DB_NAME=go_gin_clean
DB_SSLMODE=disable

MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=go_gin_clean
```

File `.env` là tùy chọn. Khi không có file này, ứng dụng dùng địa chỉ local mặc định; password và `JWT_SECRET` vẫn nên được khai báo bằng biến môi trường.

Ứng dụng kiểm tra cả PostgreSQL và MongoDB khi khởi động. PostgreSQL user cần quyền tạo extension `uuid-ossp` trong database đã chọn.

## Chạy local

```bash
go mod tidy
go run ./cmd/main.go --migrate:run
go run ./cmd/main.go --seed
go run ./cmd/main.go
```

API mặc định chạy tại `http://localhost:8888`.

Nếu đã cài Air, có thể chạy hot reload bằng:

```bash
air
```

Air build binary tạm trong `tmp/`, không còn dùng đường dẫn Docker.

## Lệnh database

```bash
go run ./cmd/main.go --migrate:run
go run ./cmd/main.go --migrate:status
go run ./cmd/main.go --migrate:rollback
go run ./cmd/main.go --migrate:rollback 2
go run ./cmd/main.go --migrate:rollback:all
go run ./cmd/main.go --migrate:create:create_posts_table
go run ./cmd/main.go --seed
go run ./cmd/main.go --migrate:run --seed
```

Các lệnh tương ứng cũng có trong `Makefile`: `make run`, `make migrate`, `make seed`, `make migrate-seed` và `make test`.

## Sử dụng kết nối database trong module

PostgreSQL:

```go
db := do.MustInvokeNamed[*gorm.DB](injector, constants.PostgreSQL)
```

MongoDB:

```go
db := do.MustInvokeNamed[*mongo.Database](injector, constants.MongoDB)
```

Hai kết nối được đăng ký trong `providers/core.go` và được đóng khi injector shutdown.

## Cấu trúc chính

```text
cmd/            điểm khởi động ứng dụng
config/         cấu hình PostgreSQL, MongoDB và logger
database/       entity, migration và seeder PostgreSQL
middlewares/    Gin middleware
modules/        controller, service, repository, DTO theo domain
pkg/            helper, constant và utility dùng chung
providers/      đăng ký dependency injection
script/         command/script chạy từ CLI
```

## API base

```text
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/refresh
POST   /api/auth/logout
GET    /api/user
GET    /api/user/me
PUT    /api/user/:id
DELETE /api/user/:id
```

Xem `modules/auth/routes.go` và `modules/user/routes.go` để kiểm tra route cùng middleware hiện tại.

## Kiểm tra

```bash
go test ./...
go vet ./...
go build ./...
```

## License

MIT, xem `LICENSE`.
