# go-logindex

面向日志行的倒排索引库（无前端）。

## 能力

- `Add` / `Delete` / `Flush` / `Search`
- 查询：`Term` / `And` / `Or` / `Not`
- 分析器 → 内存缓冲 → 不可变段 → 段合并
- 索引预写日志（walix）与词频统计

## 快速开始

```go
idx, _ := logindex.Open(dir)
_ = idx.Add(1, "error timeout connecting database")
_ = idx.Flush()
hits, _ := idx.Search(logindex.And(logindex.Term("error"), logindex.Term("timeout")))
```

## 验证

```bash
go test ./... -count=1
```

模块路径：`github.com/LYH2263/go-logindex`
