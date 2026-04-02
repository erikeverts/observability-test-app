package telemetry

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/erikeverts/observability-test-app/internal/config"
	"github.com/grafana/pyroscope-go"
	otelpyroscope "github.com/grafana/otel-profiling-go"
	"go.opentelemetry.io/otel"
)

// InitProfiling starts the Pyroscope continuous profiler if enabled.
// It wraps the existing TracerProvider for span-profile correlation
// and returns a cleanup function that stops the profiler.
func InitProfiling(cfg *config.Config) (func(), error) {
	if !cfg.ProfilingEnabled {
		return func() {}, nil
	}

	// Wrap the global TracerProvider for span profiling
	tp := otelpyroscope.NewTracerProvider(otel.GetTracerProvider())
	otel.SetTracerProvider(tp)

	var authUser, authPassword string
	if cfg.PyroscopeAuthToken != "" {
		parts := strings.SplitN(cfg.PyroscopeAuthToken, ":", 2)
		if len(parts) == 2 {
			authUser = parts[0]
			authPassword = parts[1]
		} else {
			authPassword = cfg.PyroscopeAuthToken
		}
	}

	profiler, err := pyroscope.Start(pyroscope.Config{
		ApplicationName:   cfg.ServiceName,
		ServerAddress:     cfg.PyroscopeEndpoint,
		BasicAuthUser:     authUser,
		BasicAuthPassword: authPassword,
		TenantID:          cfg.PyroscopeTenantID,
		Tags:              map[string]string{"service": cfg.ServiceName},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if err != nil {
		return func() {}, fmt.Errorf("starting pyroscope profiler: %w", err)
	}

	slog.Info("profiling enabled",
		"endpoint", cfg.PyroscopeEndpoint,
		"service", cfg.ServiceName,
		"go_version", runtime.Version(),
	)

	return func() {
		if err := profiler.Stop(); err != nil {
			slog.Error("stopping profiler", "error", err)
		}
	}, nil
}
