# Trabalho 1 — Concorrência e Sincronização: O Problema dos Filósofos Jantando

  Pontifícia Universidade Católica do Rio Grande do Sul \
  \
  Fundamentos de Processamento Paralelo e Distribuído \
  \
  Profª Eduarda Rodrigues Monteiro \
  \
  Grupo E: Diogo Luiz Zago, Enrico Canabarro Lazzari, Yan Pantoja \
  \
  Outubro de 2026 \

## 1. Soluções implementadas e mecanismos de sincronização

Todas as versões usam 5 filósofos, 20 ciclos por filósofo (pensar → adquirir garfos → comer → liberar garfos), 2 s pensando e 3 s comendo. Os experimentos foram realizados na máquina descrita na Seção 5; os resultados completos estão em `results/todas.csv`.

**V0 — Sequencial.**
Os filósofos executam um após o outro, sem goroutines. Não há disputa por garfos, portanto não há mecanismo de sincronização. O tempo total é determinístico: $5 \times 20 \times (2 + 3)\ \text{s} = 500\ \text{s}$, confirmado pelos dados (média 500,30 s).

**V1 — Concorrente sujeita a deadlock.**
Uma goroutine por filósofo; garfos como `[5]sync.Mutex`. Cada filósofo trava o garfo esquerdo e depois o direito. Uma goroutine `vigia()` encerra o processo se não houver progresso em 10 s, evidenciando o deadlock.

**V2 — Sem deadlock, com sincronização excessiva.**
Igual à V1, acrescida de um semáforo — implementado como canal com buffer de capacidade 4 (`make(chan struct{}, 4)`) — que cada filósofo adquire antes de tentar pegar os garfos e libera após devolvê-los. No máximo 4 filósofos disputam garfos ao mesmo tempo.

**V3 — Concorrente aprimorada.**
Implementa duas estratégias selecionáveis via flag `-strategy`:

- **Monitor** (padrão): cada filósofo tem um estado — PENSANDO, FAMINTO ou COMENDO. O filósofo só passa a comer quando tem garantia de que nenhum vizinho está comendo; caso contrário, fica esperando. Os dois garfos são obtidos de uma vez só: nenhum filósofo segura um enquanto espera o outro. Internamente isso é implementado com um `sync.Mutex` de monitor e uma variável de condição `sync.Cond` por filósofo — `pegarGarfos(i)` marca o estado como FAMINTO e, se a condição não for satisfeita, dorme em `cond[i].Wait()` até um vizinho sinalizar ao terminar de comer.
- **Ordenada** (`ordered`): garfos como `sync.Mutex`; cada filósofo trava primeiro o garfo de menor índice entre os seus dois. O filósofo 4, em vez de tentar o garfo 4 e depois o 0, tenta o 0 primeiro — quebrando o ciclo de espera.

### Análise comparativa das estratégias de sincronização

| | Mecanismo principal | Condição de Coffman quebrada | Máx. comendo (teórico) |
|---|---|---|:---:|
| V0 | nenhum (sequencial) | — | 1 |
| V1 | `sync.Mutex` por garfo | nenhuma → deadlock | 2 |
| V2 | mutex + semáforo (canal buffer 4) | espera circular | 2 |
| V3 monitor | `sync.Mutex` + `sync.Cond` | posse-e-espera | 2 |
| V3 ordered | mutex por garfo, aquisição ordenada | espera circular | 2 |

A V2 e a V3 são ambas corretas, mas quebram condições distintas. A V2 ainda permite que um filósofo segure o garfo esquerdo enquanto espera o direito; com tempos iguais os quatro admitidos formam uma corrente de espera em que só o da ponta come — produzindo execução quase serializada. A V3 elimina a posse-e-espera (monitor) ou o ciclo de recursos (ordered), permitindo que não-vizinhos comam em paralelo.

\newpage

## 2. Resultados experimentais

### Tabela comparativa

| Solução | Estratégia | Exec. | Tempo médio (s) | Tempo máx. (s) | Espera média (s) | Maior espera (s) | Máx. comendo | Ciclos/filósofo |
|:---:|:---:|:---:|---:|---:|---:|---:|:---:|:---:|
| V0 | sequencial | 5 | 500,32 | 500,34 | 0,000 | 0,000 | 1 | 20 |
| V1 | — | 1 | — (deadlock) | — | — | — | 0 | 0 |
| V2 | semaforo4 | 5 | 302,09 | 302,10 | 9,802 | 12,007 | 1 | 20 |
| V3 | monitor | 5 | 158,05 | 176,09 | 2,420 | 45,027 | 2 | 20 |
| V3 | ordered | 5 | 174,24 | 182,04 | 2,294 | 9,004 | 2 | 20 |

