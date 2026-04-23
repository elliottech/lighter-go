# Java

JNA bindings for the lighter-go shared library, with a benchmark.

## Prerequisites

- Java 21+
- Maven
- Go (to build the shared library)

Install on macOS:
```
brew install --cask temurin
brew install maven
```

## Build

**1. Build the shared library** from the repo root:

```
go build -buildmode=c-shared -o sharedlib/lighter.dylib ./sharedlib   # macOS
go build -buildmode=c-shared -o sharedlib/lighter.so   ./sharedlib   # Linux
```

**2. Compile** from the `java/` directory:

```
cd java
mvn compile
```

## Run

```
mvn exec:java
```

## What the benchmark does

Spawns 5 threads, each of which:
1. Generates a fresh API key pair
2. Creates a client on chain 304
3. Obtains an auth token (7-hour expiry)
4. Signs 100 create-order + cancel-order pairs back to back
5. Prints elapsed time for the signing loop
