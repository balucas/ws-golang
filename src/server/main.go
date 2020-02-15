package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"encoding/json"
	"io/ioutil"
	"os"
)

type counters struct {
	sync.Mutex	`json:"-"`
	View  int	`json:"view"`
	Click int	`json:"click"`
}

var (
	//map to temporarily store counters
	m = make(map[string]*counters)

	content = []string{"sports", "entertainment", "business", "education"}
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {

	data := content[rand.Intn(len(content))] + ":" + time.Now().Format("2006-01-02 15:04")

	//check if content already registered views 
	if _, ok := m[data]; !ok {
		m[data] = &counters{}
	}

	m[data].Lock()
	m[data].View++
	m[data].Unlock()

	err := processRequest(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}

	// simulate random click call
	if rand.Intn(100) < 50 {
		processClick(data)
	}
}

func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(data string) error {
	m[data].Lock()
	m[data].Click++
	m[data].Unlock()

	return nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowed() {
		w.WriteHeader(429)
		return
	}
}

func isAllowed() bool {
	return true
}

func periodicUpload() error {
	uploadTicker := time.NewTicker( 5 * time.Second)

	for {
		select {
		case <-uploadTicker.C:
			fmt.Println("Tick")
			uploadCounters()
		}
	}
}

func uploadCounters() error {
	
	fmt.Println("Uploading")

	//read json store
	jsonFile, err := os.Open("store.json")

	if err!= nil {
		fmt.Println(err)
	}

	byteValue, err := ioutil.ReadAll(jsonFile)

	var store map[string]counters

	//create store if json file is empty
	if len(byteValue) == 0 {
		store = make(map[string]counters)
	}else{
		json.Unmarshal(byteValue, &store)
	}

	//insert values
	for k, v := range m {
		store[k] = *v
	}

	jsonString, err := json.Marshal(store)

	if err != nil {
		fmt.Println(err)
	}

	//write to store
	err = ioutil.WriteFile("store.json", jsonString, 0644)

	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func main() {
	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/stats/", statsHandler)

	//run upload routine
	go periodicUpload()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