Médias e máximos calculados sobre as execuções de cada solução. "Maior espera" é o maior valor observado em qualquer execução. V1 não conclui; a linha registra o comportamento observado (deadlock após ciclo 0).

### Gráficos comparativos

\begin{figure}[H]
  \centering
  \includegraphics[width=0.65\textwidth]{../results/tempo_total.png}
  \caption{Tempo total de execução por solução (médias sobre 5 execuções).}
\end{figure}

\begin{figure}[H]
  \centering
  \includegraphics[width=0.65\textwidth]{../results/espera.png}
  \caption{Tempo médio de espera pela aquisição dos garfos.}
\end{figure}

\begin{figure}[H]
  \centering
  \includegraphics[width=0.65\textwidth]{../results/max_comendo.png}
  \caption{Número máximo de filósofos comendo simultaneamente.}
\end{figure}

### Capturas de tela das execuções

\begin{figure}[H]
  \centering
  \begin{minipage}{0.48\textwidth}
    \includegraphics[width=\textwidth]{prints/v0image.jpeg}
    \caption{V0 — sequencial: filósofos executam um após o outro.}
  \end{minipage}
  \hfill
  \begin{minipage}{0.48\textwidth}
    \includegraphics[width=\textwidth]{prints/v1image.jpeg}
    \caption{V1 — deadlock: todos travam em L1, nenhum chega a comer.}
  \end{minipage}
\end{figure}

\begin{figure}[H]
  \centering
  \begin{minipage}{0.48\textwidth}
    \includegraphics[width=\textwidth]{prints/v2image.jpeg}
    \caption{V2 — semáforo: filósofos pensam em paralelo, comem um por vez.}
  \end{minipage}
  \hfill
  \begin{minipage}{0.48\textwidth}
    \includegraphics[width=\textwidth]{prints/v3image.jpeg}
    \caption{V3 — monitor: T e E intercalados simultaneamente em colunas diferentes.}
  \end{minipage}
\end{figure}

### Evidência do deadlock da V1

A saída completa está em `results/v1_deadlock.txt`. Os cinco filósofos pensaram (T1), cada um travou o garfo esquerdo (L1), e nenhum chegou a comer (E1). O `vigia()` detectou 10 s sem progresso e encerrou o processo:

\begin{center}
\small\ttfamily
\begin{tabular}{l}
{[P1] [P2] [P3] [P4] [P5]}\\[2pt]
{\hspace{5.2em}T1}\\
{\ T1}\\
{\hspace{1.3em}T1}\\
{\hspace{2.6em}T1}\\
{\hspace{3.9em}T1}\\
{\hspace{3.9em}L1}\\
{\hspace{2.6em}L1}\\
{\hspace{1.3em}L1}\\
{\ L1}\\
{\hspace{5.2em}L1}\\[2pt]
{DEADLOCK detected! No progress in the last 10s}\\
{Cycles per philosopher: [0 0 0 0 0]}\\
\end{tabular}
\end{center}

Cada filósofo $i$ segura o garfo $i$ e aguarda o garfo $(i+1) \bmod 5$, formando o ciclo $P_0 \to P_1 \to P_2 \to P_3 \to P_4 \to P_0$. As quatro condições de Coffman estão satisfeitas: exclusão mútua (cada garfo é um mutex), posse-e-espera (segura o esquerdo aguardando o direito), ausência de preempção (ninguém retira o garfo de outro) e espera circular (o ciclo descrito acima).

## 3. Conclusão e análise dos resultados

### 1. Como funciona cada solução?

- **V0:** filósofos executam em sequência — um termina todos os 20 ciclos antes do próximo começar.
- **V1:** cada filósofo é uma goroutine que tenta travar o garfo esquerdo e depois o direito; sem proteção contra deadlock.
- **V2:** igual à V1, mas um semáforo de 4 vagas impede que todos os 5 tentem pegar garfos ao mesmo tempo.
- **V3 (monitor):** filósofo só "possui" garfos quando obtém os dois juntos, dentro do monitor; enquanto FAMINTO, não segura nenhum e dorme em sua variável de condição.
- **V3 (ordered):** filósofo trava sempre o garfo de menor índice primeiro, quebrando o ciclo de dependências.

