# Jimdo/repos

Make repository metadata available via an HTTP API.

## API

* List of repos under the configured GitHub organization: `GET /v1/all`
* List of repos with a `.travis.yml` file under the configured GitHub organization: `GET /v1/travis`
* Healthcheck: `GET /v1/healthcheck`

## Usage

You can use the Docker containers availble in [Quay](https://quay.io/repository/jimdo/repos) to run `repos`:

```
docker run -d -p 80:3000 \
    -e GITHUB_TOKEN=<github token> \
    -e GITHUB_ORG=<github organization> \
    quay.io/jimdo/repos
```

## Configuration Options

```
$ repos -h
Usage of repos:
      --addr=":3000": IP/Port to listen on
      --github-org="": The GitHub organization to fetch repositories from
      --github-repo-type="private": Type of GitHub repos to list. Options: all, public, private, forks, sources, member
      --github-token="": The GitHub token for authentication at the GitHub API
      --poll-interval=5m0s: Interval in which GitHub repos should be polled
```
