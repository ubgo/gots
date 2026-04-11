package generator

import (
	"strings"
	"testing"

	"github.com/khanakia/gots/pkg/config"
	"github.com/khanakia/gots/pkg/parser"
)

const testFile = "../../testdata/all_types.go"

// parseTestFile is a helper that parses the test fixture.
func parseTestFile(t *testing.T) *parser.ParseResult {
	t.Helper()
	p := parser.New()
	if err := p.ParseFile(testFile); err != nil {
		t.Fatalf("failed to parse test file: %v", err)
	}
	return p.Result()
}

// generate runs the generator with the given config and returns the output string.
func generate(t *testing.T, cfg *config.Config, result *parser.ParseResult) string {
	t.Helper()
	gen := New(cfg, result)

	grouped := gen.groupConstsByType()
	output, err := gen.generateSource(cfg.Sources[0], grouped)
	if err != nil {
		t.Fatalf("generateSource failed: %v", err)
	}
	return output
}

// defaultConfig returns a config pointing at the test file with sensible defaults.
func defaultConfig() *config.Config {
	return &config.Config{
		DefaultStyle:    "union",
		StructTag:       "json",
		PointerHandling: "optional",
		EmptyInterface:  "unknown",
		Header:          "// generated",
		TypeMap: map[string]string{
			"time.Time":       "string",
			"json.RawMessage": "unknown",
		},
		Sources: []config.Source{
			{
				File:   testFile,
				Output: "/dev/null",
			},
		},
		Overrides: map[string]config.Override{},
	}
}

// --- Union type tests ---

func TestGenerateUnionType(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.DefaultStyle = "union"

	output := generate(t, cfg, result)

	expected := `export type MessageRole = "user" | "assistant" | "system";`
	if !strings.Contains(output, expected) {
		t.Errorf("expected union type:\n%s\n\ngot output:\n%s", expected, output)
	}
}

func TestGenerateUnionSingleValue(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	expected := `export type FuncToolType = "vector";`
	if !strings.Contains(output, expected) {
		t.Errorf("expected single-value union:\n%s\n\ngot output:\n%s", expected, output)
	}
}

func TestGenerateUnionManyValues(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	expected := `export type ItemGroupSourceTrainingStatus = "pending" | "chunking" | "training" | "completed" | "error";`
	if !strings.Contains(output, expected) {
		t.Errorf("expected long union:\n%s\n\ngot output:\n%s", expected, output)
	}
}

// --- Enum type tests ---

func TestGenerateEnum(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.Overrides = map[string]config.Override{
		"MessageRole": {Style: "enum"},
	}

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export enum MessageRole {") {
		t.Error("expected enum declaration")
	}
	if !strings.Contains(output, `  User = "user",`) {
		t.Error("expected stripped prefix key 'User'")
	}
	if !strings.Contains(output, `  Assistant = "assistant",`) {
		t.Error("expected stripped prefix key 'Assistant'")
	}
	if !strings.Contains(output, `  System = "system",`) {
		t.Error("expected stripped prefix key 'System'")
	}
}

func TestGenerateEnumSourceLevel(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.Sources[0].Style = "enum"

	output := generate(t, cfg, result)

	// All typed consts should be enums
	if !strings.Contains(output, "export enum MessageRole {") {
		t.Error("expected MessageRole as enum")
	}
	if !strings.Contains(output, "export enum FuncToolType {") {
		t.Error("expected FuncToolType as enum")
	}
}

// --- Const object tests ---

func TestGenerateConstObject(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.Overrides = map[string]config.Override{
		"MessageRole": {Style: "const-object"},
	}

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export const MessageRole = {") {
		t.Error("expected const object declaration")
	}
	if !strings.Contains(output, `  User: "user",`) {
		t.Error("expected key 'User'")
	}
	if !strings.Contains(output, "} as const;") {
		t.Error("expected 'as const'")
	}
	if !strings.Contains(output, "export type MessageRole = typeof MessageRole[keyof typeof MessageRole];") {
		t.Error("expected type alias from const object")
	}
}

// --- Values array tests ---

func TestGenerateValuesArray(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.Overrides = map[string]config.Override{
		"MessageRole": {ValuesArray: true},
	}

	output := generate(t, cfg, result)

	expected := `export const MessageRoleValues = ["user", "assistant", "system"] as const;`
	if !strings.Contains(output, expected) {
		t.Errorf("expected values array:\n%s\n\ngot output:\n%s", expected, output)
	}
}

func TestGenerateValuesArrayGlobal(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.EmitValuesArray = true

	output := generate(t, cfg, result)

	if !strings.Contains(output, "MessageRoleValues") {
		t.Error("expected MessageRoleValues with global emit_values_array")
	}
	if !strings.Contains(output, "FuncToolTypeValues") {
		t.Error("expected FuncToolTypeValues with global emit_values_array")
	}
}

