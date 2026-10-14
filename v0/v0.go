// Example code filo_seq.go - Sequential Dinning Philosophers
// Since is sequential, elapsed timne could be simply calculated by (roubnds*philosophers*(time_to_think+time_to_eat))
// By Cesar De Rose

package main

import (
	"fmt"
	"time"
)

func Philosopher(id, iteration int) {
	time_to_think := 2
	time_to_eat := 3

	// think
	for s := 1; s <= id*6; s++ {
		fmt.Printf(" ")
	}
	fmt.Printf(" T%d\n", iteration) // Thinking
	time.Sleep(time.Duration(time_to_think) * time.Second)

	// eat

	for s := 1; s <= id*6; s++ {
		fmt.Printf(" ")
	}
	fmt.Printf(" E%d\n", iteration) // Eating
	time.Sleep(time.Duration(time_to_eat) * time.Second)
}

func main() {
	philosophers := 5
	rounds := 20

	fmt.Println("\n[P1] [P2] [P3] [P4] [P5]\n")

	// run philosophers

	start := time.Now()

	for i := 1; i <= rounds; i++ {
		for j := 0; j < philosophers; j++ {
			Philosopher(j, i)
		}
	}
	elapsed := time.Since(start)

	fmt.Printf("\nSequential Dinner took %s\n\n", elapsed)

	// linha CSV para scripts/run_experiments.sh (sequencial: sem espera por garfos, 1 comendo por vez)
	fmt.Printf("CSV:v0,sequencial,1,%.3f,0,0,1,%d,%d,%d,%d,%d\n",
		float64(elapsed)/float64(time.Millisecond), rounds, rounds, rounds, rounds, rounds)
}
