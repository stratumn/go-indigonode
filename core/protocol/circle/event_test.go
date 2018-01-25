package circle

const (
	testNodeID          = uint64(1)
	testRemoteNodeID    = uint64(2)
	testElectionTick    = int(10)
	testHeartbeatTick   = int(1)
	testMaxSizePerMsg   = uint64(1024 * 1024)
	testMaxInflightMsgs = int(256)
	testTickerInterval  = uint64(100)
)

var (
	testLocalPeer  = []byte("localpeer")
	testRemotePeer = []byte("remotepeer")
)
