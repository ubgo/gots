package parser

import (
	"testing"
)

const testFile = "../../testdata/all_types.go"

func TestParseFile(t *testing.T) {
	p := New()
	err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	result := p.Result()

	if len(result.Types) == 0 {
		t.Fatal("expected types, got none")
	}
	if len(result.Consts) == 0 {
		t.Fatal("expected consts, got none")
	}
	if len(result.Structs) == 0 {
		t.Fatal("expected structs, got none")
	}
}

// --- Type declaration tests ---

func TestParseTypedStringType(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	td := findType(p.Result(), "MessageRole")
	if td == nil {
		t.Fatal("MessageRole type not found")
	}
	if td.Underlying != "string" {
		t.Errorf("expected underlying type 'string', got %q", td.Underlying)
	}
}

func TestParseTypedIntType(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	td := findType(p.Result(), "LockType")
	if td == nil {
		t.Fatal("LockType type not found")
	}
	if td.Underlying != "int" {
		t.Errorf("expected underlying type 'int', got %q", td.Underlying)
	}
}

func TestParseNamedPrimitiveAliases(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		underlying string
	}{
		{"UserID", "int64"},
		{"Slug", "string"},
		{"Amount", "float64"},
		{"Flag", "bool"},
	}

	for _, tt := range tests {
		td := findType(p.Result(), tt.name)
		if td == nil {
			t.Errorf("type %s not found", tt.name)
			continue
		}
		if td.Underlying != tt.underlying {
			t.Errorf("type %s: expected underlying %q, got %q", tt.name, tt.underlying, td.Underlying)
		}
	}
}

func TestParseAllExpectedTypes(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	expectedTypes := []string{
		"MessageRole",
		"FuncToolType",
		"ItemGroupSourceTrainingStatus",
		"LockType",
		"Priority",
		"UserID",
		"Slug",
		"Amount",
		"Flag",
	}

	for _, name := range expectedTypes {
		if findType(p.Result(), name) == nil {
			t.Errorf("expected type %q not found", name)
		}
	}
}

// --- Typed constant tests ---

func TestParseTypedStringConsts(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	consts := findConstsByType(p.Result(), "MessageRole")
	if len(consts) != 3 {
		t.Fatalf("expected 3 MessageRole consts, got %d", len(consts))
	}

	expected := []struct {
		name  string
		value string
	}{
		{"MessageRoleUser", "user"},
		{"MessageRoleAssistant", "assistant"},
		{"MessageRoleSystem", "system"},
	}

	for i, exp := range expected {
		if consts[i].Name != exp.name {
			t.Errorf("const[%d]: expected name %q, got %q", i, exp.name, consts[i].Name)
		}
		if consts[i].Value != exp.value {
			t.Errorf("const[%d]: expected value %q, got %q", i, exp.value, consts[i].Value)
		}
		if consts[i].TypeName != "MessageRole" {
			t.Errorf("const[%d]: expected type 'MessageRole', got %q", i, consts[i].TypeName)
		}
	}
}

func TestParseSingleTypedConst(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	consts := findConstsByType(p.Result(), "FuncToolType")
	if len(consts) != 1 {
		t.Fatalf("expected 1 FuncToolType const, got %d", len(consts))
	}
	if consts[0].Value != "vector" {
		t.Errorf("expected value 'vector', got %q", consts[0].Value)
	}
}

func TestParseManyTypedConsts(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	consts := findConstsByType(p.Result(), "ItemGroupSourceTrainingStatus")
	if len(consts) != 5 {
		t.Fatalf("expected 5 ItemGroupSourceTrainingStatus consts, got %d", len(consts))
	}

	expectedValues := []string{"pending", "chunking", "training", "completed", "error"}
	for i, exp := range expectedValues {
		if consts[i].Value != exp {
			t.Errorf("const[%d]: expected value %q, got %q", i, exp, consts[i].Value)
		}
	}
}

