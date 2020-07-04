package nnet

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
//	upgrader = websocket.Upgrader{
//		ReadBufferSize:  4096,
//		WriteBufferSize: 4096,
//		CheckOrigin: func(r *http.Request) bool {
//			return true
//		},
//	}
)

type CallbackWsPath func(http.ResponseWriter, *http.Request)

type WsServer struct {
	*Hub
	addr     string
	svr      *http.Server
	upgrader *websocket.Upgrader
	routes 	 map[string]CallbackWsPath
	ws_path  string
}

func NewWsServer(cf *HubConfig, cb ISessionCallback, p IProtocol, addr string, ws_path string, routes map[string]CallbackWsPath) *WsServer {
	s := &WsServer{
		Hub:  newHub(cf, cb, p),
		addr: addr,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  cf.ReadBufSize,
			WriteBufferSize: cf.WriteBufSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		routes:routes,
		ws_path: ws_path,
	}

	if len(s.ws_path) <= 0 {
		s.ws_path = "ws"
	}

	return s
}

func (self *WsServer) Start() error {
	router := mux.NewRouter()
	for k, v := range self.routes {
		router.HandleFunc(k, v)
	}
	router.HandleFunc("/", self.do_homepage)
	router.HandleFunc("/" + self.ws_path, func(w http.ResponseWriter, r *http.Request) {
		self.do_new_session(w, r)
	})
	self.svr = &http.Server{
		Addr:    self.addr,
		Handler: router,
	}
	err := self.svr.ListenAndServe()
	return err
}

func (self *WsServer) do_homepage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method now allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("welcome"))
}

func (self *WsServer) do_new_session(w http.ResponseWriter, r *http.Request) {
	conn, err := self.upgrader.Upgrade(w, r, nil)
	if nil != err {
		return
	}
	self.wg.Add(1)
	go func() {
		ses := newSession(NewWsConn(conn), self)
		ses.Do()
		self.wg.Done()
	}()
}

func (self *WsServer) Stop() error {
	self.svr.Close()
	close(self.chQuit)
	self.wg.Wait()
	return nil
}

func (self *WsServer) DoJob(int) {

}
