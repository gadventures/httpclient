package httpclient

import (
	"context"
	"net/http"
)

//ResponseHandler is a function that clients pass to process the response
//after the response was obtained by calling http.Do
type ResponseHandler func(context.Context, *http.Response, error) error
