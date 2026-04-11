// Package parser reads Go source files using go/ast and extracts type declarations,
// constants, and structs into a structured intermediate representation (ParseResult).
//
// The parser is the first stage of the gots pipeline:
//
//	Go source files → [parser] → ParseResult → [generator] → TypeScript
//
// It uses Go's standard library go/parser and go/ast packages to build an AST
// (Abstract Syntax Tree) from source files. This means:
//   - No regex hacking — exact type info from the compiler's own parser
//   - Handles all valid Go syntax including iota, multi-line strings, embedded structs
//   - Comments are preserved (used for documentation in future features)
//
// # How AST parsing works
//
// A Go source file is parsed into an *ast.File, which contains a list of
// declarations (ast.Decl). Each declaration is either:
//   - *ast.GenDecl: type, const, var, or import declaration
//   - *ast.FuncDecl: function declaration (ignored by gots)
//
// For GenDecl, we check the token type:
//   - token.TYPE  → type declarations (type MessageRole string, type User struct{...})
//   - token.CONST → constant declarations (const Foo = "bar")
//
// Each is further broken down into specs (ast.TypeSpec, ast.ValueSpec) which
// contain the actual names, types, and values.
//
// # Type resolution for struct fields
//
// Struct fields can have complex types: *string, []User, map[string]int,
// time.Time, etc. The resolveFieldType function recursively walks the AST
// expression to extract the base type name and flags (isPointer, isSlice, isMap).
package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// TypeDef represents a Go named type declaration.
//
// Example: "type MessageRole string" produces:
//
//	TypeDef{Name: "MessageRole", Underlying: "string"}
//
// Only simple type aliases are captured here (ident-based).
// Struct types are captured separately in StructDef.
type TypeDef struct {
	Name       string      // The Go type name (e.g., "MessageRole")
	Underlying string      // The underlying type (e.g., "string", "int")
	FuncSig    *FuncSig    // Function signature, only set when Underlying="Function"
	TypeParams []TypeParam // Generic type parameters (Go 1.18+), empty for non-generic types
	Comment    string      // Doc comment from Go source (without // prefix)
	File       string      // Source file path this was parsed from
}

// FuncSig represents a parsed Go function type signature.
//
// Examples:
//
//	func(a, b int) int           → Params: [{a, int}, {b, int}], Results: [{, int}]
//	func(err error)              → Params: [{err, error}], Results: []
//	func()                       → Params: [], Results: []
//	func(string) (string, error) → Params: [{, string}], Results: [{, string}, {, error}]
type FuncSig struct {
	Params  []FuncParam // Function parameters
	Results []FuncParam // Return types
}

// FuncParam represents a single parameter or return value in a function signature.
type FuncParam struct {
	Name string // Parameter name, may be empty for unnamed params/returns
	Type string // Full Go type string (e.g., "string", "int", "[]byte", "error")
}

// ConstDef represents a single Go constant declaration.
//
// For typed constants (e.g., MessageRoleUser MessageRole = "user"):
//   - TypeName is set to "MessageRole"
//   - This is how we group constants by their enum type
//
// For untyped constants (e.g., DefaultTimeout = 30):
//   - TypeName is empty string ""
//   - These are grouped separately and exported individually
type ConstDef struct {
	Name        string // The constant name (e.g., "MessageRoleUser")
	TypeName    string // The Go type name, empty for untyped constants
	Value       string // The constant value with quotes stripped (e.g., "user", not "\"user\""")
	Comment     string // Inline or doc comment from Go source
	IsMultiLine bool   // True if the original Go literal was a backtick string (contains newlines)
	File        string // Source file path this was parsed from
}

// StructDef represents a Go struct type declaration.
//
// Example:
//
//	type User struct {
//	    ID   int64  `json:"id"`
//	    Name string `json:"name"`
//	}
type StructDef struct {
	Name       string        // The struct name (e.g., "User")
	Fields     []StructField // All fields including embedded
	TypeParams []TypeParam   // Generic type parameters (Go 1.18+), empty for non-generic structs
	Comment    string        // Doc comment from Go source
	File       string        // Source file path
}

