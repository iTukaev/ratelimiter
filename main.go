package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	//период для Х
	cycle    = 8 * time.Second
	//Х при запуске без флага
	defaultX = 10
	//N при запуске без флага
	defaultN = 1
	//минимальное время выполнения имитации долгой работы (мс)
	randBase = 200
	//основание рандомайзера имитации долгой работы (мс)
	randMutable = 400
	//количество рабочих функций для выполнения
	workersCount = 20
)

type job func(*sync.WaitGroup, int, chan struct{})

func main() {
	//не более N команд одновременно
	//не более X команд за один период
	var N,X int
	flag.IntVar(&N,"n", defaultN, "no more then \"-n\" commands at the same time")
	flag.IntVar(&X,"x", defaultX, "no more then \"-x\" commands per minute")
	flag.Parse()
	rand.Seed(time.Now().Unix())

	ticker := time.NewTicker(cycle)

	//здесь я заполняю пул воркеров одинаковыми функциями
	workersPool := make([]job, workersCount)
	for i := range workersPool {
		workersPool[i] = workerFunc
	}

	RateLimiter(X, N, ticker, workersPool...)
}

func RateLimiter(X,N int, ticker *time.Ticker, workers ...job) {
	var wg sync.WaitGroup
	//канал N для контроля количества одновременно запущенных воркеров
	chN := make(chan struct{}, N)
	//канал X для контроля команд за один период
	chX := make(chan struct{})

	for i := 0; i < N; i++ {
		chN <- struct{}{}
	}

	//добавляем в канал свободный слот для запуска воркера
	//если время периода вышло - счетчик начинается заново
	go func() {
		for {
			for i := 0; i < X; i++ {
				select {
				case chX <- struct{}{}:
				case <-ticker.C:
					continue
				}
			}
			<-ticker.C
		}
	}()

	//имитирую номер новой команды
	//для наглядности взял просто цифру
	work := 0

	for _, doWork := range workers{
		<-chN
		<-chX
		wg.Add(1)
		go doWork(&wg, work, chN)
		work++
	}
	wg.Wait()
}

func workerFunc(wg *sync.WaitGroup, work int, chN chan struct{}) {
	defer wg.Done()
	defer func() {
		chN <- struct{}{}
	}()

	t := rand.Intn(randMutable) + randBase
	fmt.Printf("worker %d started at %v duration %vms\n", work, time.Now().UTC(), t)
	time.Sleep(time.Duration(t) * time.Millisecond)
	fmt.Printf("worker %d ended %v\n", work, time.Now().UTC())
}