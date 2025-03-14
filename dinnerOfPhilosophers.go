package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	numberOfPhilosophers = 5
)

const (
	THINKING = "Thinking"
	WAITING  = "Waiting"
	EATING   = "Eating"
	DONE     = "Done"
)

type Philosopher struct {
	id        int
	left      *sync.Mutex
	right     *sync.Mutex
	eatingNum int
	thinkTime time.Duration
	waitTime  time.Duration
	eatTime   time.Duration
	state     string
	logger    *log.Logger
}

func (p *Philosopher) dine(wg *sync.WaitGroup) {
	defer wg.Done()
	for p.eatingNum < 10 {
		p.think()
		p.eat()
	}
	p.state = DONE
	p.logger.Printf("Philosopher %d: Done eating\n", p.id)
}

func (p *Philosopher) think() {
	p.state = THINKING
	startThink := time.Now()
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	p.thinkTime += time.Since(startThink)
	p.logger.Printf("Philosopher %d: Thinking\n", p.id)
}

func (p *Philosopher) eat() {
	p.logger.Printf("Philosopher %d: Waiting\n", p.id)

	p.left.Lock()
	p.state = WAITING
	startWait := time.Now()
	if !p.right.TryLock() {
		p.left.Unlock()
		return
	}
	// p.right.Lock()
	p.waitTime += time.Since(startWait)

	p.state = EATING
	startEat := time.Now()
	p.logger.Printf("Philosopher %d: Eating\n", p.id)
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	p.eatTime += time.Since(startEat)

	p.eatingNum++

	p.left.Unlock()
	p.right.Unlock()
}

func monitor(philosophers []*Philosopher, logger *log.Logger) {
	for {
		time.Sleep(100 * time.Millisecond)
		logger.Println("------------------------")
		for _, p := range philosophers {
			logger.Printf("Philosopher %d: %s\n", p.id, p.state)
		}
	}
}

func main() {
	logFile, err := os.OpenFile("philosophers.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer logFile.Close()

	// MultiWriter: Log to both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(multiWriter, "", log.LstdFlags|log.Lmicroseconds)

	chopticks := make([]*sync.Mutex, numberOfPhilosophers)
	for i := range chopticks {
		chopticks[i] = &sync.Mutex{}
	}

	philosophers := make([]*Philosopher, numberOfPhilosophers)
	for i := 0; i < numberOfPhilosophers; i++ {
		philosophers[i] = &Philosopher{
			id:     i,
			left:   chopticks[i],
			right:  chopticks[(i+1)%numberOfPhilosophers],
			state:  THINKING,
			logger: logger,
		}
	}

	var wg sync.WaitGroup
	wg.Add(numberOfPhilosophers)

	// Start monitoring logs
	go monitor(philosophers, logger)

	// Start philosophers dining
	for _, p := range philosophers {
		go p.dine(&wg)
	}

	wg.Wait()

	logger.Println("------------ Dinner Done ------------")
	for _, p := range philosophers {
		logger.Printf("Philosopher %d: Thinking = %v, Waiting = %v, Eating = %v\n", p.id, p.thinkTime, p.waitTime, p.eatTime)
	}

	averageThinkTime := time.Duration(0)
	averageWaitTime := time.Duration(0)
	averageEatTime := time.Duration(0)

	for _, p := range philosophers {
		averageThinkTime += p.thinkTime
		averageWaitTime += p.waitTime
		averageEatTime += p.eatTime
	}

	averageThinkTime /= time.Duration(numberOfPhilosophers)
	averageWaitTime /= time.Duration(numberOfPhilosophers)
	averageEatTime /= time.Duration(numberOfPhilosophers)

	logger.Printf("Average Think Time = %v, Average Wait Time = %v, Average Eat Time = %v\n", averageThinkTime, averageWaitTime, averageEatTime)

	fmt.Println("Logs saved to philosophers.log")
}
