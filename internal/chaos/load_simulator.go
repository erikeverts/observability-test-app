package chaos

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

type LoadSimulator struct {
	cpuEnabled bool
	cpuPercent int
	memEnabled bool
	memMB      int
}

func NewLoadSimulator(cpuEnabled bool, cpuPercent int, memEnabled bool, memMB int) *LoadSimulator {
	return &LoadSimulator{
		cpuEnabled: cpuEnabled,
		cpuPercent: cpuPercent,
		memEnabled: memEnabled,
		memMB:      memMB,
	}
}

func (l *LoadSimulator) Start(ctx context.Context) {
	if l.cpuEnabled && l.cpuPercent > 0 {
		slog.Info("chaos: starting CPU load simulation", "percent", l.cpuPercent)
		go l.simulateCPU(ctx)
	}
	if l.memEnabled && l.memMB > 0 {
		slog.Info("chaos: starting memory load simulation", "mb", l.memMB)
		go l.simulateMemory(ctx)
	}
}

func (l *LoadSimulator) simulateCPU(ctx context.Context) {
	numCPU := runtime.NumCPU()
	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Burn CPU for a proportion of time based on target percent.
					busyDuration := time.Duration(l.cpuPercent) * time.Millisecond / 10
					idleDuration := time.Duration(100-l.cpuPercent) * time.Millisecond / 10

					start := time.Now()
					for time.Since(start) < busyDuration {
						// Busy loop
					}
					time.Sleep(idleDuration)
				}
			}
		}()
	}
}

func (l *LoadSimulator) simulateMemory(ctx context.Context) {
	// Allocate memory in 1MB chunks
	ballast := make([][]byte, 0, l.memMB)
	for i := 0; i < l.memMB; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			chunk := make([]byte, 1024*1024) // 1MB
			// Touch memory to ensure it's allocated
			for j := range chunk {
				chunk[j] = byte(j % 256)
			}
			ballast = append(ballast, chunk)
		}
	}
	slog.Info("chaos: memory allocated", "mb", l.memMB)

	// Hold memory until context is cancelled
	<-ctx.Done()
	runtime.KeepAlive(ballast)
}
