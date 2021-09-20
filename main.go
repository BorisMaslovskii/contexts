package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
	tomb "gopkg.in/tomb.v2"
)

// Service Структура, чтобы передавать канал в хэндлеры
type Service struct {
	ch   chan string
	tomb *tomb.Tomb
}

func (s *Service) Handler(c echo.Context) error {
	s.ch <- c.Request().Host
	return c.String(http.StatusOK, "Hello, this is a contexts sample!")
}

type Server struct {
	echo *echo.Echo
	tomb *tomb.Tomb
}

func (s *Server) Start() error {
	return s.echo.Start(":8080")
}

func (s *Server) Stop() error {
	if err := s.echo.Shutdown(context.Background()); err != nil {
		s.echo.Logger.Error(err)
	}
	s.tomb.Kill(nil)
	return s.tomb.Wait()
}

func (s *Service) Start() error {
	for {
		select {
		case x := <-s.ch:
			fmt.Print(x + "\n")
			<-time.After(time.Second)
		case <-s.tomb.Dying():
			if len(s.ch) == 0 {
				return errors.New(fmt.Sprint("server stopped, channel is empty, finish" + "\n"))
			}
			fmt.Print("server stopped, channel is not empty, continue" + "\n")
		default:
		}
	}
}

func main() {
	// Every handler writes to this channel and one background goroutine reads from it
	ch := make(chan string, 100)
	t := tomb.Tomb{}

	// background service
	service := &Service{ch, &t}

	e := echo.New()
	e.GET("/", service.Handler)

	server := &Server{e, &t}

	t.Go(service.Start)
	t.Go(server.Start)

	// stop server after 5 seconds
	<-time.After(5 * time.Second)
	err := server.Stop()
	fmt.Println(err)
}
