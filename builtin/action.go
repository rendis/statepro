package builtin

import (
	"context"
	"fmt"
	"github.com/rendis/abslog/v3"
	"github.com/rendis/statepro/instrumentation"
	"strings"
)

// LogBasicInfo builtin action (builtin:action:logBasicInfo)
// Logs basic info about the action.
// Valid args:
//   - none
func LogBasicInfo(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
	realityName := args.GetRealityName()
	universeName := args.GetUniverseCanonicalName()
	universeId := args.GetUniverseId()
	actionType := args.GetActionType()

	abslog.InfoCtxf(ctx, "%s action executed in reality '%s' from universe '%s' (universe id: %s)", actionType, realityName, universeName, universeId)

	return nil
}

// LogArgs builtin action (builtin:action:logArgs)
// Logs the action args.
// Valid args:
//   - map[string]any (key: arg name, value: any)
func LogArgs(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
	realityName := args.GetRealityName()
	universeName := args.GetUniverseCanonicalName()
	universeId := args.GetUniverseId()
	actionType := args.GetActionType()

	var argsStr []string
	for k, v := range args.GetAction().Args {
		argsStr = append(argsStr, fmt.Sprintf("%s = %v", k, v))
	}

	abslog.InfoCtxf(
		ctx,
		"%s action executed in reality %s from universe '%s' (universe id: %s) with args: %s",
		actionType, realityName, universeName, universeId, strings.Join(argsStr, ", "),
	)
	return nil
}

// LogArgsWithoutKeys builtin action (builtin:action:logArgsWithoutKeys)
// Logs the action args without keys.
// Valid args:
//   - map[string]any (key: arg name, value: any) -> keys will be ignored
func LogArgsWithoutKeys(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
	realityName := args.GetRealityName()
	universeName := args.GetUniverseCanonicalName()
	universeId := args.GetUniverseId()
	actionType := args.GetActionType()

	var argsStr []string
	for _, v := range args.GetAction().Args {
		argsStr = append(argsStr, fmt.Sprintf("'%v'", v))
	}

	abslog.InfoCtxf(
		ctx,
		"%s action executed in reality %s from universe '%s' (universe id: %s) with args: %s",
		actionType, realityName, universeName, universeId, strings.Join(argsStr, ", "),
	)
	return nil
}

// LogJustArgsValues builtin action (builtin:action:logJustArgs)
// Logs the action args without keys and info about the action.
// Valid args:
//   - map[string]any (key: arg name, value: any) -> keys will be ignored
func LogJustArgsValues(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
	var argsStr []string
	for _, v := range args.GetAction().Args {
		argsStr = append(argsStr, fmt.Sprintf("'%v'", v))
	}
	abslog.InfoCtxf(ctx, "%s", strings.Join(argsStr, ", "))
	return nil
}