// TypeParam represents a generic type parameter on a struct or type.
//
// Example: type Box[T any] → TypeParam{Name: "T", Constraint: "any"}
// Example: type Pair[K comparable, V any] → two TypeParams
// Example: type NumberBox[T int | int64 | float64] → TypeParam{Name: "T", Constraint: "int | int64 | float64"}
type TypeParam struct {
	Name       string // The type parameter name (e.g., "T", "K", "V")
	Constraint string // The constraint (e.g., "any", "comparable", "int | int64 | float64")
}

// StructField represents a single field in a Go struct.
//
// The field's type information is decomposed into the base type name
// plus boolean flags, since TypeScript handles these orthogonally:
//   - *string    → GoType="string", IsPointer=true
//   - []User     → GoType="User",   IsSlice=true
//   - time.Time  → GoType="time.Time"
//   - BaseModel  → IsEmbedded=true (no field name, it's an embedded struct)
type StructField struct {
	Name         string        // Field name (e.g., "ID"), or embedded type name
	GoType       string        // Base type name (e.g., "string", "time.Time", "User")
	FullType     string        // Full Go type expression (e.g., "[][]int", "map[string][]int", "*string")
	Tag          string        // Raw struct tag including backticks (e.g., `json:"id"`)
	IsPointer    bool          // Field type is a pointer (*T)
	IsSlice      bool          // Field type is a slice ([]T)
	IsMap        bool          // Field type is a map (map[K]V)
	MapKeyType   string        // Map key type (e.g., "string", "int") — only set when IsMap=true
	MapValType   string        // Map value type (e.g., "int", "User") — only set when IsMap=true
	IsEmbedded   bool          // Field is an embedded struct (no name, just type)
	IsExported   bool          // Field name starts with uppercase
	IsInlineStruct bool          // Field type is an anonymous inline struct
	InlineFields   []StructField // Fields of the inline struct (only set when IsInlineStruct=true)
	FuncSig        *FuncSig      // Function signature, only set when GoType="Function"
	Comment        string        // Inline or doc comment from Go source
}

// ParseResult holds all declarations extracted from one or more Go source files.
//
// This is the intermediate representation passed from parser to generator.
// Types, Consts, and Structs maintain their order of appearance in the source.
type ParseResult struct {
	Types   []TypeDef   // Named type declarations (type Foo string)
	Consts  []ConstDef  // Constants (typed and untyped)
	Structs []StructDef // Struct type declarations
}

// Parser reads Go source files and builds a ParseResult.
//
// It maintains a token.FileSet for position tracking (used in error messages)
// and accumulates results across multiple ParseFile calls.
type Parser struct {
	fset   *token.FileSet // Tracks file positions for error reporting
	result *ParseResult   // Accumulated parse results
}

// New creates a new Parser ready to parse Go source files.
func New() *Parser {
	return &Parser{
		fset:   token.NewFileSet(),
		result: &ParseResult{},
	}
}

// Result returns the accumulated ParseResult from all parsed files.
// Call this after all ParseFile calls are complete.
func (p *Parser) Result() *ParseResult {
	return p.result
}

// ParseFile reads a single Go source file from disk, parses it into an AST,
// and extracts all type declarations, constants, and structs.
//
// Can be called multiple times to accumulate results from multiple files.
// The filename is stored with each declaration so the generator can filter
// by source file later.
func (p *Parser) ParseFile(filename string) error {
	// parser.ParseComments preserves comments in the AST.
	// We don't use them yet but they'll be needed for:
	// - Documentation in generated TS
	// - Future //gots: comment directives
	f, err := parser.ParseFile(p.fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", filename, err)
	}

	// Walk all top-level declarations in the file.
	// Go files contain: imports, types, consts, vars, and funcs.
	// We only care about types and consts.
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue // Skip function declarations
		}

		switch genDecl.Tok {
		case token.TYPE:
			// type MessageRole string
			// type User struct { ... }
			p.parseTypeDecl(genDecl, filename)
		case token.CONST:
			// const ( MessageRoleUser MessageRole = "user" ... )
			// const DefaultTimeout = 30
			p.parseConstDecl(genDecl, filename)
		}
	}

	return nil
}