func TestParseIotaConstsSkipBlank(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	// LockType uses _ = iota as first entry, so it should be skipped
	consts := findConstsByType(p.Result(), "LockType")
	if len(consts) != 3 {
		t.Fatalf("expected 3 LockType consts (blank _ skipped), got %d", len(consts))
	}

	expectedNames := []string{"LockReadonly", "LockUpdate", "LockDelete"}
	for i, exp := range expectedNames {
		if consts[i].Name != exp {
			t.Errorf("const[%d]: expected name %q, got %q", i, exp, consts[i].Name)
		}
	}
}

func TestParseIotaConstsNoSkip(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	consts := findConstsByType(p.Result(), "Priority")
	if len(consts) != 3 {
		t.Fatalf("expected 3 Priority consts, got %d", len(consts))
	}

	expectedNames := []string{"PriorityLow", "PriorityMedium", "PriorityHigh"}
	for i, exp := range expectedNames {
		if consts[i].Name != exp {
			t.Errorf("const[%d]: expected name %q, got %q", i, exp, consts[i].Name)
		}
	}
}

// --- Untyped constant tests ---

func TestParseUntypedStringConsts(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	untyped := findConstsByType(p.Result(), "")

	wsActive := findConstByName(untyped, "WorkspaceSubStatusActive")
	if wsActive == nil {
		t.Fatal("WorkspaceSubStatusActive not found")
	}
	if wsActive.Value != "active" {
		t.Errorf("expected value 'active', got %q", wsActive.Value)
	}
	if wsActive.TypeName != "" {
		t.Errorf("expected empty type name for untyped const, got %q", wsActive.TypeName)
	}
}

func TestParseStandaloneStringConst(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	untyped := findConstsByType(p.Result(), "")
	c := findConstByName(untyped, "DefaultAIModelID")
	if c == nil {
		t.Fatal("DefaultAIModelID not found")
	}
	if c.Value != "openai_gpt_4" {
		t.Errorf("expected value 'openai_gpt_4', got %q", c.Value)
	}
}

func TestParseStandaloneIntConst(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	untyped := findConstsByType(p.Result(), "")
	c := findConstByName(untyped, "DefaultVectorDimensions")
	if c == nil {
		t.Fatal("DefaultVectorDimensions not found")
	}
	if c.Value != "768" {
		t.Errorf("expected value '768', got %q", c.Value)
	}
}

func TestParseStandaloneBoolConsts(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	untyped := findConstsByType(p.Result(), "")

	fe := findConstByName(untyped, "FeatureEnabled")
	if fe == nil {
		t.Fatal("FeatureEnabled not found")
	}
	if fe.Value != "true" {
		t.Errorf("expected value 'true', got %q", fe.Value)
	}

	dm := findConstByName(untyped, "DebugMode")
	if dm == nil {
		t.Fatal("DebugMode not found")
	}
	if dm.Value != "false" {
		t.Errorf("expected value 'false', got %q", dm.Value)
	}
}

func TestParseMultiLineStringConst(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	untyped := findConstsByType(p.Result(), "")
	c := findConstByName(untyped, "DefaultAIInstruction")
	if c == nil {
		t.Fatal("DefaultAIInstruction not found")
	}
	if !c.IsMultiLine {
		t.Error("expected IsMultiLine=true for backtick string")
	}
	if c.Value == "" {
		t.Error("expected non-empty value for multi-line string")
	}
}

// --- Struct tests ---

func TestParseBasicStruct(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "User")
	if sd == nil {
		t.Fatal("User struct not found")
	}
	if len(sd.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(sd.Fields))
	}

	tests := []struct {
		name     string
		goType   string
		exported bool
	}{
		{"ID", "int64", true},
		{"Name", "string", true},
		{"Email", "string", true},
	}

	for i, tt := range tests {
		f := sd.Fields[i]
		if f.Name != tt.name {
			t.Errorf("field[%d]: expected name %q, got %q", i, tt.name, f.Name)
		}
		if f.GoType != tt.goType {
			t.Errorf("field[%d]: expected type %q, got %q", i, tt.goType, f.GoType)
		}
		if f.IsExported != tt.exported {
			t.Errorf("field[%d]: expected exported=%v, got %v", i, tt.exported, f.IsExported)
		}
	}
}

