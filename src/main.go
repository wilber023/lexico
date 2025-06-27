package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Estructuras para el análisis
type TokenGroup struct {
	Count  int      `json:"count"`
	Tokens []string `json:"tokens"`
}

type LexicalAnalysis struct {
	ReservedWords TokenGroup `json:"reserved_words"`
	Identifiers   TokenGroup `json:"identifiers"`
	Numbers       TokenGroup `json:"numbers"`
	Symbols       TokenGroup `json:"symbols"`
	Strings       TokenGroup `json:"strings"`
	Errors        TokenGroup `json:"errors"`
	TotalTokens   int        `json:"total_tokens"`
}

type SyntaxAnalysis struct {
	Valid   bool     `json:"valid"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

type SemanticAnalysis struct {
	Valid    bool     `json:"valid"`
	Message  string   `json:"message"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type AnalysisResult struct {
	Lexical  LexicalAnalysis  `json:"lexical"`
	Syntax   SyntaxAnalysis   `json:"syntax"`
	Semantic SemanticAnalysis `json:"semantic"`
}

type CodeRequest struct {
	Code string `json:"code"`
}

// Palabras reservadas de Python
var reservedWords = map[string]bool{
	"def": true, "if": true, "else": true, "return": true, "print": true,
	"for": true, "while": true, "in": true, "and": true, "or": true,
	"not": true, "True": true, "False": true, "None": true,
}

// Símbolos y operadores
var symbols = []string{
	"<=", ">=", "==", "!=", "<<", ">>", "**",
	"(", ")", "[", "]", "{", "}", ",", ":", "=",
	"+", "-", "*", "/", "%", "<", ">", "!", "&", "|", "^", "~",
}

func main() {
	http.HandleFunc("/analyze", enableCORS(analyzeHandler))
	fmt.Println("Servidor corriendo en puerto 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		next(w, r)
	}
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	result := analyzeCode(req.Code)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func analyzeCode(code string) AnalysisResult {
	lexical := performLexicalAnalysis(code)
	syntax := performSyntaxAnalysis(code)
	semantic := performSemanticAnalysis(code)

	return AnalysisResult{
		Lexical:  lexical,
		Syntax:   syntax,
		Semantic: semantic,
	}
}

func performLexicalAnalysis(code string) LexicalAnalysis {
	var reservedTokens, identifiers, numbers, symbolTokens, stringTokens, errors []string
	
	// Remover comentarios y espacios extra
	lines := strings.Split(code, "\n")
	cleanCode := ""
	for _, line := range lines {
		if commentIndex := strings.Index(line, "#"); commentIndex != -1 {
			line = line[:commentIndex]
		}
		cleanCode += line + "\n"
	}

	// Expresiones regulares para tokens
	stringRegex := regexp.MustCompile(`"([^"]*)"`)
	numberRegex := regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	identifierRegex := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)

	// Extraer strings
	stringMatches := stringRegex.FindAllString(cleanCode, -1)
	for _, match := range stringMatches {
		stringTokens = append(stringTokens, match)
	}

	// Remover strings para análisis posterior
	cleanCode = stringRegex.ReplaceAllString(cleanCode, " ")

	// Extraer números
	numberMatches := numberRegex.FindAllString(cleanCode, -1)
	for _, match := range numberMatches {
		numbers = append(numbers, match)
	}

	// Extraer identificadores y palabras reservadas
	identifierMatches := identifierRegex.FindAllString(cleanCode, -1)
	for _, match := range identifierMatches {
		if reservedWords[match] {
			reservedTokens = append(reservedTokens, match)
		} else {
			identifiers = append(identifiers, match)
		}
	}

	// Extraer símbolos
	for _, symbol := range symbols {
		count := strings.Count(cleanCode, symbol)
		for i := 0; i < count; i++ {
			symbolTokens = append(symbolTokens, symbol)
		}
	}

	// Buscar errores léxicos (caracteres no reconocidos)
	// Simplificado: buscar caracteres especiales no reconocidos
	invalidChars := regexp.MustCompile(`[^\w\s\(\)\[\]\{\},:=+\-*/%<>!&|^~".]`)
	errorMatches := invalidChars.FindAllString(cleanCode, -1)
	for _, match := range errorMatches {
		errors = append(errors, match)
	}

	totalTokens := len(reservedTokens) + len(identifiers) + len(numbers) + len(symbolTokens) + len(stringTokens)

	return LexicalAnalysis{
		ReservedWords: TokenGroup{Count: len(reservedTokens), Tokens: removeDuplicates(reservedTokens)},
		Identifiers:   TokenGroup{Count: len(identifiers), Tokens: removeDuplicates(identifiers)},
		Numbers:       TokenGroup{Count: len(numbers), Tokens: removeDuplicates(numbers)},
		Symbols:       TokenGroup{Count: len(symbolTokens), Tokens: removeDuplicates(symbolTokens)},
		Strings:       TokenGroup{Count: len(stringTokens), Tokens: removeDuplicates(stringTokens)},
		Errors:        TokenGroup{Count: len(errors), Tokens: removeDuplicates(errors)},
		TotalTokens:   totalTokens,
	}
}

func performSyntaxAnalysis(code string) SyntaxAnalysis {
	var errors []string
	
	lines := strings.Split(code, "\n")
	
	// Verificar estructura básica de función
	hasFunctionDef := false
	hasReturn := false
	hasIf := false
	parenthesesCount := 0
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		lineNum := i + 1
		
		// Verificar definición de función
		if strings.HasPrefix(line, "def ") {
			hasFunctionDef = true
			if !strings.Contains(line, ":") {
				errors = append(errors, fmt.Sprintf("Línea %d: Falta ':' al final de la definición de función", lineNum))
			}
		}
		
		// Verificar estructura de if
		if strings.HasPrefix(line, "if ") {
			hasIf = true
			if !strings.Contains(line, ":") {
				errors = append(errors, fmt.Sprintf("Línea %d: Falta ':' al final de la declaración if", lineNum))
			}
		}
		
		// Verificar else
		if strings.HasPrefix(line, "else:") {
			if !hasIf {
				errors = append(errors, fmt.Sprintf("Línea %d: 'else' sin 'if' correspondiente", lineNum))
			}
		}
		
		// Verificar return
		if strings.Contains(line, "return") {
			hasReturn = true
		}
		
		// Contar paréntesis
		parenthesesCount += strings.Count(line, "(") - strings.Count(line, ")")
		
		// Verificar indentación básica
		if (strings.HasPrefix(line, "if ") || strings.HasPrefix(line, "else:") || strings.HasPrefix(line, "def ")) && i+1 < len(lines) {
			nextLine := lines[i+1]
			if strings.TrimSpace(nextLine) != "" && !strings.HasPrefix(nextLine, "    ") && !strings.HasPrefix(nextLine, "\t") {
				errors = append(errors, fmt.Sprintf("Línea %d: Indentación incorrecta", i+2))
			}
		}
	}
	
	// Verificar estructura general
	if !hasFunctionDef {
		errors = append(errors, "No se encontró definición de función")
	}
	
	if !hasReturn && hasFunctionDef {
		errors = append(errors, "La función no tiene declaración return")
	}
	
	if parenthesesCount != 0 {
		errors = append(errors, "Paréntesis no balanceados")
	}
	
	valid := len(errors) == 0
	message := "Estructura sintáctica válida"
	if !valid {
		message = "Se encontraron errores sintácticos"
	}
	
	return SyntaxAnalysis{
		Valid:   valid,
		Message: message,
		Errors:  errors,
	}
}

