package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
)

var wg sync.WaitGroup

func readFile() []string {
	filePath := os.Args[1]
	readFile, err := os.Open(filePath)

	if err != nil {
		log.Println(err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}

	readFile.Close()

	return fileLines
}

func getRepositoryName(repositoryURL string) string {
	repositoryURLSplitted := strings.Split(repositoryURL, "/")
	return repositoryURLSplitted[len(repositoryURLSplitted)-1]
}

func saveError(msg string) {
	f, err := os.OpenFile("error.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(msg); err != nil {
		log.Println(err)
	}
}

func downloadRepository(ch chan string) string {
	if len(ch) == 0 {
		return "Finished"
	} else {
		defer wg.Done()
		url := <-ch
		ctx, _ := context.WithTimeout(context.Background(), 15*time.Minute)
		log.Printf("Downloading %s ...", url)
		_, err := git.PlainCloneContext(ctx, "/tmp/foo/"+getRepositoryName(url), false, &git.CloneOptions{
			URL:      url,
			Progress: os.Stdout,
		})
		if err != nil {
			log.Println("ERROR -", url)
			saveError(url + "\n")
		}

		return downloadRepository(ch)
	}
}

func main() {
	urls := readFile()

	max := len(urls)

	ch := make(chan string, max)

	for _, url := range urls {
		ch <- url
	}

	wg.Add(max)
	for i := 0; i < 4; i++ {
		go downloadRepository(ch)
	}

	wg.Wait()
}
