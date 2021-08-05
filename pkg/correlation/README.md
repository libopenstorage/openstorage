## Correlation logging instructions

There are three ways for using this package. All methods will result in a correlation ID being 
added to the log lines. Each method may make sense in different scenarios. We've provided
all three methods as they can be used interchangeably depending on the developers preference. 

### Method 1: Per-package correlation logger
This method is nice if a package maintainer wants to register a global 
logging hook at a per-package level. This is nice as you only need to
create the logger once per package. However, each log line can be quite
verbose as you must provide WithContext(ctx)

```
var (
    clogger := correlation.NewPackageLogger("test")
)
	
func testFunc() {
	ctx := correlation.NewContext(context.Background(), "test_origin")

    clogger.WithContext(ctx).Info("test info log")
}
```

### Method 2: Per-function correlation logger
This method is great for reducing the amount of count needed per-line, 
as you do not need to pass in the context at every log line. However,
you must remember to instantiate the logger at every function.

```
ctx := correlation.NewContext(context.Background(), "test_origin")
clogger := correlation.NewFunctionLogger(ctx, "test")

clogger.Info("test info log")
```

### Method 3: Global correlation logging hook
This is a nice catch-all for covering packages where a
package-level or function-level logger is not created. However,
it does not support component metadata.

```
correlation.RegisterGlobalHook()
ctx := correlation.NewContext(context.Background(), "test_origin")

logrus.WithContext(ctx).Info("test info log")
```