// parseTypeDecl extracts type declarations from a GenDecl.
//
// A single GenDecl can contain multiple type specs:
//
//	type (
//	    Foo string
//	    Bar int
//	)
//
// Each spec is either:
//   - *ast.Ident  → named type alias (type Foo string) → stored as TypeDef
//   - *ast.StructType → struct (type User struct{...}) → stored as StructDef
func (p *Parser) parseTypeDecl(decl *ast.GenDecl, filename string) {
	for _, spec := range decl.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		// Extract doc comment. Prefer the spec-level comment (on the type itself),
		// fall back to the GenDecl-level comment (on the type block).
		comment := extractComment(ts.Doc)
		if comment == "" {
			comment = extractComment(decl.Doc)
		}

		// Parse generic type parameters (available on all type specs in Go 1.18+)
		typeParams := parseTypeParams(ts.TypeParams)

		switch t := ts.Type.(type) {
		case *ast.Ident:
			p.result.Types = append(p.result.Types, TypeDef{
				Name:       ts.Name.Name,
				Underlying: t.Name,
				TypeParams: typeParams,
				Comment:    comment,
				File:       filename,
			})
		case *ast.StructType:
			sd := StructDef{
				Name:       ts.Name.Name,
				TypeParams: typeParams,
				Comment:    comment,
				File:       filename,
			}
			for _, field := range t.Fields.List {
				sf := parseStructField(field)
				sd.Fields = append(sd.Fields, sf...)
			}
			p.result.Structs = append(p.result.Structs, sd)
		case *ast.ArrayType:
			p.result.Types = append(p.result.Types, TypeDef{
				Name:       ts.Name.Name,
				Underlying: fullTypeString(ts.Type),
				TypeParams: typeParams,
				Comment:    comment,
				File:       filename,
			})
		case *ast.MapType:
			p.result.Types = append(p.result.Types, TypeDef{
				Name:       ts.Name.Name,
				Underlying: fullTypeString(ts.Type),
				TypeParams: typeParams,
				Comment:    comment,
				File:       filename,
			})
		case *ast.ChanType:
			p.result.Types = append(p.result.Types, TypeDef{
				Name:       ts.Name.Name,
				Underlying: fullTypeString(t.Value),
				TypeParams: typeParams,
				Comment:    comment,
				File:       filename,
			})
		case *ast.FuncType:
			p.result.Types = append(p.result.Types, TypeDef{
				Name:       ts.Name.Name,
				Underlying: "Function",
				FuncSig:    parseFuncType(t),
				TypeParams: typeParams,
				Comment:    comment,
				File:       filename,
			})
		case *ast.InterfaceType:
			p.result.Types = append(p.result.Types, TypeDef{
				Name:       ts.Name.Name,
				Underlying: "interface",
				TypeParams: typeParams,
				Comment:    comment,
				File:       filename,
			})
		case *ast.SelectorExpr:
			if x, ok := t.X.(*ast.Ident); ok {
				p.result.Types = append(p.result.Types, TypeDef{
					Name:       ts.Name.Name,
					Underlying: x.Name + "." + t.Sel.Name,
					TypeParams: typeParams,
					File:       filename,
				})
			}
		case *ast.StarExpr:
			p.result.Types = append(p.result.Types, TypeDef{
				Name:       ts.Name.Name,
				Underlying: fullTypeString(t.X),
				TypeParams: typeParams,
				Comment:    comment,
				File:       filename,
			})
		default:
			_ = t
			p.result.Types = append(p.result.Types, TypeDef{
				Name:       ts.Name.Name,
				Underlying: fullTypeString(ts.Type),
				TypeParams: typeParams,
				Comment:    comment,
				File:       filename,
			})
		}
	}
}

