package transport

import "context"

type HandlerFunc func(Context)

type Middleware func(HandlerFunc) HandlerFunc

type RouteRegistrar interface {
	Group(string, ...Middleware) RouteRegistrar
	GET(string, HandlerFunc, ...Middleware)
	POST(string, HandlerFunc, ...Middleware)
	PUT(string, HandlerFunc, ...Middleware)
	PATCH(string, HandlerFunc, ...Middleware)
	DELETE(string, HandlerFunc, ...Middleware)
	Any(string, HandlerFunc, ...Middleware)
}

type Router interface {
	RouteRegistrar
	Use(...Middleware)
}

type Context interface {
	RequestContext() context.Context
	SetRequestContext(context.Context)
	Method() string
	Path() string
	Header(string) string
	SetHeader(string, string)
	Param(string) string
	Query(string) string
	Set(string, any)
	Get(string) (any, bool)
	ShouldBindJSON(any) error
	ShouldBindQuery(any) error
	BindJSON(any) error
	JSON(int, any)
	FileAttachment(string, string)
	AbortWithStatusJSON(int, any)
}

type Request interface {
	Method() string
	Path() string
	Header(string) string
}

type Response interface {
	Status(int)
	JSON(int, any)
}