// --- Type rename tests ---

func TestGenerateTypeRename(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.Overrides = map[string]config.Override{
		"MessageRole": {Style: "union", Name: "Role"},
	}

	output := generate(t, cfg, result)

	expected := `export type Role = "user" | "assistant" | "system";`
	if !strings.Contains(output, expected) {
		t.Errorf("expected renamed type:\n%s\n\ngot output:\n%s", expected, output)
	}
	// Should NOT contain the original name as a type
	if strings.Contains(output, "export type MessageRole") {
		t.Error("original name should not appear when renamed")
	}
}

// --- Untyped constant tests ---

func TestGenerateStandaloneStringConst(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	if !strings.Contains(output, `export const DefaultAIModelID = "openai_gpt_4";`) {
		t.Error("expected standalone string const")
	}
}

func TestGenerateStandaloneIntConst(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	// Int consts are output without quotes
	if !strings.Contains(output, `export const DefaultVectorDimensions = 768;`) {
		t.Error("expected standalone int const without quotes")
	}
}

func TestGenerateStandaloneBoolConst(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	// Bool consts are output without quotes
	if !strings.Contains(output, `export const FeatureEnabled = true;`) {
		t.Error("expected bool const without quotes")
	}
	if !strings.Contains(output, `export const DebugMode = false;`) {
		t.Error("expected bool const without quotes")
	}
}

func TestGenerateJsonDashFieldsSkipped(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export interface SecureUser {") {
		t.Error("expected SecureUser interface")
	}
	// json:"-" fields should be skipped
	if strings.Contains(output, "Password") {
		t.Error("Password (json:\"-\") should be skipped")
	}
	if strings.Contains(output, "Token") {
		t.Error("Token (json:\"-\") should be skipped")
	}
	// Non-hidden fields should be present
	if !strings.Contains(output, "  id: number;") {
		t.Error("expected id field")
	}
	if !strings.Contains(output, "  name: string;") {
		t.Error("expected name field")
	}
}

func TestGenerateNumericUnion(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	// LockType (iota with _ skip) should be numeric union without quotes
	if !strings.Contains(output, "export type LockType = 1 | 2 | 3;") {
		t.Error("expected numeric union for LockType")
	}
}

func TestGenerateNumericEnum(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.Overrides = map[string]config.Override{
		"LockType": {Style: "enum"},
	}

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export enum LockType {") {
		t.Error("expected LockType enum")
	}
	// LockReadonly doesn't start with "LockType" so prefix is not stripped
	if !strings.Contains(output, "  LockReadonly = 1,") {
		t.Error("expected numeric enum value without quotes")
	}
}

func TestGenerateUintptrField(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	if !strings.Contains(output, "  addr: number;") {
		t.Error("expected uintptr mapped to number")
	}
}

func TestGenerateMapField(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	if !strings.Contains(output, "  labels: Record<string, string>;") {
		t.Error("expected map field as Record<string, string>")
	}
}

func TestGenerateAnyField(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.EmptyInterface = "unknown"

	output := generate(t, cfg, result)

	// AnyPayload has `any` field — should respect empty_interface config
	if !strings.Contains(output, "export interface AnyPayload {") {
		t.Error("expected AnyPayload interface")
	}
}

func TestGenerateMultiLineConst(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	// Multi-line should use backticks, not double quotes
	if !strings.Contains(output, "export const DefaultAIInstruction = `") {
		t.Error("expected multi-line const with backtick template literal")
	}
	// Should NOT have double quotes wrapping a multi-line string
	if strings.Contains(output, `export const DefaultAIInstruction = "`) {
		t.Error("multi-line const should not use double quotes")
	}
}

func TestGenerateHeaderConsts(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	if !strings.Contains(output, `export const HeaderWorkspaceId = "x-workspace-id";`) {
		t.Error("expected HeaderWorkspaceId const")
	}
	if !strings.Contains(output, `export const HeaderRequestId = "x-request-id";`) {
		t.Error("expected HeaderRequestId const")
	}
}

// --- Struct / interface tests ---

func TestGenerateBasicInterface(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export interface User {") {
		t.Error("expected User interface")
	}
	if !strings.Contains(output, "  id: number;") {
		t.Error("expected id field mapped to number")
	}
	if !strings.Contains(output, "  name: string;") {
		t.Error("expected name field")
	}
	if !strings.Contains(output, "  email: string;") {
		t.Error("expected email field")
	}
}

