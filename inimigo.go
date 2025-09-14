package main

func jogoMoverInimigo(jogo *Jogo) {
	jogoMutex.Lock()
	defer jogoMutex.Unlock()

	direcaoX := []int{1, 0, -1, 0} // Direções: direita, baixo, esquerda, cima
	direcaoY := []int{0, 1, 0, -1}
	novaDirecoes := make(map[[2]int]int)
	novaUltimoVisitado := make(map[[2]int]Elemento)
	visitados := make(map[[2]int]bool)

	mapaCopia := make([][]Elemento, len(jogo.Mapa))
	for i := range jogo.Mapa {
		mapaCopia[i] = make([]Elemento, len(jogo.Mapa[i]))
		copy(mapaCopia[i], jogo.Mapa[i])
	}

	// Percorre todo o mapa para encontrar inimigos
	for y, linha := range mapaCopia {
		for x, elem := range linha {
			posicao := [2]int{x, y}

			if (elem == Inimigo || elem == Inimigo2) && !visitados[posicao] {
				dir := inimigoDirecao[posicao]
				ultimo := inimigoUltimoVisitado[posicao]
				moveu := false
				if elem == Inimigo2 {
					// Inimigo Perseguidor
					// move em direção do jogador

					if inimigoModo == "perseguidor" {
						personagemPosicaoX, personagemPosicaoY := jogo.PosX, jogo.PosY
						var mdx, mdy int
						if x < personagemPosicaoX {
							mdx = 1
						} else if x > personagemPosicaoX {
							mdx = -1
						}
						if y < personagemPosicaoY {
							mdy = 1
						} else if y > personagemPosicaoY {
							mdy = -1
						}
						// Tenta mover primeiro no eixo X e depois no Y
						tentativas := [][2]int{{mdx, 0}, {0, mdy}}
						for _, t := range tentativas {
							novoX, novoY := x+t[0], y+t[1]
							novaposicao := [2]int{novoX, novoY}
							if jogoPodeMoverPara(jogo, novoX, novoY) && jogo.Mapa[novoY][novoX] != elem {
								jogo.Mapa[y][x] = ultimo
								novaUltimoVisitado[novaposicao] = jogo.Mapa[novoY][novoX]
								jogo.Mapa[novoY][novoX] = elem
								novaDirecoes[novaposicao] = dir
								visitados[novaposicao] = true
								moveu = true
								break

							}
						}
					} else {
						inimigoPatrulheiro(jogo, x, y, posicao, direcaoX, direcaoY, novaUltimoVisitado, novaDirecoes, visitados, moveu)
					}

				} else {
					inimigoPatrulheiro(jogo, x, y, posicao, direcaoX, direcaoY, novaUltimoVisitado, novaDirecoes, visitados, moveu)
				}
				if !moveu {
					// Se não conseguiu mover, inverte a direção
					inimigoDirecao[posicao] = (dir + 2) % 4
					novaDirecoes[posicao] = inimigoDirecao[posicao]
					novaUltimoVisitado[posicao] = inimigoUltimoVisitado[posicao]
					visitados[posicao] = true
				}
			}
		}
	}
	inimigoDirecao = novaDirecoes
	inimigoUltimoVisitado = novaUltimoVisitado

}

func inimigoPatrulheiro(jogo *Jogo, x, y int, posicao [2]int, direcaoX, direcaoY []int, novaUltimoVisitado map[[2]int]Elemento, novaDirecoes map[[2]int]int, visitados map[[2]int]bool, moveu bool) {
	// Inimigo Patrulheiro
	dir := inimigoDirecao[posicao]
	for i := 0; i < 4; i++ {
		d := (dir + i) % 4
		ultimo := inimigoUltimoVisitado[posicao]
		nx, ny := x+direcaoX[d], y+direcaoY[d]
		novaPosicao := [2]int{nx, ny}
		// Verifica se o movimento é permitido e realiza a movimentação
		if jogoPodeMoverPara(jogo, nx, ny) && jogo.Mapa[ny][nx] != Inimigo {
			// Move o inimigo
			jogo.Mapa[y][x] = ultimo
			novaUltimoVisitado[novaPosicao] = jogo.Mapa[ny][nx]
			jogo.Mapa[ny][nx] = Inimigo // move o elemento
			novaDirecoes[novaPosicao] = d
			visitados[novaPosicao] = true
			moveu = true
			break
		}
	}
}
