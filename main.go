package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bhavyagada/xeneinterpreter/runtime"
	"github.com/bhavyagada/xeneinterpreter/token"
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
	http.HandleFunc("/", interpretHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func interpretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InterpretRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	code, err := fromString(req.Code)
	if err != nil {
		res := InterpretResponse{
			Success: false,
			Message: fmt.Sprintf("Code parsing failed: %v", err),
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	callableCode, ok := code.(runtime.Callable)
	if !ok {
		res := InterpretResponse{
			Success: false,
			Message: "Parsed code is not callable",
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	log.Printf("Processing code: %s", req.Code)
	tokensChan := iterateExecutableTokens(callableCode)
	var tokens []*token.Token
	for t := range tokensChan {
		// Print token details
		log.Printf("Token found: Lit=%s, Type=%d, Pos=(Offset=%d, Line=%d, Column=%d)",
			string(t.Lit), t.Type, t.Pos.Offset, t.Pos.Line, t.Pos.Column)
		tokens = append(tokens, t)
	}
	log.Printf("Total tokens collected: %d", len(tokens))
	if tokens == nil {
		log.Printf("Tokens slice is nil")
	} else {
		log.Printf("Collected tokens: %v", tokens)
	}

	// Convert to string slice for response
	tokenStrings := make([]string, len(tokens))
	for i, t := range tokens {
		tokenStrings[i] = string(t.Lit)
	}

	var testCaseResults []TestCaseResult
	for _, tc := range req.Params {
		result := runTestCase(code, tc)
		testCaseResults = append(testCaseResults, result)
	}

	var hiddenTestCaseResults []TestCaseResult
	for _, tc := range req.HiddenParams {
		result := runTestCase(code, tc)
		hiddenTestCaseResults = append(hiddenTestCaseResults, result)
	}

	res := InterpretResponse{
		Success:              true,
		TestCaseResult:       testCaseResults,
		HiddenTestCaseResult: hiddenTestCaseResults,
		Tokens:               tokenStrings,
		Message:              "",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func runTestCase(code runtime.Value, tc TestCase) TestCaseResult {
	callableCode, ok := code.(runtime.Callable)
	if !ok {
		return TestCaseResult{
			Result:  false,
			Message: "Parsed code is not callable",
		}
	}

	var inpVal runtime.Value = 0
	if tc.Input != "" {
		inpCode, err := fromString(tc.Input)
		if err != nil {
			return TestCaseResult{
				Result:  false,
				Message: fmt.Sprintf("Input parsing failed: %v", err),
			}
		}
		inpVal, err = exec(inpCode, 100*time.Millisecond)
		if err != nil {
			return TestCaseResult{
				Result:  false,
				Message: fmt.Sprintf("Input execution failed: %v", err),
			}
		}
	}

	ctx := runtime.NewContext(runtime.DefaultTimeout)
	ctx.SetVariable("input", inpVal)
	output, err := ctx.Call(callableCode)
	if err != nil {
		return TestCaseResult{
			Result:  false,
			Message: fmt.Sprintf("Code execution failed: %v", err),
		}
	}

	outputStr := runtime.ToString(output, true)
	if outputStr == tc.Expected {
		return TestCaseResult{
			Result:  true,
			Message: "",
		}
	}
	return TestCaseResult{
		Result:  false,
		Message: fmt.Sprintf("Expected %q, got %q", tc.Expected, outputStr),
	}
}
