package store

import (
	"fmt"
	"time"

	"github.com/windmilleng/tilt/internal/model"
	"github.com/windmilleng/wmclient/pkg/analytics"
	v1 "k8s.io/api/core/v1"
)

type ErrorAction struct {
	Error error
}

func (ErrorAction) Action() {}

func NewErrorAction(err error) ErrorAction {
	return ErrorAction{Error: err}
}

type LogAction interface {
	Action
	model.LogEvent

	// Ideally, all logs should be associated with a source.
	//
	// In practice, not all logs have an obvious source identifier,
	// so this might be empty.
	//
	// Right now, that source is a ManifestName. But in the future,
	// this might make more sense as another kind of identifier.
	//
	// (As of this writing, we have TargetID as an abstract build-time
	// source identifier, but no generic run-time source identifier)
	Source() model.ManifestName
}

type LogEvent struct {
	mn        model.ManifestName
	timestamp time.Time
	msg       []byte
}

func (LogEvent) Action() {}

func (le LogEvent) Source() model.ManifestName {
	return le.mn
}

func (le LogEvent) Time() time.Time {
	return le.timestamp
}

func (le LogEvent) Message() []byte {
	return le.msg
}

func NewLogEvent(mn model.ManifestName, b []byte) LogEvent {
	return LogEvent{
		mn:        mn,
		timestamp: time.Now(),
		msg:       append([]byte{}, b...),
	}
}

func NewGlobalLogEvent(b []byte) LogEvent {
	return LogEvent{
		mn:        "",
		timestamp: time.Now(),
		msg:       append([]byte{}, b...),
	}
}

type K8sEventAction struct {
	Event *v1.Event
}

func (K8sEventAction) Action() {}

func NewK8sEventAction(event *v1.Event) K8sEventAction {
	return K8sEventAction{event}
}

func (kEvt K8sEventAction) ToLogAction(mn model.ManifestName) LogAction {
	msg := fmt.Sprintf("[K8s EVENT: %s] %s\n",
		objRefHumanReadable(kEvt.Event.InvolvedObject), kEvt.Event.Message)

	return LogEvent{
		mn:        mn,
		timestamp: kEvt.Event.LastTimestamp.Time,
		msg:       []byte(msg),
	}
}

func objRefHumanReadable(obj v1.ObjectReference) string {
	s := fmt.Sprintf("%s/%s %s", obj.APIVersion, obj.Kind, obj.Name)
	if obj.Namespace != "default" {
		s += fmt.Sprintf(" (ns: %s)", obj.Namespace)
	}
	return s
}

type AnalyticsOptAction struct {
	Opt analytics.Opt
}

func (AnalyticsOptAction) Action() {}

type AnalyticsNudgeSurfacedAction struct{}

func (AnalyticsNudgeSurfacedAction) Action() {}