func TestGeneratePointerFieldsOptional(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.PointerHandling = "optional"

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export interface Post {") {
		t.Error("expected Post interface")
	}
	// Non-pointer field — no ?
	if !strings.Contains(output, "  title: string;") {
		t.Error("expected non-optional title")
	}
	// Pointer field — should have ?
	if !strings.Contains(output, "  content?: string;") {
		t.Error("expected optional content field")
	}
	if !strings.Contains(output, "  author?: User;") {
		t.Error("expected optional author field")
	}
}

func TestGenerateAllPrimitiveTypeMappings(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export interface FullModel {") {
		t.Error("expected FullModel interface")
	}

	// All numeric types should map to number
	numberFields := []string{
		"int_field", "int8_field", "int16_field", "int32_field", "int64_field",
		"uint_field", "uint8_field", "uint16_field", "uint32_field", "uint64_field",
		"float32_field", "float64_field", "byte_field", "rune_field",
	}
	for _, f := range numberFields {
		expected := f + ": number;"
		if !strings.Contains(output, expected) {
			t.Errorf("expected %q in output", expected)
		}
	}

	// String fields
	if !strings.Contains(output, "string_field: string;") {
		t.Error("expected string_field: string")
	}

	// Bool fields
	if !strings.Contains(output, "bool_field: boolean;") {
		t.Error("expected bool_field: boolean")
	}
}

func TestGenerateExternalTypeMapping(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export interface Event {") {
		t.Error("expected Event interface")
	}
	// time.Time → string (from type_map)
	if !strings.Contains(output, "  created_at: string;") {
		t.Error("expected created_at mapped to string via type_map")
	}
	// *time.Time → optional string
	if !strings.Contains(output, "  updated_at?: string;") {
		t.Error("expected updated_at as optional string")
	}
	// json.RawMessage → unknown (from type_map)
	if !strings.Contains(output, "  payload: unknown;") {
		t.Error("expected payload mapped to unknown via type_map")
	}
}

func TestGenerateSliceFields(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	if !strings.Contains(output, "  posts: Post[];") {
		t.Error("expected posts: Post[]")
	}
	if !strings.Contains(output, "  tags: string[];") {
		t.Error("expected tags: string[]")
	}
	if !strings.Contains(output, "  scores: number[];") {
		t.Error("expected scores: number[]")
	}
}

func TestGenerateUnexportedFieldsSkipped(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export interface Config {") {
		t.Error("expected Config interface")
	}
	if !strings.Contains(output, "  database_url: string;") {
		t.Error("expected exported database_url field")
	}
	// Unexported fields should NOT appear
	if strings.Contains(output, "debug") {
		t.Error("unexported field 'debug' should not appear in output")
	}
	if strings.Contains(output, "internal") {
		t.Error("unexported field 'internal' should not appear in output")
	}
}

func TestGenerateNoTagUsesGoName(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export interface NoTagStruct {") {
		t.Error("expected NoTagStruct interface")
	}
	// No json tag → use Go field name as-is
	if !strings.Contains(output, "  Name: string;") {
		t.Error("expected 'Name' (Go field name) when no tag")
	}
	if !strings.Contains(output, "  Age: number;") {
		t.Error("expected 'Age' when no tag")
	}
	if !strings.Contains(output, "  IsAdmin: boolean;") {
		t.Error("expected 'IsAdmin' when no tag")
	}
}

func TestGenerateInterfaceFieldUnknown(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.EmptyInterface = "unknown"

	output := generate(t, cfg, result)

	if !strings.Contains(output, "  data: unknown;") {
		t.Error("expected interface{} mapped to unknown")
	}
}

func TestGenerateInterfaceFieldAny(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.EmptyInterface = "any"

	output := generate(t, cfg, result)

	if !strings.Contains(output, "  data: any;") {
		t.Error("expected interface{} mapped to any when configured")
	}
}

func TestGenerateEmptyStruct(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()

	output := generate(t, cfg, result)

	// Empty struct should still generate an empty interface
	if !strings.Contains(output, "export interface Empty {\n}") {
		t.Error("expected empty interface for Empty struct")
	}
}

// --- Header tests ---

func TestGenerateHeader(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.Header = "// Custom header"

	output := generate(t, cfg, result)

	if !strings.HasPrefix(output, "// Custom header\n") {
		t.Error("expected output to start with custom header")
	}
}

// --- Include/exclude filter tests ---

func TestGenerateIncludeFilter(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.Sources[0].Include = []string{"MessageRole"}

	output := generate(t, cfg, result)

	if !strings.Contains(output, "MessageRole") {
		t.Error("expected MessageRole to be included")
	}
	if strings.Contains(output, "FuncToolType") {
		t.Error("FuncToolType should be excluded when not in include list")
	}
}

