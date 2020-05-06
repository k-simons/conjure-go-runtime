package errors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/palantir/conjure-go-runtime/conjure-go-contract/codecs"
	wparams "github.com/palantir/witchcraft-go-params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func BenchmarkUnmarshalErrorName(b *testing.B) {
	runBench := func(b *testing.B, bodies [][]byte, unmarshalName func(*testing.B, []byte) string) {
		for i := 0; i < b.N; i++ {
			for _, body := range bodies {
				_ = unmarshalName(b, body)
			}
		}
	}

	makeBodies := func(count int) [][]byte {
		bodies := make([][]byte, count)
		for i := 0; i < count; i++ {
			safeParams := map[string]interface{}{}
			for j := 0; j < count; j++ {
				safeParams[fmt.Sprintf("param_%d", count)] = strings.Repeat("a", count)
			}
			payload := NewError(MustErrorType(Internal, fmt.Sprintf("MyNamespace:Error%d", i)), wparams.NewSafeParamStorer(safeParams))
			body, err := codecs.JSON.Marshal(payload)
			require.NoError(b, err)
			bodies[i] = body
		}
		return bodies
	}

	for _, count := range []int{1, 10, 100} {
		bodies := makeBodies(count)
		b.Run(fmt.Sprintf("count=%d", count), func(b *testing.B) {
			b.Run("stdlib", func(b *testing.B) {
				b.ReportAllocs()
				runBench(b, bodies, func(b *testing.B, body []byte) string {
					var name struct {
						Name string `json:"errorName"`
					}
					assert.NoError(b, codecs.JSON.Unmarshal(body, &name))
					return name.Name
				})
			})
			b.Run("gjson", func(b *testing.B) {
				b.ReportAllocs()
				runBench(b, bodies, func(b *testing.B, body []byte) string {
					return gjson.GetBytes(body, "errorName").String()
				})
			})
		})
	}
}
