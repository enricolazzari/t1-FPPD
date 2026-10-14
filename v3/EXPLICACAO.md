# V3 explicada — material de estudo pra apresentação

## A ideia em uma frase

Um filósofo **só começa a comer quando os dois vizinhos não estão comendo**, e pega os dois garfos **de uma vez**. Assim ele nunca fica segurando um garfo enquanto espera o outro. Essa é a solução clássica de Dijkstra/Tanenbaum com estados, implementada como um **monitor** (uma trava + variáveis de condição).

## A comparação central (V2 × V3)

| | V2 (semáforo de 4) | V3 (monitor com estados) |
|---|---|---|
| Condição de Coffman que quebra | **Espera circular**: no máximo 4 disputam, então o círculo de 5 não fecha | **Posse-e-espera**: ninguém segura um garfo esperando o outro |
| Ainda deixa alguém segurar 1 garfo e esperar? | **Sim**: entra no semáforo, pega o esquerdo e espera o direito | **Não**: tem os dois ou nenhum |
| Efeito com tempos iguais (2 s / 3 s) | Os 4 que entram pegam o esquerdo ao mesmo tempo e formam uma **corrente**. Só o da ponta come, depois o próximo, e assim por diante: **1 comendo** | Dois não-vizinhos comem juntos: **2 comendo**, que é o máximo possível com 5 garfos (⌊5/2⌋) |

A V2 está correta e não trava, mas só quebrar a espera circular **não impede a corrente de espera**. A V3 ataca a causa da corrente, que é segurar um recurso enquanto espera outro.

## Linha a linha ([v3.go](v3.go))

**Flags (linhas ~28–37).** `-think=2s`, `-eat=3s` e `-cycles=20` são os mesmos valores de V0–V2. Viraram flags para eu testar rápido com `-think=20ms -eat=30ms` sem mudar o código. `-strategy` escolhe `monitor` (padrão) ou `ordered`.

**Estados (42–46).** `pensando`, `faminto` e `comendo`. "Faminto" é o estado que a V1 não tem: o filósofo **quer** comer, mas ainda **não pegou nada**.

**Monitor (48–52).**
- `mu` é **uma** trava que protege o vetor `state`. Toda leitura e escrita de estado acontece com ela travada, então não existe condição de corrida entre "olhar os vizinhos" e "mudar pra comendo".
- `cond[i]` é onde o filósofo *i* dorme enquanto não pode comer. São 5 `sync.Cond`, **todas criadas sobre a mesma `mu`** (`sync.NewCond(&mu)` no `main`). Isso é obrigatório, porque `Wait` precisa soltar exatamente a trava que protege a condição.

**`testar(i)` (59–64).** É o coração da solução. Se *i* está faminto e nenhum vizinho está comendo, *i* passa a `comendo` e recebe um `Signal` pra acordar. Só é chamado com `mu` travada.

**`pegarGarfos(i)` (69–77).**
1. Trava `mu` e marca `faminto`.
2. Chama `testar(i)`. Se os vizinhos estão livres, já vira `comendo` ali mesmo e nem dorme.
3. Senão, executa `for state[i] != comendo { cond[i].Wait() }`. O `Wait` **solta `mu` enquanto dorme** (senão ninguém conseguiria soltar garfos) e retrava ao acordar. É `for` e não `if` porque esse é o padrão de variável de condição: ao acordar, a condição é checada de novo.

**`soltarGarfos(i)` (81–87).** Marca `pensando` e chama `testar` nos **dois vizinhos**. Se um deles estava faminto esperando justamente por *i*, ele é liberado e acordado. Quem libera é sempre quem solta. Ninguém fica em espera ativa.

**`ordered` (91–110).** É a variante com `[5]sync.Mutex`.
- `pegarOrdenado` pega primeiro o garfo de **menor índice**. Para os filósofos 0–3 isso é esquerdo e depois direito. Para o 4, é garfo 0 e depois garfo 4. O 4 disputa o garfo 0 com o filósofo 0 **antes** de pegar o 4, então a cadeia nunca fecha. Isso quebra a **espera circular**, com a mesma ideia da V2 e outro mecanismo.
- A ordem dos `Lock` é o que importa. A dos `Unlock` não, porque soltar nunca bloqueia e portanto não cria espera.

**`philosopher` (116–144).** É o ciclo do enunciado: pensar → `pegar` → comer → `soltar`.
- A espera é medida de "pediu" (`waitStart` antes de `pegar`) até "tem os dois" (depois de `pegar`).
- **Não tem `Sleep` entre as aquisições.** A V2 tem 1 ms ali, o que aumenta a chance de todos pegarem o esquerdo antes de alguém pegar o direito.
- `currentlyEating--` acontece **antes** de `soltar`, então o contador nunca mostra mais gente comendo do que realmente está.

**Medição.** Usa as mesmas variáveis e o mesmo cálculo da V1/V2 (`statsMu`, `totalWait/(5*ciclos)`, `maxWait`, `maxEating`, `cycles`), só que a saída é uma linha `CSV:` por execução.

