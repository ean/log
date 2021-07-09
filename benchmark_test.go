package log

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"ngrd.no/log/control"
	"ngrd.no/log/utils"
)

func BenchmarkShouldLog(b *testing.B) {
	for i := 1; i < 100000; i *= 10 {
		b.Run(fmt.Sprintf("%d", i), func(b *testing.B) {
			b.StopTimer()
			component := ""
			f, err := ioutil.TempFile("", "logctrl.*")
			require.Nil(b, err)
			f.Close()
			defer os.Remove(f.Name())
			c := control.NewLogControl(f.Name())
			// defer c.Reset()
			for j := 0; j < i; j++ {
				comp := strings.Join([]string{
					utils.RandStringRunes(10),
					utils.RandStringRunes(10),
					utils.RandStringRunes(10),
					utils.RandStringRunes(10),
				}, "/")
				buf := &bytes.Buffer{}
				_, err = New(WithComponentName(comp), WithLogControl(c), WithWriter(buf))
				require.Nil(b, err)
				component = comp
			}

			key := ApplicationName + ":" + component
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				c.ShouldLog(key, INFO)
			}
		})
	}
}

func BenchmarkLog(b *testing.B) {
	for i := 1; i < 17; i *= 2 {
		for j := 8; j < 65; j *= 2 {
			b.Run(fmt.Sprintf("#%d-len(%d)", i, j), func(b *testing.B) {
				b.StopTimer()
				f, err := ioutil.TempFile("", "logctrl.*")
				require.Nil(b, err)
				f.Close()
				defer os.Remove(f.Name())
				c := control.NewLogControl(f.Name())
				// defer c.Reset()
				elements := []interface{}{}
				buf := &bytes.Buffer{}
				component := utils.RandStringRunes(10) + "/" + utils.RandStringRunes(20)
				l, err := New(WithComponentName(component), WithLogControl(c), WithWriter(buf))
				require.Nil(b, err)
				for k := 0; k < i; k++ {
					elements = append(elements, utils.RandStringRunes(j))
				}

				b.StartTimer()
				for i := 0; i < b.N; i++ {
					l.Print(elements...)
				}
			})
		}
	}
}
