package leicache

type ByteView struct {
	b []byte
}

func (b ByteView) Len() int {
	return len(b.b)
}

func (b ByteView) String() string {
	return string(b.b)
}

func (b ByteView) ByteSlice() []byte {
	return cloneBytes(b.b)
}

func cloneBytes(b []byte) []byte {
	a := make([]byte, len(b))
	copy(a, b)
	return a
}
