package main

import (
	"context"
	"fmt"
	//"fmt"
	"net/http"
	"sync"
	"time"

	echo "github.com/labstack/echo/v4"
)

type Service struct {
	ch chan string
}

func (s Service) Handler(c echo.Context) error {
	s.ch <- c.Request().Host
	return c.String(http.StatusOK, "Hello, this is a contexts sample!")
}

func main() {

	// Main context, we stop it right after we stop the server, and then we catch it in the background service
	ctxMain, cancelMain := context.WithCancel(context.Background())
	// Every handler writes to this channel and only one background service reads from it
	ch := make(chan string, 100)

	s := Service{ch}

	e := echo.New()
	e.GET("/", s.Handler)

	wg := sync.WaitGroup{}

	wg.Add(1)
	// Background server
	go func(context.Context, chan string) {
		defer wg.Done()
		chDone := ctxMain.Done()
		for {
			select {
			case s := <-ch:
				fmt.Print(s + "\n")
				<-time.After(time.Second)
			// Catch the server stops and check if channel is not empty
			case <-chDone:
				if len(ch) == 0 {
					fmt.Print("server stopped, channel is empty, finish" + "\n")
					return
				}
				fmt.Print("server stopped, channel is not empty, continue" + "\n")
			default:
			}
		}
	}(ctxMain, ch)

	go func() {
		e.Logger.Error(e.Start(":8080"))
	}()

	// Stop server after 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	<-time.After(5 * time.Second)
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	cancel()
	cancelMain()

	wg.Wait()
}
