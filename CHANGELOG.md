# Changelog

## [v0.0.12](https://github.com/fujiwara/jsonnet-armed/compare/v0.0.11...v0.0.12) - 2025-09-28
- Add --stdout flag to output to both file/HTTP and stdout by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/43
- Add warning when --write-if-changed is used with HTTP output by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/45

## [v0.0.11](https://github.com/fujiwara/jsonnet-armed/compare/v0.0.10...v0.0.11) - 2025-09-28
- Add jq function for JSON data processing by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/41

## [v0.0.10](https://github.com/fujiwara/jsonnet-armed/compare/v0.0.9...v0.0.10) - 2025-09-28
- Add support for custom native functions in library usage by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/38
- feat: Add HTTP/HTTPS output support by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/40

## [v0.0.9](https://github.com/fujiwara/jsonnet-armed/compare/v0.0.8...v0.0.9) - 2025-09-20
- Use os.UserCacheDir() instead of lookup environment variables directly. by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/32
- Add regular expression functions by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/34
- Add UUID v4 and v7 generation functions by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/35
- Add comprehensive native function development guidelines by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/36
- Add UUID functions documentation to README by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/37

## [v0.0.8](https://github.com/fujiwara/jsonnet-armed/compare/v0.0.7...v0.0.8) - 2025-09-17
- feat: add cache mechanism for evaluation results by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/27
- feat: add stale cache fallback mechanism by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/29
- feat: add HTTP functions for making HTTP requests by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/30
- feat: add DNS lookup functions with HTTPS/SVCB record support by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/31

## [v0.0.7](https://github.com/fujiwara/jsonnet-armed/compare/v0.0.6...v0.0.7) - 2025-09-16
- feat: Implement atomic file writing with write-if-changed option by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/25

## [v0.0.6](https://github.com/fujiwara/jsonnet-armed/compare/v0.0.5...v0.0.6) - 2025-09-09
- Refactor exec functions to use explicit context passing by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/23

## [v0.0.5](https://github.com/fujiwara/jsonnet-armed/compare/v0.0.4...v0.0.5) - 2025-09-08
- Add env_parse native function by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/21

## [v0.0.4](https://github.com/fujiwara/jsonnet-armed/compare/v0.0.3...v0.0.4) - 2025-09-06
- Refactor native functions from slices to maps for O(1) access by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/15
- Add exec functions for external command execution by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/17
- Improve README structure and navigation by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/18
- Move Usage section to improve first-time user experience by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/19
- Optimize test performance by reducing sleep times by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/20

## [v0.0.3](https://github.com/fujiwara/jsonnet-armed/compare/v0.0.2...v0.0.3) - 2025-09-06
- Add stdin support with '-' filename by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/9
- Add base64 encoding functions by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/10
- Add time functions with format constant support by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/11
- Add file_exists function for safe file operations by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/12
- Add timeout flag for evaluation safety by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/13
- Refactor functions tests to unit tests and add integration tests by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/14

## [v0.0.2](https://github.com/fujiwara/jsonnet-armed/compare/v0.0.1...v0.0.2) - 2025-09-05
- Add hash functions (string and file) by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/4
- Add file content and metadata functions by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/6
- Add unified function access via import 'armed.libsonnet' by @fujiwara in https://github.com/fujiwara/jsonnet-armed/pull/7

## [v0.0.1](https://github.com/fujiwara/jsonnet-armed/commits/v0.0.1) - 2025-09-05
