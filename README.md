# Sand-ai on Nakama

Game backend for Sand-ai, a social multiplayer game, built on top of [Nakama](https://github.com/heroiclabs/nakama) by Heroic Labs.

This repository is a copy of the Nakama server (Apache-2.0, see `LICENSE`) with Sand-ai code added on top. Nakama itself provides users, auth, groups, chat, matchmaking, leaderboards and storage; the code under `src/`, `cmd/api/` and `deploy/sandai/` is the Sand-ai part. For anything about the Nakama server itself, see the [Nakama documentation](https://heroiclabs.com/docs).

## What Sand-ai adds

**Domain layer (`src/domain`)**, with no Nakama types inside:

- `tournament`: tournament and participant aggregates, entry rules and ranking
- `analytics`: tracked events and play sessions
- `battle`, `group`, `leaderboard`, `player`, `bot`: aggregates and repository ports

**Application layer (`src/app`)**: one service per context (auth, groups, battles, leaderboard, tournaments, analytics, bot) that works against the repository ports.

**Infrastructure (`src/infra`)**:

- `runtime/sandai_module.go`: a Nakama Go runtime module with auth hooks, leaderboard write validation and an authoritative `sandai_battle` match handler
- `tournament/`: an in-memory repository and a provider backed by Nakama tournaments
- `analytics/`: a Segment dispatcher and an in-memory session store

**REST façade (`cmd/api`)** under `/v1`: `POST /auth/login`, `POST /groups`, `POST /battles`, `POST /leaderboard/{season}`, `POST /bot/webhook`, plus `/metrics` for Prometheus and OpenTelemetry tracing on every handler.

**Deployment (`deploy/sandai`)**: Nakama config, a Docker Compose stack (CockroachDB, Nakama, façade, Prometheus) and Kubernetes manifests. The secrets and keys in these files are local development placeholders.

The design and scaling plan is in [docs/sand-ai-architecture.md](docs/sand-ai-architecture.md), and `src/README.md` explains the layer rules.

## Status

| Part | State |
|---|---|
| `src/domain`, `src/app` | builds; tournament and analytics have unit tests that pass |
| `src/infra/runtime` | does not build yet: uses `runtime.BadInputError`, which the vendored nakama-common does not have |
| `cmd/api` | does not build yet: imports `src/infra/nakama`, which is not written, and its OpenTelemetry dependencies are not vendored |

Run the parts that build:

```bash
go test ./src/domain/... ./src/app/...
```

## License

Nakama is Copyright Heroic Labs and licensed under the Apache License 2.0. The Sand-ai additions are released under the same license.
