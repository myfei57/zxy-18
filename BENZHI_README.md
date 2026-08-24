# TraceFlow

TraceFlow 是一个分布式链路追踪与采样平台：服务端接收带 trace id 与 parent id 的
span，按采样率决定是否落盘，批量上报写入本地存储，聚合器把同一 trace 的 span
归并成完整调用链，查询按索引代际检索，span 总量受命名空间配额约束，上报与聚合
事件全部留审计。

## 构建与运行

环境要求：Go 1.23（离线构建使用 vendor）。

```bash
go build -mod=vendor ./...
go run -mod=vendor ./cmd/traceflowd -data ./data -addr 127.0.0.1:7790
```

启动后访问：

- 采样台：http://127.0.0.1:7790/spans
- 上报台：http://127.0.0.1:7790/report
- 聚合台：http://127.0.0.1:7790/aggregate
- 审计台：http://127.0.0.1:7790/audit

健康检查：GET /api/health。

`-seed=true` 会在启动时预置三条演示链路，方便直接观察采样、上报、聚合与索引
流水线。
