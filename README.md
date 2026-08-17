# awesome

Generates RSS, Atom and JSON feeds for links added to `README.md` files in GitHub repositories.

Useful for personal dashboards and feed readers when GitHub's commit feeds are too broad.

## Install

Download a pre-built binary for your platform from [GitHub Releases](https://github.com/coalaura/awesome/releases), then run it from a directory where it can store `config.yml` and `awesome.db`.

```sh
./awesome
```

On first run, it creates `config.yml` and exits because a GitHub token is required.

## Configure

Set `server.github_token` and configure the repositories to watch:

```yaml
server:
  port: 5883
  github_token: your-github-token

feeds:
  - type: go
    repository: avelino/awesome-go
    branch: main
  - type: selfhosted
    repository: awesome-selfhosted/awesome-selfhosted
    branch: master
```

The service checks configured repositories every 10 minutes. It stores processed commits in `awesome.db`.

## Feeds

With the example configuration:

| Feed | URL |
| --- | --- |
| RSS | `http://localhost:5883/go` or `http://localhost:5883/go/rss` |
| Atom | `http://localhost:5883/go/atom` |
| JSON Feed | `http://localhost:5883/go/json` |

Visit `http://localhost:5883/` to see configured feeds.