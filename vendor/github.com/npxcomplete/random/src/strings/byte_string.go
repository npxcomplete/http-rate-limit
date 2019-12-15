package random_strings

import "math/rand"

var EnglishAlphabet []byte = []byte("abcdefghijklmnopqrstuvwxyz")

type ByteStringGenerator struct {
	Alphabet  []byte
	RandomGen *rand.Rand
}

func (gen *ByteStringGenerator) String(length int) string {
	raw := make([]byte, length)
	out := make([]byte, length)

	gen.RandomGen.Read(raw)

	for i := 0; i < length; i++ {
		out[i] = gen.Alphabet[int(raw[i])%length]
	}

	return string(out)
}
