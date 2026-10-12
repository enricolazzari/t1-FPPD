package main


import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const(
     philosophers = 5
     rounds = 20
)
var (
     forks          [philosophers]sync.Mutex
     statsMu        sync.Mutex
     totalWait, maxWait time.Duration
     currentlyEating, maxEating int
     cycles         [philosophers]int
)

func Philosopher(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	time_to_think := 2
	time_to_eat := 3
	col := strings.Repeat(" ", id*5)
	leftFork, rightFork := id, (id+1)%philosophers

	for round := 1; round <= rounds; round++ {
		// Pensando
		fmt.Printf("%s T%d\n", col, round)
		time.Sleep(time.Duration(time_to_think) * time.Second)

		// Pega Garfo : Primeiro da esquerda e depois da direita
		waitStart := time.Now()
		forks[leftFork].Lock()
		fmt.Printf("%s L%d\n", col, round)
		time.Sleep(time.Millisecond)
		forks[rightFork].Lock()
		waitDuration := time.Since(waitStart)

		// registra quantos estão comendo
		statsMu.Lock()
		totalWait += waitDuration
		maxWait = max(maxWait, waitDuration)
		currentlyEating++
		maxEating = max(maxEating, currentlyEating)
		statsMu.Unlock()

		// come
		fmt.Printf("%s E%d\n", col, round)
		time.Sleep(time.Duration(time_to_eat) * time.Second)

		statsMu.Lock()
		currentlyEating--
		cycles[id]++
		statsMu.Unlock()

		forks[rightFork].Unlock()
		forks[leftFork].Unlock()
	}
}

// vigia: Se a função demorar mais que 10 segundos parada, é considerado deadlock e mata a função
func vigia() {
	last := 0
	for {
		time.Sleep(10 * time.Second)
		statsMu.Lock()
		total := 0
		for _, c := range cycles {
			total += c
		}
		statsMu.Unlock()
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