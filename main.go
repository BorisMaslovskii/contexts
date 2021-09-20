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
	// Канал в который пишут все хэндлеры и с которого читает одна постоянная горутина
	ch := make(chan string, 100)
	t := tomb.Tomb{}

	// background сервис который взаимодействует с http запросами
	service := &Service{ch, &t}

	e := echo.New()
	e.GET("/", service.Handler)

	server := &Server{e, &t}

	t.Go(service.Start)
	t.Go(server.Start)

	// Остановка сервера через 5 секунд
	<-time.After(5 * time.Second)
	err := server.Stop()
	fmt.Println(err)
}