// parseConstDecl extracts constants from a GenDecl.
//
// In Go's AST, a const block is a single GenDecl with multiple ValueSpecs:
//
//	const (                          ← GenDecl
//	    MessageRoleUser MessageRole = "user"     ← ValueSpec with Type
//	    MessageRoleAssistant        = "assistant" ← ValueSpec, Type inherited
//	)
//
// For typed constants, the Type field on the ValueSpec tells us the Go type.
// This is how we automatically group constants by their enum type.
//
// Blank identifiers (_) are skipped — they're used in Go for iota gaps.
//
// Important: In a Go const block, only the first spec may have an explicit type.
// Subsequent specs inherit the type from the previous one:
//
//	const (
//	    _ LockType = iota   ← explicit type "LockType"
//	    LockReadonly         ← Type is nil in AST, but inherits "LockType"
//	    LockUpdate           ← same inheritance
//	)
//
// We track the "current type" across specs in a block to handle this.
func (p *Parser) parseConstDecl(decl *ast.GenDecl, filename string) {
	// currentTypeName tracks the inherited type across a const block.
	// Reset for each new const block (each GenDecl is one const(...) block).
	currentTypeName := ""

	// iotaCounter tracks the current iota value within the const block.
	// iota starts at 0 and increments by 1 for each ValueSpec (not each name).
	// isIotaBlock tracks whether this const block uses iota.
	iotaCounter := 0
	isIotaBlock := false

	for _, spec := range decl.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		// Extract the type name if this constant has an explicit type.
		// In "MessageRoleUser MessageRole = "user"", Type is *ast.Ident("MessageRole").
		// For untyped consts like "Foo = 42", Type is nil.
		//
		// If Type is nil but we're in a typed const block, inherit the previous type.
		// If Type is explicitly set, update the current type for subsequent specs.
		if vs.Type != nil {
			if ident, ok := vs.Type.(*ast.Ident); ok {
				currentTypeName = ident.Name
			}
		}
		typeName := currentTypeName

		// Detect if this spec uses iota. If the first value in the block is "iota"
		// or an expression containing iota, mark the block as iota-based.
		if len(vs.Values) > 0 {
			if containsIota(vs.Values[0]) {
				isIotaBlock = true
			}
		}

		// Extract comment — prefer inline comment, fall back to doc comment
		constComment := extractComment(vs.Comment)
		if constComment == "" {
			constComment = extractComment(vs.Doc)
		}

		// A single ValueSpec can declare multiple names: const A, B = 1, 2
		// (rare, but valid Go)
		for i, name := range vs.Names {
			// Skip blank identifier — used for iota gaps
			if name.Name == "_" {
				continue
			}

			// Extract the value as a string.
			// For string constants, quotes are stripped: "user" → user
			// For iota or complex expressions, value may be empty.
			// isMultiLine is true when the Go source used backtick strings.
			value := ""
			isMultiLine := false
			if i < len(vs.Values) {
				value, isMultiLine = exprToString(vs.Values[i])
			}

			// For iota-based const blocks, resolve the value to the numeric iota counter.
			// "iota" → "0", empty (inherited) → current counter value.
			if isIotaBlock && (value == "iota" || value == "") {
				value = fmt.Sprintf("%d", iotaCounter)
			}

			p.result.Consts = append(p.result.Consts, ConstDef{
				Name:        name.Name,
				TypeName:    typeName,
				Value:       value,
				Comment:     constComment,
				IsMultiLine: isMultiLine,
				File:        filename,
			})
		}

		// iota increments once per ValueSpec (line in const block), not per name.
		iotaCounter++
	}
}

// parseTypeParams extracts generic type parameters from a TypeSpec.
//
// In Go 1.18+, ast.TypeSpec has a TypeParams field (*ast.FieldList).
// Each field in the list is one type parameter group:
//
//	type Box[T any]                  → [{T, any}]
//	type Pair[K comparable, V any]   → [{K, comparable}, {V, any}]
//	type Num[T int | int64 | float64] → [{T, int | int64 | float64}]
//
// Returns nil for non-generic types (TypeParams is nil).
func parseTypeParams(typeParams *ast.FieldList) []TypeParam {
	if typeParams == nil {
		return nil
	}

	var params []TypeParam
	for _, field := range typeParams.List {
		// Extract the constraint as a string representation
		constraint := typeExprToString(field.Type)

		// Each field can have multiple names: [A, B any] → two params with same constraint
		for _, name := range field.Names {
			params = append(params, TypeParam{
				Name:       name.Name,
				Constraint: constraint,
			})
		}
	}
	return params
}

