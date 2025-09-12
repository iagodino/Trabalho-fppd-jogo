// personagem.go - Funções para movimentação e ações do personagem
package main

import "fmt"

// Atualiza a posição do personagem com base na tecla pressionada (WASD)
func personagemMover(tecla rune, jogo *Jogo) {
	jogoMutex.Lock()
	dx, dy := 0, 0
	switch tecla {
	case 'w':
		dy = -1 // Move para cima
	case 'a':
		dx = -1 // Move para a esquerda
	case 's':
		dy = 1 // Move para baixo
	case 'd':
		dx = 1 // Move para a direita
	}

	nx, ny := jogo.PosX+dx, jogo.PosY+dy
	// Verifica se o movimento é permitido e realiza a movimentação
	if jogoPodeMoverPara(jogo, nx, ny) {
		// Verifica se está pisando em um portal aberto
		if jogo.Mapa[ny][nx] == PortalAberto {
			// Teleporta para uma posição segura no mapa
			personagemTeleportar(jogo)
		} else {
			// Movimento normal
			jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
			jogo.PosX, jogo.PosY = nx, ny
		}
	}
	defer jogoMutex.Unlock()
}

// Teleporta o personagem para uma posição aleatória segura
func personagemTeleportar(jogo *Jogo) {
	// Procura uma posição vazia para teleportar
	for tentativas := 0; tentativas < 100; tentativas++ {
		// Gera coordenadas aleatórias
		novoX := 10 + (tentativas*7)%50  // Varia X
		novoY := 5 + (tentativas*3)%20   // Varia Y
		
		// Verifica se a posição é válida e vazia
		if novoY >= 0 && novoY < len(jogo.Mapa) && 
		   novoX >= 0 && novoX < len(jogo.Mapa[novoY]) &&
		   jogo.Mapa[novoY][novoX] == Vazio {
			
			// Teleporta o personagem
			jogo.Mapa[jogo.PosY][jogo.PosX] = jogo.UltimoVisitado // Restaura posição anterior
			jogo.UltimoVisitado = jogo.Mapa[novoY][novoX]          // Guarda elemento da nova posição
			jogo.Mapa[novoY][novoX] = Personagem                   // Move personagem
			jogo.PosX, jogo.PosY = novoX, novoY                    // Atualiza coordenadas
			
			jogo.StatusMsg = fmt.Sprintf("Teleportado para (%d, %d)!", novoX, novoY)
			return
		}
	}
	
	// Se não encontrou posição, só move normalmente
	jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, 0, 0)
	jogo.StatusMsg = "Teleporte falhou - sem espaço disponível!"
}

// Define o que ocorre quando o jogador pressiona a tecla de interação
// Neste exemplo, apenas exibe uma mensagem de status
// Você pode expandir essa função para incluir lógica de interação com objetos
func personagemInteragir(jogo *Jogo) {
    // fmt.Printf("DEBUG: Tentando interagir em (%d, %d)\n", jogo.PosX, jogo.PosY)
    
    // Procura alavancas próximas (adjacentes ou na mesma posição)
    if jogo.SistemaConcorrencia != nil {
        for _, elem := range jogo.SistemaConcorrencia.elementos {
            if alavanca, ok := elem.(*Alavanca); ok {
                // Verifica se está próximo (1 casa de distância)
                dx := abs(alavanca.x - jogo.PosX)
                dy := abs(alavanca.y - jogo.PosY)
                
                if dx <= 1 && dy <= 1 {
                    // fmt.Printf("DEBUG: Ativando alavanca!\n")
                    // Envia mensagem para a alavanca
                    jogo.SistemaConcorrencia.EnviarMensagem(Mensagem{
                        Tipo:    "ativar",
                        Origem:  "jogador",
                        Destino: alavanca.id,
                    })
                    return
                }
            }
        }
    }
    
	// Se não encontrou alavanca próxima
	jogo.StatusMsg = fmt.Sprintf("Nada para interagir em (%d, %d)", jogo.PosX, jogo.PosY)
}

// Função auxiliar para valor absoluto
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Processa o evento do teclado e executa a ação correspondente
func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	switch ev.Tipo {
	case "sair":
		// Retorna false para indicar que o jogo deve terminar
		return false
	case "interagir":
		// Executa a ação de interação
		personagemInteragir(jogo)
	case "mover":
		// Move o personagem com base na tecla
		personagemMover(ev.Tecla, jogo)
	}
	return true // Continua o jogo
}