## Por que não tem deadlock

As 4 condições de Coffman precisam valer **juntas**:
- **Exclusão mútua:** continua, porque um garfo só serve a um filósofo.
- **Sem preempção:** continua, porque ninguém tira o garfo de quem está comendo.
- **Posse-e-espera:** **quebrada.** O filósofo só passa a "ter" garfos quando tem os dois. Enquanto espera (`faminto`), não tem nenhum.
- **Espera circular:** sem posse-e-espera, não há como montar o ciclo "P0 segura e espera P1, que segura e espera P2…".

Além disso, quem está esperando sempre é acordado. Só esperam filósofos com um vizinho comendo, e todo filósofo que come termina e chama `testar` nos dois vizinhos.

## Starvation: a resposta honesta

**Possível em teoria, improvável na prática com tempos iguais.**
- **Monitor.** Suponha P1 faminto. P0 come. Antes de P0 terminar, P2 começa a comer. P0 termina, mas `testar(1)` falha porque P2 está comendo. P0 volta a ter fome e come de novo antes de P2 terminar, e assim por diante. Os vizinhos se **alternam** e P1 nunca encontra os dois livres. Nada no algoritmo impede isso, porque ele não tem fila nem prioridade por idade. Com pensar = 2 s e comer = 3 s iguais pra todos, essa alternância perfeita não se sustenta. Os dados mostram isso: **todos concluem os 20 ciclos** em todas as execuções, e a "maior espera" mostra o pior caso real.
- **Ordered.** `sync.Mutex` **não é justo por padrão**: quem chega pode "furar" a fila de quem está dormindo. Depois de ~1 ms de espera, o mutex entra em *modo starvation* e passa a entregar em ordem FIFO. Por isso starvation também é improvável, mas não impossível por construção.
- **Pra eliminar de verdade**, seria preciso fila/ticket por idade de pedido (por exemplo, um vizinho faminto há mais tempo tem prioridade). Isso custa mais código e menos concorrência.

## E com 4 filósofos e 4 garfos?

- **Monitor:** funciona igual. Máximo de 2 comendo (P0+P2 ou P1+P3).
- **V1:** trava igual, com um círculo de 4.
- **Garçom (V2):** o semáforo teria que ser **N−1 = 3**. Com 4 seria igual a não ter semáforo, e voltaria a travar.
- **Ordered:** funciona igual, porque o último pega o garfo 0 primeiro.

## Por que o semáforo da V2 é 4 e não 5

Com 5 vagas, todos os 5 entram, cada um pega o esquerdo e todos esperam o direito: é exatamente a V1. Com **N−1 = 4**, sobra sempre um garfo livre entre os 4 competidores, então pelo menos um consegue o segundo garfo, come e libera.

## `monitor` × `ordered`

| | monitor (padrão) | ordered |
|---|---|---|
| Quebra | posse-e-espera | espera circular |
| Máx. comendo | 2 (garantido sempre que dá) | 2 é possível, mas ainda pode formar corrente parcial |
| Prós | Ninguém segura garfo à toa. É o máximo de concorrência do problema | Muito simples (2 linhas a mais que a V1), sem estado global |
| Contras | Uma trava global (o monitor) serializa as **decisões**, mas não o comer. Mais código | Garfos ficam presos enquanto se espera o segundo. Mutex do Go não é FIFO |

## Perguntas que a professora pode fazer

1. **Por que a V3 não trava?** Porque ninguém segura um garfo enquanto espera o outro: posse-e-espera está quebrada. Sem isso, o ciclo de espera não se forma.
2. **A trava global `mu` não deixa tudo sequencial como a V2?** Não. `mu` só fica travada nas decisões (microssegundos). O comer (3 s) acontece fora dela, e dois não-vizinhos comem em paralelo.
3. **Por que `for` e não `if` no `Wait`?** É o padrão de variável de condição: ao acordar, a condição é checada de novo. É seguro mesmo se alguém acordar sem a condição valer.
4. **Por que todas as `Cond` usam a mesma `mu`?** Porque a condição ("meus vizinhos não estão comendo") depende do `state` de vários filósofos, protegido por uma única trava. `Wait` precisa soltar essa trava.
5. **Por que 2 comendo é o máximo?** Cada um precisa de 2 dos 5 garfos, e ⌊5/2⌋ = 2. Um terceiro sempre seria vizinho de alguém.
6. **Existe starvation?** Em teoria sim (vizinhos alternando), mas é improvável com tempos iguais. Nos dados, todos fizeram 20 ciclos.
7. **Qual a diferença pra V2?** A V2 quebra espera circular, mas deixa a corrente "segura um e espera o outro", e por isso ficou com 1 comendo. A V3 quebra posse-e-espera e chega a 2 comendo.
8. **Por que a espera média cai, mas a maior espera pode não cair?** A média reflete a concorrência (2 comendo). O máximo depende de azar no escalonamento: um filósofo cujos vizinhos se revezaram algumas vezes. É o "starvation parcial" que a teoria prevê.
