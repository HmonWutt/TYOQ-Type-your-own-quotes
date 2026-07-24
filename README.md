# TYOQ — Type Your Own Quotes

![Go 1.26.4](https://img.shields.io/badge/Go-1.26.4+-00ADD8?logo=go&logoColor=white)
![License: MIT](https://img.shields.io/badge/license-MIT-green)

A MonkeyType-inspired typing CLI

## Usage

### Typing practice (Go)
```bash
cd ./tools
go run ./cmd/typing/main.go # random quotes
go run ./cmd/typing/main.go -i     # paste your own text
```

<!-- ### Typing practice (Docker) -->
<!-- ```bash -->
<!-- docker compose run --rm -it app    #random quotes -->
<!-- ``` -->
## Components

| Component | Path | Description |
|-----------|------|-------------|
| Typing CLI | `tools/internal/typing/`, `tools/cmd/typing` | `Bubbletea TUI` with WPM/accuracy |
| Scraper | `tools/internal/scraper/`, `tools/cmd/scraper/` | Scrapes quotes to JSONL |
| Seed generator | `tools/internal/genseed/`, `tools/cmd/genseed/` | Generates sqlite seed file from JSONL |
| Database | `data/seed.db` | sqlite3 schema + seed |

## Data format

`quotes.jsonl` is [JSON Lines](https://jsonlines.org) — one JSON object per line:

```jsonl
{"text":"Be yourself; everyone else is already taken.","author":"Oscar Wilde","source":"","tags":null}
{"text":"So many books, so little time.","author":"Frank Zappa","source":"","tags":["books","humor"]}
```

| Field | Type | Notes |
|-------|------|-------|
| `text` | string | the quote body |
| `author` | string | e.g. `"Oscar Wilde"` |
| `source` | string | book/work title, often `""` |
| `tags` | array of strings | quote tags, or `null` |

## Roadmap
- [x] ~~Typing UI built in python using `curses` library~~
- [x] Webscraping quotes
- [x] ~~PostgreSQL database with seed data~~
- [x] Migrate to sqlite 
- [x] Wire the typing CLI to the database
- [x] Replace UI with Bubbletea
- [ ] Clean quotes
- [ ] Filter quotes by author / category / tag
- [ ] Different themes for the CLI

## License

MIT — see [LICENSE](LICENSE).
