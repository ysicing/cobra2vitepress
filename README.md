## 说明

```go
import (
doc "github.com/ysicing/cobra2vitepress"
)
 err := doc.GenMarkdownTree(q, "./docs")
 if err != nil {
  panic(err)
 }
```

生成vitepress文档
