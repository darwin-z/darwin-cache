package core

// darwin-cache支持的主要数据结构是ByteView，存储真实的缓存值。
// 择 byte 类型是为了能够支持任意的数据类型的存储，例如字符串、图片等。
type ByteView struct {
	b []byte
}

// ByteView的Len()方法返回字节切片的长度，实现了Value接口的Len()方法。
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteView的ByteSlice()方法返回一个拷贝的字节切片，防止缓存值被外部程序修改。
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// ByteView的String()方法返回字节切片的字符串表示，实现了Value接口的String()方法。
func (v ByteView) String() string {
	return string(v.b)
}

// 克隆字节切片，返回一个新的字节切片。
func cloneBytes(b []byte) []byte {
	copied := make([]byte, len(b))
	copy(copied, b)
	return copied
}
