# Vibe Wallet

Self-hosted transaction tracker for importing bank CSVs, managing accounts, and categorizing spending with rules first and an optional OpenAI-compatible LLM fallback.

## Features

- Import bank transaction CSVs across multiple accounts.
- Browse, filter, edit, delete, and manually categorize transactions.
- Manage income/expense categories with colors and icons.
- Create rules for automatic categorization.
- Run rule-based categorization, with optional LLM categorization for unmatched transactions.
- Generate AI-suggested rules from previous LLM categorizations.
- Visualize income, expenses, net balance, and category breakdowns.
- Export/import rules and export account transactions.
- Configure light/dark mode and LLM settings in the UI.

## Setup

### Docker

```sh
curl -sSL https://raw.githubusercontent.com/yazdipour/vibe-wallet/main/docker-compose.yml -o docker-compose.yml
docker compose up -d
```

### Development

Requirements:

- Go 1.26+
- Node 22+
- npm

```sh
cd web
npm install
npm run build
cd ..
go run .
```

Open `http://localhost:8080`.

By default the app stores data in `vibe-wallet.db`. Override with:

```sh
DB_PATH=/path/to/vibe-wallet.db ADDR=:8080 go run .
```