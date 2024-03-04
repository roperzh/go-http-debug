package httpdebug

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
)

type message struct {
	RawHeaders string `json:"raw_headers"`
	Body       string `json:"body"`
}

type transaction struct {
	Status   int     `json:"status"`
	Path     string  `json:"path"`
	Method   string  `json:"method"`
	Request  message `json:"request"`
	Response message `json:"response"`
}

type WebUIHandler struct {
	transactions  []*transaction
	mu            sync.RWMutex
	addr          string
	serverStarted sync.Once
	skipMessage   bool
}

func NewWebUIHandler(opts ...func(*WebUIHandler)) *WebUIHandler {
	h := &WebUIHandler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func WithAddress(addr string) func(*WebUIHandler) {
	return func(h *WebUIHandler) {
		h.addr = addr
	}
}

func WithoutMessage() func(*WebUIHandler) {
	return func(h *WebUIHandler) {
		h.skipMessage = true
	}
}

func (h *WebUIHandler) Wrap(next http.HandlerFunc) http.HandlerFunc {
	go h.serveUI()

	return func(w http.ResponseWriter, req *http.Request) {
		t, err := generateTransaction(next, w, req)
		if err != nil {
			log.Printf("ERR: %s", err)
			return
		}
		h.mu.Lock()
		h.transactions = append(h.transactions, t)
		h.mu.Unlock()
	}
}

var defaultHookUI = WebUIHandler{addr: ":3141"}

func WebUI(next http.HandlerFunc) http.HandlerFunc {
	return defaultHookUI.Wrap(next)
}

//go:embed assets/**
var assets embed.FS

func (h *WebUIHandler) serveUI() {
	h.serverStarted.Do(func() {
		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("GET /data", func(w http.ResponseWriter, req *http.Request) {
				h.mu.RLock()
				res, err := json.Marshal(h.transactions)
				h.mu.RUnlock()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Write(res)
			})

			assetsFS, err := fs.Sub(assets, "assets")
			if err != nil {
				log.Fatal(err)
			}

			mux.Handle("GET /", http.FileServer(http.FS(assetsFS)))

			if !h.skipMessage {
				fmt.Printf(`
            .                .  *       - )-         .
         .      *       o       .       *    o    .   
   o          |                             |         
             -O-.                                     
  .           |              *      .     -O-         -
                .         .        |      *   .       
     *             *              -O-          . *     
           .             *         |     ,              
          .---.                    o       '       .
    =   _/__~0_\_     .  *                   *    .
   = = (_________)             .          *    o
         *               - ) -       *             
  +------------------------------------------------+
  |      VEx is connected and ready to explore.    |
  |      Visit http://localhost%s to start      |
  +------------------------------------------------+

`, h.addr)
			}

			http.ListenAndServe(h.addr, mux)
		}()
	})
}