// typeExprToString converts a type constraint AST expression to a string.
//
// Handles:
//   - *ast.Ident: "any", "comparable", "int", etc.
//   - *ast.BinaryExpr: union constraints like "int | int64 | float64"
//   - *ast.UnaryExpr: tilde constraints like "~int"
//   - *ast.InterfaceType: inline interface constraints
func typeExprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.BinaryExpr:
		// Union constraint: int | int64 | float64
		left := typeExprToString(t.X)
		right := typeExprToString(t.Y)
		return left + " | " + right
	case *ast.UnaryExpr:
		// Tilde constraint: ~int
		return "~" + typeExprToString(t.X)
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.InterfaceType:
		// Inline interface constraint — just return "any" as fallback
		return "any"
	default:
		return "any"
	}
}

// containsIota checks if an AST expression contains the identifier "iota".
// Used to detect iota-based const blocks.
func containsIota(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name == "iota"
	case *ast.BinaryExpr:
		// Handles expressions like "iota + 1", "1 << iota"
		return containsIota(v.X) || containsIota(v.Y)
	case *ast.ParenExpr:
		return containsIota(v.X)
	case *ast.CallExpr:
		// Handles expressions like int(iota)
		for _, arg := range v.Args {
			if containsIota(arg) {
				return true
			}
		}
	}
	return false
}

// parseStructField converts an *ast.Field into one or more StructField values.
//
// Why "one or more"? In Go, a single field line can declare multiple names:
//
//	type Point struct {
//	    X, Y float64  ← one ast.Field, two names → two StructFields
//	}
//
// Embedded fields have no names — len(field.Names) == 0:
//
//	type Post struct {
//	    BaseModel     ← embedded, Names is empty, type is BaseModel
//	    Title string
//	}
func parseStructField(field *ast.Field) []StructField {
	var fields []StructField

	// Extract field comment — prefer inline comment, fall back to doc comment
	fieldComment := extractComment(field.Comment)
	if fieldComment == "" {
		fieldComment = extractComment(field.Doc)
	}

	// Struct tags are stored as raw string literals including backticks:
	// `json:"name,omitempty"`
	tag := ""
	if field.Tag != nil {
		tag = field.Tag.Value
	}

	// Resolve the field's type into base type name + flags
	goType, isPointer, isSlice, isMap := resolveFieldType(field.Type)

	// Capture the full type expression for nested composite types.
	// e.g., "[][]int", "map[string][]int", "*[]string"
	fullType := fullTypeString(field.Type)

	// For map types, extract the key and value types from the AST.
	// resolveFieldType returns "map" as the type name — we need more detail.
	mapKeyType, mapValType := "", ""
	if isMap {
		mapKeyType, mapValType = resolveMapTypes(field.Type)
	}

	// For inline anonymous struct fields, recursively parse the inner fields.
	// e.g., Data struct { User User; Token string } `json:"data"`
	isInlineStruct := false
	var inlineFields []StructField
	if st, ok := field.Type.(*ast.StructType); ok {
		isInlineStruct = true
		for _, innerField := range st.Fields.List {
			inlineFields = append(inlineFields, parseStructField(innerField)...)
		}
	}

	// For function-type fields, parse the full signature.
	var funcSig *FuncSig
	if ft, ok := field.Type.(*ast.FuncType); ok {
		funcSig = parseFuncType(ft)
	}

	// Handle embedded fields (no explicit name)
	if len(field.Names) == 0 {
		name := goType // For embedded fields, the type name is the field name
		fields = append(fields, StructField{
			Name:           name,
			GoType:         goType,
			FullType:       fullType,
			Tag:            tag,
			IsPointer:      isPointer,
			IsSlice:        isSlice,
			IsMap:          isMap,
			MapKeyType:     mapKeyType,
			MapValType:     mapValType,
			IsEmbedded:     true,
			IsExported:     ast.IsExported(name),
			IsInlineStruct: isInlineStruct,
			InlineFields:   inlineFields,
			FuncSig:        funcSig,
			Comment:        fieldComment,
		})
		return fields
	}

	// Handle named fields (one or more names per ast.Field)
	for _, name := range field.Names {
		fields = append(fields, StructField{
			Name:           name.Name,
			GoType:         goType,
			FullType:       fullType,
			Tag:            tag,
			IsPointer:      isPointer,
			IsSlice:        isSlice,
			IsMap:          isMap,
			MapKeyType:     mapKeyType,
			MapValType:     mapValType,
			IsEmbedded:     false,
			IsExported:     name.IsExported(),
			IsInlineStruct: isInlineStruct,
			InlineFields:   inlineFields,
			FuncSig:        funcSig,
			Comment:        fieldComment,
		})
	}

	return fields
}

