# Learning Notes — Giải thích code đã viết sẵn

> Đọc file này SAU khi đã skim qua code, hoặc khi gặp đoạn chưa hiểu.
> Mỗi section giải thích "tại sao viết như vậy" — đây là chỗ thấm idiom Go.

---

## 1. `cmd/server/main.go` — Graceful Shutdown

### Câu hỏi tôi đã trả lời khi viết file này

**Q: Tại sao dùng `signal.NotifyContext` chứ không phải `signal.Notify` cũ?**

`signal.Notify(chan)` là API cũ:
```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
<-sigCh  // block đợi signal
```

`signal.NotifyContext` (Go 1.16+) tích hợp với context — cancel context khi nhận signal:
```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
<-ctx.Done()  // block, nhưng giờ ctx propagate được xuống các goroutine khác
```

Lợi: bạn có thể pass `ctx` này xuống các worker, database connection pool, etc. Khi user Ctrl+C, mọi thứ cancel cùng lúc.

**Q: Tại sao start server trong goroutine?**

`srv.ListenAndServe()` block forever cho đến khi server stopped. Nếu gọi trực tiếp trong `main()`, code sau nó (xử lý shutdown) sẽ không bao giờ chạy.

Pattern chuẩn:
```go
go func() { srv.ListenAndServe() }()  // chạy nền
<-ctx.Done()                           // main đợi signal
srv.Shutdown(...)                      // shutdown sau khi nhận signal
```

**Q: Tại sao `srv.Shutdown` cần context riêng với timeout?**

`Shutdown` đợi tất cả request đang xử lý xong rồi mới đóng. Nếu có request "treo" 5 phút, bạn không muốn đợi 5 phút. Timeout 10s là balance hợp lý: cho request đang chạy thời gian finish, nhưng không đợi quá lâu.

**Q: Vì sao check `errors.Is(err, http.ErrServerClosed)`?**

Khi gọi `Shutdown()`, `ListenAndServe()` trả về error `http.ErrServerClosed` — đây không phải lỗi thật, chỉ là signal "server đã đóng bình thường". Phải filter ra để log không spam giả error.

**Q: Vì sao 4 timeout khác nhau cho server?**

```go
ReadHeaderTimeout: 5s    // đọc HTTP headers
ReadTimeout:       10s   // đọc cả body
WriteTimeout:      15s   // ghi response
IdleTimeout:       60s   // giữ keep-alive connection
```

Mỗi timeout chống một kiểu attack/lỗi khác nhau. `ReadHeaderTimeout` đặc biệt quan trọng — chống Slowloris attack (gửi headers cực chậm để chiếm connection).

### Self-check
- Bỏ goroutine wrapper quanh `ListenAndServe()` → code sẽ làm gì? (Trả lời: hang ở `ListenAndServe`, không bao giờ chạy phần shutdown)
- Bỏ timeout context của `Shutdown()` → vấn đề gì? (Treo vô hạn nếu có request slow)

---

## 2. `internal/config/config.go` — Envconfig pattern

**Q: Tại sao không dùng Viper?**

Viper mạnh nhưng nặng (~30 transitive deps). Cho project size này, `envconfig` (~0 deps) là đủ. Quy tắc Go: **chọn package nhỏ nhất đủ dùng**.

**Q: Struct tags `envconfig:"HTTP_PORT" default:"8080"` hoạt động thế nào?**

Go có reflection. `envconfig` đọc tags lúc runtime, map field name với env var, apply default nếu env var rỗng. Đây là pattern phổ biến trong Go: dùng struct tags làm "config-by-convention" — JSON cũng dùng pattern này (`json:"field_name"`).

**Q: Vì sao một số field `required:"true"`, một số `default:"..."`?**

Có gì không thể có default an toàn (JWT_SECRET, DATABASE_URL) → required. Có gì optional có thể fallback → default.

`required:"true"` mà thiếu → `envconfig.Process` trả error. App fail-fast at startup thay vì crash giữa chừng. Đây là idiom quan trọng: **fail at startup, not at request time**.

### Self-check
- Tại sao `LogLevel` là `string` không phải `slog.Level`? (Vì envconfig không tự convert; bạn parse string thành Level sau)
- Thêm field mới `MaxConnections int` cần làm gì? (Thêm field với tag, thế thôi — không config thêm gì)

---

## 3. `internal/domain/errors.go` — Sentinel Errors

**Q: Vì sao có folder `domain` riêng?**

Đây là **Clean Architecture**. Layer `domain` chứa types + errors thuần — không import HTTP, không import database, không import bất cứ gì ngoài stdlib. Mọi layer khác phụ thuộc vào `domain`, không ngược lại.

