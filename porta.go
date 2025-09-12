// porta.go - Porta Semáforo
package main

import (
	"strconv"
)

// Porta que abre/fecha em ciclos com override
type Porta struct {
	id       string
	x, y     int
	aberta   bool
	ticks    int  // Contador de ticks
	override bool // Se está em modo override
	jogo     *Jogo
	parar    chan bool
}

// Cria nova porta semáforo
func novaPorta(x, y int, jogo *Jogo) *Porta {
	return &Porta{
		id:     "porta_" + strconv.Itoa(x) + "_" + strconv.Itoa(y),
		x:      x,
		y:      y,
		aberta: false,
		ticks:  0,
		jogo:   jogo,
		parar:  make(chan bool),
	}
}

func (p *Porta) ObterID() string {
	return p.id
}

func (p *Porta) Iniciar() {
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

func (p *Porta) Parar() {
	close(p.parar)
}

// Escuta múltiplos canais: tick + override da alavanca
func (p *Porta) ProcessarMensagem(msg Mensagem) {
	switch msg.Tipo {
	case "tick":
		p.processarTick()
	case "ativar", "desativar":
		p.processarOverride(msg.Tipo == "ativar")
	}
}

func (p *Porta) processarTick() {
	if !p.override {
		p.ticks++
		// Alterna a cada 2 ticks (6 segundos)
		if p.ticks >= 2 {
			p.ticks = 0
			p.alternar()
		}
	}
}

func (p *Porta) processarOverride(forcarAbrir bool) {
	p.override = true
	p.aberta = forcarAbrir
	p.atualizarVisual()
	
	// Remove override após um tempo
	go func() {
		// Simula timeout do override (poderia usar timer real)
		p.override = false
		p.ticks = 0 // Reset contador
	}()
}

func (p *Porta) alternar() {
	p.aberta = !p.aberta
	p.atualizarVisual()
}

func (p *Porta) atualizarVisual() {
	if p.aberta {
		p.jogo.AtualizarElemento(p.x, p.y, PortaAberta) // Espaço vazio
	} else {
		p.jogo.AtualizarElemento(p.x, p.y, PortaFechada) // Parede bloqueando
	}
}