package builtin

import (
	"context"
	"log/slog"

	"github.com/rendis/statepro/v3/instrumentation"
)

// LogBasicInfo builtin action (builtin:action:logBasicInfo)
// Logs basic info about the action.
// Valid args:
//   - none
func LogBasicInfo(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
	slog.InfoContext(ctx, "action executed",
		"actionType", args.GetActionType(),
		"reality", args.GetRealityName(),
		"universe", args.GetUniverseCanonicalName(),
		"universeId", args.GetUniverseId(),
	)
	return nil
}

// LogArgs builtin action (builtin:action:logArgs)
// Logs the action args.
// Valid args:
//   - map[string]any (key: arg name, value: any)
func LogArgs(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
	actionArgs := args.GetAction().Args
	attrs := make([]any, 0, 8+2*len(actionArgs))
	attrs = append(attrs,
		"actionType", args.GetActionType(),
		"reality", args.GetRealityName(),
		"universe", args.GetUniverseCanonicalName(),
		"universeId", args.GetUniverseId(),
	)
	for k, v := range actionArgs {
		attrs = append(attrs, k, v)
	}
	slog.InfoContext(ctx, "action executed with args", attrs...)
	return nil
}

// LogArgsWithoutKeys builtin action (builtin:action:logArgsWithoutKeys)
// Logs the action args without keys.
// Valid args:
//   - map[string]any (key: arg name, value: any) -> keys will be ignored
func LogArgsWithoutKeys(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
	actionArgs := args.GetAction().Args
	vals := make([]any, 0, len(actionArgs))
	for _, v := range actionArgs {
		vals = append(vals, v)
	}
	slog.InfoContext(ctx, "action executed with args",
		"actionType", args.GetActionType(),
		"reality", args.GetRealityName(),
		"universe", args.GetUniverseCanonicalName(),
		"universeId", args.GetUniverseId(),
		"args", vals,
	)
	return nil
}

// LogJustArgsValues builtin action (builtin:action:logJustArgs)
// Logs the action args without keys and info about the action.
// Valid args:
//   - map[string]any (key: arg name, value: any) -> keys will be ignored
func LogJustArgsValues(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
	actionArgs := args.GetAction().Args
	vals := make([]any, 0, len(actionArgs))
	for _, v := range actionArgs {
		vals = append(vals, v)
	}
	slog.InfoContext(ctx, "action args", "args", vals)
	return nil
}
