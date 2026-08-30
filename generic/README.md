# Generic Call Sample

[English](README.md) | [中文](README_zh.md)

This sample demonstrates generic invocation over the Triple protocol for Go-Java interoperability. Generic invocation allows calling remote services without generating stubs or having the service interface locally.

## Layout

```
generic/
├── go-server/      # Go provider (Triple protocol, port 50052)
├── go-client/      # Go consumer with generic invocation (direct URL)
├── java-server/    # Java provider (Triple protocol, port 50052)
└── java-client/    # Java consumer with generic invocation (direct URL)
```


## Run the Go Server

```bash
cd generic/go-server/cmd
go run .
```

The server exposes the Triple protocol on port `50052` and serves `UserProvider` with version `1.0.0` and group `triple`. No registry is required.

## Run the Go Client

```bash
cd generic/go-client/cmd
go run .
```

The client passes `client.WithURL("tri://127.0.0.1:50052")` to `cli.NewGenericService(...)` for a per-service direct connection and performs generic calls through that generic service.

## Generic mode runtime checks

The Go client uses `client.WithGenericType(...)` to select the generic format. This is independent from `client.WithSerialization(...)`, which selects the transport encoding on the wire.

| Generic mode | Meaning | Runtime sample coverage |
|--------------|---------|-------------------------|
| `true` | Map-based generic result (default) | Complete `User` check, including `Time` |
| `gson` | JSON generic result | JSON shape check with an accepted Map fallback |
| `bean` | JavaBean descriptor result | Typed DTO check for `ID`, `Name`, and `Age` |
| `protobuf-json` | Protobuf JSON result | Not used by the current Hessian POJO service |
| `protobuf` | Legacy compatibility alias | Preserved for compatibility |
| `false` or empty | Disable generic invocation | No generic call is made |

When a `gson` result is a JSON string, it must decode to a complete `User` (`ID`, `Name`, `Age`, and `Time`). A Hessian Map fallback from either provider is accepted with an explicit warning. The `true` mode also checks every observable `User` field, including `Time`.

The Bean generalizer represents exported bean properties but cannot round-trip the unexported state inside Go's `time.Time`. The `bean` mode therefore uses an explicit DTO containing the supported `ID`, `Name`, and `Age` fields instead of accepting a partially populated `User`. The client also confirms that an unknown mode is rejected. These are runtime sample checks rather than `go test` unit tests. The current `User` service is a Hessian POJO rather than a `proto.Message`, so `protobuf-json` is not included in this flow.

## Run the Java Server

Build and run from the java-server directory:

```bash
cd generic/java-server
mvn clean compile exec:java -Dexec.mainClass="org.apache.dubbo.samples.ApiProvider"
```

## Run the Java Client

```bash
cd generic/java-client
mvn clean compile exec:java -Dexec.mainClass="org.apache.dubbo.samples.ApiTripleConsumer"
```

The client uses `reference.setGeneric("true")` and `reference.setUrl("tri://127.0.0.1:50052")` to perform generic calls via direct connection.

## Tested Methods

| Method | Parameters | Return |
|--------|------------|--------|
| GetUser1 | String | User |
| GetUser2 | String, String | User |
| GetUser3 | int | User |
| GetUser4 | int, String | User |
| GetOneUser | - | User |
| GetUsers | String[] | User[] |
| GetUsersMap | String[] | Map<String, User> |
| QueryUser | User | User |
| QueryUsers | User[] | User[] |
| QueryAll | - | Map<String, User> |

## Expected Output

Server log:

```
Generic Go server started on port 50052
```

Client log:

```
[Triple] GetUser1(userId string) res: {id=A003, name=Joe, age=48, ...}
[Triple] GetUser2(userId string, name string) res: {id=A003, name=lily, age=48, ...}
...
All generic call tests completed
```

## Notes

- Do NOT start Go Server and Java Server at the same time. Both listen on port 50052.
- Neither the Go server nor the Java server requires ZooKeeper; both listen directly on their configured ports.
- The Java client uses direct connection via `tri://127.0.0.1:50052` (`reference.setUrl(...)`).
- The Go client uses direct connection via `tri://127.0.0.1:50052`.
- Unknown generic modes fail during service creation instead of silently falling back to Map.
