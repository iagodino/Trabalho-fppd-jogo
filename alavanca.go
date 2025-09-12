// alavanca.go - Alavanca/Chaveador
package main

import (
	"fmt"
	"strconv"
)

// Alavanca que envia broadcast quando ativada
type Alavanca struct {
	id     string
	x, y   int
	ligada bool
	jogo   *Jogo
	parar  chan bool
}

// Cria nova alavanca
func novaAlavanca(x, y int, jogo *Jogo) *Alavanca {
	return &Alavanca{
		id:     "alavanca_" + strconv.Itoa(x) + "_" + strconv.Itoa(y),
		x:      x,
		y:      y,
		ligada: false,
		jogo:   jogo,
		parar:  make(chan bool),
	}
}

func (a *Alavanca) ObterID() string {
	return a.id
}

func (a *Alavanca) Iniciar() {
	go func() {
		canal := a.jogo.SistemaConcorrencia.ObterCanal(a.id)
		for {
			select {
			case msg := <-canal:
				a.ProcessarMensagem(msg)
			case <-a.parar:
				return
			}
		}
	}()
}

func (a *Alavanca) Parar() {
	close(a.parar)
}

func (a *Alavanca) ProcessarMensagem(msg Mensagem) {
	switch msg.Tipo {
	case "ativar":
		a.alternar()
	case "tick":
		// Poderia alternar automaticamente a cada X ticks
		// Implementação opcional para demonstrar múltiplos canais
	}
}

func (a *Alavanca) alternar() {
	a.ligada = !a.ligada
	a.atualizarVisual()
	
	// Broadcast para todos os elementos
	var tipoMsg string
	if a.ligada {
		tipoMsg = "ativar"
	} else {
		tipoMsg = "desativar"
	}
	
	// Envia mensagem para todos os elementos
	a.jogo.SistemaConcorrencia.EnviarMensagem(Mensagem{
		Tipo:    tipoMsg,
		Origem:  a.id,
		Destino: "*", // Broadcast
		Dados:   a.ligada,
	})
	
	status := "desligada"
	if a.ligada {
		status = "ligada"
	}
	a.jogo.StatusMsg = fmt.Sprintf("Alavanca %s!", status)
}

func (a *Alavanca) atualizarVisual() {
	if a.ligada {
		a.jogo.AtualizarElemento(a.x, a.y, AlavancaOn)
	} else {
		a.jogo.AtualizarElemento(a.x, a.y, AlavancaOff)
	}
}