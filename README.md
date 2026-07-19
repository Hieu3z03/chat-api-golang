# Chat API Golang

Chat service viết bằng Gin theo cấu trúc controller - service - repository. Service chạy độc lập, dùng PostgreSQL cho dữ liệu quan hệ và MongoDB cho message.

## Identity contract

Chat service không xác thực JWT và không đọc header `Authorization`. Upstream gateway/auth service truyền hai header tin cậy:

```text
x-user-id: <user UUID>
x-user-role: <role UUID>
```

Middleware chỉ kiểm tra hai giá trị là UUID hợp lệ rồi đưa vào request context. `x-user-role` chưa được lưu vào database; nó được giữ trong context cho các rule nghiệp vụ sau này.

## Data model

PostgreSQL:

- `users`: `id`, `first_name`, `last_name`, `username`, `avatar_id`, timestamps
- `channels`: `id`, `name`, `created_by`, timestamps
- `channel_members`: `channel_id`, `user_id`, `joined_at`

MongoDB collection `messages`:

- `_id`
- `channel_id`
- `user_id`
- `content`: JSON object bắt buộc có `type`
- `created_at`

Chat service chỉ lưu `avatar_id`; metadata/file avatar thuộc file service.

## Cách hệ thống hoạt động

1. Auth/gateway xác định người gọi rồi chuyển `user UUID` và `role UUID` qua hai header.
2. Chat service chỉ đọc header; không kiểm tra token và không lưu password.
3. Hồ sơ user được đồng bộ vào PostgreSQL qua `PUT /api/users/me`.
4. Khi tạo channel, service lấy creator từ `x-user-id`, kiểm tra các user đã tồn tại local rồi ghi `channels` và `channel_members` trong một transaction PostgreSQL.
5. Khi gửi/đọc message, service kiểm tra `channel_members` trong PostgreSQL trước, sau đó ghi/đọc collection `messages` trong MongoDB.
6. `content` của message linh hoạt theo `type`; chat service chỉ giữ ID của tài nguyên thuộc service khác.

## Cấu hình local

Sao chép `.env.example` thành `.env` và nhập thông tin database thật:

```env
APP_NAME=chat-api
GOLANG_PORT=8888

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=your_postgres_password
DB_NAME=chat_api
DB_SSLMODE=disable

MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=chat_api
```

Tạo database PostgreSQL `chat_api` trước bằng pgAdmin. Ứng dụng yêu cầu cả PostgreSQL và MongoDB truy cập được khi khởi động.

```powershell
go run ./cmd/main.go --migrate:run
go run ./cmd/main.go --seed
go run ./cmd/main.go
```

Server mặc định chạy tại `http://localhost:8888`.

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
  "first_name": "Hieu",
  "last_name": "Nguyen",
  "username": "hieu",
  "avatar_id": null
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
  first_name = "Hieu"
  last_name = "Nguyen"
  username = "hieu"
  avatar_id = $null
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Put `
  -Uri "$baseUrl/api/users/me" `
  -Headers $headers1 `
  -Body $user1Body

$user2Body = @{
  first_name = "An"
  last_name = "Tran"
  username = "an"
  avatar_id = $null
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
