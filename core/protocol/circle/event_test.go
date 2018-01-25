package circle

const (
	testNodeID          uint64 = 1
	testRemoteNodeID    uint64 = 2
	testElectionTick    int    = 10
	testHeartbeatTick   int    = 1
	testMaxSizePerMsg   uint64 = 1024 * 1024
	testMaxInflightMsgs int    = 256
	testTickerInterval  uint64 = 100
	testConfigChangeID  uint64 = 10
	testLastNodeID      uint64 = 20
)

var (
	testLocalPeer         = []byte("localpeer")
	testRemotePeer        = []byte("remotepeer")
	testInternodeMessage  = []byte("heyya")
	testEntryOne          = []byte("dead")
	testEntryTwo          = []byte("beef")
	testConfigChangeEntry = []byte("configchange")
	testEntries           = [][]byte{testEntryOne, testEntryTwo}
)
