package infrastructure

import "github.com/sendelius/go/server"

type Handler struct {
	response server.Response
	request  server.Request
}
