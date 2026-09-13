package http

import "net/http"

type StaticContentHandler struct{}

func NewStaticContentHandler() *StaticContentHandler {
	return &StaticContentHandler{}
}

func (handler *StaticContentHandler) GetStaticContentURL(w http.ResponseWriter, r *http.Request) error {
	return nil
}
