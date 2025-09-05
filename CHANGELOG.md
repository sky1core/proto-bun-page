# Changelog

[한국어 버전](CHANGELOG.ko.md)

All notable changes to this project will be documented in this file.

## [Unreleased]
- Additional metrics/log fields (skipped keys count)
- README examples for more OR-chain variations
- MySQL tuple optimization flag implementation

## [v0.1.0] - 2025-09-05
- Offset + cursor pagination core
- Cursor policy: PK-only token, anchor fetch, strict exclusive OR-chain
- Ordering: single plan for both modes; always append PK; default to PK DESC
- Composite PK support: append all PKs; cursor includes PK tuple
- Options: AllowedOrderKeys, DefaultOrder, DefaultLimit, MaxLimit (clamp + defaulting)
- Logging: warn for disallowed order keys and limit clamp/defaulting
- Proto: schema file added, adapter (`ApplyAndScan`) implemented
- Example: offset, cursor, proto adapter demos
- Tests: unit + sqlite in-memory integration; boundary cases (page/limit, stale cursor, AllowedOrderKeys)
