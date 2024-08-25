# go-bench

Go HTTP client benchmark tool, supporting H2C and HTTP/1.1.

TLDR H2C is significantly faster than HTTP/1.1.

Example command:

```
./go-bench ./configfiles/macmini.toml

{"time":"2024-08-25T04:42:34.143103-05:00","level":"INFO","msg":"end main","mergedStatusCodeCount":{"200":200000},"totalCallsPerSecond":31835.99646062061}
```