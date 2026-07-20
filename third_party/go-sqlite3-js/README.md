# go-sqlite3-js

## Quantlete vendoring note

This module is vendored because upstream has been frozen since 2022 and its
`database/sql` scan path truncates every JavaScript number to an integer. The
source is from upstream commit `28aa791a1c91` (module version
`v0.0.0-20220419092513-28aa791a1c91`).

Quantlete's patch preserves fractional JavaScript numbers as `float64` while
returning mathematically integral numbers as `int64`. Keeping the upstream
module path unchanged allows the repository's relative `replace` directive to
provide reproducible WASM builds without relying on a hosted fork.

The original upstream README follows.

Experimental SQL driver for sql.js (in-browser sqlite) from Go WASM.

Only implements the subset of the SQL API required by Dendrite.

To run tests in Docker and Node:
```
$ docker build -t gsj .
$ docker run gsj
```

To run tests locally:

```bash
$ yarn install
$ GOOS=js GOARCH=wasm go test -exec="./go_sqlite_js_wasm_exec" .
```
