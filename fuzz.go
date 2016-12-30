//+build gofuzz

package rtnetlink

func Fuzz(data []byte) int {
	return fuzzLinkMessage(data)
}

func fuzzLinkMessage(data []byte) int {
	m := &LinkMessage{}
	if err := (m).UnmarshalBinary(data); err != nil {
		return 0
	}

	if _, err := m.MarshalBinary(); err != nil {
		panic(err)
	}

	return 1
}
