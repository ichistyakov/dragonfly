# KidsCraft

A family-focused Minecraft: Bedrock server project built on the Dragonfly server framework.

**Maintainer:** nsakela <akellot@mail.ru>  
**Upstream:** [hashimthearab/dragonfly](https://github.com/hashimthearab/dragonfly)  
**Base commit:** `01c91ae0e5edf00470f9ce6bae2b66f6529415b3`

## Project status

This repository is the starting point for developing a Dragonfly version suitable for a family Bedrock server. Family-specific gameplay, permissions, safety rules, and content are not implemented yet; changes should be added and tested here rather than in the Go module cache.

The original Dragonfly README is preserved as [`README.dragonfly.md`](README.dragonfly.md), and the upstream MIT license is retained.

## Development

The `worldgen-lab` steel-server consumes this checkout through its local Go module replacement (`../KidsCraft`). Its Go module path remains `github.com/df-mc/dragonfly` for compatibility with existing imports.

Run the Dragonfly tests from this directory with:

```sh
go test ./...
```

The repository-local Git author is configured as `nsakela <akellot@mail.ru>`. The `upstream` remote points to the HashimTheArab Dragonfly fork; this local checkout is pinned to the base commit above.
