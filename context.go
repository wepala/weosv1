package weos

import "golang.org/x/net/context"

type ContextKey string

//based on recommendations here https://www.calhoun.io/pitfalls-of-context-values-and-how-to-avoid-or-mitigate-them/

//add more keys here if needed
const ACCOUNT_ID ContextKey = "ACCOUNT_ID"
const USER_ID ContextKey = "USER_ID"
const LOG_LEVEL ContextKey = "LOG_LEVEL"

//---- Context Getters

//Get account info from context
func GetAccount(ctx context.Context) string {
	if value, ok := ctx.Value(ACCOUNT_ID).(string); ok {
		return value
	}
	return ""
}

//Get user info from context
func GetUser(ctx context.Context) string {
	if value, ok := ctx.Value(USER_ID).(string); ok {
		return value
	}
	return ""
}

//Get user info from context
func GetLogLevel(ctx context.Context) string {
	if value, ok := ctx.Value(LOG_LEVEL).(string); ok {
		return value
	}
	return ""
}
