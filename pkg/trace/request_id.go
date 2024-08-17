package trace

import (
	"context"

	"github.com/gin-gonic/gin"
)

type contextKey string

const requestIdKey = contextKey("requestId")

func WithRequestID(ctx context.Context, requestId string) context.Context {
	if gCtx, ok := ctx.(*gin.Context); ok {
		ctx = gCtx.Request.Context()
	}
	return context.WithValue(ctx, requestIdKey, requestId)
}

func RequestIDFromContext(ctx context.Context) string{
	if ctx == nil {
		return ""
	}
	if gCtx, ok := ctx.(*gin.Context); ok {
		ctx = gCtx.Request.Context()
	}
	if requestId, ok := ctx.Value(requestIdKey).(string); ok {
		return requestId
	}
	return ""
}