Lợi: bạn có thể test domain logic mà không cần DB, không cần HTTP server. Và khi swap PostgreSQL → MySQL, domain không đổi.

**Q: Vì sao dùng `errors.New("...")` không phải `errors.New(...)` với mã lỗi?**

Pattern Go: errors là **values**, không phải enum/code. So sánh bằng `errors.Is(err, ErrNotFound)` — không phải `err.Code == 404`.

Khi nào dùng custom struct error? Khi cần kèm metadata:
```go
type ValidationError struct {
    Field string
    Value any
    Rule  string
}
func (e *ValidationError) Error() string { ... }
```

Nhưng cho domain errors generic (NotFound, Unauthorized...) → sentinel error đơn giản hơn.

**Q: `errors.Is` vs `errors.As` khác gì?**

- `errors.Is(err, target)`: "err có **phải là** target không?" (so sánh sentinel)
- `errors.As(err, &target)`: "err có thể **gán vào** target không?" (typed error extraction)

```go
if errors.Is(err, ErrNotFound) { ... }  // check loại

var validationErr *ValidationError
if errors.As(err, &validationErr) {
    fmt.Println(validationErr.Field)  // truy cập metadata
}
```

### Self-check
- Tại sao biến error đặt tên bắt đầu bằng `Err` (vd `ErrNotFound`)? (Go convention)
- Khi nào nên tạo custom error struct thay vì sentinel? (Khi cần kèm data ngoài message)

---

## 4. `internal/auth/jwt.go` — JWT Manager

**Q: Vì sao `Claims` embed `jwt.RegisteredClaims`?**

```go
type Claims struct {
    UserID int64
    Email  string
    jwt.RegisteredClaims  // ← embedded, không có field name
}
```

Đây là **struct embedding** — composition của Go (thay cho inheritance). `Claims` "kế thừa" tất cả field và method của `RegisteredClaims`. Bạn có thể truy cập `claims.ExpiresAt` trực tiếp.

`RegisteredClaims` đã có các field chuẩn theo RFC 7519: `iss`, `exp`, `iat`, `sub`, `aud`. Bạn không cần tự define lại.

**Q: Vì sao secret là `[]byte` không phải `string`?**

JWT library cần raw bytes để HMAC. String và []byte trong Go khác nhau (string immutable). Convert một lần ở constructor → reuse mãi → tránh allocation mỗi request.

**Q: Đoạn này quan trọng — vì sao check signing method?**

```go
token, err := jwt.ParseWithClaims(..., func(t *jwt.Token) (any, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return m.secret, nil
})
```

Đây là **defense chống "alg=none" attack**: attacker tạo JWT với header `{"alg": "none"}`, không cần biết secret. Nếu code không check method type, library sẽ "accept" token đó.

Đây là loại bug bảo mật cực phổ biến trong các implementation tự viết. Library hiện tại đã có protection, nhưng explicit check là best practice.

**Q: Vì sao trả về `*Claims` không phải `Claims`?**

Nếu trả `Claims` (value), khi func return, Go phải copy toàn bộ struct. `*Claims` chỉ copy pointer (8 bytes). Với struct lớn → đáng kể.

Rule of thumb: struct >3 fields hoặc có pointer trong đó → return pointer.

### Self-check
- Tại sao access token TTL 15 phút, refresh token TTL 30 ngày? (Bảo mật: nếu access token leak, hết hạn nhanh; refresh token lưu trữ an toàn hơn)
- Có cần lưu access token vào DB không? Vì sao? (Không — stateless. DB chỉ lưu refresh token để revoke được)

---

## 5. `internal/auth/password.go` — Bcrypt

**Q: Vì sao cost 12 không phải 10?**

Bcrypt cost là số mũ — cost 10 = 2^10 iterations, cost 12 = 2^12 = 4x lâu hơn. Năm 2024 trên CPU thường:
- Cost 10: ~60ms
- Cost 12: ~250ms

Cost cao hơn → hash chậm hơn → brute-force tốn hơn. Nhưng nếu quá chậm → user khó chịu khi login. 12 là sweet spot hiện tại. OWASP recommend cost ≥ 10.

**Q: Vì sao không lưu salt riêng?**

Bcrypt tự sinh salt 16-byte mỗi lần hash, EMBED salt vào output:
```
$2a$12$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
       ^^                                    ^^
       cost                                  hash + salt encoded
```

Khi verify, bcrypt parse salt từ hash string. Bạn không bao giờ phải xử lý salt.

**Q: Tại sao `CompareHashAndPassword` không trả `bool`?**

```go
func VerifyPassword(password, hash string) error
```

