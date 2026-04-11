package testdata

import (
	"encoding/json"
	"time"
)

// --- 1. Typed string constants (enum pattern) ---

// MessageRole represents the role of a message sender in a conversation.
type MessageRole string

const (
	// MessageRoleUser is the human user
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant" // AI assistant
	MessageRoleSystem    MessageRole = "system"
)

// --- 2. Typed string constants (single value) ---

type FuncToolType string

const (
	FuncToolTypeVector FuncToolType = "vector"
)

// --- 3. Typed string constants (many values, long prefix) ---

type ItemGroupSourceTrainingStatus string

const (
	ItemGroupSourceTrainingStatusPending  ItemGroupSourceTrainingStatus = "pending"
	ItemGroupSourceTrainingStatusChunking ItemGroupSourceTrainingStatus = "chunking"
	ItemGroupSourceTrainingStatusTraining ItemGroupSourceTrainingStatus = "training"
	ItemGroupSourceTrainingStatusCompleted ItemGroupSourceTrainingStatus = "completed"
	ItemGroupSourceTrainingStatusError    ItemGroupSourceTrainingStatus = "error"
)

// --- 4. Typed int constants (iota) ---

type LockType int

const (
	_ LockType = iota
	LockReadonly
	LockUpdate
	LockDelete
)

// --- 5. Typed int constants (iota, no skip) ---

type Priority int

const (
	PriorityLow    Priority = iota
	PriorityMedium
	PriorityHigh
)

// --- 6. Untyped string constants (loose group) ---

const (
	WorkspaceSubStatusActive   = "active"
	WorkspaceSubStatusInactive = "inactive"
)

// --- 7. Untyped string constants (mixed group) ---

const (
	PrefixPostType   = "a"
	PrefixPostStatus = "b"
	SiteCheckURL     = "url"
	SiteStatusActive = "active"
)

// --- 8. Standalone string constant ---

const DefaultAIModelID = "openai_gpt_4"

// --- 9. Standalone int constant ---

const DefaultVectorDimensions = 768

// --- 10. Standalone bool constants ---

const FeatureEnabled = true
const DebugMode = false

// --- 11. Multi-line string constant (backtick) ---

const DefaultAIInstruction = `You are a helpful assistant.
You are knowledgeable and friendly.`

// --- 12. Header constants ---

const (
	HeaderWorkspaceId = "x-workspace-id"
	HeaderRequestId   = "x-request-id"
)

// --- 13. Basic struct ---

// User represents a registered user in the system.
type User struct {
	// ID is the unique identifier
	ID    int64  `json:"id"`
	// Name is the user's display name
	Name  string `json:"name"`
	Email string `json:"email"` // primary email address
}

// --- 14. Struct with pointer fields ---

type Post struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Content   *string `json:"content,omitempty"`
	Author    *User   `json:"author,omitempty"`
}

// --- 15. Struct with various Go types ---

type FullModel struct {
	StringField    string          `json:"string_field"`
	IntField       int             `json:"int_field"`
	Int8Field      int8            `json:"int8_field"`
	Int16Field     int16           `json:"int16_field"`
	Int32Field     int32           `json:"int32_field"`
	Int64Field     int64           `json:"int64_field"`
	UintField      uint            `json:"uint_field"`
	Uint8Field     uint8           `json:"uint8_field"`
	Uint16Field    uint16          `json:"uint16_field"`
	Uint32Field    uint32          `json:"uint32_field"`
	Uint64Field    uint64          `json:"uint64_field"`
	Float32Field   float32         `json:"float32_field"`
	Float64Field   float64         `json:"float64_field"`
	BoolField      bool            `json:"bool_field"`
	ByteField      byte            `json:"byte_field"`
	RuneField      rune            `json:"rune_field"`
}

// --- 16. Struct with external package types ---

