package main

import (
	"log"
	"net/http"
	"sort"

	"golang.org/x/oauth2"

	"github.com/Luzifer/rconfig"
	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
)

var (
	config struct {
		Address        string `flag:"addr" default:":3000" description:"IP/Port to listen on"`
		GitHubToken    string `flag:"github-token" env:"GITHUB_TOKEN" description:"The GitHub token for authentication at the GitHub API"`
		GitHubOrg      string `flag:"github-org" env:"GITHUB_ORG" description:"The GitHub organization to fetch repositories from"`
		GitHubRepoType string `flag:"github-repo-type" env:"GITHUB_REPO_TYPE" default:"private" description:"Type of GitHub repos to list. Options: all, public, private, forks, sources, member"`
	}
	ghClient *github.Client
)

func init() {
	if err := rconfig.Parse(&config); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	ghClient = github.NewClient(oauth2.NewClient(
		oauth2.NoContext,
		oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GitHubToken},
		),
	))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/all", handleAllReposRequest)
	r.HandleFunc("/travis", handleTravisReposRequest)
	r.HandleFunc("/healthcheck", func(res http.ResponseWriter, r *http.Request) { http.Error(res, "OK", http.StatusOK) })
	http.ListenAndServe(config.Address, r)
}

func handleAllReposRequest(w http.ResponseWriter, req *http.Request) {
	repos, err := fetchRepoNames()
	if err != nil {
		jsonError(w, err)
		return
	}
	jsonResponse(w, repos)
}

func handleTravisReposRequest(w http.ResponseWriter, req *http.Request) {
	repos, err := fetchRepoNames()
	if err != nil {
		jsonError(w, err)
		return
	}
	repos, err = filterTravisRepos(repos)
	if err != nil {
		jsonError(w, err)
		return
	}
	jsonResponse(w, repos)
}

func fetchRepoNames() ([]string, error) {
	var (
		opt = &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{PerPage: 30},
			Type:        config.GitHubRepoType,
		}
		result []string
	)

	for {
		repos, resp, err := ghClient.Repositories.ListByOrg(config.GitHubOrg, opt)
		if err != nil {
			return []string{}, err
		}
		for _, repo := range repos {
			result = append(result, *repo.Name)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	sort.Strings(result)

	return result, nil
}

func filterTravisRepos(repos []string) ([]string, error) {
	var result []string
	for _, repo := range repos {
		_, _, _, err := ghClient.Repositories.GetContents(config.GitHubOrg, repo, ".travis.yml", &github.RepositoryContentGetOptions{})
		if err == nil {
			result = append(result, repo)
		}
	}
	return result, nil
}
