	// analizador_java_completo.go
	package main

	import (
		"encoding/json"
		"fmt"
		"log"
		"net/http"
		"regexp"
		"strings"
	)

		// Palabras reservadas de Java
		var keywords = map[string]struct{}{
			"abstract": {}, "assert": {}, "boolean": {}, "break": {}, "byte": {},
			"case": {}, "catch": {}, "char": {}, "class": {}, "const": {},
			"continue": {}, "default": {}, "do": {}, "double": {}, "else": {},
			"enum": {}, "extends": {}, "final": {}, "finally": {}, "float": {},
			"for": {}, "goto": {}, "if": {}, "implements": {}, "import": {},
			"instanceof": {}, "int": {}, "interface": {}, "long": {}, "native": {},
			"new": {}, "package": {}, "private": {}, "protected": {}, "public": {},
			"return": {}, "short": {}, "static": {}, "strictfp": {}, "super": {},
			"switch": {}, "synchronized": {}, "this": {}, "throw": {}, "throws": {},
			"transient": {}, "try": {}, "void": {}, "volatile": {}, "while": {},
			"true": {}, "false": {}, "null": {},
		}

		var systemWords = map[string]bool{
			"System": true, "out": true, "println": true, "print": true, "printf": true,
			"String": true, "main": true, "args": true, "length": true, "Object": true,
			"Integer": true, "Double": true, "Boolean": true, "Character": true,
			"Scanner": true, "Math": true, "Arrays": true, "List": true,
			"ArrayList": true, "HashMap": true, "Set": true, "Map": true,
			"Exception": true, "RuntimeException": true, "IOException": true,
			"Collections": true, "Comparator": true, "Iterator": true,
		}

		// Métodos comunes del sistema
		var systemMethods = map[string][]string{
			"System": {"out", "in", "err", "exit", "currentTimeMillis", "gc"},
			"out":    {"println", "print", "printf", "flush"},
			"String": {"length", "charAt", "substring", "indexOf", "toLowerCase", "toUpperCase", "trim", "split", "equals", "compareTo"},
			"Math":   {"abs", "max", "min", "sqrt", "pow", "random", "floor", "ceil", "round"},
			"Arrays": {"sort", "toString", "copyOf", "fill", "binarySearch"},
		}

		// Token / Error / Result
		type Token struct {
			Type  string `json:"type"`
			Value string `json:"value"`
			Line  int    `json:"line"`
		}

		type Error struct {
			Line    int    `json:"line"`
			Message string `json:"message"`
			Type    string `json:"type"` // "lexical", "syntactic", "semantic"
		}

		type AnalysisStats struct {
			TotalTokens     int `json:"total_tokens"`
			Keywords        int `json:"keywords"`
			Identifiers     int `json:"identifiers"`
			Numbers         int `json:"numbers"`
			Strings         int `json:"strings"`
			Symbols         int `json:"symbols"`
			Comments        int `json:"comments"`
		}

		type Result struct {
			Tokens      []Token       `json:"tokens"`
			LexErrors   []Error       `json:"lex_errors"`
			SynErrors   []Error       `json:"syn_errors"`
			SemErrors   []Error       `json:"sem_errors"`
			Stats       AnalysisStats `json:"stats"`
			IsLexValid  bool          `json:"is_lex_valid"`
			IsSynValid  bool          `json:"is_syn_valid"`
			IsSemValid  bool          `json:"is_sem_valid"`
		}

		// Lexer patterns para Java (orden importa)
		var patterns = []struct {
			typ string
			reg *regexp.Regexp
		}{
			{"whitespace", regexp.MustCompile(`^\s+`)},
			{"comment_line", regexp.MustCompile(`^//.*`)},
			{"comment_block", regexp.MustCompile(`^/\*[\s\S]*?\*/`)},
			{"string", regexp.MustCompile(`^"([^"\\]|\\.)*"`)},
			{"char", regexp.MustCompile(`^'([^'\\]|\\.)'`)},
			{"number", regexp.MustCompile(`^\d+(\.\d+)?([eE][+-]?\d+)?[fFdDlL]?`)},
			{"increment", regexp.MustCompile(`^\+\+`)},
			{"decrement", regexp.MustCompile(`^--`)},
			{"assign_op", regexp.MustCompile(`^(\+=|-=|\*=|/=|%=|&=|\|=|\^=|<<=|>>=|>>>=)`)},
			{"compare_op", regexp.MustCompile(`^(==|!=|<=|>=|<<|>>|>>>)`)},
			{"logical_op", regexp.MustCompile(`^(&&|\|\|)`)},
			{"arrow", regexp.MustCompile(`^->`)},
			{"scope", regexp.MustCompile(`^::`)},
			{"identifier", regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*`)},
			{"symbol", regexp.MustCompile(`^[\(\)\{\}\[\];,=<>+\-*/%.!&|^~?:@]`)},
		}

		// Función para calcular distancia de Levenshtein
		func levenshteinDistance(s1, s2 string) int {
			if len(s1) == 0 {
				return len(s2)
			}
			if len(s2) == 0 {
				return len(s1)
			}

			matrix := make([][]int, len(s1)+1)
			for i := range matrix {
				matrix[i] = make([]int, len(s2)+1)
				matrix[i][0] = i
			}
			for j := range matrix[0] {
				matrix[0][j] = j
			}

			for i := 1; i <= len(s1); i++ {
				for j := 1; j <= len(s2); j++ {
					cost := 0
					if s1[i-1] != s2[j-1] {
						cost = 1
					}
					matrix[i][j] = min3(
						matrix[i-1][j]+1,     // deletion
						matrix[i][j-1]+1,     // insertion
						matrix[i-1][j-1]+cost, // substitution
					)
				}
			}
			return matrix[len(s1)][len(s2)]
		}

		func min3(a, b, c int) int {
			if a < b {
				if a < c {
					return a
				}
				return c
			}
			if b < c {
				return b
			}
			return c
		}
		// Sugerir corrección para identificadores mal escritos
		func suggestCorrection(word string, candidates map[string]bool) string {
			minDist := 3
			suggestion := ""
			
			for candidate := range candidates {
				if dist := levenshteinDistance(word, candidate); dist <= minDist && dist < len(word)/2 {
					minDist = dist
					suggestion = candidate
				}
			}
			
		
			
			return suggestion
		}
		// Lexer para Java con detección de errores mejorada
		func lex(input string) ([]Token, []Error, AnalysisStats) {
			var toks []Token
			var errs []Error
			stats := AnalysisStats{}
			lines := strings.Split(input, "\n")

			for row, line := range lines {
				rest := line
				for len(rest) > 0 {
					matched := false
					for _, p := range patterns {
						if m := p.reg.FindString(rest); m != "" {
							matched = true
							if p.typ != "whitespace" {
								typ := p.typ
								if typ == "identifier" {
									if _, ok := keywords[m]; ok {
										typ = "keyword"
										stats.Keywords++
									} else {
										stats.Identifiers++
										
										// Verificar si es un identificador mal escrito
										if !systemWords[m] && len(m) > 2 {
											if suggestion := suggestCorrection(m, systemWords); suggestion != "" {
												errs = append(errs, Error{
													row + 1, 
													fmt.Sprintf("'%s' no reconocido. ¿Quisiste decir '%s'?", m, suggestion),
													"lexical",
												})
											}
										}
									}
								} else if typ == "number" {
									stats.Numbers++
								} else if typ == "string" || typ == "char" {
									stats.Strings++
									
									// Verificar strings mal formados
									if typ == "string" && !strings.HasSuffix(m, "\"") {
										errs = append(errs, Error{row + 1, "String sin cerrar", "lexical"})
									}
									if typ == "char" && !strings.HasSuffix(m, "'") {
										errs = append(errs, Error{row + 1, "Caracter sin cerrar", "lexical"})
									}
								} else if typ == "comment_line" || typ == "comment_block" {
									stats.Comments++
								} else {
									stats.Symbols++
								}
								toks = append(toks, Token{typ, m, row + 1})
								stats.TotalTokens++
							}
							rest = rest[len(m):]
							break
						}
					}
					if !matched {
						char := rest[0]
						errs = append(errs, Error{
							row + 1, 
							fmt.Sprintf("Carácter inválido '%c' (código ASCII: %d)", char, char),
							"lexical",
						})
						rest = rest[1:]
					}
				}
			}
			return toks, errs, stats
		}
		// Parser mejorado para Java
		type parser struct {
			toks []Token
			pos  int
			errs []Error
		}

		func (p *parser) cur() *Token {
			if p.pos >= len(p.toks) {
				return nil
			}
			return &p.toks[p.pos]
		}

		func (p *parser) peek(offset int) *Token {
			pos := p.pos + offset
			if pos >= len(p.toks) {
				return nil
			}
			return &p.toks[pos]
		}

		func (p *parser) consume() {
			if p.pos < len(p.toks) {
				p.pos++
			}
		}
		func (p *parser) consumeIf(typ, val string) bool {
			t := p.cur()
			if t != nil && t.Type == typ && (val == "" || t.Value == val) {
				p.consume()
				return true
			}
			return false
		}
		func (p *parser) match(typ, val string) bool {
			t := p.cur()
			return t != nil && t.Type == typ && (val == "" || t.Value == val)
		}
		func (p *parser) expect(typ, val string, errMsg string) bool {
			if p.consumeIf(typ, val) {
				return true
			}
			
			line := 1
			if t := p.cur(); t != nil {
				line = t.Line
			}
			
			p.errs = append(p.errs, Error{line, errMsg, "syntactic"})
			return false
		} 
		// Verificar estructura básica de Java
	func (p *parser) hasValidStructure() bool {
		hasClass := false
		for _, tok := range p.toks {
			if tok.Type == "keyword" && tok.Value == "class" {
				hasClass = true
				break
			}
		}
		return hasClass
	}

	// Skipear tokens hasta encontrar uno específico
	func (p *parser) skipTo(typ, val string) bool {
		for p.cur() != nil {
			if p.match(typ, val) {
				return true
			}
			p.consume()
		}
		return false
	}

	// Skipear bloque balanceado con verificación de balance
	func (p *parser) skipBalanced(open, close string) bool {
		if !p.consumeIf("symbol", open) {
			return false
		}
		
		count := 1
		line := 1
		if p.cur() != nil {
			line = p.cur().Line
		}
		
		for p.cur() != nil && count > 0 {
			if p.match("symbol", open) {
				count++
			} else if p.match("symbol", close) {
				count--
			}
			p.consume()
		}
		
		if count > 0 {
			p.errs = append(p.errs, Error{line, fmt.Sprintf("'%s' sin cerrar", open), "syntactic"})
			return false
		}
		
		return true
	}

	// Parse package declaration
	func (p *parser) parsePackage() {
		if p.consumeIf("keyword", "package") {
			if !p.match("identifier", "") {
				p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre de paquete después de 'package'", "syntactic"})
				return
			}
			
			// Validar estructura del paquete (puede tener puntos)
			p.consume() // consume identifier
			for p.consumeIf("symbol", ".") {
				if !p.match("identifier", "") {
					p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre después de '.' en el paquete", "syntactic"})
					return
				}
				p.consume()
			}
			
			if !p.consumeIf("symbol", ";") {
				p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba ';' después de la declaración del paquete", "syntactic"})
			}
		}
	}

	// Parse import statements
	func (p *parser) parseImports() {
		for p.match("keyword", "import") {
			p.consume()
			if p.consumeIf("keyword", "static") {}
			
			if !p.match("identifier", "") {
				p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre de clase después de 'import'", "syntactic"})
				continue
			}
			
			// Validar estructura del import
			p.consume() // consume identifier
			for p.consumeIf("symbol", ".") {
				if p.consumeIf("symbol", "*") {
					break // import con wildcard
				} else if !p.match("identifier", "") {
					p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre después de '.' en el import", "syntactic"})
					break
				} else {
					p.consume()
				}
			}
			
			if !p.consumeIf("symbol", ";") {
				p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba ';' después de la declaración de import", "syntactic"})
			}
		}
	}

	// Parse annotations
	func (p *parser) parseAnnotations() {
		for p.match("symbol", "@") {
			p.consume()
			if p.match("identifier", "") {
				p.consume()
			} else {
				p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre de anotación después de '@'", "syntactic"})
			}
			if p.match("symbol", "(") {
				if !p.skipBalanced("(", ")") {
					p.errs = append(p.errs, Error{p.cur().Line, "Paréntesis desbalanceados en anotación", "syntactic"})
				}
			}
		}
	}

	// Parse modifiers
	func (p *parser) parseModifiers() {
		modifiers := []string{"public", "private", "protected", "static", "final", 
			"abstract", "synchronized", "native", "strictfp", "transient", "volatile"}
		
		for p.cur() != nil {
			found := false
			for _, mod := range modifiers {
				if p.consumeIf("keyword", mod) {
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
	}

	// Parse type
	func (p *parser) parseType() bool {
		if p.match("keyword", "") || p.match("identifier", "") {
			p.consume()
			
			// Generics
			if p.match("symbol", "<") {
				if !p.skipBalanced("<", ">") {
					p.errs = append(p.errs, Error{p.cur().Line, "Símbolos de generics desbalanceados", "syntactic"})
					return false
				}
			}
			
			// Arrays
			for p.match("symbol", "[") {
				if !p.skipBalanced("[", "]") {
					p.errs = append(p.errs, Error{p.cur().Line, "Corchetes desbalanceados en declaración de array", "syntactic"})
					return false
				}
			}
			return true
		}
		return false
	}

	// Parse parameters
	func (p *parser) parseParameters() {
		if p.consumeIf("symbol", "(") {
			// Parámetros vacíos
			if p.match("symbol", ")") {
				p.consume()
				return
			}
			
			for p.cur() != nil && !p.match("symbol", ")") {
				p.parseAnnotations()
				if p.consumeIf("keyword", "final") {}
				
				if !p.parseType() {
					p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba tipo de parámetro", "syntactic"})
					break
				}
				
				// Varargs
				if p.consumeIf("symbol", ".") {
					if !p.consumeIf("symbol", ".") || !p.consumeIf("symbol", ".") {
						p.errs = append(p.errs, Error{p.cur().Line, "Sintaxis incorrecta para varargs, se esperaba '...'", "syntactic"})
					}
				}
				
				if !p.match("identifier", "") {
					p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre de parámetro", "syntactic"})
					break
				} else {
					p.consume()
				}
				
				if p.match("symbol", ",") {
					p.consume()
					if p.match("symbol", ")") {
						p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba parámetro después de ','", "syntactic"})
						break
					}
				} else if !p.match("symbol", ")") {
					p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba ',' o ')' en lista de parámetros", "syntactic"})
					break
				}
			}
			
			if !p.consumeIf("symbol", ")") {
				p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba ')' para cerrar parámetros", "syntactic"})
			}
		}
	}

	// Parse method body
	func (p *parser) parseBlock() {
		if p.match("symbol", "{") {
			if !p.skipBalanced("{", "}") {
				p.errs = append(p.errs, Error{p.cur().Line, "Llaves desbalanceadas en bloque", "syntactic"})
			}
		} else if p.match("symbol", ";") {
			p.consume() // método abstracto
		} else {
			p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba '{' para el cuerpo del método o ';' para método abstracto", "syntactic"})
		}
	}

	// Verificar si es método main válido
	func (p *parser) isMainMethod(methodName string) bool {
		return methodName == "main"
	}

	// Parse field or method
	func (p *parser) parseClassMember() {
		p.parseAnnotations()
		
		
		hasPublic := false
		hasStatic := false
		
		// Capturar modificadores específicos
		modifiers := []string{"public", "private", "protected", "static", "final", 
			"abstract", "synchronized", "native", "strictfp", "transient", "volatile"}
		
		for p.cur() != nil {
			found := false
			for _, mod := range modifiers {
				if p.consumeIf("keyword", mod) {
					if mod == "public" {
						hasPublic = true
					}
					if mod == "static" {
						hasStatic = true
					}
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
		
		// Constructor
		if p.match("identifier", "") {
			next := p.peek(1)
			if next != nil && next.Value == "(" {
				constructorName := p.cur().Value
				fmt.Println("Constructor encontrado:", constructorName)

				p.consume() // constructor name
				p.parseParameters()
				
				if p.consumeIf("keyword", "throws") {
					if !p.match("identifier", "") {
						p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre de excepción después de 'throws'", "syntactic"})
					}
					for p.cur() != nil && !p.match("symbol", "{") {
						if p.match("identifier", "") {
							p.consume()
							if p.match("symbol", ",") {
								p.consume()
							}
						} else {
							break
						}
					}
				}
				p.parseBlock()
				return
			}
		}
		
		// Reset position para parsear método o campo
	
		p.parseModifiers()
		
		// Method or field
		if !p.parseType() {
			p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba tipo", "syntactic"})
			return
		}
		
		if p.match("identifier", "") {
			methodName := p.cur().Value
			p.consume()
			
			if p.match("symbol", "(") {
				// Method
				p.parseParameters()
				
				// Verificar método main específicamente
				if p.isMainMethod(methodName) {
					if !hasPublic {
						p.errs = append(p.errs, Error{p.cur().Line, "El método main debe ser público", "syntactic"})
					}
					if !hasStatic {
						p.errs = append(p.errs, Error{p.cur().Line, "El método main debe ser estático", "syntactic"})
					}
				}
				
				if p.consumeIf("keyword", "throws") {
					if !p.match("identifier", "") {
						p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre de excepción después de 'throws'", "syntactic"})
					}
					for p.cur() != nil && !p.match("symbol", "{") && !p.match("symbol", ";") {
						if p.match("identifier", "") {
							p.consume()
							if p.match("symbol", ",") {
								p.consume()
							}
						} else {
							break
						}
					}
				}
				p.parseBlock()
			} else {
				// Field
				// Verificar inicialización
				if p.match("symbol", "=") {
					p.consume()
					p.parseExpression()
				}
				
				if !p.consumeIf("symbol", ";") {
					p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba ';' después de la declaración del campo", "syntactic"})
				}
			}
		} else {
			p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre de método o campo", "syntactic"})
		}
	}

	// Parse expression (simplified but more robust)
	func (p *parser) parseExpression() {
		parenCount := 0
		bracketCount := 0
		braceCount := 0
		
		for p.cur() != nil && !p.match("symbol", ";") && !p.match("symbol", ")") && 
			!p.match("symbol", "}") && !p.match("symbol", ",") {
			
			if p.match("symbol", "(") {
				parenCount++
				p.consume()
			} else if p.match("symbol", ")") {
				parenCount--
				if parenCount < 0 {
					break // Salir si hay paréntesis desbalanceados
				}
				p.consume()
			} else if p.match("symbol", "[") {
				bracketCount++
				p.consume()
			} else if p.match("symbol", "]") {
				bracketCount--
				if bracketCount < 0 {
					break
				}
				p.consume()
			} else if p.match("symbol", "{") {
				braceCount++
				p.consume()
			} else if p.match("symbol", "}") {
				braceCount--
				if braceCount < 0 {
					break
				}
				p.consume()
			} else {
				p.consume()
			}
		}
		
		// Verificar balance
		if parenCount > 0 {
			p.errs = append(p.errs, Error{p.cur().Line, "Paréntesis sin cerrar en expresión", "syntactic"})
		}
		if bracketCount > 0 {
			p.errs = append(p.errs, Error{p.cur().Line, "Corchetes sin cerrar en expresión", "syntactic"})
		}
		if braceCount > 0 {
			p.errs = append(p.errs, Error{p.cur().Line, "Llaves sin cerrar en expresión", "syntactic"})
		}
	}

	// Parse class
	func (p *parser) parseClass() {
		p.parseAnnotations()
		p.parseModifiers()
		
		classType := ""
		if p.consumeIf("keyword", "class") {
			classType = "class"
		} else if p.consumeIf("keyword", "interface") {
			classType = "interface"
		} else if p.consumeIf("keyword", "enum") {
			classType = "enum"
		}
		
		if classType != "" {
			if !p.match("identifier", "") {
				p.errs = append(p.errs, Error{p.cur().Line, fmt.Sprintf("Se esperaba nombre de %s", classType), "syntactic"})
				return
			} else {
				className := p.cur().Value
				p.consume()
				
				// Verificar que el nombre de la clase empiece con mayúscula
				if len(className) > 0 && className[0] >= 'a' && className[0] <= 'z' {
					p.errs = append(p.errs, Error{p.cur().Line, fmt.Sprintf("El nombre de la clase '%s' debe comenzar con mayúscula", className), "syntactic"})
				}
			}
			
			// Generics
			if p.match("symbol", "<") {
				if !p.skipBalanced("<", ">") {
					p.errs = append(p.errs, Error{p.cur().Line, "Símbolos de generics desbalanceados en clase", "syntactic"})
				}
			}
			
			// Extends
			if p.consumeIf("keyword", "extends") {
				if !p.parseType() {
					p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre de clase padre después de 'extends'", "syntactic"})
				}
			}
			
			// Implements
			if p.consumeIf("keyword", "implements") {
				if !p.parseType() {
					p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre de interfaz después de 'implements'", "syntactic"})
				}
				for p.consumeIf("symbol", ",") {
					if !p.parseType() {
						p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba nombre de interfaz después de ','", "syntactic"})
						break
					}
				}
			}
			
			if !p.match("symbol", "{") {
				p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba '{' para abrir el cuerpo de la clase", "syntactic"})
			} else {
				p.consume()
				
				// Verificar que el cuerpo no esté vacío para clases principales
				isEmpty := true
			
				
				for p.cur() != nil && !p.match("symbol", "}") {
					isEmpty = false
					if p.match("keyword", "class") || p.match("keyword", "interface") || p.match("keyword", "enum") {
						p.parseClass() // Clase anidada
					} else {
						p.parseClassMember()
					}
				}
				
				if isEmpty && classType == "class" {
					p.errs = append(p.errs, Error{p.cur().Line, "La clase no puede estar vacía", "syntactic"})
				}
				
				if !p.consumeIf("symbol", "}") {
					p.errs = append(p.errs, Error{p.cur().Line, "Se esperaba '}' para cerrar el cuerpo de la clase", "syntactic"})
				}
			}
		}
	}

	// Parser principal - robusto y permisivo
	func (p *parser) parseAll() []Error {
		if !p.hasValidStructure() {
			p.errs = append(p.errs, Error{1, "No se encontró declaración de clase", "syntactic"})
			return p.errs
		}
		
		p.parsePackage()
		p.parseImports()
		
		hasMainClass := false
		
		for p.cur() != nil {
			if p.match("keyword", "class") || p.match("keyword", "interface") || 
				p.match("keyword", "enum") || p.match("symbol", "@") || 
				p.match("keyword", "public") || p.match("keyword", "private") || 
				p.match("keyword", "protected") || p.match("keyword", "final") || 
				p.match("keyword", "abstract") {
				
				// Verificar si es una clase pública (posible clase principal)
				if p.match("keyword", "public") {
					hasMainClass = true
					_ = hasMainClass
				}
				
				p.parseClass()
			} else {
				// Token inesperado
				p.errs = append(p.errs, Error{p.cur().Line, fmt.Sprintf("Token inesperado: '%s'", p.cur().Value), "syntactic"})
				p.consume()
			}
		}
		
		return p.errs
	}

	// HTTP handler
	func handler(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var in struct{ Code string }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), 400)
			log.Println("❌ JSON", err)
			return
		}

		log.Println("📥 Entrada Java:\n", in.Code)

		// Análisis
		toks, lexErr, stats := lex(in.Code)
		p := &parser{toks, 0, nil}
		synErr := p.parseAll()
		semErr := semantic(toks)

		res := Result{
			Tokens:     toks,
			LexErrors:  lexErr,
			SynErrors:  synErr,
			SemErrors:  semErr,
			Stats:      stats,
			IsLexValid: len(lexErr) == 0,
			IsSynValid: len(lexErr) == 0 && len(synErr) == 0,
			IsSemValid: len(lexErr) == 0 && len(synErr) == 0 && len(semErr) == 0,
		}

		log.Printf("📤 Resultado Java: tokens=%d lex=%d syn=%d sem=%d\n",
			stats.TotalTokens, len(lexErr), len(synErr), len(semErr))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}

 func semantic(tokens []Token) []Error {
	declared := make(map[string]bool)
	used := make(map[string]bool)
	varTypes := make(map[string]string) // Mapa para guardar tipos de variables
	var semErrs []Error
	
	// Variables del sistema siempre disponibles
	for word := range systemWords {
		declared[word] = true
		used[word] = true // Marcamos como usadas para evitar errores
	}
	
	// Tipos primitivos y comunes
	primitives := []string{"int", "double", "float", "long", "short", "byte", 
		"boolean", "char", "String", "Object", "void"}
	for _, p := range primitives {
		declared[p] = true
		used[p] = true
	}
	
	// Tipos válidos para bucles for
	validForTypes := map[string]bool{
		"int": true, "long": true, "short": true, "byte": true,
		"double": true, "float": true, "char": true,
	}
	
	// Primera pasada: identificar declaraciones
	for i, t := range tokens {
		// Verificar secuencia System.out.println específicamente
		if t.Type == "identifier" && t.Value == "System" {
			if i+4 < len(tokens) {
				if tokens[i+1].Value == "." && tokens[i+2].Value == "out" && 
				tokens[i+3].Value == "." && tokens[i+4].Value == "println" {
					continue
				}
			}
			
			// Verificar errores comunes en System.out.println
			if i+1 < len(tokens) && tokens[i+1].Value == "." {
				if i+2 < len(tokens) {
					second := tokens[i+2].Value
					if second != "out" {
						if suggestion := suggestCorrection(second, map[string]bool{"out": true}); suggestion != "" {
							semErrs = append(semErrs, Error{
								t.Line, 
								fmt.Sprintf("Error en System.%s - ¿Quisiste decir 'System.out'?", second),
								"semantic",
							})
						}
					} else if i+3 < len(tokens) && tokens[i+3].Value == "." {
						if i+4 < len(tokens) {
							method := tokens[i+4].Value
							if method != "println" && method != "print" && method != "printf" {
								if suggestion := suggestCorrection(method, map[string]bool{"println": true, "print": true, "printf": true}); suggestion != "" {
									semErrs = append(semErrs, Error{
										tokens[i+4].Line, 
										fmt.Sprintf("Error en System.out.%s - ¿Quisiste decir 'System.out.%s'?", method, suggestion),
										"semantic",
									})
								}
							}
						}
					}
				}
			}
		}
		
		// Declaración de variables y arreglos
		if i > 0 && tokens[i-1].Type == "keyword" {
			prevVal := tokens[i-1].Value
			if prevVal == "int" || prevVal == "String" || prevVal == "double" || 
				prevVal == "boolean" || prevVal == "float" || prevVal == "long" || 
				prevVal == "short" || prevVal == "byte" || prevVal == "char" {
				if t.Type == "identifier" {
					declared[t.Value] = true
					varTypes[t.Value] = prevVal // Guardar el tipo de la variable
				}
			}
		}
		
		// Declaración de arreglos (tipo[] variable o tipo variable[])
		if i > 0 && t.Type == "identifier" {
			prevVal := tokens[i-1].Value
			// Caso: tipo[] variable
			if i >= 2 && tokens[i-2].Type == "keyword" && prevVal == "[]" {
				declared[t.Value] = true
				varTypes[t.Value] = tokens[i-2].Value + "[]" // Tipo de arreglo
			}
			// Caso: tipo variable[]
			if i+1 < len(tokens) && tokens[i+1].Value == "[]" {
				if i > 0 && tokens[i-1].Type == "keyword" {
					declared[t.Value] = true
					varTypes[t.Value] = tokens[i-1].Value + "[]" // Tipo de arreglo
				}
			}
		}
		
		// Validación de estructura for
		if t.Type == "keyword" && t.Value == "for" {
			semErrs = append(semErrs, validateForLoop(tokens, i, declared, varTypes, validForTypes)...)
		}
		
		// Declaración de clases y métodos
		if i > 0 && tokens[i-1].Value == "class" && t.Type == "identifier" {
			declared[t.Value] = true
			used[t.Value] = true // Las clases se consideran usadas por defecto
		}
		
		if i > 1 && (tokens[i-1].Value == "void" || tokens[i-1].Value == "int" || 
			tokens[i-1].Value == "String") && t.Type == "identifier" {
			if tokens[i-2].Value == "static" || tokens[i-2].Value == "public" || 
				tokens[i-2].Value == "private" {
				declared[t.Value] = true
				used[t.Value] = true // Los métodos se consideran usados por defecto
			}
		}
		
		// Parámetros de métodos
		if i > 0 && t.Type == "identifier" {
			if i+1 < len(tokens) {
				next := tokens[i+1]
				if next.Value == ")" || next.Value == "," {
					declared[t.Value] = true
				}
			}
		}
	}
	
	// Segunda pasada: identificar uso de variables
	for i, t := range tokens {
		if t.Type == "identifier" && declared[t.Value] {
			// Verificar que no sea parte de una declaración
			isDeclaring := false
			
			// Verificar si es una declaración de variable
			if i > 0 && tokens[i-1].Type == "keyword" {
				prevVal := tokens[i-1].Value
				if prevVal == "int" || prevVal == "String" || prevVal == "double" || 
					prevVal == "boolean" || prevVal == "float" || prevVal == "long" || 
					prevVal == "short" || prevVal == "byte" || prevVal == "char" {
					isDeclaring = true
				}
			}
			
			// Verificar si es declaración de arreglo
			if i > 0 && tokens[i-1].Value == "[]" {
				if i >= 2 && tokens[i-2].Type == "keyword" {
					isDeclaring = true
				}
			}
			if i+1 < len(tokens) && tokens[i+1].Value == "[]" {
				if i > 0 && tokens[i-1].Type == "keyword" {
					isDeclaring = true
				}
			}
			
			// Verificar si es declaración de clase o método
			if i > 0 && (tokens[i-1].Value == "class" || 
				(i > 1 && (tokens[i-2].Value == "static" || tokens[i-2].Value == "public" || 
				tokens[i-2].Value == "private"))) {
				isDeclaring = true
			}
			
			if !isDeclaring {
				used[t.Value] = true
			}
		}
	}
	
	// Tercera pasada: verificar variables no declaradas y generar errores
	for i, t := range tokens {
		if t.Type == "identifier" && !declared[t.Value] {
			// Ignorar si es parte de System.out.println u otros casos especiales
			if i > 0 && tokens[i-1].Value == "." {
				continue
			}
			if i+1 < len(tokens) && tokens[i+1].Value == "." {
				continue
			}
			if i+1 < len(tokens) && tokens[i+1].Value == "(" {
				continue // Llamada a método
			}
			
			// Ignorar keywords después de ciertos tokens
			prev := ""
			if i > 0 {
				prev = tokens[i-1].Value
			}
			if prev == "class" || prev == "extends" || prev == "implements" || 
				prev == "new" || prev == "instanceof" {
				continue
			}
			
			// Sugerir corrección para variables no declaradas
			if suggestion := suggestCorrection(t.Value, declared); suggestion != "" {
				semErrs = append(semErrs, Error{
					t.Line, 
					fmt.Sprintf("Variable '%s' no está declarada. ¿Quisiste decir '%s'?", t.Value, suggestion),
					"semantic",
				})
			} else {
				semErrs = append(semErrs, Error{
					t.Line, 
					fmt.Sprintf("Variable '%s' no está declarada", t.Value),
					"semantic",
				})
			}
		}
	}
	
	// Verificar variables declaradas pero no utilizadas
	for varName := range declared {
		if !used[varName] {
			// Buscar la línea donde se declaró la variable
			line := 1
			for i, t := range tokens {
				if t.Type == "identifier" && t.Value == varName {
					// Verificar si es una declaración
					if i > 0 && tokens[i-1].Type == "keyword" {
						prevVal := tokens[i-1].Value
						if prevVal == "int" || prevVal == "String" || prevVal == "double" || 
							prevVal == "boolean" || prevVal == "float" || prevVal == "long" || 
							prevVal == "short" || prevVal == "byte" || prevVal == "char" {
							line = t.Line
							break
						}
					}
					// Verificar declaración de arreglo
					if (i > 0 && tokens[i-1].Value == "[]") || 
						(i+1 < len(tokens) && tokens[i+1].Value == "[]") {
						if (i >= 2 && tokens[i-2].Type == "keyword") || 
							(i > 0 && tokens[i-1].Type == "keyword") {
							line = t.Line
							break
						}
					}
				}
			}
			
			semErrs = append(semErrs, Error{
				line,
				fmt.Sprintf("Variable '%s' está declarada pero no se utiliza", varName),
				"semantic",
			})
		}
	}
	
	return semErrs
}

// Función para validar la estructura del bucle for
func validateForLoop(tokens []Token, forIndex int, declared map[string]bool, varTypes map[string]string, validForTypes map[string]bool) []Error {
	var errors []Error
	
	// Buscar el paréntesis de apertura después de 'for'
	openParenIndex := -1
	for i := forIndex + 1; i < len(tokens) && i < forIndex + 3; i++ {
		if tokens[i].Value == "(" {
			openParenIndex = i
			break
		}
	}
	
	if openParenIndex == -1 {
		errors = append(errors, Error{
			tokens[forIndex].Line,
			"Estructura for inválida: falta paréntesis de apertura '('",
			"semantic",
		})
		return errors
	}
	
	// Buscar el paréntesis de cierre
	closeParenIndex := -1
	parenthesesCount := 0
	for i := openParenIndex; i < len(tokens); i++ {
		if tokens[i].Value == "(" {
			parenthesesCount++
		} else if tokens[i].Value == ")" {
			parenthesesCount--
			if parenthesesCount == 0 {
				closeParenIndex = i
				break
			}
		}
	}
	
	if closeParenIndex == -1 {
		errors = append(errors, Error{
			tokens[forIndex].Line,
			"Estructura for inválida: falta paréntesis de cierre ')'",
			"semantic",
		})
		return errors
	}
	
	// Extraer tokens dentro del for
	forContent := tokens[openParenIndex+1 : closeParenIndex]
	
	// Dividir por punto y coma para obtener las tres partes del for
	parts := [][]Token{}
	currentPart := []Token{}
	
	for _, token := range forContent {
		if token.Value == ";" {
			parts = append(parts, currentPart)
			currentPart = []Token{}
		} else {
			currentPart = append(currentPart, token)
		}
	}
	parts = append(parts, currentPart) // Agregar la última parte
	
	// Validar que tenga exactamente 3 partes
	if len(parts) != 3 {
		errors = append(errors, Error{
			tokens[forIndex].Line,
			fmt.Sprintf("Estructura for inválida: debe tener exactamente 3 partes separadas por ';', encontradas %d partes", len(parts)),
			"semantic",
		})
		return errors
	}
	
	// Validar la primera parte (inicialización)
	initPart := parts[0]
	if len(initPart) >= 3 {
		// Verificar si es declaración de variable: tipo variable = valor
		if initPart[0].Type == "keyword" {
			varType := initPart[0].Value
			if !validForTypes[varType] {
				errors = append(errors, Error{
					tokens[forIndex].Line,
					fmt.Sprintf("Tipo '%s' no es válido para bucle for. Tipos válidos: int, long, short, byte, double, float, char", varType),
					"semantic",
				})
			}
		} else if initPart[0].Type == "identifier" {
			// Verificar si es asignación a variable existente
			varName := initPart[0].Value
			if declared[varName] {
				if varType, exists := varTypes[varName]; exists {
					if !validForTypes[varType] {
						errors = append(errors, Error{
							tokens[forIndex].Line,
							fmt.Sprintf("Variable '%s' de tipo '%s' no es válida para bucle for. Tipos válidos: int, long, short, byte, double, float, char", varName, varType),
							"semantic",
						})
					}
				}
			}
		}
	}
	
	// Validar la segunda parte (condición)
	conditionPart := parts[1]
	if len(conditionPart) == 0 {
		errors = append(errors, Error{
			tokens[forIndex].Line,
			"Estructura for inválida: la condición no puede estar vacía",
			"semantic",
		})
	} else {
		// Verificar que las variables en la condición sean de tipos válidos
		for _, token := range conditionPart {
			if token.Type == "identifier" && declared[token.Value] {
				if varType, exists := varTypes[token.Value]; exists {
					if !validForTypes[varType] && varType != "boolean" {
						errors = append(errors, Error{
							tokens[forIndex].Line,
							fmt.Sprintf("Variable '%s' de tipo '%s' no es apropiada para condición de bucle for", token.Value, varType),
							"semantic",
						})
					}
				}
			}
		}
	}
	
	// Validar la tercera parte (incremento/decremento)
	incrementPart := parts[2]
	if len(incrementPart) == 0 {
		errors = append(errors, Error{
			tokens[forIndex].Line,
			"Estructura for inválida: el incremento/decremento no puede estar vacío",
			"semantic",
		})
	} else {
		// Verificar que las variables en el incremento sean de tipos válidos
		for _, token := range incrementPart {
			if token.Type == "identifier" && declared[token.Value] {
				if varType, exists := varTypes[token.Value]; exists {
					if !validForTypes[varType] {
						errors = append(errors, Error{
							tokens[forIndex].Line,
							fmt.Sprintf("Variable '%s' de tipo '%s' no es válida para incremento/decremento en bucle for", token.Value, varType),
							"semantic",
						})
					}
				}
			}
		}
	}
	
	return errors
}
	func main() {
		http.HandleFunc("/analyze", handler)
		log.Println("🚀 Analizador Java mejorado corriendo en :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}	