func TestParseStructWithPointers(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Post")
	if sd == nil {
		t.Fatal("Post struct not found")
	}

	content := findField(sd, "Content")
	if content == nil {
		t.Fatal("Content field not found")
	}
	if !content.IsPointer {
		t.Error("Content should be a pointer")
	}
	if content.GoType != "string" {
		t.Errorf("Content type: expected 'string', got %q", content.GoType)
	}

	author := findField(sd, "Author")
	if author == nil {
		t.Fatal("Author field not found")
	}
	if !author.IsPointer {
		t.Error("Author should be a pointer")
	}
	if author.GoType != "User" {
		t.Errorf("Author type: expected 'User', got %q", author.GoType)
	}
}

func TestParseStructAllPrimitiveTypes(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "FullModel")
	if sd == nil {
		t.Fatal("FullModel struct not found")
	}

	tests := []struct {
		name   string
		goType string
	}{
		{"StringField", "string"},
		{"IntField", "int"},
		{"Int8Field", "int8"},
		{"Int16Field", "int16"},
		{"Int32Field", "int32"},
		{"Int64Field", "int64"},
		{"UintField", "uint"},
		{"Uint8Field", "uint8"},
		{"Uint16Field", "uint16"},
		{"Uint32Field", "uint32"},
		{"Uint64Field", "uint64"},
		{"Float32Field", "float32"},
		{"Float64Field", "float64"},
		{"BoolField", "bool"},
		{"ByteField", "byte"},
		{"RuneField", "rune"},
	}

	for _, tt := range tests {
		f := findField(sd, tt.name)
		if f == nil {
			t.Errorf("field %q not found", tt.name)
			continue
		}
		if f.GoType != tt.goType {
			t.Errorf("field %q: expected type %q, got %q", tt.name, tt.goType, f.GoType)
		}
	}
}

func TestParseStructExternalTypes(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Event")
	if sd == nil {
		t.Fatal("Event struct not found")
	}

	createdAt := findField(sd, "CreatedAt")
	if createdAt == nil {
		t.Fatal("CreatedAt field not found")
	}
	if createdAt.GoType != "time.Time" {
		t.Errorf("CreatedAt type: expected 'time.Time', got %q", createdAt.GoType)
	}
	if createdAt.IsPointer {
		t.Error("CreatedAt should not be a pointer")
	}

	updatedAt := findField(sd, "UpdatedAt")
	if updatedAt == nil {
		t.Fatal("UpdatedAt field not found")
	}
	if updatedAt.GoType != "time.Time" {
		t.Errorf("UpdatedAt type: expected 'time.Time', got %q", updatedAt.GoType)
	}
	if !updatedAt.IsPointer {
		t.Error("UpdatedAt should be a pointer")
	}

	payload := findField(sd, "Payload")
	if payload == nil {
		t.Fatal("Payload field not found")
	}
	if payload.GoType != "json.RawMessage" {
		t.Errorf("Payload type: expected 'json.RawMessage', got %q", payload.GoType)
	}
}

func TestParseStructWithSlices(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "PostList")
	if sd == nil {
		t.Fatal("PostList struct not found")
	}

	posts := findField(sd, "Posts")
	if posts == nil {
		t.Fatal("Posts field not found")
	}
	if !posts.IsSlice {
		t.Error("Posts should be a slice")
	}
	if posts.GoType != "Post" {
		t.Errorf("Posts type: expected 'Post', got %q", posts.GoType)
	}

	tags := findField(sd, "Tags")
	if tags == nil {
		t.Fatal("Tags field not found")
	}
	if !tags.IsSlice {
		t.Error("Tags should be a slice")
	}
	if tags.GoType != "string" {
		t.Errorf("Tags type: expected 'string', got %q", tags.GoType)
	}
}

