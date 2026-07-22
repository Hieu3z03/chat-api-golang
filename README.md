# Chat API Golang

Chat service viết bằng Gin theo cấu trúc controller - service - repository. Service chạy độc lập, dùng PostgreSQL cho dữ liệu quan hệ và MongoDB cho message.

## Identity contract

Chat service không xác thực JWT và không đọc header `Authorization`. Client truyền `x-user-id`; `x-user-role` là tùy chọn cho frontend development:

```text
x-user-id: <user UUID>
x-user-role: <role UUID, optional>
```

Middleware luôn kiểm tra `x-user-id` là UUID hợp lệ. Nếu có `x-user-role`, giá trị này cũng phải là UUID hợp lệ; role chưa được lưu vào database.

## Data model

PostgreSQL:

- `users`: `id`, `username`, `name`, `avatar_url`
- `channels`: `id`, `name`, `created_by`, timestamps
- `channel_members`: `channel_id`, `user_id`, `joined_at`, `last_read_sequence`, `last_read_at`

MongoDB collection `messages`:

- `_id`
- `channel_id`
- `user_id`
- `sequence`: số thứ tự tăng dần trong từng channel
- `content`: JSON object bắt buộc có `type`
- `created_at`

Collection nội bộ `message_counters` cấp `sequence` atomically theo channel. Chat service chỉ lưu URL avatar cần để hiển thị chat.

## Cách hệ thống hoạt động

1. Auth/gateway xác định người gọi rồi chuyển `user UUID` và `role UUID` qua hai header.
2. Chat service chỉ đọc header; không kiểm tra token và không lưu password.
3. Hồ sơ user được đồng bộ vào PostgreSQL qua `PUT /api/users/me`.
4. Khi tạo channel, service lấy creator từ `x-user-id`, kiểm tra các user đã tồn tại local rồi ghi `channels` và `channel_members` trong một transaction PostgreSQL.
5. Khi gửi message, service kiểm tra membership rồi cấp `sequence` và ghi MongoDB. Khi đọc, service lấy messages một lần, cập nhật read pointer và lấy toàn bộ member + user trong một câu SQL, sau đó dựng `seen_by` trong Go.
6. `content` của message linh hoạt theo `type`; chat service chỉ giữ ID của tài nguyên thuộc service khác.

## Cấu hình local

Sao chép `.env.example` thành `.env` và nhập thông tin database thật:

```env
APP_NAME=chat-api
GOLANG_PORT=8888

LOG_LEVEL=info
LOG_SQL=true
LOG_HTTP=true

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=your_postgres_password
DB_NAME=chat_api
DB_SSLMODE=disable

MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=chat_api
```

Logs are emitted as one JSON object per line on stdout. `LOG_LEVEL` accepts
`trace`, `debug`, `info`, `warn`, `error`, or `disabled`. `LOG_SQL` controls
GORM query logs and `LOG_HTTP` controls request completion logs; request error
logs remain enabled according to `LOG_LEVEL`.

Tạo database PostgreSQL `chat_api` trước bằng pgAdmin. Ứng dụng yêu cầu cả PostgreSQL và MongoDB truy cập được khi khởi động.

```powershell
go run ./cmd/main.go --migrate:run
go run ./cmd/main.go --seed
go run ./cmd/main.go
```

Server mặc định chạy tại `http://localhost:8888`.

## Chạy bằng Docker Compose

Docker Compose chỉ khởi động Centrifugo và backend Go. Backend dùng thông tin PostgreSQL và MongoDB có sẵn trong `.env`.

```powershell
if (-not (Test-Path .env)) { Copy-Item .env.example .env }
docker compose up --build -d
docker compose ps
```

Nếu database chạy trên cùng máy với Docker, Compose tự dùng `host.docker.internal`. Có thể ghi rõ hoặc đổi sang địa chỉ database server khác trong `.env` bằng:

```env
DOCKER_DB_HOST=host.docker.internal
DOCKER_MONGO_URI=mongodb://host.docker.internal:27017
```

Nếu database nằm trên server khác, dùng hostname hoặc IP mà container truy cập được. Có thể chạy migration thủ công bằng:

```powershell
docker compose run --rm backend --migrate:run
```

Các endpoint local:

- Backend API: `http://localhost:8888`
- Centrifugo WebSocket: `ws://localhost:8000/connection/websocket`
- Centrifugo admin UI: `http://localhost:8000` (dùng `CENTRIFUGO_ADMIN_PASSWORD` trong `.env`)

Xem log và dừng stack:

```powershell
docker compose logs -f backend centrifugo
docker compose down
```

Config Centrifugo tại `deploy/centrifugo/config.json` đang cho phép mọi origin để tiện phát triển local. Khi deploy cần giới hạn `allowed_origins` và thay toàn bộ secret `change_me_*`. Backend cấp JWT cho protected channel `$personal_<user-id>` và publish event `message_added` tới personal channel của từng thành viên.

Frontend lấy token realtime qua:

```text
GET /api/realtime/connection-token
GET /api/realtime/subscription-token?channel=$personal_<user-id>
```

