package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const workerCount = 5

type Job struct {
	ID  int
	URL string
}

type Result struct {
	Job      Job
	Status   string
	Duration time.Duration
}

func simulateRequest() {
	delay := time.Duration(100+rand.Intn(400)) * time.Millisecond
	time.Sleep(delay)
}

func worker(jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		start := time.Now()
		simulateRequest()
		results <- Result{
			Job:      job,
			Status:   "обработан",
			Duration: time.Since(start),
		}
	}
}

func main() {
	urls := []string{
		"https://example.com",
		"https://go.dev",
		"https://pkg.go.dev",
		"https://github.com",
		"https://golang.org",
		"https://netology.ru",
		"https://example.org/health",
		"https://example.org/api",
	}

	jobs := make(chan Job, len(urls))
	results := make(chan Result, len(urls))

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	for i, url := range urls {
		jobs <- Job{ID: i + 1, URL: url}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	processed := make([]Result, 0, len(urls))
	var total time.Duration
	success := 0

	fmt.Println("Результаты:")
	for result := range results {
		processed = append(processed, result)
		total += result.Duration
		if result.Status == "обработан" {
			success++
		}
		fmt.Printf("%d. %s — %s, %s\n", result.Job.ID, result.Job.URL, result.Status, result.Duration.Round(time.Millisecond))
	}

	var average time.Duration
	if len(processed) > 0 {
		average = total / time.Duration(len(processed))
	}

	fmt.Println()
	fmt.Printf("Обработано: %d из %d\n", len(processed), len(urls))
	fmt.Printf("Успешно: %d\n", success)
	fmt.Printf("Среднее время: %s\n", average.Round(time.Millisecond))
	fmt.Printf("Воркеров: %d\n", workerCount)
}
