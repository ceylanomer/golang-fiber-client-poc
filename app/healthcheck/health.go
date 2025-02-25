package healthcheck

import (
	"context"
	"runtime"
	"time"
)

type Request struct {
}

type Response struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Memory    *MemoryStats      `json:"memory"`
	System    *SystemStats      `json:"system"`
	Checks    map[string]string `json:"checks,omitempty"`
}

type MemoryStats struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"totalAlloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"numGC"`
}

type SystemStats struct {
	GoVersion    string `json:"goVersion"`
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	NumCPU       int    `json:"numCPU"`
	NumGoroutine int    `json:"numGoroutine"`
}

type Handler struct {
	startTime time.Time
	version   string
	checks    map[string]func() string
}

func NewHealthCheckHandler() *Handler {
	return &Handler{
		startTime: time.Now(),
		version:   "1.0.0", // Uygulama versiyonunu burada belirtin
		checks:    make(map[string]func() string),
	}
}

// RegisterCheck özel sağlık kontrolü eklemek için kullanılabilir
func (h *Handler) RegisterCheck(name string, check func() string) {
	h.checks[name] = check
}

func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Çalışma süresi hesaplama
	uptime := time.Since(h.startTime).String()

	// Sağlık kontrol sonuçlarını toplama
	checks := make(map[string]string)
	for name, check := range h.checks {
		checks[name] = check()
	}

	return &Response{
		Status:    "OK",
		Timestamp: time.Now(),
		Version:   h.version,
		Uptime:    uptime,
		Memory: &MemoryStats{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			NumGC:      memStats.NumGC,
		},
		System: &SystemStats{
			GoVersion:    runtime.Version(),
			GOOS:         runtime.GOOS,
			GOARCH:       runtime.GOARCH,
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
		},
		Checks: checks,
	}, nil
}
