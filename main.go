package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/urfave/negroni"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade error:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write error:", err)
			break
		}
	}
}

func middleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("Content-Type", "application/json")
	next(w, r)
}

func main() {
	port := "5000"

	// root
	rootRouter := mux.NewRouter()

	// websocketRouter
	websocketRouter := rootRouter.NewRoute().PathPrefix("/ws").Subrouter()
	websocketRouter.HandleFunc("", websocketHandler)

	// health
	healthRouter := mux.NewRouter().PathPrefix("/health").Subrouter()
	healthRouter.HandleFunc("", func(w http.ResponseWriter, req *http.Request) {
		data := map[string]string{
			"asdf": "Asdf",
		}
		json.NewEncoder(w).Encode(data)
	})

	// set middleware to healthRouter
	rootRouter.PathPrefix("/health").Handler(negroni.New(
		negroni.HandlerFunc(middleware),
		negroni.Wrap(healthRouter),
	))

	n := negroni.Classic()
	n.UseHandler(rootRouter)

	log.Printf("Server started at port %v", port)
	log.Fatal(http.ListenAndServe(":"+port, n))
}
