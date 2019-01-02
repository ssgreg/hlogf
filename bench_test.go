package main

import (
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ssgreg/logf"
	"github.com/ssgreg/logftext"
)

var goldenN = []byte(`{"level":"debug","ts":"2018-12-13T22:21:26.84954039+03:00","logger":"token","msg":"request done","caller":"token/resource.go:66","scope-type":"root","scope-value":"","attempt":1,"request-id":"SL2DYF5L6XGT4BGQ","status":200}` + "\n")
var golden = []byte(`{"level":"debug","ts":"2018-12-13T22:21:26.84954039+03:00","logger":"token","msg":"request done","caller":"token/resource.go:66","scope-type":"root","scope-value":"","attempt":1,"request-id":"SL2DYF5L6XGT4BGQ","status":200}`)

type testReader struct {
	count int
}

func (r *testReader) Read(p []byte) (n int, err error) {
	if r.count <= 0 {
		return 0, io.EOF
	}
	r.count--
	copy(p, goldenN)
	return len(goldenN), nil
}

func BenchmarkDefault(b *testing.B) {
	r := &testReader{b.N}
	scan(r, ioutil.Discard, Options{
		NoColor:        true,
		TimeFormat:     time.StampMilli,
		StartingNumber: 1,
		BufferSize:     4096,
	})
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parse(golden)
	}
}

func BenchmarkFormat(b *testing.B) {
	eseq := logftext.EscapeSequence{NoColor: true}
	buf := logf.NewBufferWithCapacity(4096)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		e, _ := parse(golden)
		adoptEntry(&e)
		format(buf, eseq, &e, time.StampMilli)
	}
}
