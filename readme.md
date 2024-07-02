# Concurrency Utility Lib

A Generic pattern for handing concurrent processing of objects.
Could be used for any use case that requires parallel execution.

The go Wait Group can be tricky to get right.  This approach tries to
make writing concurrent processing easier by encapsulating the complexity
and providing a go generic strategy.  This provides consistency in how this pattern
can be used and helps prevent common issues, such as deadlocks, from occurring.

Exmaple usage:

```go

type itemData struct{}


// Provide timeout for clean shutdown
ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
defer cancel()

processor := MyItemProcessor()

// Provide delegate function for processing items and concurrency level
processJobPool := NewJobPool(processor.ProcessItem, 2)

items := []itemData{} //..Items that need processing

for range items {
    processJob := &itemData{}
    processJobPool.Process(processJob)
}

processJobPool.Wait(ctx)
// Check if the context was timed out
err := ctx.Err()
if err != nil {
    // All Jobs completed before timeout
}


```
