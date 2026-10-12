# V2 — differences from V1 marked with [NEW]

```go
package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	philosophers = 5
	rounds       = 20
)

var (
	forks              [philosophers]sync.Mutex
	mu                 sync.Mutex
	totalWait, maxWait time.Duration
	eating, maxEating  int
	cycles             [philosophers]int
	sem                = make(chan struct{}, philosophers-1) // [NEW] semaforo
)

func Philosopher(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	time_to_think := 2
	time_to_eat := 3
	col := strings.Repeat(" ", id*5)
	left, right := id, (id+1)%philosophers

	for i := 1; i <= rounds; i++ {
		fmt.Printf("%s T%d\n", col, i)
		time.Sleep(time.Duration(time_to_think) * time.Second)

		start := time.Now()
		sem <- struct{}{}        // [NEW] entra no semaforo antes de pegar garfo
		forks[left].Lock()
		fmt.Printf("%s L%d\n", col, i)
		time.Sleep(time.Millisecond)
		forks[right].Lock()
		wait := time.Since(start)

		mu.Lock()
		totalWait += wait
		maxWait = max(maxWait, wait)
		eating++
		maxEating = max(maxEating, eating)
		mu.Unlock()

		fmt.Printf("%s E%d\n", col, i)
		time.Sleep(time.Duration(time_to_eat) * time.Second)

		mu.Lock()
		eating--
		cycles[id]++
		mu.Unlock()

		forks[right].Unlock()
		forks[left].Unlock()
		<-sem                    // [NEW] libera o semaforo depois de soltar os garfos
	}
}

func vigia() {
	last := 0
	for {
		time.Sleep(10 * time.Second)
		mu.Lock()
		total := 0
		for _, c := range cycles {
			total += c
		}
		mu.Unlock()
		if total == last {
			fmt.Println("\nDEADLOCK detected! No progress in the last 10s")
			fmt.Println("Cycles per philosopher:", cycles)
			os.Exit(1)
		}
		last = total
	}
}

func main() {
	fmt.Printf("\n[P1] [P2] [P3] [P4] [P5]\n\n")

	var wg sync.WaitGroup
	start := time.Now()
	go vigia()

	for j := 0; j < philosophers; j++ {
		wg.Add(1)
		go Philosopher(j, &wg)
	}
	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("\nDinner took %s\n", elapsed)
	fmt.Println("Average wait:", totalWait/(philosophers*rounds))
	fmt.Println("Max wait:", maxWait)
	fmt.Println("Max eating at the same time:", maxEating)
	fmt.Println("Cycles per philosopher:", cycles)
}
```
