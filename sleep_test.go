package gosync

import (
	. "github.com/echocat/gocheck-addons"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

type SleepTest struct{}

func (s *SleepTest) TestInterrupt(c *C) {
	mainGroup, err := NewGroup()
	c.Assert(err, IsNil)
	sg, err := mainGroup.NewGroup()
	c.Assert(err, IsNil)
	start := time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		sg.Interrupt()
	}()
	sg.Sleep(10 * time.Second)
	duration := time.Since(start)
	c.Assert(duration, IsLessThan, time.Duration(50*time.Millisecond))
}

func Test(t *testing.T) {
	TestingT(t)
}

func init() {
	Suite(&SleepTest{})
}
