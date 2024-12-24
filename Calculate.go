package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "strconv"
    "strings"
)

type Request struct {
    Expression string `json:"expression"`
}
type Response struct {
    Result string `json:"result,omitempty"`
    Error  string `json:"error,omitempty"`
}
func calculateHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    result, err := Calc(req.Expression)
    var resp Response
    if err != nil {
        if errors.Is(err, errors.New("invalid character")) || errors.Is(err, errors.New("mismatched parentheses")) {
            resp.Error = "Expression is not valid"
            w.WriteHeader(http.StatusUnprocessableEntity)
        } else {
            resp.Error = "Internal server error"
            w.WriteHeader(http.StatusInternalServerError)
        }
    } else {
        resp.Result = fmt.Sprintf("%f", result)
        w.WriteHeader(http.StatusOK)
    }
    json.NewEncoder(w).Encode(resp)
}
func Calc(expression string) (float64, error) {
    tokens := tokenize(expression)
    postfix, err := infixToPostfix(tokens)
    if err != nil {
        return 0, err
    }
    return evaluatePostfix(postfix)
}
func tokenize(expr string) []string {
    var tokens []string
    var currentToken strings.Builder
    for _, char := range expr {
        switch char {
        case ' ':
            continue
        case '+', '-', '*', '/', '(', ')':
            if currentToken.Len() > 0 {
                tokens = append(tokens, currentToken.String())
                currentToken.Reset()
            }
            tokens = append(tokens, string(char))
        default:
            currentToken.WriteRune(char)
        }
    }
    if currentToken.Len() > 0 {
        tokens = append(tokens, currentToken.String())
    }
    return tokens
}
func infixToPostfix(tokens []string) ([]string, error) {
    var output []string
    var operators []string
    for _, token := range tokens {
        if isNumber(token) {
            output = append(output, token)
        } else if token == "(" {
            operators = append(operators, token)
        } else if token == ")" {
            for len(operators) > 0 && operators[len(operators)-1] != "(" {
                output = append(output, operators[len(operators)-1])
                operators = operators[:len(operators)-1]
            } 
            if len(operators) == 0 {
                return nil, errors.New("mismatched parentheses")
            }
            operators = operators[:len(operators)-1]
        } else if isOperator(token) {
            for len(operators) > 0 && precedence(operators[len(operators)-1]) >= precedence(token) {
                output = append(output, operators[len(operators)-1])
                operators = operators[:len(operators)-1]
            }
            operators = append(operators, token)
        } else {
            return nil, fmt.Errorf("invalid character")
        }
    }
    for len(operators) > 0 {
        if operators[len(operators)-1] == "(" {
            return nil, errors.New("mismatched parentheses")
        }
        output = append(output, operators[len(operators)-1])
        operators = operators[:len(operators)-1]
    }
    return output, nil
}
func evaluatePostfix(postfix []string) (float64, error) {
    var stack []float64
    for _, token := range postfix {
        if isNumber(token) {
            num, _ := strconv.ParseFloat(token, 64)
            stack = append(stack, num)
        } else if isOperator(token) {
            if len(stack) < 2 {
                return 0, errors.New("invalid expression")
            }
            b := stack[len(stack)-1]
            a := stack[len(stack)-2]
            stack = stack[:len(stack)-2]
            switch token {
            case "+":
                stack = append(stack, a+b)
            case "-":
                stack = append(stack, a-b)
            case "*":
                stack = append(stack, a*b)
            case "/":
                if b == 0 {
                    return 0, errors.New("division by zero")
                }
                stack = append(stack, a/b)
            default:
                return 0, fmt.Errorf("unknown operator: %s", token)
            }
        } else {
            return 0, fmt.Errorf("invalid token: %s", token)
        }
    }
    if len(stack) != 1 {
        return 0, errors.New("invalid expression")
    }
    return stack[0], nil
}
func isNumber(token string) bool {
    if _, err := strconv.ParseFloat(token, 64); err == nil {
        return true
    }
    return false
}
func isOperator(token string) bool {
    return token == "+" || token == "-" || token == "*" || token == "/"
}
func precedence(op string) int {
    switch op {
    case "+", "-":
        return 1
    case "*", "/":
        return 2
    default:
        return 0
    }
}
func main() {
    http.HandleFunc("/api/v1/calculate", calculateHandler)
    http.ListenAndServe(":8080", nil)
}
