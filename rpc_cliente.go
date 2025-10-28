package main

import (
    "fmt"
    "net/rpc"
    "time"
)

type RPCClient struct {
    client *rpc.Client
    nome   string
    seq    int
}

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

func NovoRPCClient(nome string) *RPCClient {
    c, err := rpc.Dial("tcp", "localhost:12345")
    if err != nil {
        panic(err)
    }
    client := &RPCClient{client: c, nome: nome}
    var ok bool
    c.Call("Servidor.RegistrarJogador", nome, &ok)
    return client
}

func (r *RPCClient) EnviarPosicao(x, y int) {
    r.seq++
    cmd := Comando{Nome: r.nome, X: x, Y: y, SeqNum: r.seq}
    var ok bool
    err := r.client.Call("Servidor.AtualizarPosicao", cmd, &ok)
    if err != nil {
        fmt.Println("Erro RPC:", err)
    }
}

func (r *RPCClient) LoopAtualizacoes(callback func(EstadoGlobal)) {
    go func() {
        for {
            var estado EstadoGlobal
            err := r.client.Call("Servidor.ObterEstadoGlobal", true, &estado)
            if err == nil {
                callback(estado)
            }
            time.Sleep(500 * time.Millisecond)
        }
    }()
}
