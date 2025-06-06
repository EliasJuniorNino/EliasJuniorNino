package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/joho/godotenv"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

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
