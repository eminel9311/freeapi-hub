# Tuần 1 - Go Foundations + First Endpoint

> Mục tiêu cuối tuần: endpoint `GET /v1/weather?city=Hanoi` chạy được, trả JSON đẹp.

Mỗi buổi ~1-2h. Đừng vội — Go cực kỳ đơn giản nhưng cần "nói chuyện" với compiler nhiều lần để quen.

## Buổi 1: Setup môi trường

### Checklist

- [ ] Cài Go 1.22+: https://go.dev/dl
- [ ] Verify: `go version`
- [ ] Cài VSCode + extension "Go" (của Google), hoặc GoLand
- [ ] Khi mở file `.go` đầu tiên, VSCode sẽ prompt cài gopls, dlv, gotests... → Install All
- [ ] Cài tools dev:
  ```bash
  cd freeapi-hub
  make install-tools
  ```

### Hiểu Go module

```bash
go mod tidy        # tải dependencies
go run ./cmd/server  # chạy server
curl http://localhost:8080/health
```

Nếu thấy `{"status":"ok"}` → môi trường OK.

### Bài tập

1. Đổi `main.go` để in `"Hello, [tên bạn]"` khi start. Save → tự reload với `make dev`.
2. Đọc kỹ file `go.mod`. Hỏi Claude: "giải thích từng dòng go.mod cho tôi".

---

## Buổi 2: Types, Structs, Slices, Maps

Tạo file `playground/main.go` (folder mới ngoài project, chỉ để học syntax):

```go
package main

import "fmt"

type Person struct {
    Name string
    Age  int
}

func main() {
    // Variables
    var name string = "An"   // explicit
    age := 25                 // type inference

    // Struct
    p := Person{Name: "Bình", Age: 30}
    fmt.Println(p.Name)

    // Slice
    nums := []int{1, 2, 3}
    nums = append(nums, 4)

    // Map
    scores := map[string]int{"An": 90}
    scores["Bình"] = 85

    // Pointer
    pp := &p
    pp.Age = 31  // tự dereference, không cần (*pp).Age

    fmt.Println(name, age, nums, scores, p)
}
```

### Bài tập

1. Khai báo struct `Task` có ID, Title, Done.
2. Tạo slice 5 Task, in ra số task chưa done.
3. Function `MarkDone(t *Task)` — hiểu vì sao phải truyền pointer.

---

## Buổi 3: Error Handling - "The Go Way"

Khác với try/catch của các ngôn ngữ khác. Go dùng **value-based errors**:

```go
package main

import (
    "errors"
    "fmt"
)

var ErrDivByZero = errors.New("division by zero")

func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, ErrDivByZero
    }
    return a / b, nil
}

func main() {
    result, err := divide(10, 0)
    if err != nil {
        // wrap error với context bằng %w
        wrapped := fmt.Errorf("calculate ratio: %w", err)
        fmt.Println(wrapped)

        // check loại lỗi
        if errors.Is(wrapped, ErrDivByZero) {
            fmt.Println("đúng là lỗi div by zero")
        }
        return
    }
    fmt.Println(result)
}
```

### Bài tập

1. Viết function `ReadConfig(path string) (Config, error)` đọc file, return lỗi nếu file không tồn tại.
2. Định nghĩa sentinel error `ErrConfigNotFound`. Check bằng `errors.Is`.

---

## Buổi 4: Gọi API ngoài (HTTP Client)

Tạo file `playground/weather.go`:

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type GeoResult struct {
    Results []struct {
        Latitude  float64 `json:"latitude"`
        Longitude float64 `json:"longitude"`
        Name      string  `json:"name"`
    } `json:"results"`
}

