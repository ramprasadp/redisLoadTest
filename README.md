# redisLoadTest

A Redis long duration benchmarking tool.

## Setup Instructions

1. Initialize the Go module:
```bash
go mod init redisLoadTest
go mod tidy
```

2. Build the benchmark tool:
```bash
cd pushBench
go build pushPop.go
```

## Usage

Run the Redis push/pop test with the following command:
```bash
./pushPop [options]
```

### Available Options

| Flag | Description | Default |
|------|-------------|---------|
| `-h` | Redis server host | `localhost` |
| `-P` | Redis port | `6379` |
| `-p` | Redis password | - |
| `-n` | Number of elements to push/pop | `10` |
| `-s` | Packet size | `100` |
| `-t` | Number of concurrent threads | `3` |
| `-i` | Interval between test runs (seconds) | `10` |
| `-m` | Number of times to run the tests | `2` |