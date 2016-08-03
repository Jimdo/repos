package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"github.com/gregjones/httpcache"
)

const (
	defaultPollInterval = 5 * time.Minute
	defaultRepoType     = "private"
)

func NewRepoMetadataService(config RepoMetadataServiceConfig) *RepoMetadataService {
	pollInterval := defaultPollInterval
	if config.PollInterval != 0 {
		pollInterval = config.PollInterval
	}

	repoType := defaultRepoType
	if config.GitHubRepoType != "" {
		repoType = config.GitHubRepoType
	}

	return &RepoMetadataService{
		ghClient: github.NewClient(&http.Client{
			Transport: &oauth2.Transport{
				Base: httpcache.NewMemoryCacheTransport(),
				Source: oauth2.ReuseTokenSource(nil, oauth2.StaticTokenSource(
					&oauth2.Token{AccessToken: config.GitHubToken},
				)),
			},
		}),
		ghOrg:        config.GitHubOrg,
		ghRepoType:   repoType,
		pollInterval: pollInterval,
	}
}

type RepoMetadataServiceConfig struct {
	GitHubOrg      string
	GitHubRepoType string
	GitHubToken    string
	PollInterval   time.Duration
}

type RepoMetadataService struct {
	cacheLock sync.RWMutex
	cache     struct {
		allRepos    []string
		travisRepos []string
	}
	ghClient     *github.Client
	ghOrg        string
	ghRepoType   string
	pollInterval time.Duration
}

func (s *RepoMetadataService) StartPolling() <-chan error {
	t := time.NewTicker(s.pollInterval)
	errors := make(chan error)
	go func() {
		// Trigger polling initially
		if err := s.pollRepos(); err != nil {
			errors <- err
		}
		for {
			select {
			case <-t.C:
				if err := s.pollRepos(); err != nil {
					errors <- err
				}
			}
		}
	}()

	return errors
}

func (s *RepoMetadataService) pollRepos() error {
	var (
		opt = &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{PerPage: 30},
			Type:        config.GitHubRepoType,
		}
		allRepos    []string
		travisRepos []string
	)

	for {
		repos, resp, err := s.ghClient.Repositories.ListByOrg(config.GitHubOrg, opt)
		if err != nil {
			return err
		}
		for _, repo := range repos {
			allRepos = append(allRepos, *repo.Name)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	sort.Strings(allRepos)

	for _, repo := range allRepos {
		hasFile, err := s.hasFile(config.GitHubOrg, repo, ".travis.yml")
		if err != nil {
			return err
		}
		if hasFile {
			travisRepos = append(travisRepos, repo)
		}
	}

	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	s.cache.allRepos = s.prefixOrg(allRepos)
	s.cache.travisRepos = s.prefixOrg(travisRepos)
	return nil
}

func (s *RepoMetadataService) AllRepos() []string {
	s.cacheLock.RLock()
	defer s.cacheLock.RUnlock()
	return s.cache.allRepos
}

func (s *RepoMetadataService) TravisRepos() []string {
	s.cacheLock.RLock()
	defer s.cacheLock.RUnlock()
	return s.cache.travisRepos
}

// hasFile requests a file from a repository in raw format and returns true if
// the file exists.
// The reason for using the raw format is that these responses can be cached.
// Regular JSON responses contain one-time tokens and therefore different etags.
//
// Code is an adapted version of https://github.com/google/go-github/blob/56add9d4071bb1a06003a1239ebbc1500692b55d/github/repos_contents.go#L158-L169
func (s *RepoMetadataService) hasFile(owner, repo, filename string) (bool, error) {
	escapedPath := (&url.URL{Path: filename}).String()
	u := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, escapedPath)
	req, err := s.ghClient.NewRequest("GET", u, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3.raw")

	resp, err := s.ghClient.Do(req, nil)
	if resp != nil {
		return resp.StatusCode == http.StatusOK, nil
	}
	return false, err
}

func (s *RepoMetadataService) prefixOrg(repos []string) []string {
	var result []string
	for _, repo := range repos {
		result = append(result, fmt.Sprintf("%s/%s", s.ghOrg, repo))
	}
	return result
}
