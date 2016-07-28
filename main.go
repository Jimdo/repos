package main

import (
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/Luzifer/rconfig"
	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
)

var (
	config struct {
		Address        string        `flag:"addr" default:":3000" description:"IP/Port to listen on"`
		GitHubToken    string        `flag:"github-token" env:"GITHUB_TOKEN" description:"The GitHub token for authentication at the GitHub API"`
		GitHubOrg      string        `flag:"github-org" env:"GITHUB_ORG" description:"The GitHub organization to fetch repositories from"`
		GitHubRepoType string        `flag:"github-repo-type" env:"GITHUB_REPO_TYPE" default:"private" description:"Type of GitHub repos to list. Options: all, public, private, forks, sources, member"`
		PollInterval   time.Duration `flag:"poll-interval" env:"POLL_INTERVAL" description:"Interval in which GitHub repos should be polled"`
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
	repoMetadataService := NewRepoMetadataService(RepoMetadataServiceConfig{
		GitHubToken:    config.GitHubToken,
		GitHubOrg:      config.GitHubOrg,
		GitHubRepoType: config.GitHubRepoType,
		PollInterval:   config.PollInterval,
	})
	s := NewServer(ServerConfig{
		Address:             config.Address,
		RepoMetadataService: repoMetadataService,
	})

	errors := repoMetadataService.StartPolling()
	go func() {
		err := <-errors
		log.Fatalf("Error polling GitHub repos: %s", err)
	}()

	r := mux.NewRouter()
	r.HandleFunc("/v1/all", s.HandleAllReposRequest)
	r.HandleFunc("/v1/travis", s.HandleTravisReposRequest)
	r.HandleFunc("/v1/healthcheck", s.HandleHealthCheckRequest)
	log.Fatal(http.ListenAndServe(config.Address, r))
}
