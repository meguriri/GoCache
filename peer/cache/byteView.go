package cache

type ByteView []byte

func (v ByteView) Len() int {
	return len(v)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (v ByteView) ToByte() []byte {
	return cloneBytes(v)
}

func (v ByteView) String() string {
	return string(v)
}
