package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var strToFind string = "go"

type task struct {
	url       string
	count     int
	errorText string
}

//--------------------getCount-----------------------------
func getCount(t *task) {

	resp, err := http.Get(t.url)
	if err != nil {
		t.errorText = err.Error()
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.errorText = err.Error()
		return
	}
	t.count = strings.Count(string(body), strToFind)
}

//----------------------main----------------------------
func main() {

	maxWorkers := 5
	workers := 0

	file, err := os.Open("url.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	taskCh := make(chan task)
	doneCh := make(chan bool)
	processedTaskCh := make(chan task)
	processedDoneCh := make(chan bool)

	// -------------------------------------
	go func() {
		totalKolvo := 0
		for pTask := range processedTaskCh {
			errText := ""
			if pTask.errorText != "" {
				errText = "Error: " + pTask.errorText
			}
			fmt.Println("Count for", pTask.url, "-", pTask.count, errText)
			totalKolvo += pTask.count
		}
		fmt.Println("Total: ", totalKolvo)
		processedDoneCh <- true
	}()

	for scanner.Scan() {
		if workers < maxWorkers {
			workers++
			go func() {
				for task := range taskCh {
					getCount(&task)
					processedTaskCh <- task
				}
				doneCh <- true
			}()
		}
		url := scanner.Text()
		taskCh <- task{url: url}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
	close(taskCh)

	for i := 0; i < workers; i++ {
		<-doneCh
	}
	close(processedTaskCh)
	<-processedDoneCh
}
