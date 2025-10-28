package main

import (
    "fmt"
    "net"
    "net/rpc"
    "sync"
)

type Jogador struct {
    Nome   string
    X, Y   int
    SeqNum int
}

type EstadoGlobal struct {
    Jogadores map[string]Jogador
}

type Comando struct {
    Nome   string
    X, Y   int
    SeqNum int
}

type Servidor struct {
    mu        sync.Mutex
    jogadores map[string]Jogador
}

func (s *Servidor) RegistrarJogador(nome string, reply *bool) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, existe := s.jogadores[nome]; !existe {
        s.jogadores[nome] = Jogador{Nome: nome, X: 1, Y: 1, SeqNum: 0}
        fmt.Println("Novo jogador:", nome)
    }
    *reply = true
    return nil
}

func (s *Servidor) AtualizarPosicao(cmd Comando, reply *bool) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    j := s.jogadores[cmd.Nome]
    if cmd.SeqNum <= j.SeqNum {
        *reply = true
        return nil
    }

    j.X, j.Y = cmd.X, cmd.Y
    j.SeqNum = cmd.SeqNum
    s.jogadores[cmd.Nome] = j

    fmt.Printf("Movimento %s → (%d, %d)\n", j.Nome, j.X, j.Y)
    *reply = true
    return nil
}

func (s *Servidor) ObterEstadoGlobal(vazio bool, estado *EstadoGlobal) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    estado.Jogadores = make(map[string]Jogador)
    for k, v := range s.jogadores {
        estado.Jogadores[k] = v
    }
    return nil
}

func main() {
    srv := &Servidor{jogadores: make(map[string]Jogador)}
    rpc.Register(srv)
    l, _ := net.Listen("tcp", ":12345")
    fmt.Println("Servidor RPC escutando na porta 12345...")
    rpc.Accept(l)
}
