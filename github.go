package main

import (
	"sort"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
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
		ghClient: github.NewClient(oauth2.NewClient(
			oauth2.NoContext,
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: config.GitHubToken},
			),
		)),
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
		_, _, _, err := s.ghClient.Repositories.GetContents(config.GitHubOrg, repo, ".travis.yml", &github.RepositoryContentGetOptions{})
		if err == nil {
			travisRepos = append(travisRepos, repo)
		}
	}

	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	s.cache.allRepos = allRepos
	s.cache.travisRepos = travisRepos
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