type Event struct {
	ID        string          `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

// --- 17. Struct with slices ---

type PostList struct {
	Posts    []Post   `json:"posts"`
	Tags     []string `json:"tags"`
	Scores   []int    `json:"scores"`
}

// --- 18. Struct with embedded struct ---

type BaseModel struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type Article struct {
	BaseModel
	Title   string `json:"title"`
	Content string `json:"content"`
}

// --- 18b. Struct with multiple embedding ---

type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SoftDelete struct {
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type BlogPost struct {
	Timestamps
	SoftDelete
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// --- 19. Struct with unexported fields ---

type Config struct {
	DatabaseURL string `json:"database_url"`
	Port        int    `json:"port"`
	debug       bool
	internal    string
}

// --- 20. Struct with json:"-" (hidden fields) ---

type SecureUser struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Password string `json:"-"`
	Token    string `json:"-"`
}

// --- 21. Struct with no json tags ---

type NoTagStruct struct {
	Name    string
	Age     int
	IsAdmin bool
}

// --- 22. Struct with omitempty ---

type OptionalFields struct {
	Required string  `json:"required"`
	Optional *string `json:"optional,omitempty"`
	AlsoOpt  *int    `json:"also_opt,omitempty"`
}

// --- 23. Struct with interface{} / any field ---

type DynamicPayload struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// --- 24. Struct with map field ---

type Metadata struct {
	Labels map[string]string `json:"labels"`
}

// --- 25. Named type aliases ---

type UserID int64
type Slug string
type Amount float64
type Flag bool

// --- 26. Named slice type ---

type Tags []string

// --- 27. Named map type ---

type StringMap map[string]string

// --- 28. Empty struct ---

type Empty struct{}

// --- 29. Struct with single field ---

type Wrapper struct {
	Value string `json:"value"`
}

// --- 30. Struct with uintptr field ---

type PointerInfo struct {
	Addr uintptr `json:"addr"`
}

// --- 31. Struct with complex types ---

type ComplexNumbers struct {
	C64  complex64  `json:"c64"`
	C128 complex128 `json:"c128"`
}

// --- 32. Struct with fixed-size array ---

type Matrix struct {
	Row [4]float64 `json:"row"`
}

// --- 33. Struct with channel fields ---

type Worker struct {
	Name   string         `json:"name"`
	Input  chan string     `json:"input"`
	Output chan<- string   `json:"output"`
	Signal <-chan struct{} `json:"signal"`
}

// --- 34. Function types ---

type Handler func(request string) (response string, err error)
type Callback func()
type Predicate func(item int) bool
type Middleware func(next Handler) Handler
type MultiReturn func(a, b int) (int, int, error)

// --- 35. Interface types ---

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type ReadWriter interface {
	Reader
	Writer
}

type Stringer interface {
	String() string
}

type EmptyInterface interface{}

// --- 36. Type aliases (= syntax) ---

type Name = string
type Age = int
type Table = map[string]int
type Bytes = []byte

// --- 37. Nested composite types ---

type NestedTypes struct {
	Matrix      [][]int                `json:"matrix"`
	SliceOfMaps []map[string]int       `json:"slice_of_maps"`
	MapOfSlices map[string][]int       `json:"map_of_slices"`
	MapOfMaps   map[string]map[string]string `json:"map_of_maps"`
	PtrToSlice  *[]string              `json:"ptr_to_slice"`
	PtrToPtr    **int                  `json:"ptr_to_ptr"`
	SliceOfPtrs []*User                `json:"slice_of_ptrs"`
}

// --- 38. Generic struct (single type param) ---

type Box[T any] struct {
	Value T `json:"value"`
}

// --- 39. Generic struct (multiple type params) ---

type Pair[K comparable, V any] struct {
	Key   K `json:"key"`
	Value V `json:"value"`
}

// --- 40. Generic struct with constraint ---

type NumberBox[T int | int64 | float64] struct {
	Value T    `json:"value"`
	Min   T    `json:"min"`
	Max   T    `json:"max"`
}

// --- 41. Generic paginated response ---

type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
}

// --- 42. Generic result type ---

type Result[T any, E any] struct {
	Data  *T     `json:"data,omitempty"`
	Error *E     `json:"error,omitempty"`
	Ok    bool   `json:"ok"`
}

// --- 43. Generic with nested generics ---

type Cache[K comparable, V any] struct {
	Items map[K]V `json:"items"`
	TTL   int     `json:"ttl"`
}

// --- 44. Struct with any field (Go 1.18+) ---

type AnyPayload struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// --- 45. Struct with multi-name fields (X, Y float64) ---

type Point struct {
	X, Y float64
}

type Rect struct {
	Left, Top, Right, Bottom int
}

// --- 46. Struct with array of structs ---

type Team struct {
	Name    string    `json:"name"`
	Members [5]User   `json:"members"`
	Scores  [3]int    `json:"scores"`
}

// --- 47. Deeply nested struct ---

type APIResponse struct {
	Status string `json:"status"`
	Data   struct {
		User    User   `json:"user"`
		Token   string `json:"token"`
		Expires int64  `json:"expires"`
	} `json:"data"`
}

// --- 48. Struct with function field ---

type Hooks struct {
	OnCreate func(id int64)         `json:"-"`
	OnUpdate func(id int64, data any) `json:"-"`
}

// --- 49. Named channel types ---

type EventStream chan Event
type SendOnlyStream chan<- string
type ReceiveOnlyStream <-chan int

// --- 50. Named function types ---

type Reducer func(accumulator, current int) int
type ErrorHandler func(err error)
type Comparator[T any] func(a, b T) int

// --- 51. Named array type ---

type IPv4 [4]byte
type Matrix3x3 [3][3]float64

// --- 52. Named interface types with generics ---

type Container[T any] interface {
	Get() T
	Set(value T)
	Len() int
}

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

// --- 53. Type alias with underlying composite ---

type JSONMap = map[string]interface{}
type HandlerFunc = func(string) error

// Keep these to prevent unused import errors
var _ time.Time
var _ json.RawMessage