func TestParseStructWithEmbedding(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Article")
	if sd == nil {
		t.Fatal("Article struct not found")
	}

	// First field should be the embedded BaseModel
	if len(sd.Fields) < 1 {
		t.Fatal("expected at least 1 field")
	}

	embedded := sd.Fields[0]
	if !embedded.IsEmbedded {
		t.Error("first field should be embedded")
	}
	if embedded.GoType != "BaseModel" {
		t.Errorf("embedded type: expected 'BaseModel', got %q", embedded.GoType)
	}
}

func TestParseStructUnexportedFieldsSkipped(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Config")
	if sd == nil {
		t.Fatal("Config struct not found")
	}

	// Should have 4 fields total (2 exported + 2 unexported)
	if len(sd.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(sd.Fields))
	}

	exportedCount := 0
	unexportedCount := 0
	for _, f := range sd.Fields {
		if f.IsExported {
			exportedCount++
		} else {
			unexportedCount++
		}
	}
	if exportedCount != 2 {
		t.Errorf("expected 2 exported fields, got %d", exportedCount)
	}
	if unexportedCount != 2 {
		t.Errorf("expected 2 unexported fields, got %d", unexportedCount)
	}
}

func TestParseStructJsonDash(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "SecureUser")
	if sd == nil {
		t.Fatal("SecureUser struct not found")
	}

	password := findField(sd, "Password")
	if password == nil {
		t.Fatal("Password field not found")
	}
	// Parser should still capture the field; generator decides to skip json:"-"
	if password.Tag == "" {
		t.Error("Password should have a tag")
	}
}

func TestParseStructNoTags(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "NoTagStruct")
	if sd == nil {
		t.Fatal("NoTagStruct struct not found")
	}

	for _, f := range sd.Fields {
		if f.Tag != "" {
			t.Errorf("field %q: expected empty tag, got %q", f.Name, f.Tag)
		}
	}
}

func TestParseStructInterfaceField(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "DynamicPayload")
	if sd == nil {
		t.Fatal("DynamicPayload struct not found")
	}

	data := findField(sd, "Data")
	if data == nil {
		t.Fatal("Data field not found")
	}
	if data.GoType != "interface{}" {
		t.Errorf("Data type: expected 'interface{}', got %q", data.GoType)
	}
}

func TestParseStructMapField(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Metadata")
	if sd == nil {
		t.Fatal("Metadata struct not found")
	}

	labels := findField(sd, "Labels")
	if labels == nil {
		t.Fatal("Labels field not found")
	}
	if !labels.IsMap {
		t.Error("Labels should be a map")
	}
}

func TestParseEmptyStruct(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Empty")
	if sd == nil {
		t.Fatal("Empty struct not found")
	}
	if len(sd.Fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(sd.Fields))
	}
}

func TestParseAllExpectedStructs(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	expectedStructs := []string{
		"User",
		"Post",
		"FullModel",
		"Event",
		"PostList",
		"BaseModel",
		"Article",
		"Config",
		"SecureUser",
		"NoTagStruct",
		"OptionalFields",
		"DynamicPayload",
		"Metadata",
		"Empty",
		"Wrapper",
		"PointerInfo",
		"ComplexNumbers",
		"Matrix",
		"Worker",
		"NestedTypes",
		"Box",
		"Pair",
		"NumberBox",
		"PaginatedResponse",
		"Result",
		"Cache",
		"AnyPayload",
		"Point",
		"Rect",
		"Team",
		"APIResponse",
		"Hooks",
	}

	for _, name := range expectedStructs {
		if findStruct(p.Result(), name) == nil {
			t.Errorf("expected struct %q not found", name)
		}
	}
}

// --- exprToString tests ---

