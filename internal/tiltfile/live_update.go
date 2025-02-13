package tiltfile

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"go.starlark.net/syntax"

	"go.starlark.net/starlark"

	"github.com/windmilleng/tilt/internal/model"
)

// when adding a new type of `liveUpdateStep`, make sure that any tiltfile functions that create them also call
// `s.recordLiveUpdateStep`
type liveUpdateStep interface {
	starlark.Value
	liveUpdateStep()
	declarationPos() string
}

type liveUpdateFallBackOnStep struct {
	files    []string
	position syntax.Position
}

var _ starlark.Value = liveUpdateFallBackOnStep{}
var _ liveUpdateStep = liveUpdateFallBackOnStep{}

func (l liveUpdateFallBackOnStep) String() string {
	return fmt.Sprintf("fall_back_on step: %v'", l.files)
}
func (l liveUpdateFallBackOnStep) Type() string         { return "live_update_fall_back_on_step" }
func (l liveUpdateFallBackOnStep) Freeze()              {}
func (l liveUpdateFallBackOnStep) Truth() starlark.Bool { return len(l.files) > 0 }
func (l liveUpdateFallBackOnStep) Hash() (uint32, error) {
	t := starlark.Tuple{}
	for _, path := range l.files {
		t = append(t, starlark.String(path))
	}
	return t.Hash()
}
func (l liveUpdateFallBackOnStep) liveUpdateStep()        {}
func (l liveUpdateFallBackOnStep) declarationPos() string { return l.position.String() }

type liveUpdateSyncStep struct {
	localPath, remotePath string
	position              syntax.Position
}

var _ starlark.Value = liveUpdateSyncStep{}
var _ liveUpdateStep = liveUpdateSyncStep{}

func (l liveUpdateSyncStep) String() string {
	return fmt.Sprintf("sync step: '%s'->'%s'", l.localPath, l.remotePath)
}
func (l liveUpdateSyncStep) Type() string { return "live_update_sync_step" }
func (l liveUpdateSyncStep) Freeze()      {}
func (l liveUpdateSyncStep) Truth() starlark.Bool {
	return len(l.localPath) > 0 || len(l.remotePath) > 0
}
func (l liveUpdateSyncStep) Hash() (uint32, error) {
	return starlark.Tuple{starlark.String(l.localPath), starlark.String(l.remotePath)}.Hash()
}
func (l liveUpdateSyncStep) liveUpdateStep()        {}
func (l liveUpdateSyncStep) declarationPos() string { return l.position.String() }

type liveUpdateRunStep struct {
	command  string
	triggers []string
	position syntax.Position
}

var _ starlark.Value = liveUpdateRunStep{}
var _ liveUpdateStep = liveUpdateRunStep{}

func (l liveUpdateRunStep) String() string {
	s := fmt.Sprintf("run step: %s", strconv.Quote(l.command))
	if len(l.triggers) > 0 {
		s = fmt.Sprintf("%s (triggers: %s)", s, strings.Join(l.triggers, "; "))
	}
	return s
}

func (l liveUpdateRunStep) Type() string { return "live_update_run_step" }
func (l liveUpdateRunStep) Freeze()      {}
func (l liveUpdateRunStep) Truth() starlark.Bool {
	return len(l.command) > 0
}
func (l liveUpdateRunStep) Hash() (uint32, error) {
	t := starlark.Tuple{starlark.String(l.command)}
	for _, trigger := range l.triggers {
		t = append(t, starlark.String(trigger))
	}
	return t.Hash()
}
func (l liveUpdateRunStep) declarationPos() string { return l.position.String() }

func (l liveUpdateRunStep) liveUpdateStep() {}

type liveUpdateRestartContainerStep struct {
	position syntax.Position
}

var _ starlark.Value = liveUpdateRestartContainerStep{}
var _ liveUpdateStep = liveUpdateRestartContainerStep{}

func (l liveUpdateRestartContainerStep) String() string         { return "restart_container step" }
func (l liveUpdateRestartContainerStep) Type() string           { return "live_update_restart_container_step" }
func (l liveUpdateRestartContainerStep) Freeze()                {}
func (l liveUpdateRestartContainerStep) Truth() starlark.Bool   { return true }
func (l liveUpdateRestartContainerStep) Hash() (uint32, error)  { return 0, nil }
func (l liveUpdateRestartContainerStep) declarationPos() string { return l.position.String() }
func (l liveUpdateRestartContainerStep) liveUpdateStep()        {}

func (s *tiltfileState) recordLiveUpdateStep(step liveUpdateStep) {
	s.unconsumedLiveUpdateSteps[step.declarationPos()] = step
}

