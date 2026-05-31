# Bài Tập Rebuild-From-Scratch

> Triết lý: bạn chỉ thực sự hiểu một đoạn code khi tự viết được nó **mà không nhìn**.
> File này có các bài tập "xóa và viết lại" theo từng tuần, kèm self-check questions.

## Hướng dẫn cách dùng

Cuối mỗi tuần, làm theo thứ tự:

1. **Pre-check questions**: trả lời bằng miệng/giấy, không Google. Nếu trả lời sai → đọc lại LEARNING_NOTES.md section liên quan.
2. **Rebuild challenge**: xóa file → tự viết lại từ đầu, chỉ nhìn comment/signature gốc.
3. **Diff & reflect**: so sánh với code gốc, ghi note lại điểm khác.
4. **Extension challenge**: bài tập mở rộng, không có đáp án sẵn.

Dùng `git stash` để backup code gốc trước khi xóa:
```bash
git stash push -m "backup before rebuild week1"
# ... làm bài tập ...
git stash show -p  # so sánh sau khi xong
git stash pop      # restore nếu cần
```

---

## Tuần 1: Foundations

### Pre-check questions

Trả lời từng câu trước khi đọc tiếp:

1. Giải thích `defer resp.Body.Close()` — vì sao cần, nếu thiếu thì sao?
2. Khi nào dùng `var x int = 5` vs `x := 5`? Có lúc nào BẮT BUỘC dùng `var`?
3. Khác biệt giữa `func f(p Person)` và `func f(p *Person)` khi gọi với một biến `person Person`?
4. `json.Unmarshal` cần `*T` hay `T`? Vì sao?
5. `fmt.Errorf("op: %w", err)` khác `fmt.Errorf("op: %v", err)` ở chỗ nào?
6. Slice có thể nil không? Nil slice append được không?

<details>
<summary>Đáp án</summary>

1. `defer` schedule cleanup khi func return. `Body` là `io.ReadCloser`, không close → connection không trả về pool, leak. Sau ~hàng ngàn request không close → app crash vì hết file descriptor.
2. `:=` chỉ trong function. Package-level vars phải dùng `var`. Cũng phải dùng `var` khi muốn explicit zero value: `var x int` (= 0).
3. `f(p Person)` copy toàn bộ struct, modify trong f không ảnh hưởng caller. `f(p *Person)` truyền pointer, modify ảnh hưởng caller. Call site: `f(person)` cho value, `f(&person)` cho pointer.
4. `*T`. Unmarshal cần modify struct → cần pointer. Truyền value → Go copy, modify trên copy, caller không thấy.
5. `%w` wrap error preserve identity (`errors.Is` work được). `%v` chỉ in message (mất khả năng check loại error).
6. Có. Nil slice có thể append (Go tự alloc): `var s []int; s = append(s, 1)` — OK.

</details>

### Rebuild challenge

**Mục tiêu**: tự viết lại `cmd/server/main.go` với graceful shutdown.

```bash
git stash push -m "backup main.go"
echo 'package main; func main() {}' > cmd/server/main.go
```

Yêu cầu (KHÔNG nhìn file gốc):
- [ ] Load `.env` file (silent fail nếu không có)
- [ ] Setup `slog` với JSON handler
- [ ] HTTP server với 4 timeouts khác nhau
- [ ] Endpoint `/health` trả `{"status":"ok"}`
- [ ] Graceful shutdown khi nhận SIGINT/SIGTERM, timeout 10s

Khi xong, so sánh:
```bash
git stash show -p stash@{0} -- cmd/server/main.go
```

Note lại 3 thứ bạn viết khác và lý do.

### Extension challenge

Thêm vào server:
- Endpoint `/version` trả version từ build flag: `go build -ldflags="-X main.version=v0.1.0"`
- Middleware đếm số request đang in-flight, hiện ở `/metrics`
- Health check kiểm tra connectivity đến Open-Meteo API (return 503 nếu fail)

---

## Tuần 2: Architecture

### Pre-check questions

1. Interface trong Go có cần `implements` keyword không? Cách Go biết struct X implement interface Y?
2. Nếu interface có 1 method `Read() string`, và bạn có struct `func (f *File) Read() string` — `*File` hay `File` thỏa interface?
3. Vì sao Go ưa interface nhỏ (1-3 methods)?
4. `internal/` folder có ý nghĩa gì với Go compiler?
5. Khác biệt `pkg/` và `internal/`?
6. Dependency injection trong Go làm thế nào không cần framework như Spring?

