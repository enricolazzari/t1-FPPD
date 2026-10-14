// V3 – Filósofos jantando, versão concorrente aprimorada.
//
// Estratégia padrão "monitor" (solução de Dijkstra/Tanenbaum com estados):
// cada filósofo é PENSANDO, FAMINTO ou COMENDO, e só passa a COMENDO quando
// nenhum dos dois vizinhos está comendo. Os dois garfos são "pegos" juntos,
// dentro do monitor: ou o filósofo tem os dois, ou não tem nenhum.
//
// Variante "-strategy=ordered": garfos como sync.Mutex, cada filósofo pega
// primeiro o garfo de MENOR índice (ordenação de recursos).
//
// As métricas (totalWait, maxWait, maxEating, cycles) são calculadas do mesmo
// jeito que na V1/V2, para a tabela comparar a mesma coisa.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

const philosophers = 5

// Mesmos tempos e ciclos de V0–V2 (2 s pensando, 3 s comendo, 20 ciclos).
var (
	strategy = flag.String("strategy", "monitor", "monitor | ordered")
	think    = flag.Duration("think", 2*time.Second, "tempo pensando")
	eat      = flag.Duration("eat", 3*time.Second, "tempo comendo")
	rounds   = flag.Int("cycles", 20, "ciclos por filósofo")
	runs     = flag.Int("runs", 1, "execuções no mesmo processo (uma linha CSV cada)")
	csvPath  = flag.String("csv", "", "também grava as linhas CSV neste arquivo")
	verbose  = flag.Bool("v", false, "log por ciclo (distorce o tempo; desligado por padrão)")
	env      = flag.Bool("env", false, "imprime o ambiente de execução e sai")
)

// ---------- Estratégia "monitor" ----------

const (
	pensando = iota
	faminto
	comendo
)

var (
	mu    sync.Mutex               // trava do monitor: protege state
	state [philosophers]int        // estado de cada filósofo
	cond  [philosophers]*sync.Cond // onde cada filósofo dorme até poder comer
)

func esq(i int) int { return (i + philosophers - 1) % philosophers }
func dir(i int) int { return (i + 1) % philosophers }

// testar libera o filósofo i para comer se ele está com fome e nenhum vizinho
// está comendo. Só é chamado com mu travado.
func testar(i int) {
	if state[i] == faminto && state[esq(i)] != comendo && state[dir(i)] != comendo {
		state[i] = comendo
		cond[i].Signal()
	}
}

// pegarGarfos: o filósofo nunca segura UM garfo esperando o outro.
// Isso quebra a condição de Coffman "posse-e-espera" — e sem ela não existe
// deadlock, mesmo com exclusão mútua, sem preempção e com a mesa em círculo.
func pegarGarfos(i int) {
	mu.Lock()
	state[i] = faminto
	testar(i)
	for state[i] != comendo { // padrão de Cond: ao acordar, re-checa a condição
		cond[i].Wait() // solta mu enquanto dorme e retrava ao acordar
	}
	mu.Unlock()
}

// soltarGarfos: ao parar de comer, dá a chance a cada vizinho que estava com
// fome e esperava justamente por este filósofo.
func soltarGarfos(i int) {
	mu.Lock()
	state[i] = pensando
	testar(esq(i))
	testar(dir(i))
	mu.Unlock()
}

// ---------- Estratégia "ordered" ----------

var forks [philosophers]sync.Mutex

// A ORDEM dos Lock é o que importa: se todos pegam primeiro o menor índice,
// o filósofo 4 disputa o garfo 0 com o filósofo 0 em vez de pegar o 4 —
// a cadeia "cada um segura um e espera o do lado" não fecha o círculo.
// Isso quebra a condição de Coffman "espera circular".
func pegarOrdenado(i int) {
	a, b := i, dir(i)
	if b < a {
		a, b = b, a
	}
	forks[a].Lock()
	forks[b].Lock()
}

// A ordem dos Unlock não importa: soltar nunca bloqueia, então não cria espera.
func soltarOrdenado(i int) {
	forks[i].Unlock()
	forks[dir(i)].Unlock()
}

// ---------- Filósofo ----------

var pegar, soltar func(int)

