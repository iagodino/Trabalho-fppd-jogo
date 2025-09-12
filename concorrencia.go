// concorrencia.go - Sistema base de concorrência
package main

import (
	"sync"
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
	mutex     sync.RWMutex
}

// Cria novo sistema de concorrência
func novoSistemaConcorrencia() *SistemaConcorrencia {
	return &SistemaConcorrencia{
		elementos: make(map[string]ElementoConcorrente),
		canais:    make(map[string]chan Mensagem),
		ticker:    time.NewTicker(3 * time.Second), // Tick a cada 3 segundos
		parar:     make(chan bool),
	}
}

// Adiciona elemento ao sistema
func (s *SistemaConcorrencia) AdicionarElemento(elem ElementoConcorrente) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	id := elem.ObterID()
	s.elementos[id] = elem
	s.canais[id] = make(chan Mensagem, 10)
	
	elem.Iniciar()
}

// Envia mensagem para elemento específico ou broadcast
func (s *SistemaConcorrencia) EnviarMensagem(msg Mensagem) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if msg.Destino == "*" {
		// Broadcast para todos
		for id, canal := range s.canais {
			if id != msg.Origem {
				select {
				case canal <- msg:
				default: // Canal cheio, ignora
				}
			}
		}
	} else {
		// Mensagem específica
		if canal, existe := s.canais[msg.Destino]; existe {
			select {
			case canal <- msg:
			default: // Canal cheio, ignora
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
				// Envia tick para todos
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
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	for _, elem := range s.elementos {
		elem.Parar()
	}
	
	for _, canal := range s.canais {
		close(canal)
	}
}

// Obtém canal de um elemento
func (s *SistemaConcorrencia) ObterCanal(id string) chan Mensagem {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.canais[id]
}