<details>
<summary>Đáp án</summary>

1. Không. Implicit satisfaction — bất kỳ type nào có đủ method signatures là thỏa. Compiler check khi assign.
2. `*File` thỏa. Method với pointer receiver chỉ available trên pointer. (Method với value receiver thì cả 2 đều OK.)
3. Composability. Interface nhỏ dễ implement, dễ mock test, dễ combine: `io.ReadWriter = io.Reader + io.Writer`. Interface lớn = bloat = ai cũng phải implement đầy đủ.
4. Package trong `internal/X/...` chỉ import được bởi code trong cùng module + path "above" internal. Ngăn external dùng API nội bộ.
5. `pkg/`: public reusable. `internal/`: private, không cho external import. Không phải convention bắt buộc nhưng phổ biến.
6. Constructor injection: `New(deps...)` nhận dependencies, return struct. Main wire mọi thứ ở `main.go`. Không cần annotation, không cần container.

</details>

### Rebuild challenge

**Mục tiêu**: refactor weather provider thành full interface-driven.

Bước 1: viết lại `internal/providers/provider.go` interface (đừng nhìn gốc).

Bước 2: tạo `MockProvider` cho test:
```go
// internal/providers/mock.go
type MockProvider struct {
    NameFunc  func() string
    FetchFunc func(ctx context.Context, params map[string]string) (any, error)
}

func (m *MockProvider) Name() string { return m.NameFunc() }
func (m *MockProvider) Fetch(...) (...) { return m.FetchFunc(...) }
```

Bước 3: viết test cho `aggregator.Service.FetchAll` dùng MockProvider — không gọi API thật.

Bước 4: refactor `weather.Provider` để dùng `interface HTTPClient` thay vì `*resty.Client` cụ thể:
```go
type HTTPClient interface {
    R() *resty.Request
}
```
Vì sao? Để mock HTTP cho test. (Hint: bạn sẽ thấy đây là **bad abstraction** — resty không thực sự mockable qua interface này. Đây là bài học về "premature interface".)

### Extension challenge

