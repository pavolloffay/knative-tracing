package httplogging

import (
	"log"
	"net/http"
)

type LoggingHandler struct {
	Wrapped http.Handler
}

var _ http.Handler = (*LoggingHandler)(nil)

func (l *LoggingHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Println("Request headers:")
	for k, v := range request.Header {
		log.Printf("\t%s: %s\n", k, v)
	}

	l.Wrapped.ServeHTTP(writer, request)

	log.Println("Response headers:")
	for k, v := range writer.Header() {
		log.Printf("\t%s: %s\n", k, v)
	}
	log.Printf("\n\n")
}
