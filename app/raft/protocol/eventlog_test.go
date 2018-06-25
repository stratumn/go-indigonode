package protocol

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/go-indigonode/app/raft/pb"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestLogNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	var entries []pb.Entry
	for i, e := range testEntries {
		entries = append(entries, pb.Entry{Index: uint64(i), Data: e})
	}

	c := circleProcess{
		committed: testEntries,
	}

	entriesChan := make(chan pb.Entry)
	doneChan := make(chan struct{})

	go func() {
		var ee []pb.Entry
		for e := range entriesChan {
			ee = append(ee, e)
		}
		assert.Equal(t, entries, ee)
		doneChan <- struct{}{}
	}()

	msg := hubCallLog{EntriesChan: entriesChan}

	c.eventLog(ctx, msg)

	select {
	case <-doneChan:
	case <-time.NewTimer(1 * time.Second).C:
		t.Error("timeout")
	}

}
