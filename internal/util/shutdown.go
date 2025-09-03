package util


import (
"context"
"os"
"os/signal"
"syscall"
)


func WithSignals(ctx context.Context) (context.Context, context.CancelFunc) {
ctx, cancel := context.WithCancel(ctx)
sig := make(chan os.Signal, 1)
signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
go func() {
<-sig
cancel()
}()
return ctx, cancel
}