package providers

import "context"

// Provider là interface chung mà mọi external API wrapper phải thỏa mãn.
//
// Triết lý thiết kế:
//   - Interface NHỎ (chỉ vài method) — Go idiom: "accept interfaces, return structs"
//   - Mọi method nhận context.Context đầu tiên — cancel/timeout luôn được support
//   - Trả về `any` không phải lý tưởng nhất, nhưng ở giai đoạn đầu giúp đơn giản hóa.
//     Sau tuần 5 khi học generics, ta sẽ refactor thành Provider[T].
//
// Bạn sẽ thấy "aha moment" ở Tuần 2 khi có 3-4 providers cùng implement interface này.
type Provider interface {
	// Name trả về định danh của provider (vd: "weather", "crypto").
	Name() string

	// Fetch gọi API ngoài, trả về kết quả đã parse.
	// params là map các tham số (vd: {"city": "Hanoi"}).
	Fetch(ctx context.Context, params map[string]string) (any, error)
}
