// concorrencia.go - Sistema base de concorrência
package main

import (
	"time"
)

// Mensagem representa comunicação entre elementos
type Mensagem struct {
	Tipo    string
	Origem  string
	Destino string
	Dados   interface{}
}

// ElementoConcorrente interface para todos os elementos
type ElementoConcorrente interface {
	Iniciar()
	Parar()
	ProcessarMensagem(msg Mensagem)
	ObterID() string
}

// SistemaConcorrencia gerencia todos os elementos
type SistemaConcorrencia struct {
	elementos map[string]ElementoConcorrente
	canais    map[string]chan Mensagem
	ticker    *time.Ticker
	parar     chan bool

	acaoChan chan func() // novo: canal para exclusão mútua
}

// Cria novo sistema de concorrência
func novoSistemaConcorrencia() *SistemaConcorrencia {
	s := &SistemaConcorrencia{
		elementos: make(map[string]ElementoConcorrente),
		canais:    make(map[string]chan Mensagem),
		ticker:    time.NewTicker(3 * time.Second),
		parar:     make(chan bool),
		acaoChan:  make(chan func(), 100),
	}

	// Goroutine que garante exclusão mútua
	go func() {
		for acao := range s.acaoChan {
			acao()
		}
	}()

	return s
}

// Adiciona elemento ao sistema
func (s *SistemaConcorrencia) AdicionarElemento(elem ElementoConcorrente) {
	s.acaoChan <- func() {
		id := elem.ObterID()
		s.elementos[id] = elem
		s.canais[id] = make(chan Mensagem, 10)
		elem.Iniciar()
	}
}

// Envia mensagem para elemento específico ou broadcast
func (s *SistemaConcorrencia) EnviarMensagem(msg Mensagem) {
	s.acaoChan <- func() {
		if msg.Destino == "*" {
			// Broadcast
			for id, canal := range s.canais {
				if id != msg.Origem {
					select {
					case canal <- msg:
					default:
					}
				}
			}
		} else {
			// Mensagem específica
			if canal, existe := s.canais[msg.Destino]; existe {
				select {
				case canal <- msg:
				default:
				}
			}
		}
	}
}

// Inicia o sistema (loop de tick)
func (s *SistemaConcorrencia) Iniciar() {
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.EnviarMensagem(Mensagem{
					Tipo:    "tick",
					Origem:  "sistema",
					Destino: "*",
				})
			case <-s.parar:
				return
			}
		}
	}()
}

// Para o sistema
func (s *SistemaConcorrencia) Parar() {
	s.parar <- true
	s.ticker.Stop()

	s.acaoChan <- func() {
		for _, elem := range s.elementos {
			elem.Parar()
		}
		for _, canal := range s.canais {
			close(canal)
		}
	}

	close(s.acaoChan)
}

// Obtém canal de um elemento
func (s *SistemaConcorrencia) ObterCanal(id string) chan Mensagem {
	res := make(chan chan Mensagem, 1)

	s.acaoChan <- func() {
		if canal, existe := s.canais[id]; existe {
			res <- canal
		} else {
			res <- nil
		}
	}

	return <-res
}
