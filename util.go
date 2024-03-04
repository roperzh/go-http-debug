package httpdebug

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
)

func formatXML(in []byte) []byte {
	b := &bytes.Buffer{}
	decoder := xml.NewDecoder(bytes.NewReader(in))
	encoder := xml.NewEncoder(b)
	encoder.Indent("", "  ")
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			encoder.Flush()
			return b.Bytes()
		}
		if err != nil {
			log.Printf("ERR: decoding XML: %s", err)
			return in
		}
		err = encoder.EncodeToken(token)
		if err != nil {
			log.Printf("ERR: decoding XML: %s", err)
			return in
		}
	}
}

func formatJSON(in []byte) []byte {
	var data map[string]any
	if err := json.Unmarshal(in, &data); err != nil {
		return in
	}

	var result bytes.Buffer
	enc := json.NewEncoder(&result)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return in
	}

	return result.Bytes()
}

// Duplicate and reset the request body.
func duplicateRequestBody(req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

// Copy from the response recorder to the actual response writer.
func forwardResponse(rec *httptest.ResponseRecorder, w http.ResponseWriter) {
	for k, values := range rec.Header() {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(rec.Code)
	rec.Body.WriteTo(w)
}

func generateTransaction(next http.HandlerFunc, w http.ResponseWriter, req *http.Request) (*transaction, error) {
	resp := httptest.NewRecorder()

	reqBody, err := duplicateRequestBody(req)
	if err != nil {
		next(resp, req)
		forwardResponse(resp, w)
		return nil, err
	}

	next(resp, req)
	defer forwardResponse(resp, w)

	reqHeaders, reqBodyDump, err := dump(req.Header, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	respHeaders, respBodyDump, err := dump(resp.Header(), bytes.NewBuffer(resp.Body.Bytes()))
	if err != nil {
		return nil, err
	}

	return &transaction{
		Status: resp.Code,
		Method: req.Method,
		Path:   req.URL.String(),
		Request: message{
			RawHeaders: reqHeaders,
			Body:       string(reqBodyDump),
		},
		Response: message{
			RawHeaders: respHeaders,
			Body:       string(respBodyDump),
		},
	}, nil
}

func dump(headers http.Header, body *bytes.Buffer) (string, []byte, error) {
	var headerDump bytes.Buffer
	if err := headers.WriteSubset(&headerDump, nil); err != nil {
		return "", nil, err
	}

	content, err := io.ReadAll(body)
	if err != nil {
		return "", nil, err
	}

	ct := headers.Get("Content-Type")
	if ct == "" {
		buf := make([]byte, 512)
		n := copy(buf, body.Bytes())
		ct = http.DetectContentType(buf[:n])
	}

	var bodyDump []byte
	switch ct {
	case "text/xml; charset=utf-8":
		bodyDump = formatXML(content)
	case "text/plain; charset=utf-8",
		"application/x-www-form-urlencoded":
		bodyDump = content
	case "application/json":
		bodyDump = formatJSON(content)
	default:
		bodyDump = []byte(fmt.Sprintf("content-type %s preview not supported", ct))
	}

	return headerDump.String(), bodyDump, nil
}