func (s *tiltfileState) liveUpdateFallBackOn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var files starlark.Value
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "paths", &files); err != nil {
		return nil, err
	}
	filesSlice := starlarkValueOrSequenceToSlice(files)
	var fileStrings []string
	for _, f := range filesSlice {
		switch fStr := f.(type) {
		case starlark.String:
			fileStrings = append(fileStrings, string(fStr))
		default:
			return nil, fmt.Errorf("fall_back_on step contained value '%s' of type '%s'. it may only contain strings", fStr, fStr.Type())
		}
	}

	ret := liveUpdateFallBackOnStep{
		files:    fileStrings,
		position: thread.TopFrame().Position(),
	}
	s.recordLiveUpdateStep(ret)
	return ret, nil
}

func (s *tiltfileState) liveUpdateSync(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var localPath, remotePath string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "local_path", &localPath, "remote_path", &remotePath); err != nil {
		return nil, err
	}

	ret := liveUpdateSyncStep{
		localPath:  s.absPath(localPath),
		remotePath: remotePath,
		position:   thread.TopFrame().Position(),
	}
	s.recordLiveUpdateStep(ret)
	return ret, nil
}

func (s *tiltfileState) liveUpdateRun(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var command string
	var triggers starlark.Value
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "cmd", &command, "trigger?", &triggers); err != nil {
		return nil, err
	}

	triggersSlice := starlarkValueOrSequenceToSlice(triggers)
	var triggerStrings []string
	for _, t := range triggersSlice {
		switch t2 := t.(type) {
		case starlark.String:
			triggerStrings = append(triggerStrings, string(t2))
		default:
			return nil, fmt.Errorf("run cmd '%s' triggers contained value '%s' of type '%s'. it may only contain strings", command, t.String(), t.Type())
		}
	}

	ret := liveUpdateRunStep{
		command:  command,
		triggers: triggerStrings,
		position: thread.TopFrame().Position(),
	}
	s.recordLiveUpdateStep(ret)
	return ret, nil
}

func (s *tiltfileState) liveUpdateRestartContainer(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs); err != nil {
		return nil, err
	}

	ret := liveUpdateRestartContainerStep{
		position: thread.TopFrame().Position(),
	}
	s.recordLiveUpdateStep(ret)
	return ret, nil
}

func (s *tiltfileState) liveUpdateStepToModel(l liveUpdateStep) (model.LiveUpdateStep, error) {
	switch x := l.(type) {
	case liveUpdateFallBackOnStep:
		return model.LiveUpdateFallBackOnStep{Files: x.files}, nil
	case liveUpdateSyncStep:
		if !filepath.IsAbs(x.remotePath) {
			return nil, fmt.Errorf("sync destination '%s' (%s) is not absolute", x.remotePath, x.position.String())
		}
		return model.LiveUpdateSyncStep{Source: x.localPath, Dest: x.remotePath}, nil
	case liveUpdateRunStep:
		return model.LiveUpdateRunStep{
			Command: model.ToShellCmd(x.command),
			Triggers: model.PathSet{
				Paths:         x.triggers,
				BaseDirectory: s.absWorkingDir(),
			},
		}, nil
	case liveUpdateRestartContainerStep:
		return model.LiveUpdateRestartContainerStep{}, nil
	default:
		return nil, fmt.Errorf("internal error - unknown liveUpdateStep '%v' of type '%T', declared at %s", l, l, l.declarationPos())
	}
}

func (s *tiltfileState) liveUpdateFromSteps(maybeSteps starlark.Value) (model.LiveUpdate, error) {
	var modelSteps []model.LiveUpdateStep
	stepSlice := starlarkValueOrSequenceToSlice(maybeSteps)

	for _, v := range stepSlice {
		step, ok := v.(liveUpdateStep)
		if !ok {
			return model.LiveUpdate{}, fmt.Errorf("'steps' must be a list of live update steps - got value '%v' of type '%s'", v.String(), v.Type())
		}

		ms, err := s.liveUpdateStepToModel(step)
		if err != nil {
			return model.LiveUpdate{}, err
		}
		s.consumeLiveUpdateStep(step)

		modelSteps = append(modelSteps, ms)
	}

	return model.NewLiveUpdate(modelSteps, s.absWorkingDir())
}

func (s *tiltfileState) consumeLiveUpdateStep(stepToConsume liveUpdateStep) {
	delete(s.unconsumedLiveUpdateSteps, stepToConsume.declarationPos())
}

func (s *tiltfileState) checkForUnconsumedLiveUpdateSteps() error {
	if len(s.unconsumedLiveUpdateSteps) > 0 {
		var errorStrings []string
		for _, step := range s.unconsumedLiveUpdateSteps {
			errorStrings = append(errorStrings, fmt.Sprintf("value '%s' of type '%s' declared at %s", step.String(), step.Type(), step.declarationPos()))
		}
		return fmt.Errorf("found %d live_update steps that were created but not used in a live_update: %s",
			len(s.unconsumedLiveUpdateSteps), strings.Join(errorStrings, "\n\t"))
	}

	return nil
}
