package main

import (
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "os"
  "time"
  "github.com/bhavyagada/xeneinterpreter/runtime"
)

type TestCase struct {
  Input    string `json:"I"`
  Expected string `json:"O"`
}

type InterpretRequest struct {
  Params       []TestCase `json:"Params"`
  HiddenParams []TestCase `json:"HiddenParams"`
  Code         string     `json:"Code"`
}

type TestCaseResult struct {
  Result  bool   `json:"Result"`
  Message string `json:"Message"`
}

type InterpretResponse struct {
  Success              bool             `json:"Success"`
  TestCaseResult       []TestCaseResult `json:"TestCaseResult"`
  HiddenTestCaseResult []TestCaseResult `json:"HiddenTestCaseResult"`
  Tokens               []string         `json:"Tokens"`
  Message              string           `json:"Message"`
}

func main() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
      return
    }

    var req InterpretRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      http.Error(w, "Invalid request", http.StatusBadRequest)
      return
    }

    code, err := fromString(req.Code)
    if err != nil {
      json.NewEncoder(w).Encode(InterpretResponse{Success: false, Message: err.Error()})
      return
    }

    callable, ok := code.(runtime.Callable)
    if !ok {
      json.NewEncoder(w).Encode(InterpretResponse{Success: false, Message: "Code not callable"})
      return
    }

    tokens := getTokens(callable)
    testResults := runTests(callable, req.Params)
    hiddenResults := runTests(callable, req.HiddenParams)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(InterpretResponse{
      Success:              true,
      TestCaseResult:       testResults,
      HiddenTestCaseResult: hiddenResults,
      Tokens:               tokens,
    })
  })

  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
  }
  log.Printf("Starting server on :%s", port)
  log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getTokens(code runtime.Callable) []string {
  tokensChan := iterateExecutableTokens(code)
  var tokens []string
  for t := range tokensChan {
    tokens = append(tokens, string(t.Lit))
  }
  return tokens
}

func runTests(code runtime.Callable, cases []TestCase) []TestCaseResult {
  var results []TestCaseResult
  for _, tc := range cases {
    ctx := runtime.NewContext(runtime.DefaultTimeout)

    var input runtime.Value = 0
    if tc.Input != "" {
      inpCode, err := fromString(tc.Input)
      if err != nil {
        results = append(results, TestCaseResult{Result: false, Message: err.Error()})
        continue
      }
      input, err = exec(inpCode, 100*time.Millisecond)
      if err != nil {
        results = append(results, TestCaseResult{Result: false, Message: err.Error()})
        continue
      }
    }

    ctx.SetVariable("input", input)
    output, err := ctx.Call(code)
    if err != nil {
      results = append(results, TestCaseResult{Result: false, Message: err.Error()})
      continue
    }

    outputStr := runtime.ToString(output, true)
    results = append(results, TestCaseResult{
      Result:  outputStr == tc.Expected,
      Message: fmt.Sprintf("Expected %q, got %q", tc.Expected, outputStr),
    })
  }
  return results
}
