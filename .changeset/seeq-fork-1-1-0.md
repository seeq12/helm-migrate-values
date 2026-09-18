---
"helm-migrate-values": minor
---

First Seeq release. No functional change to the migration logic.

The Go module path is now `github.com/seeq12/helm-migrate-values`. Release
archives are published as `.tgz` (`.zip` on Windows) and now also carry
`NOTICE` and `install-binary.sh`, so installing and updating the plugin both
work from a published archive.

Dependencies updated: `helm.sh/helm/v3` to v3.22.0, which removes
`github.com/containerd/containerd` from the module graph entirely and carries
`oras.land/oras-go/v2` to v2.6.2 and `golang.org/x/crypto` to v0.57.0;
`golang.org/x/net` bumped explicitly to v0.59.0. No `replace` directives are
required.

CI now runs `go vet`, a `gofmt` check, and tests with `-race`, with the Go
toolchain taken from `go.mod`.

Also fixes an inverted debug-log condition that could only ever log a nil
value, removes two unused exported types from `pkg`, strips UTF-8 BOMs from
four source files, moves test-only helpers out of the shipped `pkg` package,
and repairs an always-true version comparison in the release-versioning
workflow.
