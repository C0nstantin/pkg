package notificator

import (
	"github.com/C0nstantin/pkg/errors"
	"github.com/airbrake/gobrake/v5"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/C0nstantin/pkg/log"
)

type Notificator interface {
	Notify(err interface{}, r *http.Request)
}
type errbitNotificator struct {
	Notifier *gobrake.Notifier
}

func (d errbitNotificator) Notify(err interface{}, r *http.Request) {
	if err == nil {
		return
	}
	// get request_id and sent to notify params
	if d.Notifier == nil {
		log.Errorf("notifier is nil")
		return
	}

	n := d.Notifier.Notice(err, r, 0)
	if e, ok := err.(errors.StackTracer); ok {
		frames := make([]gobrake.StackFrame, 0, len(e.StackTrace()))
		stackTrace := e.StackTrace()
		var pcs []uintptr
		for _, f := range stackTrace {
			pcs = append(pcs, uintptr(f))
		}

		ff := runtime.CallersFrames(pcs)
		var firstPkg string

		for {
			f, ok := ff.Next()
			if !ok {
				break
			}

			pkg, fn := splitPackageFuncName(f.Function)
			if firstPkg == "" {
				firstPkg = pkg
			}

			frames = append(frames, gobrake.StackFrame{
				File: f.File,
				Line: f.Line,
				Func: fn,
			})
		}
		n.Errors[0].Backtrace = frames
		n.Context["component"] = firstPkg
	}

	if os.Getenv("ENV") == "production" {
		d.Notifier.SendNoticeAsync(n)
		return
	}
	res, err1 := d.Notifier.SendNotice(n)
	log.Debugf("notify res = %s", res)
	log.Tracef("notify message %#v", n)
	if err1 != nil {
		log.Errorf("send notify  error: err: %s", err1)
	}
}

func NewErrbitNotificator(ErrbitProjectId int64, ErrbitProjectKey, ErrbitHost, env, proxy string) Notificator {

	configNotificator := &gobrake.NotifierOptions{
		ProjectId:                 ErrbitProjectId,
		ProjectKey:                ErrbitProjectKey,
		Host:                      ErrbitHost,
		DisableRemoteConfig:       true,
		Environment:               env,
		DisableCodeHunks:          true,
		DisableErrorNotifications: false,
		DisableAPM:                true,
	}
	if env == "development" {
		var pr func(*http.Request) (*url.URL, error)
		if proxy != "" && proxy != "false" && proxy != "none" {
			proxy, err := url.Parse(proxy)
			if err == nil {
				pr = http.ProxyURL(proxy)
			}
		}

		client := http.Client{
			Transport: &http.Transport{
				Proxy: pr,
			},
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       0,
		}
		configNotificator.HTTPClient = &client
	}
	log.Debugf("configNotificator: %+v", configNotificator)
	return &errbitNotificator{
		Notifier: gobrake.NewNotifierWithOptions(configNotificator),
	}
}
func splitPackageFuncName(funcName string) (string, string) {
	var packageName string
	if ind := strings.LastIndex(funcName, "/"); ind > 0 {
		packageName += funcName[:ind+1]
		funcName = funcName[ind+1:]
	}
	if ind := strings.Index(funcName, "."); ind > 0 {
		packageName += funcName[:ind]
		funcName = funcName[ind+1:]
	}
	return packageName, funcName
}
