// personagem.go - Funções para movimentação e ações do personagem
package main

import "fmt"

// Atualiza a posição do personagem com base na tecla pressionada (WASD)
func personagemMover(tecla rune, jogo *Jogo) {
	jogo.executar(func() {
		dx, dy := 0, 0
		switch tecla {
		case 'w':
			dy = -1
		case 'a':
			dx = -1
		case 's':
			dy = 1
		case 'd':
			dx = 1
		}

		nx, ny := jogo.PosX+dx, jogo.PosY+dy
		if jogoPodeMoverPara(jogo, nx, ny) {
			if jogo.Mapa[ny][nx] == PortalAberto {
				personagemTeleportar(jogo)
			} else {
				jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
				jogo.PosX, jogo.PosY = nx, ny
			}
		}
	})
}

// Teleporta o personagem para uma posição aleatória segura
func personagemTeleportar(jogo *Jogo) {
	jogo.executar(func() {
		for tentativas := 0; tentativas < 100; tentativas++ {
			novoX := 10 + (tentativas*7)%50
			novoY := 5 + (tentativas*3)%20
			if novoY >= 0 && novoY < len(jogo.Mapa) &&
				novoX >= 0 && novoX < len(jogo.Mapa[novoY]) &&
				jogo.Mapa[novoY][novoX] == Vazio {

				jogo.Mapa[jogo.PosY][jogo.PosX] = jogo.UltimoVisitado
				jogo.UltimoVisitado = jogo.Mapa[novoY][novoX]
				jogo.Mapa[novoY][novoX] = Personagem
				jogo.PosX, jogo.PosY = novoX, novoY

				jogo.StatusMsg = fmt.Sprintf("Teleportado para (%d, %d)!", novoX, novoY)
				return
			}
		}
		jogo.StatusMsg = "Teleporte falhou - sem espaço disponível!"
	})
}

// Define o que ocorre quando o jogador pressiona a tecla de interação
func personagemInteragir(jogo *Jogo) {
	if jogo.SistemaConcorrencia != nil {
		for _, elem := range jogo.SistemaConcorrencia.elementos {
			if alavanca, ok := elem.(*Alavanca); ok {
				dx := abs(alavanca.x - jogo.PosX)
				dy := abs(alavanca.y - jogo.PosY)
				if dx <= 1 && dy <= 1 {
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
	jogo.StatusMsg = fmt.Sprintf("Nada para interagir em (%d, %d)", jogo.PosX, jogo.PosY)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	switch ev.Tipo {
	case "sair":
		return false
	case "interagir":
		personagemInteragir(jogo)
	case "mover":
		personagemMover(ev.Tecla, jogo)
	}
	return true
}
