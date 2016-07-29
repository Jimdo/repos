package main

import (
	"encoding/json"
	"net/http"
)

func NewServer(config ServerConfig) *Server {
	return &Server{
		address:             config.Address,
		repoMetadataService: config.RepoMetadataService,
	}
}

type ServerConfig struct {
	Address             string
	RepoMetadataService *RepoMetadataService
}

type Server struct {
	address             string
	repoMetadataService *RepoMetadataService
}

func (s *Server) HandleAllReposRequest(w http.ResponseWriter, req *http.Request) {
	jsonResponse(w, s.repoMetadataService.AllRepos())
}

func (s *Server) HandleTravisReposRequest(w http.ResponseWriter, req *http.Request) {
	jsonResponse(w, s.repoMetadataService.TravisRepos())
}

func (s *Server) HandleHealthCheckRequest(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("OK"))
}

type apiError struct {
	Error string `json:"error"`
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		jsonError(w, err)
	}
}

func jsonError(w http.ResponseWriter, err error) {
	apiErr := apiError{err.Error()}
	jsonResponse(w, apiErr)
}
