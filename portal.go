// portal.go - Portal temporizado
package main

import (
	"fmt"
	"strconv"
	"time"
)

// Portal que abre/fecha com timeout
type Portal struct {
	id      string
	x, y    int
	aberto  bool
	jogo    *Jogo
	parar   chan bool
	timeout *time.Timer
}

// Cria novo portal
func novoPortal(x, y int, jogo *Jogo) *Portal {
	return &Portal{
		id:    "portal_" + strconv.Itoa(x) + "_" + strconv.Itoa(y),
		x:     x,
		y:     y,
		aberto: false,
		jogo:  jogo,
		parar: make(chan bool),
	}
}

func (p *Portal) ObterID() string {
	return p.id
}

func (p *Portal) Iniciar() {
	go func() {
		canal := p.jogo.SistemaConcorrencia.ObterCanal(p.id)
		for {
			select {
			case msg := <-canal:
				p.ProcessarMensagem(msg)
			case <-p.parar:
				return
			}
		}
	}()
}

func (p *Portal) Parar() {
	close(p.parar)
	if p.timeout != nil {
		p.timeout.Stop()
	}
}

func (p *Portal) ProcessarMensagem(msg Mensagem) {
	switch msg.Tipo {
	case "ativar":
		p.abrir()
	case "desativar":
		p.fechar()
	}
}

func (p *Portal) abrir() {
	if !p.aberto {
		p.aberto = true
		p.atualizarVisual()
		
		// Para timeout anterior
		if p.timeout != nil {
			p.timeout.Stop()
		}
		
		// Fecha automaticamente em 5 segundos
		p.timeout = time.AfterFunc(5*time.Second, func() {
			p.fechar()
		})
		
		p.jogo.StatusMsg = fmt.Sprintf("Portal aberto em (%d, %d)!", p.x, p.y)
	}
}

func (p *Portal) fechar() {
	if p.aberto {
		p.aberto = false
		p.atualizarVisual()
		p.jogo.StatusMsg = fmt.Sprintf("Portal fechado em (%d, %d)", p.x, p.y)
	}
}

func (p *Portal) atualizarVisual() {
	if p.aberto {
		p.jogo.AtualizarElemento(p.x, p.y, PortalAberto)
	} else {
		p.jogo.AtualizarElemento(p.x, p.y, PortalFechado)
	}
}