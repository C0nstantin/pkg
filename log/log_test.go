package log

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {

	buf := bytes.NewBufferString("")
	logrus.SetOutput(buf)

	t.Run("infof", func(t *testing.T) {
		logrus.SetLevel(logrus.DebugLevel)
		Infof("test %s", "infof")
		if !strings.Contains(buf.String(), "call:") {
			t.Errorf(" %s should contain call:", buf.String())
		}
		buf.Reset()
		logrus.SetLevel(logrus.InfoLevel)
		Infof("test %s", "printf")
		if strings.Contains(buf.String(), "call:") {
			t.Errorf(" %s should not contain call:", buf.String())
		}
	})

	t.Run("errorf", func(t *testing.T) {
		logrus.SetLevel(logrus.DebugLevel)
		Errorf("test %s", "errorf")
		if !strings.Contains(buf.String(), "call:") {
			t.Errorf(" %s should contain call:", buf.String())
		}
		buf.Reset()
		logrus.SetLevel(logrus.InfoLevel)
		Errorf("test %s", "printf")
		if strings.Contains(buf.String(), "call:") {
			t.Errorf(" %s should not contain call:", buf.String())
		}
	})
	t.Run("printf", func(t *testing.T) {
		logrus.SetLevel(logrus.DebugLevel)
		Printf("test %s", "printf")
		if !strings.Contains(buf.String(), "call:") {
			t.Errorf(" %s should contain call:", buf.String())
		}
		buf.Reset()
		logrus.SetLevel(logrus.InfoLevel)
		Printf("test %s", "printf")
		if strings.Contains(buf.String(), "call:") {
			t.Errorf(" %s should not contain call:", buf.String())
		}
	})
	t.Run("debugf", func(t *testing.T) {
		logrus.SetLevel(logrus.DebugLevel)
		Debugf("test %s", "infof")
		if !strings.Contains(buf.String(), "call:") {
			t.Errorf(" %s should contain call:", buf.String())
		}
		buf.Reset()
		logrus.SetLevel(logrus.InfoLevel)
		Debugf("test %s", "printf")
		if strings.Contains(buf.String(), "call:") {
			t.Errorf(" %s should not contain call:", buf.String())
		}

	})
	t.Run("tracef", func(t *testing.T) {
		logrus.SetLevel(logrus.TraceLevel)
		Tracef("test %s", "tracef")
		if !strings.Contains(buf.String(), "call:") {
			t.Errorf(" %s should contain call:", buf.String())
		}
	})

	t.Run("warnf", func(t *testing.T) {
		logrus.SetLevel(logrus.DebugLevel)
		Warnf("test %s", "printf")
		if !strings.Contains(buf.String(), "call:") {
			t.Errorf(" %s should contain call:", buf.String())
		}
		//assert.Contains(t, buf.String(), "call:")
		buf.Reset()
		logrus.SetLevel(logrus.InfoLevel)
		Warnf("test %s", "printf")
		//assert.NotContains(t, buf.String(), "call:")
		if strings.Contains(buf.String(), "call:") {
			t.Errorf(" %s should not contain call:", buf.String())
		}
	})
}