func TestExprToString(t *testing.T) {
	// These are tested indirectly through ParseFile above,
	// but we verify specific behaviors here.

	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	untyped := findConstsByType(p.Result(), "")

	// String values should have quotes stripped
	c := findConstByName(untyped, "HeaderWorkspaceId")
	if c == nil {
		t.Fatal("HeaderWorkspaceId not found")
	}
	if c.Value != "x-workspace-id" {
		t.Errorf("expected 'x-workspace-id', got %q", c.Value)
	}

	// Int values should be raw
	c = findConstByName(untyped, "DefaultVectorDimensions")
	if c == nil {
		t.Fatal("DefaultVectorDimensions not found")
	}
	if c.Value != "768" {
		t.Errorf("expected '768', got %q", c.Value)
	}
}

// --- New type tests (types 30-53 from fixture) ---

func TestParseUintptrField(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "PointerInfo")
	if sd == nil {
		t.Fatal("PointerInfo struct not found")
	}
	f := findField(sd, "Addr")
	if f == nil {
		t.Fatal("Addr field not found")
	}
	if f.GoType != "uintptr" {
		t.Errorf("expected type 'uintptr', got %q", f.GoType)
	}
}

func TestParseComplexFields(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "ComplexNumbers")
	if sd == nil {
		t.Fatal("ComplexNumbers struct not found")
	}

	c64 := findField(sd, "C64")
	if c64 == nil {
		t.Fatal("C64 field not found")
	}
	if c64.GoType != "complex64" {
		t.Errorf("expected type 'complex64', got %q", c64.GoType)
	}

	c128 := findField(sd, "C128")
	if c128 == nil {
		t.Fatal("C128 field not found")
	}
	if c128.GoType != "complex128" {
		t.Errorf("expected type 'complex128', got %q", c128.GoType)
	}
}

func TestParseFixedArrayField(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Matrix")
	if sd == nil {
		t.Fatal("Matrix struct not found")
	}
	f := findField(sd, "Row")
	if f == nil {
		t.Fatal("Row field not found")
	}
	// Fixed arrays are parsed like slices — IsSlice=true with element type
	if !f.IsSlice {
		t.Error("Row should be detected as slice/array")
	}
	if f.GoType != "float64" {
		t.Errorf("expected element type 'float64', got %q", f.GoType)
	}
}

func TestParseChannelFields(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Worker")
	if sd == nil {
		t.Fatal("Worker struct not found")
	}
	// Channels should be parsed (even though they don't have a TS equivalent)
	if len(sd.Fields) != 4 {
		t.Errorf("expected 4 fields in Worker, got %d", len(sd.Fields))
	}
}

func TestParseFunctionTypes(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	// Function types are not captured as TypeDef (they're not ident-based)
	// but they should not cause parse errors.
	// The parser currently skips them silently — this is expected behavior.
}

func TestParseInterfaceTypes(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	// Interface types are not captured as TypeDef or StructDef currently.
	// They should not cause parse errors.
	// Future: may be added as a separate InterfaceDef type.
}

func TestParseTypeAliases(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	// Type aliases (type Name = string) use TypeSpec.Assign != 0 in the AST.
	// Current parser treats them like type definitions — the underlying type is captured.
	// This is acceptable for TS generation since both produce the same output.
	td := findType(p.Result(), "Name")
	if td == nil {
		t.Fatal("type alias 'Name' not found")
	}
	if td.Underlying != "string" {
		t.Errorf("expected underlying 'string', got %q", td.Underlying)
	}

	td = findType(p.Result(), "Age")
	if td == nil {
		t.Fatal("type alias 'Age' not found")
	}
	if td.Underlying != "int" {
		t.Errorf("expected underlying 'int', got %q", td.Underlying)
	}
}

