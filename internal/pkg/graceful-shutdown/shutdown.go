package graceful

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type Shutdown struct {
	logger *zap.Logger
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewShutdown(logger *zap.Logger) *Shutdown {
	ctx, cancel := context.WithCancel(context.Background())
	return &Shutdown{
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Shutdown) Wait(timeout time.Duration) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	s.logger.Info("Shutting down...")
	s.cancel()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return context.DeadlineExceeded
	}
}

func (s *Shutdown) Add(f func(context.Context) error) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := f(s.ctx); err != nil {
			s.logger.Error("Error during shutdown", zap.Error(err))
		}
	}()
}