func philosopher(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for round := 1; round <= *rounds; round++ {
		logf(id, "T%d", round)
		time.Sleep(*think) // pensando

		waitStart := time.Now()
		pegar(id) // sem Sleep entre aquisições (a V2 tem 1 ms ali)
		waitDuration := time.Since(waitStart)

		// registra a espera e quantos estão comendo (igual V1/V2)
		statsMu.Lock()
		totalWait += waitDuration
		maxWait = max(maxWait, waitDuration)
		currentlyEating++
		maxEating = max(maxEating, currentlyEating)
		statsMu.Unlock()

		logf(id, "E%d", round)
		time.Sleep(*eat) // comendo

		statsMu.Lock()
		currentlyEating--
		cycles[id]++
		statsMu.Unlock()

		soltar(id)
	}
}

func logf(id int, format string, a ...any) {
	if *verbose {
		fmt.Printf("%s "+format+"\n", append([]any{strings.Repeat(" ", id*5)}, a...)...)
	}
}

// ---------- Medição (mesmas variáveis da V1/V2) ----------

var (
	statsMu                    sync.Mutex
	totalWait, maxWait         time.Duration
	currentlyEating, maxEating int
	cycles                     [philosophers]int
)

func ms(d time.Duration) float64 { return float64(d) / float64(time.Millisecond) }

func main() {
	flag.Parse()
	if *env {
		printEnv()
		return
	}
	switch *strategy {
	case "monitor":
		pegar, soltar = pegarGarfos, soltarGarfos
	case "ordered":
		pegar, soltar = pegarOrdenado, soltarOrdenado
	default:
		fmt.Fprintln(os.Stderr, "strategy deve ser monitor ou ordered")
		os.Exit(2)
	}
	for i := range cond {
		cond[i] = sync.NewCond(&mu) // todas as Cond usam a MESMA trava do monitor
	}

	var out *os.File
	if *csvPath != "" {
		var err error
		if out, err = os.Create(*csvPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer out.Close()
		fmt.Fprintln(out, "versao,estrategia,execucao,tempo_total_ms,espera_media_ms,espera_max_ms,max_comendo,ciclos_f0,ciclos_f1,ciclos_f2,ciclos_f3,ciclos_f4")
	}

	for run := 1; run <= *runs; run++ {
		totalWait, maxWait, currentlyEating, maxEating = 0, 0, 0, 0
		cycles = [philosophers]int{}

		var wg sync.WaitGroup
		start := time.Now()
		for j := 0; j < philosophers; j++ {
			wg.Add(1)
			go philosopher(j, &wg)
		}
		wg.Wait()
		elapsed := time.Since(start)

		row := fmt.Sprintf("v3,%s,%d,%.3f,%.3f,%.3f,%d,%d,%d,%d,%d,%d",
			*strategy, run, ms(elapsed), ms(totalWait/time.Duration(philosophers**rounds)),
			ms(maxWait), maxEating, cycles[0], cycles[1], cycles[2], cycles[3], cycles[4])
		fmt.Println("CSV:" + row) // prefixo CSV: separa a linha do log no stdout
		if out != nil {
			fmt.Fprintln(out, row)
		}
	}
}

func printEnv() {
	cpu := run("sysctl", "-n", "machdep.cpu.brand_string") // macOS
	if cpu == "" {
		cpu = run("sh", "-c", "grep -m1 'model name' /proc/cpuinfo | cut -d: -f2") // Linux
	}
	fmt.Println("Linguagem:   Go", strings.TrimPrefix(runtime.Version(), "go"))
	fmt.Println("SO:         ", run("uname", "-srm"), "("+runtime.GOOS+"/"+runtime.GOARCH+")")
	fmt.Println("Processador:", cpu)
	fmt.Println("Cores:      ", runtime.NumCPU())
	fmt.Println("Compilar:    go build -o build/v3 v3/v3.go")
	fmt.Println("Executar:    ./build/v3 [-strategy=monitor|ordered] [-runs=5] [-csv=results/v3.csv] [-v]")
}

func run(name string, args ...string) string {
	out, _ := exec.Command(name, args...).Output()
	return strings.TrimSpace(string(out))
}
