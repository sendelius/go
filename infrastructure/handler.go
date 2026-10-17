package infrastructure

import "github.com/sendelius/go/server"

type Handler struct {
	Response server.Response
	Request  server.Request
}