- Thêm provider thứ 5: IP Geolocation (https://ip-api.com)
- Provider interface cho phép `Healthcheck() error` — implement cho mọi providers
- Endpoint `/v1/providers/status` ping mọi provider, trả tình trạng

### Self-reflection

Sau khi xong, trả lời:
- Bạn có thấy lợi ích của interface khi có 4 providers chưa? Cụ thể chỗ nào?
- Nếu chỉ có 1 provider mãi mãi → có cần interface không? (Câu trả lời: thường là không. YAGNI principle.)

---

## Tuần 3: Concurrency

### Pre-check questions

1. Goroutine khác thread OS ở chỗ nào? Một process Go có thể có bao nhiêu goroutine?
2. `chan int` buffered 5 — khi nào send block? Khi nào receive block?
3. `select` statement với 3 case ready cùng lúc — case nào được chọn?
4. `context.WithTimeout` vs `context.WithDeadline` khác gì?
5. Vì sao goroutine không có "ID" hoặc cách kill từ ngoài?
6. Race condition là gì? Phát hiện thế nào trong Go?
7. `sync.WaitGroup` vs `errgroup` — khi nào dùng cái nào?

<details>
<summary>Đáp án</summary>

1. Goroutine: ~2KB stack ban đầu, multiplex lên OS thread bởi Go scheduler. Thread OS: ~1MB stack, kernel-managed. 1 process Go có thể spawn hàng triệu goroutine (vs hàng nghìn thread).
2. Buffered chan capacity 5: send block khi buffer đầy (>5 chưa receive). Receive block khi buffer rỗng.
3. **Random**. Go intentionally randomize để tránh dependency vào order. Đừng rely vào order.
4. `WithTimeout(ctx, 5*time.Second)` = `WithDeadline(ctx, time.Now().Add(5s))`. WithTimeout là syntactic sugar.
5. Triết lý Go: goroutine không "kill" được từ ngoài. Communicate qua channel/context để goroutine **tự exit**. Force kill là design smell.
6. Race condition = 2 goroutine truy cập cùng memory, ít nhất 1 write, không sync. Phát hiện: `go test -race`, `go run -race main.go`.
7. `WaitGroup`: chỉ đợi, không handle error. `errgroup`: đợi + collect first error + cancel context. Dùng errgroup khi có lỗi cần propagate.

</details>

### Rebuild challenge

**Mục tiêu**: implement `aggregator.Service.FetchAll` từ đầu.

Bước 1: viết với `sync.WaitGroup` thuần (KHÔNG dùng errgroup):
- Spawn 1 goroutine per provider
- Mỗi goroutine ghi result vào `chan Result`
- Main goroutine collect từ channel
- Timeout 3s tổng (nếu quá → trả về kết quả partial)

Bước 2: refactor sang `errgroup`. So sánh code dài/ngắn, error handling khác chỗ nào.

Bước 3: thêm test với race detector:
```bash
go test -race ./internal/aggregator/
```

Đảm bảo test pass cả khi race detector on.

Bước 4: simulate slow provider:
```go
type slowProvider struct{}
func (s *slowProvider) Fetch(ctx context.Context, params map[string]string) (any, error) {
    select {
    case <-time.After(5 * time.Second):
        return "done", nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

Test: `FetchAll` với timeout 2s gặp slow provider → return trong 2s với error cho slow one, results cho providers nhanh.

### Extension challenge

Pipeline pattern: build "weather monitor"
- Goroutine 1: every 10s, fetch weather for 5 cities (concurrent)
- Goroutine 2: receive weather data, check if temperature changed >5°C from previous
- Goroutine 3: receive alerts, log to console
- All connected by channels
- Graceful shutdown qua context

Đây là pattern thật trong production: data pipeline với backpressure.

### Self-check questions sau Tuần 3

Sau khi xong, bạn nên hiểu:
- Tại sao `for _, x := range slice { go func() { use(x) }() }` là buggy? (Go <1.22)
- Khi nào dùng buffered channel?
- "Context cancellation" propagate xuống nested function thế nào?
- Tại sao `defer cancel()` quan trọng với mọi `context.WithX`?

---

## Tuần 4: Database + Auth

### Pre-check questions

1. Connection pool là gì? Vì sao không tạo connection mới mỗi request?
2. `pgx` vs `database/sql` khác gì? Vì sao recommend pgx?
3. Prepared statement chống được loại attack nào?
4. JWT có 3 phần: header, payload, signature. Phần nào encode (Base64), phần nào hash, phần nào ký?
5. Vì sao access token nên short-lived (15 phút)?
6. Refresh token rotation là gì?
7. Stateful vs stateless authentication?

<details>
<summary>Đáp án</summary>

1. Pool = pre-created connections, reuse. Mỗi connection setup ~3-way TCP handshake + auth → ~10-50ms. Pool tránh overhead này.
2. `database/sql` là interface chung. `pgx` native Postgres driver, faster, hỗ trợ feature riêng (LISTEN/NOTIFY, COPY, array types). `pgx` có thể dùng qua `database/sql` interface hoặc native API.
3. SQL injection. Placeholder values KHÔNG được parse là SQL — engine treat như literal data.
4. Header + payload: Base64 encode (không encrypt, ai cũng đọc được). Signature: HMAC sign over (header + payload + secret). JWT không hash payload — chỉ ký.
5. Nếu leak (XSS, MITM, log...), attacker chỉ có 15 phút. Compromise tối thiểu. Long-lived → hậu quả lâu dài.
6. Mỗi lần dùng refresh token → revoke cũ, cấp mới. Detect token theft: nếu cả attacker và user đều dùng cùng refresh token → 1 trong 2 sẽ thấy "invalid" → force logout cả hai → safer.
7. Stateful: server lưu session ID, check DB mỗi request. Stateless (JWT): không lưu state, verify chữ ký tại chỗ. Trade-off: stateful revoke ngay được, stateless phải đợi expire.

</details>

### Rebuild challenge

**Mục tiêu**: viết lại JWT manager + auth middleware không nhìn code gốc.

Bước 1: xóa `internal/auth/jwt.go`, tự viết:
- Struct `Claims` embed `jwt.RegisteredClaims`
- `GenerateAccessToken(userID, email) (string, error)`
- `Verify(token string) (*Claims, error)` — check signing method, handle expired error đúng

Bước 2: xóa `internal/middleware/auth.go`, tự viết:
- Middleware extract Bearer token
- Verify, inject userID vào context
- Helper `UserIDFromContext`

Bước 3: viết test cho cả hai:
```go
func TestJWTManager_GenerateAndVerify(t *testing.T) {
    m := NewJWTManager("secret", 15*time.Minute, 30*24*time.Hour)
    token, err := m.GenerateAccessToken(42, "test@example.com")
    require.NoError(t, err)

    claims, err := m.Verify(token)
    require.NoError(t, err)
    assert.Equal(t, int64(42), claims.UserID)
}

func TestJWTManager_RejectsExpired(t *testing.T) {
    // tạo manager với TTL âm để generate token đã expire
}

func TestJWTManager_RejectsWrongSecret(t *testing.T) {
    // sign bằng secret A, verify bằng secret B
}

func TestJWTManager_RejectsAlgNone(t *testing.T) {
    // craft token với alg="none", verify phải fail
}
```

### Extension challenge

- Password strength validator: reject password trong top 10k common passwords (download list)
- Account lockout: 5 lần login fail trong 15 phút → lock 30 phút
- Email verification: register → gửi link verify → user click → activate (fake email, log link ra console)
- Password reset flow đầy đủ với token có TTL

### Self-check questions

- Token leak qua URL query (`?token=xxx`) vs header — chỗ nào nguy hiểm hơn? Vì sao? (URL: log vào access log, gửi trong Referer header, browser history)
- Có thể dùng JWT làm session ID không? Trade-off? (Có. Lợi: stateless. Hại: revoke khó, payload to hơn cookie)

---

## Tuần 5: Cache, Rate Limit, Generics

### Pre-check questions

1. Generics trong Go 1.18+ giải quyết vấn đề gì so với `interface{}`/`any`?
2. Type parameter constraint là gì? `comparable` constraint cho phép gì?
3. Token bucket algorithm hoạt động thế nào?
4. Cache stampede là gì? Cách prevent?
5. Khi nào dùng cache, khi nào không?

<details>
<summary>Đáp án</summary>

1. Type safety + performance. `interface{}` mất type info, cần type assertion runtime. Generics giữ type, compile-time check.
2. Restrict tập hợp type T có thể nhận. `comparable` = type hỗ trợ `==`, `!=`. Cho phép dùng làm map key. `~int` = bất kỳ type underlying là int.
3. Bucket capacity N tokens, refill rate R tokens/s. Mỗi request consume 1 token. Hết token → reject. Cho phép burst N requests, sustained rate R.
4. Hàng nghìn request hit cache miss cùng lúc → đồng loạt query backend. Fix: singleflight (gộp duplicate requests), hoặc probabilistic early refresh.
5. Cache khi: data read-heavy, stale-OK ngắn hạn, query expensive. KHÔNG cache khi: data realtime critical, write-heavy, security-sensitive per-user.

</details>

### Rebuild challenge

Bước 1: refactor `MemoryCache` thành generic.

Trước:
```go
type MemoryCache struct { items map[string]item }
type item struct { value []byte; expireAt time.Time }
```

Sau:
```go
type MemoryCache[V any] struct {
    items map[string]item[V]
    mu    sync.RWMutex
}
type item[V any] struct {
    value    V
    expireAt time.Time
}

func NewMemory[V any]() *MemoryCache[V] { ... }
```

Usage:
```go
weatherCache := cache.NewMemory[weather.Response]()
weatherCache.Set(ctx, "Hanoi", resp, 5*time.Minute)
resp, ok := weatherCache.Get(ctx, "Hanoi")  // type-safe!
```

Bước 2: implement singleflight pattern manually:
```go
type Loader[V any] struct {
    cache    *MemoryCache[V]
    inflight map[string]chan V  // key → channel chờ result
    mu       sync.Mutex
}

func (l *Loader[V]) LoadOrFetch(ctx context.Context, key string, fetch func() (V, error)) (V, error) {
    // 1. Check cache, hit → return
    // 2. Check inflight, có → wait on channel
    // 3. Không có → start fetch, register inflight, broadcast khi xong
}
```

Đây là pattern phức tạp — đáng để gãy đầu tự nghĩ. Có sẵn ở `golang.org/x/sync/singleflight` nhưng tự build dạy bạn rất nhiều về concurrency.

Bước 3: rate limit middleware
```go
func RateLimit(rps float64, burst int) func(http.Handler) http.Handler {
    // Lấy IP/userID làm key
    // Mỗi key 1 token bucket
    // map[string]*rate.Limiter — nhớ cleanup expired entries
}
```

### Extension challenge

- LRU eviction: cache có max size, kick item old khi đầy
- Cache với "stale-while-revalidate": trả stale data nếu fetch background đang chạy
- Distributed rate limit qua Redis (nhiều instance share state)

---

## Tuần 6: Testing & Deploy

### Pre-check questions

1. Table-driven test là gì? Vì sao là idiom Go?
2. `t.Helper()` để làm gì?
3. `httptest.Server` vs `httptest.NewRecorder()` — khi nào dùng cái nào?
4. Mock vs Fake vs Stub khác gì?
5. Integration test với Postgres thật vs với in-memory SQLite — trade-off?
6. Build cache trong Docker hoạt động thế nào?

<details>
<summary>Đáp án</summary>

1. Tests viết dưới dạng table (slice of structs) với input + expected, loop chạy từng case. Dễ thêm case, dễ đọc.
2. Mark function là test helper. Khi assertion fail trong helper, line reported là caller (test thật), không phải helper. Giúp error message useful hơn.
3. `Server`: spin up real HTTP server, test HTTP client. `Recorder`: record response, test handler không cần server.
4. Stub: hardcoded response. Fake: working implementation simplified (in-memory DB). Mock: verify interactions (was method X called with arg Y).
5. Postgres thật: realistic, slower setup. SQLite: fast, không match prod (different SQL dialect, không có row-level lock như Postgres).
6. Layer cache: nếu instructions + files trong layer không đổi, Docker reuse cached layer. Đặt instructions ít thay đổi lên trên (vd: install OS deps trước, copy source code cuối).

</details>

### Rebuild challenge

Bước 1: convert tất cả test cũ sang table-driven:
```go
func TestPasswordHashing(t *testing.T) {
    tests := []struct {
        name     string
        password string
        wantErr  bool
    }{
        {"valid password", "MySecure!Pass123", false},
        {"empty password", "", true},
        {"too long", strings.Repeat("a", 73), true}, // bcrypt limit
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := HashPassword(tt.password)
            if (err != nil) != tt.wantErr {
                t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

Bước 2: viết integration test với testcontainers:
```go
func TestUserRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    ctx := context.Background()
    pgContainer, err := postgres.RunContainer(ctx, ...)
    require.NoError(t, err)
    defer pgContainer.Terminate(ctx)

    connStr, _ := pgContainer.ConnectionString(ctx)
    // run migrations
    // run actual test against real Postgres
}
```

Bước 3: viết end-to-end test:
- Register → Login → Call /v1/weather với token
- Assert toàn bộ flow trong 1 test

Bước 4: setup GitHub Actions CI chạy `make check` mỗi PR.

### Extension challenge

- Coverage badge trong README (codecov hoặc tự render)
- Benchmark suite: so sánh in-memory vs Redis cache performance
- Fuzz test cho JWT parsing: `go test -fuzz=FuzzJWTParse`
- Load test với `vegeta` hoặc `k6` — 100 req/s trong 1 phút, đo p50/p95/p99 latency

---

## Self-Assessment Tổng

Sau khi xong tất cả 6 tuần, đánh giá bản thân theo thang điểm 1-5 từng skill:

| Skill | Có thể giải thích | Có thể viết từ đầu | Score |
|---|---|---|---|
| Module system, package layout | _ | _ | _/5 |
| Struct, interface, embedding | _ | _ | _/5 |
| Error handling, wrapping | _ | _ | _/5 |
| HTTP server + middleware | _ | _ | _/5 |
| Goroutine, channel, select | _ | _ | _/5 |
| Context propagation | _ | _ | _/5 |
| errgroup, sync primitives | _ | _ | _/5 |
| JWT auth, password hashing | _ | _ | _/5 |
| Database với pgx + sqlc | _ | _ | _/5 |
| Cache patterns, mutex | _ | _ | _/5 |
| Generics | _ | _ | _/5 |
| Table-driven tests | _ | _ | _/5 |
| Mocking via interfaces | _ | _ | _/5 |
| Docker multi-stage build | _ | _ | _/5 |
| Graceful shutdown | _ | _ | _/5 |

Mục tiêu: 4/5 trở lên trên tất cả. Dưới 3 → quay lại tuần đó.

## Cuối cùng

Học Go không phải đua tốc độ. Tôi recommend mỗi tuần 5-7 ngày dù chỉ 1-2h/ngày. Quan trọng là **suy ngẫm sau khi code** — đọc lại sau 1 ngày, hỏi "viết lại có gọn hơn được không?". Đó là cách idiom Go thấm dần.

Một trick: viết blog/note mỗi tuần "Tuần X tôi học được gì". Sau 6 tuần bạn có 6 bài note + 1 project + nhiều câu trả lời tự viết → portfolio đẹp hơn nhiều CV bullet point.
