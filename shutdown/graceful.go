package shutdown

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func Graceful(closers ...io.Closer) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM)

	signal := <-sig

	fmt.Printf("received signal %s, shutting down", signal.String())
	for _, c := range closers {
		if err := c.Close(); err != nil {
			fmt.Printf("error closing: %v", err)
		}
	}
	os.Exit(0)
}
