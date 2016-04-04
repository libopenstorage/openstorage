# coprhd
[![MIT License](http://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/peter-edge/dlog-go/blob/master/LICENSE)

coprhd go library

###Installation 

```bash
go get github.com/ModelRocket/coprhd/...
```

###Library

The `coprhd.Client` interface provides a simple means for managing coprhd/vipr resources. 

```go
import (
 "github.com/ModelRocket/coprhd"
 "fmt"
)

func main() {
     // You only need to get your token once, or use the coprtop tool
	token, _ := corphd.GetProxyToken(host, username, password)
     
	vols, err := client.Volume().List()
	if err != nil {
		fmt.Fatalf("Failed to get volume list:", err.Error())
	}

	for i, vol := range vols {
		fmt.Printf("Volume %d: %s\n", i, vol)
	}

	// Query a volume by urn
	vol, _ := client.Volume().
		Id(urn).  // or by Name(name)
		Query()

	fmt.Printf("Volume name %s\n", vol.Name)
}
```

###Tool
You can use the `coprtop` cli tool to get a proxy token to use with the client. 

```bash
$GOPATH/bin/coprtop -u root -p pass -H 172.31.32.100 token
```