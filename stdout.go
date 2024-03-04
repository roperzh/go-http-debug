package httpdebug

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
)

func RawStdout(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		resp := httptest.NewRecorder()
		reqDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			log.Printf("ERR: dumping request body for logs: %s", err)
		}

		next(resp, req)

		var respHeaders bytes.Buffer
		resp.Header().WriteSubset(&respHeaders, nil)

		log.Printf(`

-------------------

~ REQUEST:
%s ~ RESPONSE:
Status: %d
Headers: %s

%s
-------------------

`, string(reqDump), resp.Code, respHeaders.String(), resp.Body)

		forwardResponse(resp, w)
	}
}

func JSONStdout(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		t, err := generateTransaction(next, w, req)
		if err != nil {
			log.Printf("ERR: %s", err)
			return
		}
		out, err := json.Marshal(t)
		if err != nil {
			out = []byte(fmt.Sprintf("%+v", t))
		}
		log.Println(string(out))
	}
}
