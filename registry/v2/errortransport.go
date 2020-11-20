package v2

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type HttpStatusError struct {
	Response *http.Response
	Body     []byte // Copied from `Response.Body` to avoid problems with unclosed bodies later. Nobody calls `err.Response.Body.Close()`, ever.
}

func (err *HttpStatusError) Error() string {
	return fmt.Sprintf("http: non-successful response (status=%v body=%q)", err.Response.StatusCode, err.Body)
}

var _ error = &HttpStatusError{}

type ErrorTransport struct {
	Transport http.RoundTripper
	// http.Header
}

func (t *ErrorTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	// ADD HEADERSSSSSS

	// for k, v := range t.Header {
	// 	request.Header[k] = v
	// }

	resp, err := t.Transport.RoundTrip(request)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("http: failed to read response body (status=%v, err=%q)", resp.StatusCode, err)
		}

		return nil, &HttpStatusError{
			Response: resp,
			Body:     body,
		}
	}

	return resp, err
}
