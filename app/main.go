package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	QuizUrl        string `yaml:"quizUrl"`
	QuizListUrl    string `yaml:"quizListUrl"`
	CheckerService string `yaml:"checkerServiceUrl"`
}

type Quiz struct {
	ID        string     `yaml:"id"`
	Title     string     `yaml:"title"`
	Questions []Question `yaml:"questions"`
}

type QuizSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type Question struct {
	Text    string   `yaml:"text"`
	ID      int      `yaml:"id"`
	Type    string   `yaml:"type"`
	Options []string `yaml:"options"`
	Answers []int    `yaml:"answers"`
	Code    string   `yaml:"code,omitempty"` // New field for code snippets
}

var config Config
var quizzes []Quiz

// Custom function to add two integers
func add(a, b int) int {
	return a + b
}

func loadConfig() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}
}

func fetchQuizzesFromExternalService() ([]QuizSummary, error) {
	resp, err := http.Get(config.QuizListUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quizzes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var summaries []QuizSummary
	err = json.Unmarshal(body, &summaries)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return summaries, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	summaries, err := fetchQuizzesFromExternalService()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load quizzes: %v", err), http.StatusInternalServerError)
		return
	}

	tmplPath := filepath.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, summaries)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func toJson(v interface{}) (string, error) {
	// Marshal the data into JSON format.
	jsonData, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	// Convert the JSON bytes to a string and return it.
	return string(jsonData), nil
}

func quizHandler(w http.ResponseWriter, r *http.Request) {
	// Extract quiz ID from the URL path
	quizID := r.URL.Path[len("/quiz/"):]

	// Fetch quiz data from the external service
	quiz, err := fetchQuizFromExternalService(quizID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load quiz: %v", err), http.StatusInternalServerError)
		return
	}

	// Load the quiz template
	tmpl := template.Must(template.New("quiz.html").Funcs(template.FuncMap{
		"toJson": toJson,
		"add":    add,
	}).ParseFiles("templates/quiz.html"))

	// Render the template with the quiz data
	err = tmpl.Execute(w, quiz)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}
}

func fetchQuizFromExternalService(quizID string) (*Quiz, error) {
	url := fmt.Sprintf("%s%s", config.QuizUrl, quizID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quiz from external service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var quiz Quiz
	err = yaml.Unmarshal(body, &quiz)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	return &quiz, nil
}

// isPrime checks if a number is prime.
func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// cpuIntensiveTask performs a CPU-intensive task.
func cpuIntensiveTask() {
	const max = 9000000
	for i := 2; i < max; i++ {
		isPrime(i)
	}
}

func cpuintensiveHandler(w http.ResponseWriter, r *http.Request) {
	go cpuIntensiveTask()
	fmt.Fprintf(w, "CPU-intensive task accepted")
}

func serverConfigHandler(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		http.Error(w, "Failed to get hostname", http.StatusInternalServerError)
		return
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		http.Error(w, "Failed to get IP address", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Hostname: %s\n", hostname)
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Fprintf(w, "IP Address: %s\n", ipnet.IP.String())
			}
		}
	}

	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		http.Error(w, "Failed to read config file", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "\nConfig File:\n%s\n", configData)
}

func evaluateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var submittedAnswers map[string][]string

	// Parse the incoming JSON data
	err := json.NewDecoder(r.Body).Decode(&submittedAnswers)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Extract the quizId
	quizId, ok := submittedAnswers["quizId"]
	if !ok || len(quizId) == 0 {
		http.Error(w, "Quiz ID is missing", http.StatusBadRequest)
		return
	}

	delete(submittedAnswers, "quizId")
	// Prepare the payload for the external service
	payload := map[string]interface{}{
		"quizId":  quizId[0],
		"answers": submittedAnswers,
	}

	// Use channels to handle asynchronous processing
	resultChan := make(chan []byte)
	errorChan := make(chan error)

	go func() {
		// Make the external API call for evaluation asynchronously
		evaluationResults, err := evaluateViaExternalService(config.CheckerService, payload)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- evaluationResults
		}
	}()

	// Wait for the result or error and respond accordingly
	select {
	case evaluationResults := <-resultChan:
		w.Header().Set("Content-Type", "application/json")
		w.Write(evaluationResults)
	case err := <-errorChan:
		http.Error(w, fmt.Sprintf("Error evaluating quiz: %v", err), http.StatusInternalServerError)
	}
}

func evaluateViaExternalService(apiURL string, payload map[string]interface{}) ([]byte, error) {
	// Convert the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Create a new HTTP POST request with the JSON data
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the HTTP request to the external service
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call external service: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Check for non-200 HTTP status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	return body, nil
}

func main() {
	loadConfig()

	http.HandleFunc("/", handler)
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/quiz/", quizHandler)
	http.HandleFunc("/task/cpuintensive", cpuintensiveHandler)
	http.HandleFunc("/server/config", serverConfigHandler)
	http.HandleFunc("/evaluate", evaluateHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