func performSemanticAnalysis(code string) SemanticAnalysis {
    var errors []string
    var warnings []string

    lines := strings.Split(code, "\n")
    definedFunctions := make(map[string]int) // Cambia a map[string]int para guardar cantidad de parámetros
    definedVariables := make(map[string]bool)
    usedFunctions := make(map[string][]int) // Guarda los conteos de argumentos usados
    usedVariables := make(map[string]bool)

    // Primera pasada: identificar definiciones
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        // Funciones definidas
        if strings.HasPrefix(line, "def ") {
            funcName := extractFunctionName(line)
            paramCount := 0
            // Extraer cantidad de parámetros de la definición
            re := regexp.MustCompile(`def\s+[a-zA-Z_][a-zA-Z0-9_]*\s*\(([^)]*)\)`)
            match := re.FindStringSubmatch(line)
            if len(match) > 1 {
                params := strings.Split(match[1], ",")
                for _, p := range params {
                    if strings.TrimSpace(p) != "" {
                        paramCount++
                    }
                }
            }
            if funcName != "" {
                definedFunctions[funcName] = paramCount
            }
        }

        // Variables definidas (asignaciones)
        if strings.Contains(line, "=") && !strings.Contains(line, "==") && !strings.Contains(line, "<=") && !strings.Contains(line, ">=") && !strings.Contains(line, "!=") {
            varName := extractVariableName(line)
            if varName != "" {
                definedVariables[varName] = true
            }
        }
    }

    // Segunda pasada: identificar usos
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        // Llamadas a funciones (extraer nombre y cantidad de argumentos)
        reCall := regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*\(([^)]*)\)`)
        matches := reCall.FindAllStringSubmatch(line, -1)
        for _, match := range matches {
            funcName := match[1]
            args := strings.TrimSpace(match[2])
            argCount := 0
            if args != "" {
                // Contar argumentos, ignorando espacios vacíos
                parts := strings.Split(args, ",")
                for _, p := range parts {
                    if strings.TrimSpace(p) != "" {
                        argCount++
                    }
                }
            }
            usedFunctions[funcName] = append(usedFunctions[funcName], argCount)
        }

        // Uso de variables
        varUses := extractVariableUses(line)
        for _, varUse := range varUses {
            usedVariables[varUse] = true
        }
    }

    // Verificar que las funciones usadas estén definidas y argumentos correctos
    for funcName, calls := range usedFunctions {
        paramCount, defined := definedFunctions[funcName]
        if !defined && !isBuiltinFunction(funcName) {
            errors = append(errors, fmt.Sprintf("Función '%s' usada pero no definida", funcName))
        } else if defined {
            for _, argCount := range calls {
                if argCount != paramCount {
                    errors = append(errors, fmt.Sprintf("La función '%s' requiere %d argumento(s), pero se llamó con %d", funcName, paramCount, argCount))
                }
            }
        }
    }

    // Verificar que las variables usadas estén definidas
    for varName := range usedVariables {
        if !definedVariables[varName] && !isFunctionParameter(varName, code) {
            errors = append(errors, fmt.Sprintf("Variable '%s' usada pero no definida", varName))
        }
    }

    // Verificar tipos básicos (simplificado)
    if strings.Contains(code, "factorial(") {
        // Verificar que los argumentos pasados a factorial sean números o variables
        factorialCalls := extractFactorialCalls(code)
        for _, arg := range factorialCalls {
            if !isNumber(arg) && !definedVariables[arg] && !isFunctionParameter(arg, code) {
                warnings = append(warnings, fmt.Sprintf("Argumento '%s' pasado a factorial podría no ser un número", arg))
            }
        }
    }

    // Verificar llamadas recursivas
    if strings.Contains(code, "def factorial") && strings.Contains(code, "factorial(") {
        if !strings.Contains(code, "if") || !strings.Contains(code, "return") {
            warnings = append(warnings, "Función recursiva sin condición base clara")
        }
    }

    valid := len(errors) == 0
    message := "Análisis semántico válido"
    if !valid {
        message = "Se encontraron errores semánticos"
    } else if len(warnings) > 0 {
        message = "Análisis semántico válido "
    }

    return SemanticAnalysis{
        Valid:    valid,
        Message:  message,
        Errors:   errors,
        Warnings: warnings,
    }
}
func extractFunctionName(line string) string {
	re := regexp.MustCompile(`def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractVariableName(line string) string {
	re := regexp.MustCompile(`^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*=`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractFunctionCalls(line string) []string {
	re := regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	matches := re.FindAllStringSubmatch(line, -1)
	var calls []string
	for _, match := range matches {
		if len(match) > 1 {
			calls = append(calls, match[1])
		}
	}
	return calls
}

func extractVariableUses(line string) []string {
    // Ignorar líneas de definición de función
    if strings.HasPrefix(strings.TrimSpace(line), "def ") {
        return []string{}
    }

    // Eliminar strings de la línea
    reString := regexp.MustCompile(`"([^"]*)"`)
    line = reString.ReplaceAllString(line, "")

    // Buscar identificadores
    re := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`)
    matches := re.FindAllString(line, -1)
    var uses []string
    for i, match := range matches {
        // Ignorar palabras reservadas
        if reservedWords[match] {
            continue
        }
        // Ignorar si es asignación
        if strings.Contains(line, match+"=") {
            continue
        }
        // Ignorar si es llamada a función (seguido de '(')
        if i+1 < len(matches) && strings.HasPrefix(line, match+"(") {
            continue
        }
        // Ignorar si justo después viene un paréntesis
        idx := strings.Index(line, match)
        if idx != -1 && idx+len(match) < len(line) && line[idx+len(match)] == '(' {
            continue
        }
        uses = append(uses, match)
    }
    return uses
}

func extractFactorialCalls(code string) []string {
	re := regexp.MustCompile(`factorial\s*\(\s*([^)]+)\s*\)`)
	matches := re.FindAllStringSubmatch(code, -1)
	var args []string
	for _, match := range matches {
		if len(match) > 1 {
			args = append(args, strings.TrimSpace(match[1]))
		}
	}
	return args
}

func isBuiltinFunction(funcName string) bool {
	builtins := map[string]bool{
		"print": true, "len": true, "range": true, "int": true, "str": true, "float": true,
	}
	return builtins[funcName]
}

func isFunctionParameter(varName string, code string) bool {
	// Buscar parámetros en definiciones de funciones
	re := regexp.MustCompile(`def\s+[a-zA-Z_][a-zA-Z0-9_]*\s*\(\s*([^)]*)\s*\)`)
	matches := re.FindAllStringSubmatch(code, -1)
	for _, match := range matches {
		if len(match) > 1 {
			params := strings.Split(match[1], ",")
			for _, param := range params {
				param = strings.TrimSpace(param)
				if param == varName {
					return true
				}
			}
		}
	}
	return false
}

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}