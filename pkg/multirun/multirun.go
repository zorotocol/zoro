package multirun

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
)

type Job func(ctx context.Context) error

func Signal(sigs ...os.Signal) context.Context {
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() {
		ch := make(chan os.Signal, len(sigs))
		signal.Notify(ch, sigs...)
		cancel(errors.New((<-ch).String()))
	}()
	return ctx
}
func Run(parent context.Context, jobs ...Job) error {
	if len(jobs) == 0 {
		panic("empty jobs")
	}
	ctx, cancel := context.WithCancelCause(parent)
	wg := sync.WaitGroup{}
	wg.Add(len(jobs))
	for _, job := range jobs {
		local := job
		go func() {
			defer wg.Done()
			cancel(local(ctx))
		}()
	}
	<-ctx.Done()
	wg.Wait()
	return context.Cause(ctx)
}
