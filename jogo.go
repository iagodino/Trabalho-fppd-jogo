// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main

import (
	"bufio"
	"os"
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
	Mapa                [][]Elemento         // grade 2D representando o mapa
	PosX, PosY          int                  // posição atual do personagem
	UltimoVisitado      Elemento             // elemento que estava na posição do personagem antes de mover
	StatusMsg           string               // mensagem para a barra de status
	SistemaConcorrencia *SistemaConcorrencia // sistema de concorrência associado ao jogo

	acaoChan chan func() // canal de exclusão mútua
}

// Elementos visuais do jogo
var (
	inimigoModoChange = make(chan string)
	inimigoModo       = "patrulheiro"
	Personagem        = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo           = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede            = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao         = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio             = Elemento{' ', CorPadrao, CorPadrao, false}
	PortalAberto      = Elemento{'◯', CorVerde, CorPadrao, false}
	PortalFechado     = Elemento{'●', CorVermelho, CorPadrao, false}
	AlavancaOff       = Elemento{'⊥', CorVermelho, CorPadrao, false}
	AlavancaOn        = Elemento{'⊤', CorVerde, CorPadrao, false}
	PortaAberta       = Elemento{' ', CorPadrao, CorPadrao, false}
	PortaFechada      = Elemento{'║', CorAmarelo, CorPadrao, true}

	inimigoDirecao        = make(map[[2]int]int)
	inimigoUltimoVisitado = make(map[[2]int]Elemento)
)

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
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
			case Inimigo2.simbolo:
				e = Inimigo2
			case Vegetacao.simbolo:
				e = Vegetacao
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y
			case '●':
				e = PortalFechado
			case '◯':
				e = PortalAberto
			case '⊥':
				e = AlavancaOff
			case '║':
				e = PortaFechada
			case '⊤':
				e = AlavancaOn
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
				inimigoDirecao[[2]int{x, y}] = 0
				inimigoUltimoVisitado[[2]int{x, y}] = Vazio
			}
		}
	}

	return nil
}

// Verifica se o personagem pode se mover para a posição (x, y)
func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	if y < 0 || y >= len(jogo.Mapa) {
		return false
	}
	if x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}
	if jogo.Mapa[y][x].tangivel {
		return false
	}
	return true
}

// Move um elemento para a nova posição
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	nx, ny := x+dx, y+dy
	elemento := jogo.Mapa[y][x]
	jogo.Mapa[y][x] = jogo.UltimoVisitado
	jogo.UltimoVisitado = jogo.Mapa[ny][nx]
	jogo.Mapa[ny][nx] = elemento
}

func jogoNovoComConcorrencia(nomeArquivo string) *Jogo {
	jogo := &Jogo{
		UltimoVisitado:      Vazio,
		StatusMsg:           "Use WASD para mover, E para interagir, ESC para sair",
		SistemaConcorrencia: novoSistemaConcorrencia(),
		acaoChan:            make(chan func(), 100),
	}

	// Goroutine que garante exclusão mútua
	go func() {
		for acao := range jogo.acaoChan {
			acao()
		}
	}()

	// Carrega o mapa
	if err := jogoCarregarMapa(nomeArquivo, jogo); err != nil {
		panic(err)
	}

	jogo.inicializarElementos()
	return jogo
}

// Procura e inicializa todos os elementos concorrentes no mapa
func (j *Jogo) inicializarElementos() {
	j.SistemaConcorrencia.Iniciar()

	for y := 0; y < len(j.Mapa); y++ {
		for x := 0; x < len(j.Mapa[y]); x++ {
			elem := j.Mapa[y][x]
			switch elem {
			case PortalFechado:
				portal := novoPortal(x, y, j)
				j.SistemaConcorrencia.AdicionarElemento(portal)
			case AlavancaOff:
				alavanca := novaAlavanca(x, y, j)
				j.SistemaConcorrencia.AdicionarElemento(alavanca)
			case PortaFechada:
				porta := novaPorta(x, y, j)
				j.SistemaConcorrencia.AdicionarElemento(porta)
			}
		}
	}
}

// Executa ação exclusiva no estado do jogo
func (j *Jogo) executar(acao func()) {
	j.acaoChan <- acao
}

// Atualiza um elemento no mapa (thread-safe)
func (j *Jogo) AtualizarElemento(x, y int, novoElem Elemento) {
	j.executar(func() {
		if y >= 0 && y < len(j.Mapa) && x >= 0 && x < len(j.Mapa[y]) {
			j.Mapa[y][x] = novoElem
		}
	})
}

// Para o sistema de concorrência
func (j *Jogo) Finalizar() {
	if j.SistemaConcorrencia != nil {
		j.SistemaConcorrencia.Parar()
	}
	close(j.acaoChan)
}
