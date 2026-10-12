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
     forks [philosophers]sync.Mutex		// Garfos que os filosofos pegam
     mu sync.Mutex 						
     totalWait, maxWait time.Duration	// Soma total de espera e a maior espera do codigo
     eating, maxEating int				// Quantos filosofos estão comendo e qual foi o maior valor de eating
     cycles [philosophers]int			// Quantos ciclos cada filosofo completou
)

func Philosopher(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	time_to_think := 2
	time_to_eat := 3
	col := strings.Repeat(" ", id*5)
	left, right := id, (id+1)%philosophers

	for i := 1; i <= rounds; i++ {
		// Pensando
		fmt.Printf("%s T%d\n", col, i)
		time.Sleep(time.Duration(time_to_think) * time.Second)

		// Pega Garfo : Primeiro da esquerda e depois da direita
		start := time.Now()
		forks[left].Lock()
		fmt.Printf("%s L%d\n", col, i)
		time.Sleep(time.Millisecond)
		forks[right].Lock()
		wait := time.Since(start)

		// registra quantos estão comendo
		mu.Lock()
		totalWait += wait
		maxWait = max(maxWait, wait)
		eating++
		maxEating = max(maxEating, eating)
		mu.Unlock()

		// come
		fmt.Printf("%s E%d\n", col, i)
		time.Sleep(time.Duration(time_to_eat) * time.Second)

		mu.Lock()
		eating--
		cycles[id]++
		mu.Unlock()

		forks[right].Unlock()
		forks[left].Unlock()
	}
}

// vigia: Se a função demorar mais que 10 segundos parada, é considerado deadlock e mata a função
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