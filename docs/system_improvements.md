# Đề xuất Nâng cấp Hệ thống Sàn Thương mại điện tử (EMC LB)

Tài liệu này ghi chú lại các tính năng và điểm cải tiến cốt lõi cần phải thực hiện trong các Phase tiếp theo để hệ thống Sàn TMĐT thực sự hoàn thiện và có khả năng chịu tải (high scalability).

## 1. Tự động Huỷ Đơn Hàng (Order Expiration / TTL)
* **Mô tả vấn đề:** Khách đặt hàng, hệ thống trừ kho thành công trên Redis, nhưng khách không thanh toán. Nếu không xử lý, số hàng đó sẽ bị "giam" vĩnh viễn, khiến Shop không thể bán cho người khác.
* **Giải pháp đề xuất:** 
  - Tích hợp một Background Job Worker (như `Asynq`) hoặc sử dụng Redis Keyspace Notifications.
  - Quét các đơn hàng ở trạng thái `Pending Payment` quá 15-30 phút -> Đổi thành `Cancelled`.
  - Kích hoạt lệnh `INCRBY` trên Redis và Update trên MongoDB để hoàn trả lại số lượng tồn kho cho Shop.

## 2. Caching Chiến lược (Read-Heavy Optimization)
* **Mô tả vấn đề:** Sàn TMĐT có đặc thù là lưu lượng truy cập (Traffic) tập trung vào việc đọc/xem dữ liệu (chiếm 90-95%) thay vì mua hàng. Nếu query trực tiếp vào MongoDB cho mọi request thì DB sẽ quá tải.
* **Giải pháp đề xuất:** 
  - Lưu bộ nhớ đệm (Cache) cho các API truy xuất dữ liệu lớn vào Redis: `GET Products`, `GET Categories`, `GET Shop Profile`.
  - Thiết lập cơ chế **Cache Invalidation**: Chỉ cập nhật lại hoặc xoá cache khi Shop sửa thông tin, giá cả, hoặc khi có đơn hàng thành công làm thay đổi tồn kho thực tế.

## 3. Hệ thống Tìm kiếm Toàn văn (Search Engine)
* **Mô tả vấn đề:** Người mua hàng luôn nhập từ khoá tìm kiếm. Nếu sử dụng `$regex` của MongoDB để dò tìm theo Tên Sản Phẩm, query sẽ rất tốn CPU và chậm chạp. Nó cũng không hỗ trợ tự sửa lỗi chính tả (Typo tolerance) cho người dùng.
* **Giải pháp đề xuất:** 
  - Đồng bộ hoá dữ liệu `Product` từ MongoDB sang một Text-Search Engine chuyên dụng như **Elasticsearch** hoặc **Meilisearch**.
  - Thực hiện các logic Query tìm kiếm, lọc theo giá, lọc theo danh mục, tính điểm phù hợp (ranking) trực tiếp trên Engine này.

## 4. Kiểm thử Tự động & CI/CD (Testing Framework)
* **Mô tả vấn đề:** Hệ thống Sàn TMĐT có rất nhiều nghiệp vụ liên quan đến Tiền bạc và Tồn kho. Bất kỳ một sửa đổi nhỏ nào về code cũng có thể gây ra bugs nghiêm trọng về doanh thu.
* **Giải pháp đề xuất:** 
  - Cài đặt Mocking framework (như `gomock`).
  - Viết Unit Test cho các service xử lý lõi: Tính toán chia tiền (Group Payment Split), Thuật toán giảm giá (Coupon), Xác minh chữ ký IPN (Payment Service).
  - Viết Integration Test & Load Test (`k6`) để đảm bảo logic trừ kho Redis LUA hoạt động đúng thiết kế khi chịu áp lực 1000 requests/giây.