func TestParseNestedCompositeFields(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "NestedTypes")
	if sd == nil {
		t.Fatal("NestedTypes struct not found")
	}

	// [][]int — slice of slices
	matrix := findField(sd, "Matrix")
	if matrix == nil {
		t.Fatal("Matrix field not found")
	}
	if !matrix.IsSlice {
		t.Error("Matrix should be a slice")
	}

	// *[]string — pointer to slice
	ptrSlice := findField(sd, "PtrToSlice")
	if ptrSlice == nil {
		t.Fatal("PtrToSlice field not found")
	}
	if !ptrSlice.IsPointer {
		t.Error("PtrToSlice should be a pointer")
	}

	// **int — pointer to pointer
	ptrPtr := findField(sd, "PtrToPtr")
	if ptrPtr == nil {
		t.Fatal("PtrToPtr field not found")
	}
	if !ptrPtr.IsPointer {
		t.Error("PtrToPtr should be a pointer")
	}

	// []*User — slice of pointers
	slicePtrs := findField(sd, "SliceOfPtrs")
	if slicePtrs == nil {
		t.Fatal("SliceOfPtrs field not found")
	}
	if !slicePtrs.IsSlice {
		t.Error("SliceOfPtrs should be a slice")
	}
}

func TestParseGenericStruct(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Box")
	if sd == nil {
		t.Fatal("Box generic struct not found")
	}
	if len(sd.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(sd.Fields))
	}
	f := findField(sd, "Value")
	if f == nil {
		t.Fatal("Value field not found")
	}
	// Generic type param T appears as an identifier
	if f.GoType != "T" {
		t.Errorf("expected type 'T', got %q", f.GoType)
	}
}

func TestParseGenericStructMultipleParams(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Pair")
	if sd == nil {
		t.Fatal("Pair generic struct not found")
	}
	if len(sd.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(sd.Fields))
	}

	key := findField(sd, "Key")
	if key == nil {
		t.Fatal("Key field not found")
	}
	if key.GoType != "K" {
		t.Errorf("expected type 'K', got %q", key.GoType)
	}

	value := findField(sd, "Value")
	if value == nil {
		t.Fatal("Value field not found")
	}
	if value.GoType != "V" {
		t.Errorf("expected type 'V', got %q", value.GoType)
	}
}

func TestParseGenericPaginatedResponse(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "PaginatedResponse")
	if sd == nil {
		t.Fatal("PaginatedResponse generic struct not found")
	}

	data := findField(sd, "Data")
	if data == nil {
		t.Fatal("Data field not found")
	}
	if !data.IsSlice {
		t.Error("Data should be a slice ([]T)")
	}
	if data.GoType != "T" {
		t.Errorf("expected element type 'T', got %q", data.GoType)
	}
}

func TestParseGenericResultMultiTypeParams(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Result")
	if sd == nil {
		t.Fatal("Result generic struct not found")
	}

	data := findField(sd, "Data")
	if data == nil {
		t.Fatal("Data field not found")
	}
	if !data.IsPointer {
		t.Error("Data should be a pointer (*T)")
	}
	if data.GoType != "T" {
		t.Errorf("expected type 'T', got %q", data.GoType)
	}

	errField := findField(sd, "Error")
	if errField == nil {
		t.Fatal("Error field not found")
	}
	if errField.GoType != "E" {
		t.Errorf("expected type 'E', got %q", errField.GoType)
	}
}

func TestParseGenericCache(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Cache")
	if sd == nil {
		t.Fatal("Cache generic struct not found")
	}

	items := findField(sd, "Items")
	if items == nil {
		t.Fatal("Items field not found")
	}
	if !items.IsMap {
		t.Error("Items should be a map")
	}
}

func TestParseAnyField(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "AnyPayload")
	if sd == nil {
		t.Fatal("AnyPayload struct not found")
	}

	data := findField(sd, "Data")
	if data == nil {
		t.Fatal("Data field not found")
	}
	// "any" is an alias for interface{} — the AST sees it as an Ident("any")
	if data.GoType != "any" {
		t.Errorf("expected type 'any', got %q", data.GoType)
	}
}

