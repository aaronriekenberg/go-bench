# go-bench

Go HTTP client benchmark tool, supporting H2C and HTTP/1.1.

TLDR H2C is significantly faster than HTTP/1.1.

Example command:

```
./go-bench ./configfiles/macmini.toml

{"time":"2024-08-31T05:31:57.118436-05:00","level":"INFO","msg":"end main","mergedStatusCodeCount":{"200":500000},"totalCallsPerSecond":70170.06899671663}
```
