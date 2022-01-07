//тест проходит примерно за 11.6 секунды

package main

import (
	"sync"
	"testing"
	"time"
)

type TestCases struct {
	expectedTime time.Duration
	cycle time.Duration
	X int
	N int
	testWorker func(wg *sync.WaitGroup, work int, chN chan struct{})
}

var cases = []TestCases{
	//все воркеры выполняются раньше, чем закончится период
	{
		expectedTime: 200 * time.Millisecond,
		cycle: 5 * time.Second,
		X: 10,
		N: 5,
		testWorker: func(wg *sync.WaitGroup, work int, chN chan struct{}) {
			defer wg.Done()
			defer func() {
				chN <- struct{}{}
			}()
			workTime := 100 * time.Millisecond
			time.Sleep(workTime)
		},
	},

	//часть воркеров выполняются за первый период, оставшиеся за второй
	//при этом пул воркеров по количеству помещается в 1 период
	{
		expectedTime: 6000 * time.Millisecond,
		cycle: 5 * time.Second,
		X: 10,
		N: 2,
		testWorker: func(wg *sync.WaitGroup, work int, chN chan struct{}) {
			defer wg.Done()
			defer func() {
				chN <- struct{}{}
			}()
			workTime := 1200 * time.Millisecond
			time.Sleep(workTime)
		},
	},

	//5 воркеров выполняются за первый период, 5 за второй
	//при этом воркеры выполняются быстрее чем заканчивается период
	{
		expectedTime: 5400 * time.Millisecond,
		cycle: 5 * time.Second,
		X: 5,
		N: 4,
		testWorker: func(wg *sync.WaitGroup, work int, chN chan struct{}) {
			defer wg.Done()
			defer func() {
				chN <- struct{}{}
			}()
			workTime := 200 * time.Millisecond
			time.Sleep(workTime)
		},
	},
}


func TestRateLimiter(t *testing.T) {

	for i, testCase := range cases {
		workersPool := make([]job, 10)
		for i := range workersPool {
			workersPool[i] = testCase.testWorker
		}

		start := time.Now()
		ticker := time.NewTicker(testCase.cycle)
		RateLimiter(testCase.X, testCase.N, ticker, workersPool...)
		end := time.Since(start)

		if end < testCase.expectedTime {
			t.Errorf("Test case %d\nexecition too fast\nGot: %s\nExpected about: %s", i, end, testCase.expectedTime)
		}

		if end > testCase.expectedTime + 10 * time.Millisecond {
			t.Errorf("Test case %d\nexecition too slow\nGot: %s\nExpected about: %s", i, end, testCase.expectedTime)
		}
	}
}