func TestGenerateExcludeFilter(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.Sources[0].Exclude = []string{"MessageRole"}

	output := generate(t, cfg, result)

	if strings.Contains(output, "export type MessageRole") && !strings.Contains(output, "export type MessageRole =") {
		// It's okay if MessageRole appears as a field type reference
	}
	// The type declaration itself should not be generated
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "export type MessageRole") || strings.HasPrefix(line, "export enum MessageRole") || strings.HasPrefix(line, "export const MessageRole") {
			t.Error("MessageRole type should be excluded")
		}
	}
}

func TestGenerateIncludeWildcard(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.Sources[0].Include = []string{"Message*"}

	output := generate(t, cfg, result)

	if !strings.Contains(output, "MessageRole") {
		t.Error("expected MessageRole to match 'Message*'")
	}
	if strings.Contains(output, "export type FuncToolType") {
		t.Error("FuncToolType should not match 'Message*'")
	}
}

func TestGenerateExcludeWildcard(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.Sources[0].Exclude = []string{"*Status"}

	output := generate(t, cfg, result)

	if strings.Contains(output, "export type ItemGroupSourceTrainingStatus") {
		t.Error("ItemGroupSourceTrainingStatus should be excluded by '*Status'")
	}
	// MessageRole should still be present
	if !strings.Contains(output, "MessageRole") {
		t.Error("MessageRole should not be excluded by '*Status'")
	}
}

// --- Prefix stripping tests ---

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		constName string
		typeName  string
		expected  string
	}{
		{"MessageRoleUser", "MessageRole", "User"},
		{"MessageRoleAssistant", "MessageRole", "Assistant"},
		{"MessageRoleSystem", "MessageRole", "System"},
		{"FuncToolTypeVector", "FuncToolType", "Vector"},
		{"Active", "Status", "Active"},              // no prefix match
		{"SomeOther", "S", "SomeOther"},              // remainder starts lowercase
		{"LockReadonly", "Lock", "Readonly"},
		{"ItemGroupSourceTrainingStatusPending", "ItemGroupSourceTrainingStatus", "Pending"},
	}

	for _, tt := range tests {
		got := stripPrefix(tt.constName, tt.typeName)
		if got != tt.expected {
			t.Errorf("stripPrefix(%q, %q) = %q, want %q", tt.constName, tt.typeName, got, tt.expected)
		}
	}
}

// --- matchPattern tests ---

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"MessageRole", "MessageRole", true},
		{"MessageRole", "FuncToolType", false},
		{"Message*", "MessageRole", true},
		{"Message*", "MessageType", true},
		{"Message*", "FuncToolType", false},
		{"*Status", "OrderStatus", true},
		{"*Status", "UserStatus", true},
		{"*Status", "MessageRole", false},
		{"*", "Anything", true},
	}

	for _, tt := range tests {
		got := matchPattern(tt.pattern, tt.name)
		if got != tt.want {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
		}
	}
}

// --- lookupTag tests ---

func TestLookupTag(t *testing.T) {
	tests := []struct {
		tag      string
		key      string
		expected string
	}{
		{"`json:\"name\"`", "json", "name"},
		{"`json:\"name,omitempty\"`", "json", "name,omitempty"},
		{"`json:\"-\"`", "json", "-"},
		{"`json:\"id\" yaml:\"ID\"`", "json", "id"},
		{"`json:\"id\" yaml:\"ID\"`", "yaml", "ID"},
		{"`db:\"user_id\"`", "json", ""},
		{"``", "json", ""},
		{"`json:\"created_at,omitempty\" db:\"created_at\"`", "db", "created_at"},
	}

	for _, tt := range tests {
		got := lookupTag(tt.tag, tt.key)
		if got != tt.expected {
			t.Errorf("lookupTag(%q, %q) = %q, want %q", tt.tag, tt.key, got, tt.expected)
		}
	}
}

// --- Style cascade tests ---

func TestStyleCascadeGlobalDefault(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.DefaultStyle = "enum"

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export enum MessageRole {") {
		t.Error("expected global default style=enum to apply")
	}
}

func TestStyleCascadeSourceOverridesGlobal(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.DefaultStyle = "union"
	cfg.Sources[0].Style = "enum"

	output := generate(t, cfg, result)

	if !strings.Contains(output, "export enum MessageRole {") {
		t.Error("expected source style=enum to override global union")
	}
}

func TestStyleCascadeOverrideOverridesSource(t *testing.T) {
	result := parseTestFile(t)
	cfg := defaultConfig()
	cfg.DefaultStyle = "enum"
	cfg.Sources[0].Style = "enum"
	cfg.Overrides = map[string]config.Override{
		"MessageRole": {Style: "union"},
	}

	output := generate(t, cfg, result)

	expected := `export type MessageRole = "user" | "assistant" | "system";`
	if !strings.Contains(output, expected) {
		t.Error("expected per-type override union to override source enum")
	}
}