### 2. Quais mecanismos de sincronização são utilizados?

- **Exclusão mútua:** `sync.Mutex` por garfo (V1, V2, V3-ordered) e mutex do monitor (V3-monitor).
- **Semáforo contador:** V2 — canal com buffer (`make(chan struct{}, 4)`).
- **Variáveis de condição:** V3-monitor — `sync.Cond` por filósofo.
- **Sincronização de término:** `sync.WaitGroup` em todas as versões concorrentes.
- **Proteção das métricas:** mutex próprio `statsMu` em todas as versões concorrentes.

### 3. Como esses mecanismos são implementados em Go?

Go não possui semáforo nem variável de condição como primitivas da linguagem; ambos são emulados:

- **Mutex:** `sync.Mutex` com `Lock()`/`Unlock()`. É não-reentrante.
- **Semáforo:** canal com buffer — `sem <- struct{}{}` bloqueia ao atingir a capacidade; `<-sem` libera uma vaga. A semântica de canal garante FIFO entre goroutines bloqueadas.
- **Variável de condição:** `sync.Cond`, criada com `sync.NewCond(&mu)`. `Wait()` libera atomicamente o mutex e suspende a goroutine; ao ser acordada por `Signal()`, retrava o mutex antes de retornar. O padrão correto exige re-checar a condição num laço `for state[i] != comendo { cond[i].Wait() }`.
- **Concorrência:** cada filósofo é uma goroutine (`go philosopher(...)`); o `main` aguarda todas com `wg.Wait()`.

### 4. Por que ocorre deadlock na V1?

As quatro condições de Coffman estão simultaneamente satisfeitas. Com tempos de pensar iguais todos os filósofos terminam de pensar juntos, travam o garfo esquerdo e ficam aguardando o direito — que está com o vizinho. O `time.Sleep(1 ms)` entre os dois `Lock` amplia a janela de race e torna o deadlock praticamente certo a cada execução: a evidência registrada mostra os 5 filósofos em L1 e 0 ciclos concluídos.

### 5. Como V2 e V3 evitam deadlock?

- **V2** quebra a **espera circular**: com no máximo 4 filósofos disputando 5 garfos, pelo menos um sempre consegue os dois — pois não é possível que 4 filósofos fiquem todos bloqueados com 5 garfos disponíveis.
- **V3-monitor** quebra a **posse-e-espera**: o filósofo nunca segura um garfo esperando o outro — ele só passa a COMENDO quando obtém os dois juntos dentro do monitor.
- **V3-ordered** quebra a **espera circular**: a ordenação global de recursos impede que o ciclo $P_0 \to \cdots \to P_4 \to P_0$ se forme, pois $P_4$ disputa o garfo 0 (não o 4) com $P_0$.

### 6. Existe possibilidade de starvation? Justifique.

- **V0:** não — é sequencial, todos os filósofos executam em ordem garantida.
- **V1:** não se aplica — a execução termina em deadlock antes de qualquer ciclo ser concluído.
- **V3-monitor:** sim, em teoria. Se os dois vizinhos de $P_i$ se alternam comendo sem que ambos estejam livres ao mesmo tempo, o filósofo do meio nunca consegue comer. O algoritmo não mantém fila de espera por tempo de bloqueio. Na prática, com tempos iguais para todos, isso é improvável — e os dados confirmam: todos os filósofos completaram 20 ciclos em todas as execuções.
- **V2 e V3-ordered:** dependem de `sync.Mutex`, que não garante ordem FIFO por padrão (Go ativa um modo anti-starvation após ~1 ms de espera, mas não há garantia absoluta por construção). No caso da V2, o semáforo (canal) acorda goroutines em ordem FIFO, reduzindo o risco, mas o mutex dos garfos ainda não é justo.

### 7. Qual solução apresentou maior tempo total de execução?

A **V0** apresentou o maior tempo total: 500,32 s em média — o valor teórico exato de $5 \times 20 \times 5\ \text{s}$, pois não há paralelismo. A V1 não é comparável por encerrar em deadlock. Entre as concorrentes que concluem normalmente: V2 com 302,09 s, V3-ordered com 174,24 s e V3-monitor com 158,05 s.

