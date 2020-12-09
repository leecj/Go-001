package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	group, errCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return createServer(errCtx, &Server{addr: ":2001", name: "server1"})
	})
	group.Go(func() error {
		return createServer(errCtx, &Server{addr: ":2002", name: "server2"})
	})
	// 监听signal退出
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT)
		s := <-c
		fmt.Println("Exit Single: ", s)
		cancel()
	}()
	if err := group.Wait(); err != nil {
		fmt.Println("Get errors: ", err)
	}
}

type Server struct {
	addr string
	name string
}

func createServer(ctx context.Context, server *Server) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(server.name)
	})
	s := &http.Server{Addr: server.addr, Handler: mux}
	go func() {
		<-ctx.Done()
		fmt.Println(server.name + " shutdown")
		s.Shutdown(context.Background())
	}()
	return s.ListenAndServe()
}
