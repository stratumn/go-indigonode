# System Testing

System testing is key to developing reliable peer-to-peer applications.

The package `github.com/stratumn/alice/test/session` facilitates creating test
networks.  It makes it easy to launch nodes in separate processes and interact
with them via provided API clients.

To use the package, you simply need to invoke the `Run` function:

```go
package system

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/alice/grpc/grpcapi"
	"github.com/stratumn/alice/release"
	"github.com/stratumn/alice/test/session"
	"google.golang.org/grpc"
)

func TestInform(t *testing.T) {
	// Terminate the session if it lasts more than one minute.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	tester := func(ctx context.Context, set session.TestNodeSet, conns []*grpc.ClientConn) {
		for i, cx := range conns {
			// Run tests against each node.
			c := grpcapi.NewAPIClient(cx)

			r, err := c.Inform(ctx, &grpcapi.InformReq{})
			if err != nil {
				t.Errorf("node %d: Inform(): error: %s", i, err)
				continue
			}

			if got, want := r.Protocol, release.Protocol; got != want {
				t.Errorf("node %d: r.Protocol = %v want %v", i, got, want)
			}
		}
	}

	err := session.Run(ctx, "/tmp", 16, session.SystemCfg(), tester)
	if err != nil {
		t.Errorf("Session(): error: %+v", err)
	}
}
```

The package can also be used for benchmarking.