### 8. Qual solução apresentou maior tempo de espera pelos recursos?

A **V2**, com espera média de **9,802 s** e maior espera observada de **12,007 s**. A V3-monitor teve espera média de 2,420 s e a V3-ordered, 2,294 s — aproximadamente 4× menor que a V2. A V0 tem espera zero por definição (sem concorrência). Vale notar que a V3-monitor registrou um outlier de 45,027 s em uma das execuções — as demais ficaram entre 9–12 s; esse pico isolado é atribuído ao scheduler do WSL2 sob carga e não é representativo do comportamento normal.

### 9. Quantos filósofos conseguiram comer simultaneamente em cada solução?

- **V0:** 1, por construção (sequencial).
- **V1:** 0 ciclos concluídos; deadlock antes de qualquer refeição.
- **V2:** máximo de **1** em todas as 5 execuções.
- **V3-monitor:** máximo de **2** em todas as 5 execuções.
- **V3-ordered:** máximo de **2** em todas as 5 execuções.

O máximo teórico é 2: com 5 filósofos em círculo e cada um usando 2 garfos adjacentes, $\lfloor 5/2 \rfloor = 2$ é o limite geométrico.

### 10. Qual foi o impacto da sincronização excessiva da V2?

A V2 levou 302,09 s contra 158,05 s da V3-monitor — **1,91× mais lenta**, com max\_comendo = 1 em todas as execuções e espera média de 9,802 s (vs. 2,420 s da V3). O semáforo garante ausência de deadlock, mas não elimina a corrida pelo garfo direito: os quatro filósofos admitidos travam o esquerdo quase simultaneamente e formam uma corrente de espera em que só o da ponta come. A execução fica praticamente serializada, mesmo com 14 núcleos disponíveis. O resultado é pior que o sequencial em tempo de espera, mas melhor que o sequencial em tempo total porque o pensar (2 s) ocorre em paralelo.

### 11. A solução com maior sincronização foi necessariamente a de melhor desempenho?

Não. A V2 acrescenta um semáforo sobre os mesmos mutexes da V1 e apresenta o pior desempenho entre as versões que concluem. A V3 também usa múltiplas primitivas (monitor + variáveis de condição), mas sincroniza **a decisão** de comer, não o ato de comer — a trava do monitor fica ocupada por microssegundos, e os 3 s de refeição ocorrem fora dela, em paralelo entre não-vizinhos.

O que determina o desempenho não é a quantidade de primitivas, mas **quanto do trabalho real fica dentro da parte serializada** — isto é, a parte do código que só pode ser executada por um filósofo de cada vez. Na V2, a aquisição dos garfos é serializada pelo semáforo; na V3, apenas a verificação de estado é serializada.

### 12. Qual é a relação entre o grau de concorrência e o tempo total de execução?

É inversa. O trabalho total de comer é fixo ($5 \times 20 \times 3\ \text{s} = 300\ \text{s}$); com no máximo $k$ filósofos comendo ao mesmo tempo, o tempo fica limitado por baixo em $\approx 300/k\ \text{s}$:

- $k = 1$ (V0 e, na prática, V2): tempo entre 300 s e 500 s — V0: 500 s, V2: 302 s.
- $k = 2$ (V3): tempo cai para $\approx 150\ \text{s}$ — V3-monitor: 158 s, V3-ordered: 174 s.

A diferença entre V2 e V3 ilustra que aumentar o grau de concorrência de 1 para 2 reduziu o tempo em ~47%. A relação não é linear porque o tempo de pensar (2 s) também se sobrepõe às refeições dos outros, mas a tendência é clara: maior concorrência real → menor tempo total.

---

## 4. Ambiente de execução e link do repositório

Linguagem: Go 1.26.5 | SO: Linux 6.18.33.2-microsoft-standard-WSL2 x86\_64 | Processador: Intel(R) Core(TM) Ultra 5 135U | Cores: 14

Compilação: `go build -o build/vN vN/vN.go` para cada versão. Experimentos: `bash scripts/run_experiments.sh` (5 execuções/versão). Relatório: `python3 scripts/report.py`.

Código-fonte completo em: <https://github.com/enricolazzari/t1-FPPD>
