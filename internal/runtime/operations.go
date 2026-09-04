package runtime

import (
	"context"
	"fmt"
)

// ExecuteOperation applies one lifecycle action to an ordered service list.
// The caller controls ordering so dependency semantics remain explicit.
// Every service gets a result, including services that fail, allowing the CLI
// and UI to render partial outcomes without inventing aggregate success.
func ExecuteOperation(ctx context.Context, runtime Runtime, action string, scopes []Scope) OperationResult {
	result := OperationResult{Action: action, Services: make([]ServiceOperation, 0, len(scopes))}
	for _, scope := range scopes {
		operation := ServiceOperation{Service: ServiceStatus{Label: operationLabel(scope)}}
		before, beforeErr := runtime.Status(ctx, scope)
		if len(before) > 0 {
			operation.Service = before[0]
		}
		operation.Before = operation.Service.State
		if beforeErr != nil {
			operation.Message = beforeErr.Error()
		}

		var err error
		switch action {
		case "start":
			err = runtime.Start(ctx, scope)
		case "stop":
			err = runtime.Stop(ctx, scope)
		case "restart":
			err = runtime.Restart(ctx, scope)
		default:
			err = fmt.Errorf("unsupported runtime operation %q", action)
		}
		if err != nil {
			operation.Err = err
			if result.Err == nil {
				result.Err = err
			}
		}
		after, afterErr := runtime.Status(ctx, scope)
		if len(after) > 0 {
			operation.Service = after[0]
		}
		operation.After = operation.Service.State
		operation.Changed = operation.Before != operation.After
		if afterErr != nil && operation.Err == nil {
			operation.Err = afterErr
			if result.Err == nil {
				result.Err = afterErr
			}
		}
		result.Services = append(result.Services, operation)
	}
	return result
}

func operationLabel(scope Scope) string {
	if scope.Service == "nginx" {
		return "nginx"
	}
	label := "php-fpm " + scope.Version
	if scope.Dedicated {
		return label + " (dedicated)"
	}
	return label + " (shared)"
}