// resolveFieldType recursively walks an AST type expression and extracts:
//   - typeName: the base type (e.g., "string", "User", "time.Time")
//   - isPointer: whether the outermost type is a pointer (*T)
//   - isSlice: whether the type is a slice/array ([]T)
//   - isMap: whether the type is a map (map[K]V)
//
// Examples:
//
//	string         → ("string", false, false, false)
//	*string        → ("string", true,  false, false)
//	[]User         → ("User",   false, true,  false)
//	[]*User        → ("User",   true,  true,  false)
//	map[string]int → ("map",    false, false, true)
//	time.Time      → ("time.Time", false, false, false)
//	interface{}    → ("interface{}", false, false, false)
func resolveFieldType(expr ast.Expr) (typeName string, isPointer, isSlice, isMap bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		// Simple identifier: string, int, User, etc.
		return t.Name, false, false, false

	case *ast.StarExpr:
		// Pointer type: *string, *User
		// Recurse into the inner type, set isPointer=true
		name, _, s, m := resolveFieldType(t.X)
		return name, true, s, m

	case *ast.ArrayType:
		// Slice/array type: []string, []User, []*User
		// Recurse into element type, set isSlice=true
		name, p, _, m := resolveFieldType(t.Elt)
		return name, p, true, m

	case *ast.MapType:
		// Map type: map[string]int, map[string]User
		// We don't decompose key/value types yet — just flag it as a map
		return "map", false, false, true

	case *ast.SelectorExpr:
		// Qualified type from another package: time.Time, uuid.UUID
		// SelectorExpr has X (package ident) and Sel (type name)
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name, false, false, false
		}
		return t.Sel.Name, false, false, false

	case *ast.InterfaceType:
		// Empty interface: interface{}
		return "interface{}", false, false, false

	case *ast.ChanType:
		// Channel type: chan T, chan<- T, <-chan T
		// Extract the element type — the channel itself has no TS equivalent.
		name, p, s, m := resolveFieldType(t.Value)
		return name, p, s, m

	case *ast.FuncType:
		// Function type as a field — no direct TS equivalent from struct context
		return "Function", false, false, false

	default:
		// Unrecognized type expression
		return "unknown", false, false, false
	}
}

// extractComment extracts clean comment text from a Go AST *ast.CommentGroup.
//
// Strips comment markers (// and /* */) and trims whitespace.
// Returns empty string if the comment group is nil.
// Multi-line comments are joined with newlines.
func extractComment(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	// CommentGroup.Text() returns the text with // stripped and trimmed.
	text := strings.TrimSpace(cg.Text())
	return text
}

// parseFuncType extracts parameters and return types from an *ast.FuncType.
//
// Go AST represents function params as a FieldList where each Field can have
// multiple names sharing one type:
//
//	func(a, b int)  → one Field with Names=[a, b] and Type=int
//	func(a int, b string) → two Fields
//
// We expand multi-name fields into individual FuncParams.
func parseFuncType(ft *ast.FuncType) *FuncSig {
	sig := &FuncSig{}

	// Parse parameters
	if ft.Params != nil {
		for _, field := range ft.Params.List {
			typeName := fullTypeString(field.Type)
			if len(field.Names) == 0 {
				// Unnamed parameter: func(string, int)
				sig.Params = append(sig.Params, FuncParam{
					Name: "",
					Type: typeName,
				})
			} else {
				// Named parameters: func(a, b int) → expand to two params
				for _, name := range field.Names {
					sig.Params = append(sig.Params, FuncParam{
						Name: name.Name,
						Type: typeName,
					})
				}
			}
		}
	}

	// Parse return types
	if ft.Results != nil {
		for _, field := range ft.Results.List {
			typeName := fullTypeString(field.Type)
			if len(field.Names) == 0 {
				sig.Results = append(sig.Results, FuncParam{
					Name: "",
					Type: typeName,
				})
			} else {
				for _, name := range field.Names {
					sig.Results = append(sig.Results, FuncParam{
						Name: name.Name,
						Type: typeName,
					})
				}
			}
		}
	}

	return sig
}