## API

Các ví dụ dùng hai UUID mẫu:

```text
x-user-id: 11111111-1111-4111-8111-111111111111
x-user-role: aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa
```

### Đồng bộ user local

User phải tồn tại trong bảng `users` của chat service trước khi tạo/join channel.

```http
PUT /api/users/me
Content-Type: application/json
x-user-id: 11111111-1111-4111-8111-111111111111
x-user-role: aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa

{
  "username": "hieu",
  "name": "Hieu Nguyen",
  "avatar_url": null
}
```

```text
GET /api/users/me
GET /api/users?search=hieu&limit=20
GET /api/users/:user_id
```

### Tạo và đọc channel

Creator lấy từ `x-user-id` và tự động trở thành channel member. `member_ids` là tùy chọn.

```http
POST /api/channels
Content-Type: application/json
x-user-id: 11111111-1111-4111-8111-111111111111
x-user-role: aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa

{
  "name": "Backend team",
  "member_ids": [
    "22222222-2222-4222-8222-222222222222"
  ]
}
```

Body tối thiểu cũng hợp lệ:

```json
{
  "name": "Backend team"
}
```

```text
GET /api/channels
GET /api/channels/:channel_id
```

### Gửi và đọc message

Chỉ user có trong `channel_members` mới đọc hoặc ghi message.

```http
POST /api/channels/:channel_id/messages
Content-Type: application/json
x-user-id: 11111111-1111-4111-8111-111111111111
x-user-role: aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa

{
  "content": {
    "type": "text",
    "text": "Hello team"
  }
}
```

`content` có thể chứa payload khác theo type, ví dụ file:

```json
{
  "content": {
    "type": "file",
    "file_id": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
    "name": "spec.pdf"
  }
}
```

Đọc message mới nhất trước:

```text
GET /api/channels/:channel_id/messages?limit=50
GET /api/channels/:channel_id/messages?limit=50&before=2026-07-19T10:30:00Z
```

Mỗi message trả về có `sequence` và `seen_by`. Khi GET thành công, `last_read_sequence` của người gọi được tăng đến sequence lớn nhất trong page (không bao giờ giảm); một member được đưa vào `seen_by` khi `last_read_sequence >= message.sequence`.

## Test end-to-end bằng PowerShell

Chạy migration và server trước:

```powershell
go run ./cmd/main.go --migrate:run
go run ./cmd/main.go
```

Giữ terminal server mở, sau đó mở terminal PowerShell khác.

### 1. Khai báo ID test

```powershell
$baseUrl = "http://localhost:8888"
$roleId = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
$user1Id = "11111111-1111-4111-8111-111111111111"
$user2Id = "22222222-2222-4222-8222-222222222222"

$headers1 = @{
  "Content-Type" = "application/json"
  "x-user-id" = $user1Id
  "x-user-role" = $roleId
}

$headers2 = @{
  "Content-Type" = "application/json"
  "x-user-id" = $user2Id
  "x-user-role" = $roleId
}
```

### 2. Insert hoặc update hai user

`PUT /api/users/me` là upsert: chưa có thì insert, đã có thì update.

```powershell
$user1Body = @{
  username = "hieu"
  name = "Hieu Nguyen"
  avatar_url = $null
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Put `
  -Uri "$baseUrl/api/users/me" `
  -Headers $headers1 `
  -Body $user1Body

$user2Body = @{
  username = "an"
  name = "An Tran"
  avatar_url = $null
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Put `
  -Uri "$baseUrl/api/users/me" `
  -Headers $headers2 `
  -Body $user2Body
```

Kiểm tra user:

```powershell
Invoke-RestMethod -Method Get -Uri "$baseUrl/api/users/me" -Headers $headers1
Invoke-RestMethod -Method Get -Uri "$baseUrl/api/users?search=an" -Headers $headers1
```

### 3. Tạo channel có hai thành viên

```powershell
$channelBody = @{
  name = "Backend team"
  member_ids = @($user2Id)
} | ConvertTo-Json -Depth 5

$channelResponse = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/api/channels" `
  -Headers $headers1 `
  -Body $channelBody

$channelId = $channelResponse.data.id
$channelId
```

User 1 là creator nên được thêm tự động; `member_ids` chỉ cần chứa user 2.

### 4. Gửi message vào MongoDB

```powershell
$messageBody = @{
  content = @{
    type = "text"
    text = "Hello from Hieu"
  }
} | ConvertTo-Json -Depth 5

Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/api/channels/$channelId/messages" `
  -Headers $headers1 `
  -Body $messageBody
```

### 5. User 2 đọc message

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/api/channels/$channelId/messages?limit=50" `
  -Headers $headers2
```

Trong pgAdmin bạn sẽ thấy hai user, một channel và hai channel member. Trong MongoDB Compass hoặc `mongosh`, collection `messages` sẽ có message vừa gửi.

## Kiểm tra

```powershell
go test ./...
go vet ./...
go build ./...
```

## License

MIT, xem `LICENSE`.
