package main
 
import (
    "net/http"
)
 
func SayHello(w http.ResponseWriter, req *http.Request) {
    w.Write([]byte("abcd"))
}
 
func main() {
    http.HandleFunc("/", SayHello)
    http.ListenAndServe(":11112", nil)
 
}