Trả `error` thay vì `bool` vì có thể có nhiều loại lỗi: mismatch, malformed hash, etc. Bcrypt cũng implement **constant-time compare** — luôn mất cùng thời gian dù mismatch ở byte đầu hay byte cuối, chống timing attack.

### Self-check
- Tại sao **không** dùng SHA256 cho password? (Quá nhanh → brute-force dễ với GPU)
- Argon2 vs Bcrypt? (Argon2 mới hơn, tốt hơn về memory-hard. Nhưng bcrypt đủ dùng cho 99% case và simpler)

---

## 6. `internal/middleware/auth.go` — Middleware Pattern

**Q: Pattern `func(http.Handler) http.Handler` là gì?**

Đây là **decorator pattern** dạng Go:
```go
func Auth(...) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w, r) {
            // logic trước
            next.ServeHTTP(w, r)  // gọi handler tiếp theo
            // logic sau (nếu cần)
        })
    }
}
```

- Outer func: nhận dependencies (`jwtManager`), trả về middleware
- Middle func: nhận `next` handler, trả về handler mới bọc `next`
- Inner func: handler thật, có logic check JWT

Chain middleware:
```go
r.Use(Logger)
r.Use(RateLimit)
r.Use(Auth(jwtManager))
```

Order matter: Logger ngoài cùng (log mọi request), Auth trong cùng (chỉ chạy nếu pass các middleware trước).

**Q: Vì sao `contextKey` là type riêng?**

```go
type contextKey string
const userIDKey contextKey = "user_id"
```

Nếu dùng raw string `"user_id"` làm key, có thể conflict với package khác cũng dùng `"user_id"`. Type alias đảm bảo key unique theo type — package khác có `type myKey string; const userIDKey myKey = "user_id"` thì 2 key là khác nhau dù string giống.

Đây là idiom trong stdlib (xem `context` package docs).

**Q: Vì sao `UserIDFromContext` trả `(int64, bool)`?**

`ctx.Value(key)` trả `any`. Type assertion `v.(int64)` có thể fail. Go convention: trả tuple `(value, ok)` để caller check:

```go
userID, ok := UserIDFromContext(ctx)
if !ok {
    // handle missing
}
```

Pattern này gọi là **comma-ok idiom**, xuất hiện ở map access, type assertion, channel receive.

### Self-check
- Bỏ middleware Auth khỏi route → request không có token nhưng vẫn chạy được? (Đúng — middleware mới block)
- Có thể inject thêm field nào vào context? (Bất kỳ field nào của Claims, hoặc thông tin tracing như request ID)

---

## 7. `internal/cache/cache.go` — Mutex Patterns

**Q: `sync.RWMutex` vs `sync.Mutex`?**

- `Mutex`: chỉ 1 goroutine truy cập tại 1 thời điểm (read hoặc write).
- `RWMutex`: nhiều goroutine **read đồng thời**, nhưng write phải độc quyền.

Cache có pattern read-heavy (nhiều GET, ít SET) → RWMutex tối ưu hơn.

```go
c.mu.RLock()      // nhiều goroutine có thể RLock cùng lúc
defer c.mu.RUnlock()
```

vs

```go
c.mu.Lock()       // độc quyền
defer c.mu.Unlock()
```

**Q: Khi nào dùng channel, khi nào dùng mutex?**

Châm ngôn Go: **"Don't communicate by sharing memory; share memory by communicating"** (dùng channel) — nhưng đây là guidance, không phải dogma.

Rule thực tế:
- **State đơn giản** (counter, map, cache): dùng mutex. Đơn giản hơn, nhanh hơn.
- **Coordination giữa goroutines** (worker pool, pipeline): dùng channel.
- **Pass ownership của data**: dùng channel.

Cache là state đơn giản → mutex.

**Q: Vì sao goroutine `cleanup()` chạy mãi mãi?**

```go
go c.cleanup()  // không có stop signal!
```

Đây là **bug nhẹ trong code starter** mà tôi cố ý để bạn fix ở tuần 5. Production code cần:
```go
type MemoryCache struct {
    // ...
    done chan struct{}
}

func (c *MemoryCache) Close() {
    close(c.done)
}

func (c *MemoryCache) cleanup() {
    for {
        select {
        case <-ticker.C:
            // cleanup
        case <-c.done:
            return  // exit goroutine
        }
    }
}
```

Goroutine không tự exit khi reference đến cache bị garbage collect — nó giữ reference tới cache → cache không bao giờ GC. Đây là kiểu **goroutine leak** rất phổ biến.

