package activity

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cschleiden/go-dt/internal/converter"
	"github.com/cschleiden/go-dt/internal/payload"
	"github.com/cschleiden/go-dt/internal/workflow"
	"github.com/cschleiden/go-dt/pkg/core/task"
	"github.com/cschleiden/go-dt/pkg/history"
	"github.com/pkg/errors"
)

type Executor struct {
	r *workflow.Registry
}

func NewExecutor(r *workflow.Registry) Executor {
	return Executor{
		r: r,
	}
}
func (e *Executor) ExecuteActivity(ctx context.Context, task task.Activity) (payload.Payload, error) {
	a := task.Event.Attributes.(history.ActivityScheduledAttributes)

	activity := e.r.GetActivity(a.Name)

	activityFn := reflect.ValueOf(activity)
	if activityFn.Type().Kind() != reflect.Func {
		return nil, errors.New("activity not a function")
	}

	args, err := inputsToArgs(ctx, activityFn, a.Inputs)
	if err != nil {
		return nil, err
	}

	r := activityFn.Call(args)

	if len(r) < 1 || len(r) > 2 {
		return nil, errors.New("activity has to return either (error) or (result, error)")
	}

	var result payload.Payload

	if len(r) > 1 {
		var err error
		result, err = converter.DefaultConverter.To(r[0].Interface())
		if err != nil {
			return nil, errors.Wrap(err, "could not convert activity result")
		}
	}

	errResult := r[len(r)-1]
	if errResult.IsNil() {
		return result, nil
	}

	errInterface, ok := errResult.Interface().(error)
	if !ok {
		return nil, fmt.Errorf("activity error result does not satisfy error interface (%T): %v", errResult, errResult)
	}

	return result, errInterface
}

func inputsToArgs(ctx context.Context, activityFn reflect.Value, inputs []payload.Payload) ([]reflect.Value, error) {
	args := make([]reflect.Value, 0)
	activityFnT := activityFn.Type()

	input := 0
	for i := 0; i < activityFnT.NumIn(); i++ {
		argT := activityFnT.In(i)

		// Insert context if requested
		if i == 0 && isContext(argT) {
			args = append(args, reflect.ValueOf(ctx))
			continue
		}

		arg := reflect.New(argT).Interface()
		err := converter.DefaultConverter.From(inputs[input], arg)
		if err != nil {
			return nil, errors.Wrap(err, "could not convert activity input")
		}

		args = append(args, reflect.ValueOf(arg).Elem())

		input++
	}

	return args, nil
}

func isContext(inType reflect.Type) bool {
	contextElem := reflect.TypeOf((*context.Context)(nil)).Elem()
	return inType != nil && inType.Implements(contextElem)
}