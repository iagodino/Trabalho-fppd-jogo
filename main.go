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

	// Cria o jogo com o mapa
	jogo := jogoNovoComConcorrencia(mapaFile)
	defer jogo.Finalizar() // Garante que o sistema de concorrência será finalizado

	// Nome padrão
	nome := "Jogador1"

	// Se foi passado um segundo argumento, usa ele como nome
	if len(os.Args) > 2 {
	nome = os.Args[2]
	}

	rpcClient := NovoRPCClient(nome)
	jogo.rpc = rpcClient
	// 💬 2. Inicia goroutine para receber atualizações periódicas do servidor
	rpcClient.LoopAtualizacoes(func(estado EstadoGlobal) {
		jogo.AtualizarOutrosJogadores(estado)
	})

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

		// 💬 3. Após cada ação, envia sua posição atual para o servidor
		if jogo.rpc != nil {
			jogo.rpc.EnviarPosicao(jogo.PosX, jogo.PosY)
		}

		interfaceDesenharJogo(jogo)
	}
}