// fullTypeString renders a complete Go type expression from AST back to a string.
//
// This preserves the full nesting so the generator can recursively convert it:
//
//	[][]int             → "[][]int"
//	map[string][]int    → "map[string][]int"
//	map[string]map[string]string → "map[string]map[string]string"
//	*[]string           → "*[]string"
//	[]*User             → "[]*User"
//	**int               → "**int"
func fullTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + fullTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + fullTypeString(t.Elt)
	case *ast.MapType:
		return "map[" + fullTypeString(t.Key) + "]" + fullTypeString(t.Value)
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.ChanType:
		// Channel: extract the element type (chan string → string)
		return fullTypeString(t.Value)
	case *ast.FuncType:
		return "Function"
	case *ast.StructType:
		return "struct"
	default:
		return "unknown"
	}
}

// resolveMapTypes extracts the key and value type names from a map type expression.
// It walks through pointer/slice wrappers to find the underlying ast.MapType.
//
// Examples:
//
//	map[string]int       → ("string", "int")
//	map[string][]int     → ("string", "int")  (value is slice — IsSlice on the value not tracked here)
//	map[string]User      → ("string", "User")
//	map[int64]string     → ("int64", "string")
//	*map[string]int      → ("string", "int")  (unwraps pointer)
func resolveMapTypes(expr ast.Expr) (keyType, valType string) {
	switch t := expr.(type) {
	case *ast.MapType:
		keyType = simpleTypeName(t.Key)
		valType = simpleTypeName(t.Value)
		return keyType, valType
	case *ast.StarExpr:
		// Pointer to map
		return resolveMapTypes(t.X)
	default:
		return "string", "unknown"
	}
}

// simpleTypeName extracts a type name from an AST expression for map key/value types.
// Returns the base type name without pointer/slice/map decomposition.
func simpleTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.ArrayType:
		// []int → "int" (the caller handles the slice wrapping)
		return simpleTypeName(t.Elt)
	case *ast.StarExpr:
		return simpleTypeName(t.X)
	case *ast.MapType:
		return "map"
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "unknown"
	}
}

// exprToString converts a constant value AST expression to a plain string.
//
// For string literals, quotes are stripped:
//
//	"user"  → user
//	`hello` → hello
//
// For numeric literals, the raw value is returned:
//
//	42    → "42"
//	3.14  → "3.14"
//
// For identifiers (e.g., iota, or referencing another const), the name is returned.
// For complex expressions (binary ops, function calls), empty string and false are returned.
//
// The second return value (isMultiLine) is true when the original Go literal
// was a backtick string — these may contain newlines and need template literals
// (backticks) in TypeScript output instead of double quotes.
func exprToString(expr ast.Expr) (string, bool) {
	switch v := expr.(type) {
	case *ast.BasicLit:
		// Literal value: "user", 42, 3.14
		if v.Kind == token.STRING {
			s := v.Value
			// Backtick strings in Go can span multiple lines.
			// Track this so the generator uses TS template literals.
			isBacktick := len(s) >= 2 && s[0] == '`'
			// Strip surrounding quotes (double or backtick)
			if len(s) >= 2 && (s[0] == '"' || s[0] == '`') {
				return s[1 : len(s)-1], isBacktick
			}
		}
		return v.Value, false

	case *ast.Ident:
		// Identifier: iota, true, false, or a reference to another const
		return v.Name, false

	default:
		// Complex expression: binary ops (1 << iota), function calls, etc.
		// TODO: evaluate constant expressions (especially iota arithmetic)
		return "", false
	}
}
