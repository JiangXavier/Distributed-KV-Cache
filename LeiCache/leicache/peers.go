package leicache

import pb "leicache/leicachepb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter HTTP client
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
