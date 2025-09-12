// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main

import (
	"bufio"
	"os"
	"sync"
)

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool // Indica se o elemento bloqueia passagem
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Mapa           [][]Elemento // grade 2D representando o mapa
	PosX, PosY     int          // posição atual do personagem
	UltimoVisitado Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg      string       // mensagem para a barra de status
}

// Elementos visuais do jogo
var (
	Personagem            = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo               = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede                = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao             = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio                 = Elemento{' ', CorPadrao, CorPadrao, false}
	inimigoDirecao        = make(map[[2]int]int)      // posição [x,y] -> direção (0: direita, 1: baixo, 2: esquerda, 3: cima)
	inimigoUltimoVisitado = make(map[[2]int]Elemento) // posição [x,y] -> último elemento visitado
	jogoMutex             sync.Mutex                  // mutex para sincronizar o acesso ao estado do jogo
)

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
	// O ultimo elemento visitado é inicializado como vazio
	// pois o jogo começa com o personagem em uma posição vazia
	return Jogo{UltimoVisitado: Vazio}
}

// Lê um arquivo texto linha por linha e constrói o mapa do jogo
func jogoCarregarMapa(nome string, jogo *Jogo) error {
	arq, err := os.Open(nome)
	if err != nil {
		return err
	}
	defer arq.Close()

	scanner := bufio.NewScanner(arq)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []Elemento
		for x, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.simbolo:
				e = Parede
			case Inimigo.simbolo:
				e = Inimigo
			case Vegetacao.simbolo:
				e = Vegetacao
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y // registra a posição inicial do personagem
			}
			linhaElems = append(linhaElems, e)
		}
		jogo.Mapa = append(jogo.Mapa, linhaElems)
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			if elem == Inimigo {
				inimigoDirecao[[2]int{x, y}] = 0 // começa indo para a direita
				inimigoUltimoVisitado[[2]int{x, y}] = Vazio
			}
		}
	}

	return nil
}

// Verifica se o personagem pode se mover para a posição (x, y)
func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	// Verifica se a coordenada Y está dentro dos limites verticais do mapa
	if y < 0 || y >= len(jogo.Mapa) {
		return false
	}

	// Verifica se a coordenada X está dentro dos limites horizontais do mapa
	if x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}

	// Verifica se o elemento de destino é tangível (bloqueia passagem)
	if jogo.Mapa[y][x].tangivel {
		return false
	}

	// Pode mover para a posição
	return true
}

// Move um elemento para a nova posição
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	nx, ny := x+dx, y+dy

	// Obtem elemento atual na posição
	elemento := jogo.Mapa[y][x] // guarda o conteúdo atual da posição

	jogo.Mapa[y][x] = jogo.UltimoVisitado   // restaura o conteúdo anterior
	jogo.UltimoVisitado = jogo.Mapa[ny][nx] // guarda o conteúdo atual da nova posição
	jogo.Mapa[ny][nx] = elemento            // move o elemento
}

func jogoMoverInimigo(jogo *Jogo) {
	jogoMutex.Lock()

	dx := []int{1, 0, -1, 0} // Direções: direita, baixo, esquerda, cima
	dy := []int{0, 1, 0, -1}
	novaDirecoes := make(map[[2]int]int)
	novaUltimoVisitado := make(map[[2]int]Elemento)
	visitados := make(map[[2]int]bool)

	mapaCopia := make([][]Elemento, len(jogo.Mapa))
	for i := range jogo.Mapa {
		mapaCopia[i] = make([]Elemento, len(jogo.Mapa[i]))
		copy(mapaCopia[i], jogo.Mapa[i])
	}
	// Percorre todo o mapa para encontrar inimigos
	for y, linha := range mapaCopia {
		for x, elem := range linha {
			if elem == Inimigo && !visitados[[2]int{x, y}] {
				pos := [2]int{x, y}
				dir := inimigoDirecao[pos]
				moveu := false
				for i := 0; i < 4; i++ {
					d := (dir + i) % 4
					ultimo := inimigoUltimoVisitado[pos]
					nx, ny := x+dx[d], y+dy[d]
					// Verifica se o movimento é permitido e realiza a movimentação
					if jogoPodeMoverPara(jogo, nx, ny) && jogo.Mapa[ny][nx] != Inimigo {
						// Move o inimigo
						jogo.Mapa[y][x] = ultimo
						novaUltimoVisitado[[2]int{nx, ny}] = jogo.Mapa[ny][nx]
						jogo.Mapa[ny][nx] = Inimigo // move o elemento
						novaDirecoes[[2]int{nx, ny}] = d
						visitados[[2]int{nx, ny}] = true
						moveu = true
						break
					}

				}
				if !moveu {
					// Se não conseguiu mover, inverte a direção
					inimigoDirecao[pos] = (dir + 2) % 4
					novaDirecoes[pos] = inimigoDirecao[pos]
					novaUltimoVisitado[pos] = inimigoUltimoVisitado[pos]
					visitados[pos] = true
				}
			}
		}
	}
	inimigoDirecao = novaDirecoes
	inimigoUltimoVisitado = novaUltimoVisitado
	defer jogoMutex.Unlock()
}
