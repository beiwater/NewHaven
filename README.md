# New Haven

An online economy simulation game — build, trade, and compete in a player-driven market.

| Directory | Description |
|-----------|-------------|
| `backend/` | Go API server (game logic, trading, production, finance) |
| `client/`  | React + PixiJS frontend |
| `assets/`  | UI and game art resources |
| `docs/`    | Design docs, API contract, game wiki |
| `scripts/` | Utility scripts |

See [INDEX.md](INDEX.md) for the full project structure.

## Run locally

Requirements: Go 1.25+, Node.js 22+, and npm.

```bash
cd client/atlas-foods-client && npm install && cd ../..
node start.mjs
```

The launcher starts the API on `http://127.0.0.1:8088` and the frontend on `http://127.0.0.1:5173`. Development login: `dev` / `123`.

## Contributing

Pull requests are welcome. Before submitting, please read and sign our [Contributor License Agreement](CLA.md). All contributors must agree to the CLA — it grants the project the right to use, modify, and commercialize your contributions while you retain ownership of your work.

## So far, I’ve spent around 140 RMB in total — ChatGPT Plus (80 RMB) + DeepSeek V4 Flash and Pro API (around 60 RMB). So if I’ve really helped you, please consider giving me a star. Thank you ❤️
