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

	statsMinLimit = 10
	statsCounter = 0
	statsSync sync.Mutex

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
	statsSync.Lock()
	statsCounter++
	statsSync.Unlock()

	return statsCounter < statsMinLimit
}

func periodic() error {
	uploadTicker := time.NewTicker( 5 * time.Second)
	minuteTicker := getMinuteTicker()

	for {
		select {
		case <-uploadTicker.C:
			fmt.Println("Tick")
			uploadCounters()
		case <-minuteTicker.C:
			minuteTicker = getMinuteTicker()

			//reset map
			m = make(map[string]*counters)

			//reset stats limit
			statsCounter = 0
		}
	}
}

func getMinuteTicker() *time.Ticker {
	//return new ticker that triggers on the minute
	return time.NewTicker(time.Second * time.Duration(60-time.Now().Second()))
}

func uploadCounters() error {

	fmt.Println("Uploading map: ", m)

	//read json store
	jsonFile, err := os.OpenFile("store.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

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
	go periodic()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