### Self-check
- Hai goroutine cùng `c.mu.RLock()` — có block nhau không? (Không)
- Một goroutine `RLock()`, một goroutine `Lock()` — có block không? (Có — Lock đợi RLock release)
- Map raw (`map[K]V`) không có mutex, 2 goroutine cùng ghi → chuyện gì xảy ra? (Panic: "concurrent map writes" — Go runtime detect và crash)

---

## 8. `internal/http/server.go` — Chi Middleware Chain

**Q: Tại sao có cả `chi.Router` và `http.ServeMux`? Khác gì?**

`http.ServeMux` (stdlib) đã đủ tốt từ Go 1.22 với pattern matching `GET /tasks/{id}`. Nhưng chi cho thêm:
- Middleware chaining gọn hơn (`r.Use(...)`)
- Route grouping (`r.Route("/v1", ...)`)
- URL params helpers (`chi.URLParam(r, "id")`)
- Mount sub-routers

Cho project size này, chi đáng đầu tư.

**Q: Order của middleware matter, vì sao?**

```go
r.Use(RequestID)   // 1. gán ID đầu tiên
r.Use(RealIP)      // 2. parse IP để log
r.Use(Logger)      // 3. log start request (cần ID + IP)
r.Use(Recoverer)   // 4. catch panic
r.Use(Timeout)     // 5. áp timeout
```

Mỗi middleware bọc cái sau. Logger ở giữa thấy được request ID (set bởi RequestID). Recoverer phải sau Logger để log được panic. Logic thay đổi nếu đảo thứ tự.

**Q: Vì sao `middleware.Recoverer` cần?**

Mặc định, Go `net/http` đã có panic recovery — server không crash khi 1 handler panic. **Nhưng** mặc định trả 500 mà không log gì. `Recoverer` của chi log full stack trace → debug được.

### Self-check
- Bỏ `Timeout` middleware → request có thể chạy mãi mãi? (Có, trừ khi client disconnect)
- Đặt `Recoverer` trước `Logger` → bug gì? (Logger không log được request panic vì Recoverer "nuốt" panic trước)

---

## 9. `Dockerfile` — Multi-stage Build

**Q: Tại sao multi-stage?**

Stage 1 (`golang:1.22-alpine`) ~350MB, chứa compiler + source.
Stage 2 (`alpine:3.20`) ~7MB, chỉ chứa binary.

Final image: ~15MB (alpine + binary). Cùng binary mà single-stage build → image ~400MB. Quan trọng cho:
- Tốc độ push/pull
- Attack surface (ít file = ít CVE)
- Cold start nhanh

**Q: `CGO_ENABLED=0` để làm gì?**

CGO cho phép Go gọi C code. Mặc định Go build dynamic-linked binary (cần libc). `CGO_ENABLED=0` → static binary, không phụ thuộc libc → chạy được trên `scratch` image (image rỗng, 0 byte).

Cho project này dùng alpine (có libc nhỏ — musl) thì không bắt buộc, nhưng static binary luôn portable hơn.

**Q: `-ldflags="-s -w"` làm gì?**

- `-s`: strip symbol table
- `-w`: strip DWARF debug info

Binary nhỏ hơn ~20-30%, trade-off: không debug được production binary (`gdb`, `pprof` thiếu thông tin). Cho production OK.

**Q: Vì sao chạy bằng user `app` không phải root?**

Container chạy root → nếu attacker escape container, có root host. User non-root là defense in depth, security best practice. Hầu hết app không cần root.

### Self-check
- Skip stage 1, build trên host rồi `COPY binary` vào alpine — được không? (Được, nhưng làm CI/CD reproducibility kém — build env khác nhau cho dev)
- Vì sao `COPY go.mod` trước `COPY .`? (Docker layer cache — nếu source code đổi mà deps không đổi, không cần `go mod download` lại)

---

## 10. `Makefile` — Self-documenting

**Q: Đoạn `grep -E '^[a-zA-Z_-]+:.*?## .*$$'` làm gì?**

Đây là Makefile trick: parse comment `## ...` sau mỗi target name, in ra như help text.

```makefile
run: ## Chạy server  ← comment sau ##
```

`make help` → tự generate documentation. Đây là pattern mượn từ Python `argparse` — convention "self-documenting CLI".

**Q: Tại sao `.PHONY: run test ...`?**

`.PHONY` báo make rằng target này KHÔNG phải file. Nếu không có và bạn có file tên `run` trong folder → make skip target vì nghĩ file đã có sẵn.

**Q: Vì sao có `$$` thay vì `$`?**

Makefile dùng `$` cho biến của make. Để có literal `$` (cho shell), phải escape thành `$$`. Ví dụ `$$GOPATH` → shell nhận `$GOPATH`.

