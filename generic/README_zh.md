# 泛化调用示例

[English](README.md) | [中文](README_zh.md)

本示例演示了如何通过 Triple 协议进行泛化调用，实现 Go 和 Java 服务之间的互操作。泛化调用允许在没有服务接口定义的情况下调用远程服务。

## 目录结构

```
generic/
├── go-server/      # Go 服务端（Triple 协议，端口 50052）
├── go-client/      # Go 客户端，泛化调用（直连 URL）
├── java-server/    # Java 服务端（Triple 协议，端口 50052）
└── java-client/    # Java 客户端，泛化调用（直连 URL）
```


## 启动 Go 服务端

```bash
cd generic/go-server/cmd
go run .
```

服务端通过 Triple 协议监听 `50052` 端口，提供 `UserProvider` 服务（version=1.0.0，group=triple），无需注册中心。

## 启动 Go 客户端

```bash
cd generic/go-client/cmd
go run .
```

客户端在 `cli.NewGenericService(...)` 处通过 `client.WithURL("tri://127.0.0.1:50052")` 为该服务单独配置直连地址，并基于该泛化服务发起调用。

## 泛化模式运行时检查

Go 客户端使用 `client.WithGenericType(...)` 选择泛化格式。该配置与 `client.WithSerialization(...)` 相互独立，后者用于选择请求在网络上传输时采用的序列化编码。

| 泛化模式 | 含义 | 运行时示例覆盖范围 |
|----------|------|--------------------|
| `true` | 基于 Map 的泛化结果（默认模式） | 远程类型化结果检查 |
| `gson` | JSON 格式的泛化结果 | 远程原始结果检查 |
| `bean` | JavaBean 描述符格式的泛化结果 | 远程类型化结果检查 |
| `protobuf-json` | Protobuf JSON 格式的泛化结果 | 当前 Hessian POJO 服务不使用该模式 |
| `protobuf` | 为兼容旧版本保留的别名 | 保留兼容能力 |
| `false` 或空值 | 关闭泛化调用 | 不发起泛化调用 |

运行现有 Go 客户端时，会检查 `true` 和 `bean` 的类型化结果、保留 `gson` 的原始结果，并确认未知模式会被拒绝。这些属于样例运行时检查，而不是通过 `go test` 执行的单元测试。当前 `User` 服务是 Hessian POJO，并非 `proto.Message`，因此本流程不包含 `protobuf-json` 调用。

## 启动 Java 服务端

在 java-server 目录下构建并运行：

```bash
cd generic/java-server
mvn clean compile exec:java -Dexec.mainClass="org.apache.dubbo.samples.ApiProvider"
```

## 启动 Java 客户端

```bash
cd generic/java-client
mvn clean compile exec:java -Dexec.mainClass="org.apache.dubbo.samples.ApiTripleConsumer"
```

客户端使用 `reference.setGeneric("true")` 并通过 `reference.setUrl("tri://127.0.0.1:50052")` 直连服务端进行泛化调用。

## 测试方法

| 方法 | 参数 | 返回值 |
|------|------|--------|
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

## 预期输出

服务端日志：

```
Generic Go server started on port 50052
```

客户端日志：

```
[Triple] GetUser1(userId string) res: {id=A003, name=Joe, age=48, ...}
[Triple] GetUser2(userId string, name string) res: {id=A003, name=lily, age=48, ...}
...
All generic call tests completed
```

## 注意事项

- 不要同时启动 Go 服务端和 Java 服务端，它们都监听 50052 端口。
- Go 服务端和 Java 服务端均无需 ZooKeeper，直接监听各自配置的端口。
- Java 客户端通过 `reference.setUrl(...)` 直连 `tri://127.0.0.1:50052`。
- Go 客户端通过 `tri://127.0.0.1:50052` 直连。
- 未知泛化模式会在创建服务时明确失败，不会静默回退到 Map 模式。
