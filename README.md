# TYOQ — Type Your Own Quotes

![Go 1.26.4](https://img.shields.io/badge/Go-1.26.4+-00ADD8?logo=go&logoColor=white)
![Python 3.6+](https://img.shields.io/badge/Python-3.6+-3776AB?logo=python&logoColor=white)
![License: MIT](https://img.shields.io/badge/license-MIT-green)

A MonkeyType-inspired typing CLI, a Goodreads quote scraper, and a PostgreSQL
seed pipeline.

## Usage

### Typing practice (Python)
```bash
python3 main.py        # random hardcoded quote
python3 main.py -i     # paste your own text
```

### Scrape quotes (Go)
```bash
go run ./cmd/scraper/   # writes quotes.jsonl (~3,000 quotes)
```

### Generate database seed (Go)
```bash
go run ./cmd/genseed/   # reads quotes.jsonl, writes init-db/02_seed.sql
```

### Database (Docker)
```bash
cp init-db/.env.database .env
docker compose up -d     # PostgreSQL on localhost:5432
```

## Components

| Component | Path | Description |
|----------|------|-------------|
| Typing CLI | `main.py` | `curses`-based typing practice with WPM/accuracy |
| Scraper | `internal/scraper/`, `cmd/scraper/` | Scrapes 3,000 Goodreads quotes to JSONL |
| Seed generator | `internal/genseed/`, `cmd/genseed/` | Generates PostgreSQL seed SQL from JSONL |
| Database | `docker-compose.yml`, `init-db/` | PostgreSQL 18 with auto-loaded schema + seed |

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

- [x] Webscraping Goodreads quotes
- [x] PostgreSQL database with seed data
- [ ] Wire the typing CLI to the database
- [ ] Filter quotes by author / category / tag
- [ ] Different themes for the CLI

## License

MIT — see [LICENSE](LICENSE).
