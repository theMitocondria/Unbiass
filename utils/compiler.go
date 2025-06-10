package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"time"

)

const (
	totalWorkers    = 15
	jobQueueSize    = 100
	workerTimeout   = 50 * time.Second
	containerTTL    = 5 * time.Minute
	healthCheckPort = 8080
)

type Worker struct {
	Port        int
	Lang        string
	LastUsed    time.Time
	mu          sync.Mutex
	isAvailable bool
}

type CompileResponse struct {
    Output struct {
        CodeOutput string `json:"code_output"`
        CodeError  string `json:"code_error"`
    } `json:"output"`
    Error string `json:"error"`
}


type Job struct {
	Code     string
	Input    string
	Lang     string
	ResultCh chan<- Result
}

type Result struct {
	Response CompileResponse
	Error    error
}


var (
	workers   []*Worker
	jobChan   chan Job
	workerWg  sync.WaitGroup
	images    = []string{"lcpolice/python:latest", "lcpolice/js:latest", "lcpolice/java:latest", "lcpolice/cpp:latest"}
)

func CompilerInit() {
	// First, cleanup any existing containers on our port range
	for port := healthCheckPort; port < healthCheckPort+15; port++ {
		cleanup := exec.Command("docker", "rm", "-f", fmt.Sprintf("compiler-%d", port))
		_ = cleanup.Run() // Ignore errors as containers might not exist
	}

	// Pull Docker images in parallel
	var wg sync.WaitGroup
	for _, image := range images {
		wg.Add(1)
		go func(img string) {
			defer wg.Done()
			cmd := exec.Command("docker", "pull", img)
			if err := cmd.Run(); err != nil {
				fmt.Printf("Error pulling image %s: %v\n", img, err)
			}
		}(image)
	}
	wg.Wait()

	// Initialize workers
	workers = make([]*Worker, totalWorkers)
	for i := 0; i < totalWorkers; i++ {
		workers[i] = &Worker{
			Port:        healthCheckPort + i + 1,
			isAvailable: true,
		}
	}

	// Create job channel and start workers
	jobChan = make(chan Job, jobQueueSize)
	startWorkers()
}

func startWorkers() {
	for _, w := range workers {
		workerWg.Add(1)
		go func(w *Worker) {
			defer workerWg.Done()
			for job := range jobChan {
				processJob(w, job)
			}
		}(w)
	}
}

func processJob(w *Worker, job Job) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Always cleanup existing container first
	_ = StopContainer(w.Port) // Cleanup any existing container

	// Start new container
	if err := StartContainer(w.Port, job.Lang); err != nil {
		job.ResultCh <- Result{Error: fmt.Errorf("container startup failed: %v", err)}
		w.isAvailable = true // Mark worker as available even on error
		return
	}

	// Set worker state
	w.Lang = job.Lang
	w.LastUsed = time.Now()
	w.isAvailable = false

	// Execute compilation
	client := &http.Client{Timeout: 10 * time.Second}
	result, err := CompileCode(job.Code, job.Input, job.Lang, w.Port, client)

	// Always cleanup after compilation
	_ = StopContainer(w.Port)

	if err != nil {
		job.ResultCh <- Result{Error: err}
	} else {
		job.ResultCh <- Result{Response: result}
	}

	w.isAvailable = true
}

func StartContainer(port int, lang string) error {
    StopContainer(port) 
    image, err := GetImageName(lang)
    if err != nil {
        return err
    }

    cmd := exec.Command("docker", "run", "-d", "--rm",
        "--name", fmt.Sprintf("compiler-%d", port), "-p", fmt.Sprintf("%d:8080", port), image)

    output, err := cmd.CombinedOutput() // Capture output for debugging
    if err != nil {
        return fmt.Errorf("failed to start container: %w, output: %s", err, string(output))
    }

    return waitForContainerReady(port, 10*time.Second)
}

func StopContainer(port int) error {
	// First try to stop the container
	stopCmd := exec.Command("docker", "stop", fmt.Sprintf("compiler-%d", port))
	_ = stopCmd.Run() // Ignore error as container might not exist

	// Force remove the container if it exists
	rmCmd := exec.Command("docker", "rm", "-f", fmt.Sprintf("compiler-%d", port))
	if err := rmCmd.Run(); err != nil {
		// Only log the error but don't fail - container might not exist
		fmt.Printf("Error removing container compiler-%d: %v\n", port, err)
	}
	return nil
}

func waitForContainerReady(port int, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    client := http.Client{Timeout: 1 * time.Second}

    for time.Now().Before(deadline) {
        resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", port))
        if err != nil {
            fmt.Printf("Health check error: %v\n", err) // Add debug logging
            time.Sleep(300 * time.Millisecond)
            continue
        }
        if resp.StatusCode < 500 {
            resp.Body.Close()
            return nil
        }
        resp.Body.Close()
        time.Sleep(300 * time.Millisecond)
    }
    return fmt.Errorf("container on port %d not ready after %v", port, timeout)
}

func CompileCode(code, input, lang string, port int, client *http.Client) (CompileResponse, error) {
	resp, err := client.Post(
		fmt.Sprintf("http://localhost:%d/compile", port),
		"application/json",
		bytes.NewReader([]byte(fmt.Sprintf(`{"code":%q,"input":%q,"lang":%q}`, code, input, lang))),
	)
	if err != nil {
		return CompileResponse{}, fmt.Errorf("compile request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return CompileResponse{}, fmt.Errorf("compile error: %s", string(body))
	}

	var result CompileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return CompileResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, nil
}

func GetImageName(lang string) (string, error) {
	switch lang {
	case "python":
		return "lcpolice/python:latest", nil
	case "js":
		return "lcpolice/js:latest", nil
	case "java":
		return "lcpolice/java:latest", nil
	case "cpp":
		return "lcpolice/cpp:latest", nil
	default:
		return "", errors.New("unsupported language")
	}
}

func HandleCompile(Code string, Input string, Lang string) (CompileResponse, error) {

	if Code == "" || Input == "" || Lang == "" {
		return CompileResponse{}, errors.New("code, input, and language must be provided")
	}

	resultCh := make(chan Result)
	job := Job{
		Code:     Code,
		Input:    Input,
		Lang:     Lang,
		ResultCh: resultCh,
	}

	select {
	case jobChan <- job:
	case <-time.After(2 * time.Second):
		return CompileResponse{}, errors.New("server overloaded, please try again")
	}

	select {
	case res := <-resultCh:
		if res.Error != nil {
			return CompileResponse{}, res.Error
		}
		return res.Response, nil
	case <-time.After(workerTimeout):
		return CompileResponse{}, errors.New("compilation timeout")
	}
}
func Cleanup() {
	close(jobChan)
	workerWg.Wait()
}
