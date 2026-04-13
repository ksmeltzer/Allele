# Plugin Lifecycle & Change Map

When the Application Binary Interface (ABI) for plugins (`internal/abi/abi.go`) is modified (e.g., adding a new field to `Manifest`, `Dependency`, or `ConfigField`), a system-wide update must be triggered to ensure compatibility across all decoupled systems.

This document serves as the mandatory checklist for addressing ABI changes.

## 1. Core Engine (Go)
- [ ] Update structs in `internal/abi/abi.go`.
- [ ] Run `go build ./...` to ensure core engine compiles.
- [ ] Update mock/stub data in `internal/dashboard/broadcaster.go` if it initializes ABI structs manually.
- [ ] Update `internal/storage/db.go` if the new ABI fields affect how plugin configurations are persisted.

## 2. Documentation
- [ ] Update the JSON examples in `docs/architecture/08_plugin_api.md` (specifically the `Manifest` export example).
- [ ] Update `docs/architecture/07_user_interface.md` if the change necessitates a new UI behavior (e.g., rendering a new config type).

## 3. External Decoupled Plugins
*Because plugins run inside WASM via shared memory JSON parsing, they must export a matching JSON structure.*
Navigate to `plugins/` and update the `Manifest()` string output in the entry points for:
- [ ] `plugins/allele-exchange-polymarket/main.go`
- [ ] `plugins/allele-sensor-copilot/main.go`
- [ ] `plugins/allele-strategy-bayesian/main.go`
- [ ] `plugins/allele-strategy-completeness-go/main.go`
- [ ] `plugins/allele-strategy-cross-market/main.go`
- [ ] Commit and push changes in each sub-repository individually.

## 4. User Interface (React)
- [ ] Update the corresponding TypeScript interfaces in `ui/src/types/plugin.ts` (or wherever the `Manifest` is defined on the frontend).
- [ ] Implement UI logic to handle the new fields (e.g., rendering a download button for the `url` property).
