package embeddings

import (
	"strconv"
	"strings"
)

// FormatVector encodes a slice of float32 as the pgvector text format
// (`[v1,v2,…]`). lib/pq has no native float32-array binding, so we send
// the value as a string and let pgvector parse it on the server side.
func FormatVector(values []float32) string {
	var b strings.Builder
	b.Grow(len(values)*12 + 2)
	b.WriteByte('[')
	for i, v := range values {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}
