package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
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
}

func (p *Philosopher) dine(wg *sync.WaitGroup) {
	defer wg.Done()
	for p.eatingNum < 10 {
		p.think()
		p.eat()
	}
	p.state = DONE
}

func (p *Philosopher) think() {
	p.state = THINKING
	startThink := time.Now()
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	p.thinkTime += time.Since(startThink)
}

func (p *Philosopher) eat() {
	startWait := time.Now()
	p.state = WAITING
	p.left.Lock()
	if !p.right.TryLock() {
		p.left.Unlock()
		return
	}
	p.waitTime += time.Since(startWait)

	p.state = EATING
	startEat := time.Now()
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	p.eatTime += time.Since(startEat)

	p.eatingNum++

	p.left.Unlock()
	p.right.Unlock()
}

func monitor(philosophers []*Philosopher) {
	for {
		time.Sleep(200 * time.Millisecond)
		fmt.Println("------------------------")
		for _, p := range philosophers {
			fmt.Printf("Philosopher %d: %s\n", p.id, p.state)
		}
	}
}

func main() {
	n := 5
	chopticks := make([]*sync.Mutex, n)
	for i := range chopticks {
		chopticks[i] = &sync.Mutex{}
	}

	philosophers := make([]*Philosopher, n)
	for i := 0; i < n; i++ {
		philosophers[i] = &Philosopher{
			id:    i,
			left:  chopticks[i],
			right: chopticks[(i+1)%n],
			state: THINKING,
		}
	}

	var wg sync.WaitGroup
	wg.Add(n)

	go monitor(philosophers)

	for _, p := range philosophers {
		go p.dine(&wg)
	}

	wg.Wait()

	fmt.Println("------------ Dinner Done ------------")
	for _, p := range philosophers {
		fmt.Printf("Philosopher %d: Thinking = %v, Waiting = %v, Eating = %v\n", p.id, p.thinkTime, p.waitTime, p.eatTime)
	}
}
