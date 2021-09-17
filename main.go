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

// Service Структура, чтобы передавать канал в хэндлеры
type Service struct{
	ch chan string
}

func(s Service) Handler(c echo.Context) error {
	s.ch <- c.Request().Host
	return c.String(http.StatusOK, "Hello, this is a contexts sample!")
}

func main(){

	// Главный контекст программы, его закроет горунтина, которая остановит сервис и его зыкрытие отследит горутина, которая читает с канала
	ctxMain, cancelMain := context.WithCancel(context.Background())
	// Канал в который пишут все хэндлеры и с которого читает одна постоянная горутина
	ch := make(chan string, 100)

	s := Service{ch}

	e := echo.New()
	e.GET("/", s.Handler)

	// WaitGroup чтобы программа дождалась завершения работы постоянной горутины
	wg := sync.WaitGroup{}

	wg.Add(1)
	// Горутина, которая считывает с канала значения, которые туда помещают хэндлеры сервера
	go func(context.Context, chan string){
		defer wg.Done()
		chDone := ctxMain.Done()
		for{
			select {
			case s := <-ch:
				fmt.Print(s+"\n")
			// Ловим остановку контекста (остановку сервера) и дочитываем из канала оставшиеся сообщения если они там есть
			case <-chDone:
				fmt.Print("server stopped the main context" + "\n")
				if len(ch) == 0{
					return
				} else {
					for i:=0; i<len(ch); i++{
						s := <-ch
						fmt.Print(s+"\n")
					}
					return
				}
			default:
			}
		}
	}(ctxMain, ch)

	go func(){
		e.Logger.Error(e.Start(":8080"))
	}()

	// Остановка сервера через 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	<- time.After(5*time.Second)
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	// time.After если я просто даю таймаут, чтобы горутина считала все с канала
	//<- time.After(5*time.Second)

	cancel()
	cancelMain()

	wg.Wait()
}