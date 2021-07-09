package control

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ngrd.no/log/utils"
)

const app = "myapp"

type dataSet struct {
	first, middle, last string
	c                   *LogControl
}

func (d *dataSet) Close() {
	os.Remove(d.c.ControlPath)
	os.Remove(d.c.controlLockPath)
}

func generateDataSet(t testing.TB) *dataSet {
	f, err := ioutil.TempFile("", "logctrl.*")
	require.Nil(t, err)
	f.Close()
	defer os.Remove(f.Name())
	c := NewLogControl(f.Name())
	first := ""
	last := ""
	middle := ""
	num := 10000
	for i := 0; i < num; i++ {
		comp := strings.Join([]string{
			utils.RandStringRunes(10),
			utils.RandStringRunes(10),
			utils.RandStringRunes(10),
			utils.RandStringRunes(10),
		}, "/")
		require.Nil(t, c.Register(app, comp))
		if i == 0 {
			first = comp
		} else if i == num-1 {
			last = comp
		} else if num/2 == i {
			middle = comp
		}
	}
	return &dataSet{
		first:  first,
		middle: middle,
		last:   last,
		c:      c,
	}
}

func BenchmarkReadControlFile(b *testing.B) {
	d := generateDataSet(b)
	b.Run("ReadControlFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			require.Nil(b, d.c.ReadControlFile())
		}
	})
	b.Run("parseControl", func(b *testing.B) {
		for i := 0; i < b.N*1000; i++ {
			_, err := parseControl(d.c.memory.Data)
			require.Nil(b, err)
		}
	})
}

func BenchmarkKeyPresent(b *testing.B) {
	d := generateDataSet(b)
	b.Run("keyPresent first", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			require.True(b, d.c.keyPresent(app, d.first))
		}
	})
	b.Run("keyPresent middle", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			require.True(b, d.c.keyPresent(app, d.middle))
		}
	})
	b.Run("keyPresent last", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			assert.True(b, d.c.keyPresent(app, d.last))
		}
	})
}
