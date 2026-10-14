// Package fuzz contains Go native fuzz targets.
//
// These targets fuzz the boundary between dnsconfig.js and the Go code:
// a JSON document in the same shape that JSON.stringify(conf) produces
// is turned into a models.DNSConfig and pushed through the same
// pipeline that `dnscontrol` runs: record import, validation and
// normalization, and zonefile generation.
//
// The targets are run as ordinary tests (seed corpus only) by
// `go test` and gotestsum.  Actual fuzzing happens in the scheduled
// "Fuzzing" GitHub Actions workflow, which runs each target with
// -fuzz for a fixed time budget.
//
// The JavaScript engine itself is intentionally not fuzzed: the
// PANIC() helper calls os.Exit(1), which kills fuzz workers, and the
// require()/glob() helpers expose the filesystem.  Fuzzing the JS
// layer would need those to return errors instead.
package fuzz
