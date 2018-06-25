# System Testing

System testing is key to developing reliable peer-to-peer applications.

The package `github.com/stratumn/go-indigonode/test/session` facilitates creating test
networks. It makes it easy to launch nodes in separate processes and interact
with them via provided API clients.

To use the package, you simply need to invoke the `Run` function:

```go
package system

import (
	"context"
	"testing"
	"time"

	grpcapi "github.com/stratumn/go-indigonode/core/app/grpcapi/grpc"
	"github.com/stratumn/go-indigonode/release"
	"github.com/stratumn/go-indigonode/test/session"
	"github.com/stretch/testify/assert"
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
			if !assert.NoError(t, err) {
				continue
			}

			assert.Equalf(t, release.Protocol, r.Protocol, "node %d", i)
		}
	}

	err := session.Run(ctx, "/tmp", 16, session.SystemCfg(), tester)
	assert.NoError(t, err, "Session()")
}
```

The package can also be used for benchmarking.
