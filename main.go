// main.go - Loop principal do jogo
package main

import (
	"os"
	"time"
)

func main() {
	// Inicializa a interface (termbox)
	interfaceIniciar()
	defer interfaceFinalizar()

	// Usa "mapa.txt" como arquivo padrão ou lê o primeiro argumento
	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	// SUBSTITUIR por:
	jogo := jogoNovoComConcorrencia(mapaFile)
	defer jogo.Finalizar() // Garante que o sistema de concorrência será finalizado

	// Goroutine para movimento dos inimigos (mantém a funcionalidade existente)
	go func() {
		for {
			select {
			case modo := <-inimigoModoChange:
				inimigoModo = modo
			default:
				interfaceDesenharJogo(jogo)
				jogoMoverInimigo(jogo)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	// Desenha o estado inicial do jogo
	interfaceDesenharJogo(jogo)

	// Loop principal de entrada
	for {
		evento := interfaceLerEventoTeclado()
		if continuar := personagemExecutarAcao(evento, jogo); !continuar {
			break
		}
		interfaceDesenharJogo(jogo)
	}
}
