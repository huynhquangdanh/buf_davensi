### Upsert
use type upsert like insert,


```go
qb := util.CreateQueryBuilder(util.Upsert, "test")

qb.SetInsertField("a", "b", "c")
qb.SetInsertValues([]any{"a", "a", "a"})

fmt.Println(qb.GenerateSQL())
```

Result SQL 
```SQL
UPSERT INTO test(a, b, c) VALUES ($1, $2, $3) RETURNING *
```

Result Args 
```
[a a a]
```

Result Sel
```
for (a, b, c) VALUES (?, ?, ?) insert [a a a]
```

### Insert on conflict

```go
qb := util.CreateQueryBuilder(util.Insert, "test")

qb.SetInsertField("a", "b", "c")
qb.SetInsertValues([]any{"a", "a", "a"})
qb.OnConflict("a", "b")
qb.SetUpdate("a", nil).SetUpdate("b", nil)

fmt.Println(qb.GenerateSQL())
```

Result SQL 
```SQL
INSERT INTO test(a, b, c) VALUES ($1, $2, $3) ON CONFLICT (a, b) DO UPDATE SET a = excluded.a, b = excluded.b RETURNING *
```

Result Args 
```
[a a a]
```

Result Sel
```
for (a, b, c) VALUES (?, ?, ?) ON CONFLICT (a, b) DO UPDATE SET a = excluded.a, b = excluded.b insert [a a a]
```