package control_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ngrd.no/log"
	"ngrd.no/log/control"
)

func TestShouldLog(t *testing.T) {
	data := []struct {
		input    string
		expected [5]bool
	}{
		{"  ON  ON  ON  ON  ON", [...]bool{true, true, true, true, true}},
		{" OFF OFF OFF OFF OFF", [...]bool{false, false, false, false, false}},
	}

	for _, d := range data {
		ptr := control.ControlPtr(d.input)
		for i, expected := range d.expected {
			assert.Equal(t, expected, ptr.ShouldLog(control.Level(i+1)))
		}
	}
}

func TestDefaultShouldLog(t *testing.T) {
	f, err := ioutil.TempFile("", "logctrl.*")
	require.Nil(t, err)
	f.Close()
	defer os.Remove(f.Name())
	c := control.NewLogControl(f.Name())
	b1 := &bytes.Buffer{}
	_, err = log.New(log.WithComponentName("a"), log.WithWriter(b1), log.WithLogControl(c))
	require.Nil(t, err)

	assert.True(t, c.ShouldLog(log.ApplicationName+":"+"a", log.FATAL))
	assert.True(t, c.ShouldLog(log.ApplicationName+":a", log.ERROR))
	assert.True(t, c.ShouldLog(log.ApplicationName+":a", log.WARNING))
	assert.True(t, c.ShouldLog(log.ApplicationName+":a", log.INFO))
	assert.False(t, c.ShouldLog(log.ApplicationName+":a", log.DEBUG))
}