func TestParseMultiNameFields(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Point")
	if sd == nil {
		t.Fatal("Point struct not found")
	}
	// X, Y float64 → two separate StructFields
	if len(sd.Fields) != 2 {
		t.Fatalf("expected 2 fields (X, Y), got %d", len(sd.Fields))
	}

	x := findField(sd, "X")
	if x == nil {
		t.Fatal("X field not found")
	}
	if x.GoType != "float64" {
		t.Errorf("expected type 'float64', got %q", x.GoType)
	}

	y := findField(sd, "Y")
	if y == nil {
		t.Fatal("Y field not found")
	}
	if y.GoType != "float64" {
		t.Errorf("expected type 'float64', got %q", y.GoType)
	}

	// Rect: Left, Top, Right, Bottom int → four fields
	sd = findStruct(p.Result(), "Rect")
	if sd == nil {
		t.Fatal("Rect struct not found")
	}
	if len(sd.Fields) != 4 {
		t.Errorf("expected 4 fields, got %d", len(sd.Fields))
	}
}

func TestParseArrayOfStructsField(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Team")
	if sd == nil {
		t.Fatal("Team struct not found")
	}

	members := findField(sd, "Members")
	if members == nil {
		t.Fatal("Members field not found")
	}
	if !members.IsSlice {
		t.Error("Members ([5]User) should be detected as slice/array")
	}
	if members.GoType != "User" {
		t.Errorf("expected element type 'User', got %q", members.GoType)
	}
}

func TestParseAnonymousStructField(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "APIResponse")
	if sd == nil {
		t.Fatal("APIResponse struct not found")
	}
	// Anonymous struct fields are parsed but the inner fields are not accessible
	// through the current StructField model. The Data field type will be "unknown"
	// since it's an inline struct definition.
	if len(sd.Fields) < 2 {
		t.Errorf("expected at least 2 fields, got %d", len(sd.Fields))
	}
}

func TestParseFunctionFieldInStruct(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	sd := findStruct(p.Result(), "Hooks")
	if sd == nil {
		t.Fatal("Hooks struct not found")
	}
	if len(sd.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(sd.Fields))
	}
}

func TestParseNamedChannelTypes(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	// Named channel types are not captured as TypeDef (they're not ident-based)
	// but they should not cause parse errors.
	// The parser silently skips non-ident, non-struct type specs.
}

func TestParseNamedArrayType(t *testing.T) {
	p := New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	// Named array types like "type IPv4 [4]byte" are not captured as TypeDef
	// since the underlying type is an ArrayType, not an Ident.
	// They should not cause parse errors.
}

// --- Error case tests ---

func TestParseNonexistentFile(t *testing.T) {
	p := New()
	err := p.ParseFile("nonexistent_file.go")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestParseMultipleFiles(t *testing.T) {
	p := New()

	// Parse the same file twice — should accumulate results
	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	firstCount := len(p.Result().Types)
	if firstCount == 0 {
		t.Fatal("expected types after first parse")
	}

	if err := p.ParseFile(testFile); err != nil {
		t.Fatal(err)
	}

	secondCount := len(p.Result().Types)
	if secondCount != firstCount*2 {
		t.Errorf("expected %d types after second parse, got %d", firstCount*2, secondCount)
	}
}

// --- Helper functions ---

func findType(r *ParseResult, name string) *TypeDef {
	for _, td := range r.Types {
		if td.Name == name {
			return &td
		}
	}
	return nil
}

func findConstsByType(r *ParseResult, typeName string) []ConstDef {
	var result []ConstDef
	for _, c := range r.Consts {
		if c.TypeName == typeName {
			result = append(result, c)
		}
	}
	return result
}

func findConstByName(consts []ConstDef, name string) *ConstDef {
	for _, c := range consts {
		if c.Name == name {
			return &c
		}
	}
	return nil
}

func findStruct(r *ParseResult, name string) *StructDef {
	for _, sd := range r.Structs {
		if sd.Name == name {
			return &sd
		}
	}
	return nil
}

func findField(sd *StructDef, name string) *StructField {
	for _, f := range sd.Fields {
		if f.Name == name {
			return &f
		}
	}
	return nil
}
