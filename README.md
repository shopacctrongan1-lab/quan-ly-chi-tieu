# Quản lí chi tiêu - Go + Vue

Ứng dụng web quản lí thu/chi cá nhân gồm:

- Backend Go thuần `net/http`, lưu dữ liệu JSON tại `data/expenses.json`.
- Frontend Vue 3 + Vite.
- Chức năng thêm, sửa, xóa, tìm kiếm, lọc theo loại/tháng, tổng hợp thu nhập - chi tiêu - số dư và biểu đồ danh mục.

## Chạy khi phát triển

### 1. Chạy backend

```powershell
go run ./cmd/server
```

Backend chạy tại `http://localhost:8080`.

### 2. Chạy frontend

Do PowerShell có thể chặn `npm.ps1`, dùng `npm.cmd`:

```powershell
cd frontend
npm.cmd install
npm.cmd run dev
```

Mở URL Vite hiển thị trên màn hình. Nếu cần gọi backend khác origin, tạo file `frontend/.env`:

```env
VITE_API_BASE=http://localhost:8080
```

## Build chạy production

```powershell
cd frontend
npm.cmd install
npm.cmd run build
cd ..
go run ./cmd/server
```

Sau khi build, Go sẽ phục vụ frontend từ `frontend/dist` tại `http://localhost:8080`.

## API chính

- `GET /api/expenses?month=YYYY-MM&type=expense&search=abc`
- `POST /api/expenses`
- `PUT /api/expenses/{id}`
- `DELETE /api/expenses/{id}`
- `GET /api/summary?month=YYYY-MM`
