package server

type contextKey string

const (
	SessionKey contextKey = "session"
	ParamsKey  contextKey = "pathParams"
)
