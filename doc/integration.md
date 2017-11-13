# Integration Testing

Integration testing is key to developing reliable peer-to-peer applications.

The package `github.com/stratumn/alice/test/it` facilitates creating test networks.
It makes it easy to launch nodes in separate processes and interact with them
via provided API clients.

To use the package, you simply need to invoke the `Session` function:

```go
package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/alice/grpc/grpcapi"
	"github.com/stratumn/alice/release"
	"github.com/stratumn/alice/test/it"
	"google.golang.org/grpc"
)

func TestInform(t *testing.T) {
	// Terminate the session if it lasts more than one minute.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	err := it.Session(ctx, "/tmp", 16, it.IntegrationCfg(), func(set it.TestNodeSet, conns []*grpc.ClientConn) {
		for i, cx := range conns {
			// Run tests against each node.
			c := grpcapi.NewAPIClient(cx)

			r, err := c.Inform(ctx, &grpcapi.InformReq{})
			if err != nil {
				t.Errorf("node %d: Inform(): error: %+v", i, err)
				continue
			}

			if got, want := r.Protocol, release.Protocol; got != want {
				t.Errorf("node %d: r.Protocol = %q want %q", i, got, want)
			}
		}
	})

	if err != nil {
		t.Errorf("Session(): error: %+v", err)
	}
}
```

The package can also be used for benchmarking.

