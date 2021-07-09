package log_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ngrd.no/log"
	"ngrd.no/log/control"
)

func TestMain(m *testing.M) {
	f, err := ioutil.TempFile("", "logctrl.*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create temporary logctrl file")
	}
	f.Close()
	defer os.Remove(f.Name())

	control.DefaultControlPath = f.Name()
	os.Exit(m.Run())
}

func TestLogComponent(t *testing.T) {
	buf := &bytes.Buffer{}
	l, err := log.New(log.WithWriter(buf))
	require.Nil(t, err)
	l.Print("Hello!")
	assert.Contains(t, buf.String(), "\tngrd.no/log_test\tINFO\tHello!")
}

func TestError(t *testing.T) {
	buf := &bytes.Buffer{}
	l, err := log.New(log.WithWriter(buf))
	require.Nil(t, err)
	l.Errorf("Hello!")
	assert.Contains(t, buf.String(), "\tngrd.no/log_test\tERROR\tHello!")
}

func TestWarn(t *testing.T) {
	buf := &bytes.Buffer{}
	l, err := log.New(log.WithWriter(buf))
	require.Nil(t, err)
	l.Warnf("Hello!")
	assert.Contains(t, buf.String(), "\tngrd.no/log_test\tWARN\tHello!")
}

func TestDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	l, err := log.New(log.WithWriter(buf), log.WithComponentName("test_debug"))
	require.Nil(t, err)
	c := control.MaybeNewGlobalLogControl()
	update, err := c.OpenForUpdate()
	require.Nil(t, err)
	lines, err := update.ParseControl()
	require.Nil(t, err)
	for _, line := range lines {
		if line.Component == "test_debug" {
			assert.False(t, line.Ptr.ShouldLog(log.DEBUG))
			line.Ptr.On(log.DEBUG)
		}
	}
	require.Nil(t, update.Flush())
	require.Nil(t, update.Close())
	l.Debugf("Hello!")
	assert.Contains(t, buf.String(), "\ttest_debug\tDEBUG\tHello!")
}

func TestDoNotLogDebugByDefault(t *testing.T) {
	buf := &bytes.Buffer{}
	l, err := log.New(log.WithWriter(buf))
	require.Nil(t, err)
	l.Debugf("hello")
	assert.NotContains(t, buf.String(), "hello")
}
