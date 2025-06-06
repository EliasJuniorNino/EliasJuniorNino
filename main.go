package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/joho/godotenv"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

type LangStat struct {
	Name  string
	Bytes int
}

type LanguageColor struct {
	Name  string
	Color string
}

func main() {
	// Carrega variáveis do .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente existentes (se houver)")
	}

	token := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_USERNAME")

	if token == "" || username == "" {
		log.Fatal("GITHUB_TOKEN e/ou GITHUB_USERNAME não definidos")
	}

	// Autenticação GitHub
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Buscar todos os repositórios do usuário
	var allRepos []*github.Repository
	opt := &github.RepositoryListOptions{Type: "owner", Sort: "updated", ListOptions: github.ListOptions{PerPage: 100}}
	for {
		repos, resp, err := client.Repositories.List(ctx, username, opt)
		if err != nil {
			log.Fatalf("Erro ao buscar repositórios: %v", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	fmt.Printf("Total de repositórios: %d\n", len(allRepos))

	// Contabilizar uso de linguagens
	languageTotals := make(map[string]int)
	for _, repo := range allRepos {
		if repo.GetFork() {
			continue
		}
		langs, _, err := client.Repositories.ListLanguages(ctx, username, repo.GetName())
		if err != nil {
			log.Printf("Erro ao buscar linguagens do repositório %s: %v", repo.GetName(), err)
			continue
		}
		for lang, size := range langs {
			languageTotals[lang] += size
		}
	}

	if len(languageTotals) == 0 {
		log.Println("Nenhuma linguagem detectada.")
		return
	}

	// Ordenar linguagens por tamanho
	var stats []LangStat
	var total int
	for k, v := range languageTotals {
		stats = append(stats, LangStat{k, v})
		total += v
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Bytes > stats[j].Bytes
	})
	if len(stats) > 10 {
		stats = stats[:10]
	}

	// Gerar imagem com grafico
	GenerateTopLangsChart(stats)
}

func GenerateTopLangsChart(stats []LangStat) {
	const width = 400
	const height = 320
	const barHeight = 8
	const itemSpacing = 26
	const startY = 80
	const marginX = 20

	// Cores modernas para linguagens populares
	languageColors := map[string]string{
		"JavaScript": "#f1e05a", // Amarelo JavaScript
		"TypeScript": "#3178c6", // Azul TypeScript
		"Python":     "#3572a5", // Azul Python
		"Java":       "#ed752a", // Laranja Java
		"Go":         "#00add8", // Ciano Go
		"PHP":        "#777bb4", // Roxo PHP
		"C++":        "#f34b7d", // Rosa C++
		"C":          "#555555", // Cinza C
		"Rust":       "#dea584", // Laranja Rust
		"Swift":      "#fa7343", // Laranja Swift
		"Kotlin":     "#7f52cc", // Roxo Kotlin
		"Dart":       "#00b4f0", // Azul Dart
		"Ruby":       "#cc342d", // Vermelho Ruby
		"CSS":        "#1572b6", // Azul CSS
		"HTML":       "#e34c26", // Laranja HTML
		"Shell":      "#89e051", // Verde Shell
		"Vue":        "#41b883", // Verde Vue
		"C#":         "#92d050", // Verde C#
		"SCSS":       "#cf649a", // Rosa SCSS
		"Less":       "#1d365d", // Azul escuro Less
	}

	// Cores padrão
	defaultColors := []string{
		"#ffcd3c", "#6c63ff", "#00bfff", "#ff69b4", "#ff5722",
		"#00c853", "#ff9800", "#3f51b5", "#607d8b", "#00968a",
	}

	// Soma total para calcular percentuais
	var total float64
	for _, s := range stats {
		total += float64(s.Bytes)
	}

	// Início do SVG
	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	svg.WriteString("\n")

	// Fundo
	svg.WriteString(`<rect width="100%" height="100%" fill="#262a33"/>`)
	svg.WriteString("\n")

	// Título
	svg.WriteString(`<text x="20" y="35" fill="#73d9ca" font-family="Arial, sans-serif" font-size="18" font-weight="bold">Most Used Languages</text>`)
	svg.WriteString("\n")

	// Barra principal - fundo
	barX := marginX
	barY := 55
	barW := width - 2*marginX
	svg.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="4" fill="#404854"/>`,
		barX, barY, barW, barHeight))
	svg.WriteString("\n")

	// Segmentos da barra principal
	x := float64(barX)
	for i, s := range stats {
		pct := float64(s.Bytes) / total
		w := pct * float64(barW)

		// Escolhe cor
		var color string
		if c, exists := languageColors[s.Name]; exists {
			color = c
		} else {
			color = defaultColors[i%len(defaultColors)]
		}

		// Desenha segmento
		if w > 1 { // Só desenha se for visível
			svg.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%d" width="%.2f" height="%d" fill="%s"/>`,
				x, barY, w, barHeight, color))
			svg.WriteString("\n")
		}
		x += w
	}

	// Lista de linguagens
	for i, s := range stats {
		y := startY + i*itemSpacing
		pct := 100 * float64(s.Bytes) / total

		// Escolhe cor
		var color string
		if c, exists := languageColors[s.Name]; exists {
			color = c
		} else {
			color = defaultColors[i%len(defaultColors)]
		}

		// Círculo colorido
		svg.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="6" fill="%s"/>`,
			marginX+8, y-4, color))
		svg.WriteString("\n")

		// Nome da linguagem
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" fill="white" font-family="Arial, sans-serif" font-size="13">%s</text>`,
			marginX+25, y, s.Name))
		svg.WriteString("\n")

		// Percentual
		pctText := fmt.Sprintf("%.2f%%", pct)
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" fill="#8b949e" font-family="Arial, sans-serif" font-size="12">%s</text>`,
			width-60, y, pctText))
		svg.WriteString("\n")
	}

	// Fecha SVG
	svg.WriteString("</svg>")

	// Cria diretório se necessário
	if _, err := os.Stat("output"); os.IsNotExist(err) {
		os.Mkdir("output", 0755)
	}

	// Salva SVG
	err := os.WriteFile("output/top_languages.svg", []byte(svg.String()), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Gráfico SVG gerado: output/top_languages.svg")
}
