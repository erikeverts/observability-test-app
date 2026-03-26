package chaos

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

type LoadSimulator struct {
	mu        sync.Mutex
	parentCtx context.Context
	cancel    context.CancelFunc

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

func (l *LoadSimulator) Start(parentCtx context.Context) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.parentCtx = parentCtx
	l.startLocked()
}

func (l *LoadSimulator) Reconfigure(cpuEnabled bool, cpuPercent int, memEnabled bool, memMB int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
	l.cpuEnabled = cpuEnabled
	l.cpuPercent = cpuPercent
	l.memEnabled = memEnabled
	l.memMB = memMB
	l.startLocked()
}

func (l *LoadSimulator) GetConfig() (cpuEnabled bool, cpuPercent int, memEnabled bool, memMB int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cpuEnabled, l.cpuPercent, l.memEnabled, l.memMB
}

func (l *LoadSimulator) startLocked() {
	if l.parentCtx == nil {
		return
	}
	needsCPU := l.cpuEnabled && l.cpuPercent > 0
	needsMem := l.memEnabled && l.memMB > 0
	if !needsCPU && !needsMem {
		return
	}

	ctx, cancel := context.WithCancel(l.parentCtx)
	l.cancel = cancel

	if needsCPU {
		slog.Info("chaos: starting CPU load simulation", "percent", l.cpuPercent)
		go l.simulateCPU(ctx)
	}
	if needsMem {
		slog.Info("chaos: starting memory load simulation", "mb", l.memMB)
		go l.simulateMemory(ctx)
	}
}

func (l *LoadSimulator) simulateCPU(ctx context.Context) {
	l.mu.Lock()
	percent := l.cpuPercent
	l.mu.Unlock()

	numCPU := runtime.NumCPU()
	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					busyDuration := time.Duration(percent) * time.Millisecond / 10
					idleDuration := time.Duration(100-percent) * time.Millisecond / 10
					start := time.Now()
					for time.Since(start) < busyDuration {
					}
					time.Sleep(idleDuration)
				}
			}
		}()
	}
}

func (l *LoadSimulator) simulateMemory(ctx context.Context) {
	l.mu.Lock()
	mb := l.memMB
	l.mu.Unlock()

	ballast := make([][]byte, 0, mb)
	for i := 0; i < mb; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			chunk := make([]byte, 1024*1024)
			for j := range chunk {
				chunk[j] = byte(j % 256)
			}
			ballast = append(ballast, chunk)
		}
	}
	slog.Info("chaos: memory allocated", "mb", mb)
	<-ctx.Done()
	runtime.KeepAlive(ballast)
}
