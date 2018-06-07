package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", hello)
	port := os.Getenv("PORT")
	fmt.Printf("Listening on %s...", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Recieved request ", time.Now())
	response := fmt.Sprintf(`{"greeting": "hello", "instance_index": %q, "instance_guid": %q}`, os.Getenv("CF_INSTANCE_INDEX"), os.Getenv("INSTANCE_GUID"))

	res.WriteHeader(200)
	res.Write([]byte(response))
}
