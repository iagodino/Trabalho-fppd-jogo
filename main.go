// main.go - VERSÃO CORRIGIDA
package main

import (
    "fmt"
    "os"
    "strings"
    "time"
)

func main() {
    interfaceIniciar()
    defer interfaceFinalizar()

    mapaFile := "mapa.txt"
    var clienteRPC *RPCClient

    if len(os.Args) > 1 {
        arg := os.Args[1]
        
        if strings.Contains(arg, ":") {
            fmt.Printf("Conectando em servidor remoto\n")
            nome := fmt.Sprintf("Jogador_%d", (time.Now().UnixNano()%100))
            clienteRPC = NovoRPCClient(nome)
        } else if strings.HasSuffix(arg, ".txt") {
            mapaFile = arg
        }
    } else {
        fmt.Printf("Conectando em servidor local\n")
        nome := fmt.Sprintf("Jogador_%d", (time.Now().UnixNano()%100))
        clienteRPC = NovoRPCClient(nome)
    }

    jogo := jogoNovoComConcorrencia(mapaFile)
    defer jogo.Finalizar()

    jogo.rpc = clienteRPC

    if clienteRPC != nil {
        clienteRPC.LoopAtualizacoes(func(estado EstadoGlobal) {
            jogo.AtualizarOutrosJogadores(estado)
        })
        fmt.Printf("Sincronização RPC ativa!\n")
    }

    // CANAL PARA CONTROLAR A GOROUTINE
    done := make(chan bool)
    defer close(done)

    // GOROUTINE PROTEGIDA para movimento dos inimigos
    go func() {
        defer func() {
            if r := recover(); r != nil {
                fmt.Printf("Erro na goroutine de inimigos: %v\n", r)
            }
        }()
        
        for {
            select {
            case <-done:
                return  // ← Sai da goroutine quando programa termina
            default:
                // PROTEGER CONTRA PÂNICO
                func() {
                    defer func() {
                        if r := recover(); r != nil {
                            fmt.Printf("Erro no movimento de inimigo: %v\n", r)
                        }
                    }()
                    
                    interfaceDesenharJogo(jogo)
                    jogoMoverInimigo(jogo)
                }()
                
                time.Sleep(500 * time.Millisecond)
            }
        }
    }()

    interfaceDesenharJogo(jogo)

    for {
        evento := interfaceLerEventoTeclado()
        if continuar := personagemExecutarAcao(evento, jogo); !continuar {
            break
        }
        
        if jogo.rpc != nil {
            jogo.rpc.EnviarPosicao(jogo.PosX, jogo.PosY)
        }
        
        interfaceDesenharJogo(jogo)
    }
}