func main() {
    client := &http.Client{Timeout: 5 * time.Second}

    url := "https://geocoding-api.open-meteo.com/v1/search?name=Hanoi&count=1"
    resp, err := client.Get(url)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()  // CỰC quan trọng - không close → memory leak

    body, _ := io.ReadAll(resp.Body)

    var geo GeoResult
    if err := json.Unmarshal(body, &geo); err != nil {
        panic(err)
    }

    if len(geo.Results) == 0 {
        fmt.Println("city not found")
        return
    }
    fmt.Printf("Hanoi at %.2f, %.2f\n", geo.Results[0].Latitude, geo.Results[0].Longitude)
}
```

### Bài tập

1. Mở rộng: dùng kết quả lat/lon để gọi forecast endpoint:
   ```
   https://api.open-meteo.com/v1/forecast?latitude=21.03&longitude=105.85&current=temperature_2m,wind_speed_10m
   ```
2. Parse response, in ra nhiệt độ.

---

## Buổi 5: Implement Weather Provider thật

Bây giờ quay lại project, implement file `internal/providers/weather/weather.go`.

Đây là gợi ý đầy đủ (đừng copy ngay — gõ tay):

```go
func (p *Provider) Fetch(ctx context.Context, params map[string]string) (any, error) {
    city := params["city"]
    if city == "" {
        return nil, fmt.Errorf("missing city param")
    }

    // Step 1: geocoding
    var geoResp struct {
        Results []struct {
            Latitude  float64 `json:"latitude"`
            Longitude float64 `json:"longitude"`
            Name      string  `json:"name"`
        } `json:"results"`
    }

    _, err := p.client.R().
        SetContext(ctx).
        SetQueryParams(map[string]string{
            "name":  city,
            "count": "1",
        }).
        SetResult(&geoResp).
        Get("https://geocoding-api.open-meteo.com/v1/search")
    if err != nil {
        return nil, fmt.Errorf("geocoding: %w", err)
    }
    if len(geoResp.Results) == 0 {
        return nil, fmt.Errorf("city not found: %s", city)
    }

    loc := geoResp.Results[0]

    // Step 2: forecast
    var fcResp struct {
        Current struct {
            Time        string  `json:"time"`
            Temperature float64 `json:"temperature_2m"`
            WindSpeed   float64 `json:"wind_speed_10m"`
        } `json:"current"`
    }

    _, err = p.client.R().
        SetContext(ctx).
        SetQueryParams(map[string]string{
            "latitude":  fmt.Sprintf("%f", loc.Latitude),
            "longitude": fmt.Sprintf("%f", loc.Longitude),
            "current":   "temperature_2m,wind_speed_10m",
        }).
        SetResult(&fcResp).
        Get(p.baseURL + "/forecast")
    if err != nil {
        return nil, fmt.Errorf("forecast: %w", err)
    }

    t, _ := time.Parse("2006-01-02T15:04", fcResp.Current.Time)
    return Response{
        City:        loc.Name,
        Temperature: fcResp.Current.Temperature,
        WindSpeed:   fcResp.Current.WindSpeed,
        Time:        t,
    }, nil
}
```

---

## Buổi 6: HTTP Handler + Router

Tạo file `internal/providers/weather/handler.go`:

```go
package weather

import (
    "net/http"

    httpserver "github.com/eminel9311/freeapi-hub/internal/http"
)

// Handler trả về http.HandlerFunc cho route /v1/weather.
// Pattern: handler là method/closure để có thể bind dependencies (provider).
func (p *Provider) Handler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        city := r.URL.Query().Get("city")
        if city == "" {
            httpserver.Error(w, http.StatusBadRequest, "missing 'city' query param")
            return
        }

        data, err := p.Fetch(r.Context(), map[string]string{"city": city})
        if err != nil {
            httpserver.Error(w, http.StatusBadGateway, err.Error())
            return
        }

        httpserver.JSON(w, http.StatusOK, data)
    }
}
```

Update `internal/http/server.go` để mount route:

```go
// Trong NewRouter, sau /health:
weatherProv := weather.New("https://api.open-meteo.com/v1")
r.Get("/v1/weather", weatherProv.Handler())
```

Update `cmd/server/main.go` để hook lên router (xem comment TODO trong file).

---

## Buổi 7: Polish & Commit

- Test endpoint với curl: `curl 'http://localhost:8080/v1/weather?city=Hanoi'`
- Đọc lại toàn bộ code mình viết, comment chỗ chưa hiểu
- Viết các test case bằng tay: city không tồn tại, missing param, server không có internet...
- `git init`, commit, push lên GitHub
- Viết section "Week 1 Done" vào README mô tả những gì học được

## Cuối Tuần 1: Self-check

Bạn nên trả lời được các câu hỏi sau (không cần search):

1. Khác nhau giữa `var x int` và `x := 5`?
2. Khi nào dùng `func (p *Provider)` vs `func (p Provider)`?
3. Vì sao phải `defer resp.Body.Close()`?
4. `fmt.Errorf("...%w", err)` khác `fmt.Errorf("...%v", err)` chỗ nào?
5. Slice và array khác nhau như thế nào?
6. Map có thread-safe không?

Nếu chưa rõ câu nào → hỏi Claude, hoặc đọc Go by Example về topic đó.

**Next:** Tuần 2 — Add 3 providers nữa và refactor thành interface-driven architecture.