### Self-check
- Thêm target `migrate-status` cần gì? (Thêm dòng `migrate-status: ## ...` với command tương ứng)
- Vì sao một số target không có `.PHONY`? (Bug nhẹ — nên có cho tất cả)

---

## 11. `sqlc.yaml` & SQL queries — Type-safe SQL

**Q: Vì sao không dùng ORM như GORM?**

ORM ẩn SQL → khó debug performance, dễ N+1 queries, magic. sqlc ngược lại:
1. Bạn viết SQL thật (skill transferable)
2. sqlc parse SQL, generate Go function type-safe matching
3. Compile-time check: query sai → build fail

So sánh:
```go
// GORM (runtime error nếu sai field)
db.Where("emial = ?", email).First(&user)  // typo → runtime error

// sqlc (compile-time error)
sqlc.GetUserByEmail(ctx, email)  // method generated, typo → không compile
```

sqlc là **idiomatic Go** — Go ecosystem ưa explicit hơn magic.

**Q: Comment `-- name: CreateUser :one` là gì?**

sqlc parse comment để biết:
- Function name: `CreateUser`
- Return type: `:one` (1 row), `:many` (slice), `:exec` (no result), `:execrows` (rows affected)

Format strict — sai comment → sqlc generate sai code.

**Q: `$1`, `$2` là gì?**

Postgres prepared statement placeholders. MySQL dùng `?` thay thế. sqlc detect dialect từ config và map đúng.

Quan trọng: **luôn dùng placeholder, không bao giờ string concat**:
```sql
-- KHÔNG BAO GIỜ
"SELECT * FROM users WHERE email = '" + email + "'"  -- SQL injection!

-- LUÔN LUÔN
"SELECT * FROM users WHERE email = $1"  -- safe
```

### Self-check
- Sửa schema (add column) nhưng quên `sqlc generate` → chuyện gì? (Code cũ vẫn dùng được, query mới thêm sẽ fail)
- Khi nào nên dùng `:one` vs `:many`? (`:one` khi query có LIMIT 1 hoặc match unique key; `:many` khi return list)

---

## 12. Patterns Lặp Lại Trong Project

### Pattern: Constructor function

```go
func New(...) *T {
    return &T{...}
}
```

Convention: `New` cho default constructor. Nếu có nhiều variant: `NewWithX`, `NewFromY`.

Trả `*T` cho phép caller modify field hoặc gán nil. Trả `T` (value) khi T immutable hoặc rất nhỏ.

### Pattern: Interface ở caller, struct ở callee

```go
// Provider implementation - return concrete struct
func New() *weather.Provider { ... }

// Consumer - accept interface
type Service struct {
    provider providers.Provider  // interface
}
```

**Go idiom**: "Accept interfaces, return structs". Caller nhận interface (linh hoạt khi swap implementation, mock test); callee trả struct (concrete, type-safe).

### Pattern: Context first parameter

```go
func (s *Service) DoSomething(ctx context.Context, ...) error
```

`ctx` LUÔN là param đầu. Lý do: convention. IDE/linter expect như vậy. Đừng đặt giữa.

### Pattern: Error wrapping

```go
if err != nil {
    return fmt.Errorf("operation X: %w", err)
}
```

`%w` wrap error (preserve identity cho `errors.Is/As`). `%v` chỉ in message (mất identity).

Wrap với context: "tôi đang làm X thì gặp lỗi". Stack trace dạng:
```
handle request: load user: query database: connection refused
```

Đọc từ trái → phải, từ high-level → low-level.

---

## Tổng kết: 10 idiom Go quan trọng nhất xuất hiện trong project

1. **Error là value, return tuple `(T, error)`** — không try/catch.
2. **Accept interfaces, return structs** — flexibility ở caller, simplicity ở callee.
3. **Composition over inheritance** — struct embedding thay vì class hierarchy.
4. **Context first param** — cancel/timeout propagation.
5. **Comma-ok idiom** — `v, ok := map[k]`, `v, ok := ch.Recv()`.
6. **Defer for cleanup** — `defer resp.Body.Close()`, `defer mu.Unlock()`.
7. **Sentinel errors + `errors.Is`** — không dùng error codes/types unless cần metadata.
8. **Small interfaces** — `Reader`, `Writer` 1 method. Tránh interface dài 20 method.
9. **Implicit interface satisfaction** — không cần `implements` keyword.
10. **Concurrency primitives qua channel/mutex** — không cần thread pool framework.

Đọc lại file này mỗi cuối tuần. Nhiều idiom sẽ "click" sau khi gặp lại lần 2-3 trong context khác nhau.
