# Use Cases (Brainstorm)

All possible Go-to-TypeScript conversion patterns and their output formats. This is a design exploration document covering every imaginable pattern.

For the currently implemented conversions, see [README.md](../README.md#conversion-reference).

## References

Existing tools in this space:

- [tygo](https://github.com/gzuidhof/tygo) — Generate TypeScript from Go source files, config-driven, supports interfaces and enums
- [typescriptify-golang-structs](https://github.com/tkrajina/typescriptify-golang-structs) — Go structs to TypeScript interfaces, programmatic API
- [gotypes](https://github.com/newtoallofthis123/gotypes) — Lightweight Go to TypeScript type generator
- [guts](https://github.com/coder/guts/blob/main/convert.go) — Coder's Go to TypeScript converter, AST-based conversion logic
- [go2ts](https://github.com/StirlingMarketingGroup/go2ts) — Go types to TypeScript with support for complex types

---

## 1. Typed String Constants (Enum Pattern)

The most common Go pattern for string enums.

**Go input:**

```go
type MessageRole string

const (
    MessageRoleUser      MessageRole = "user"
    MessageRoleAssistant MessageRole = "assistant"
    MessageRoleSystem    MessageRole = "system"
)
```

**TS output formats:**

### A) Union Type

No runtime value. Lightweight.

```typescript
export type MessageRole = "user" | "assistant" | "system";
```

### B) Const Object + Type

Runtime lookup + type safety. Best of both worlds.

```typescript
export const MessageRole = {
  User: "user",
  Assistant: "assistant",
  System: "system",
} as const;

export type MessageRole = typeof MessageRole[keyof typeof MessageRole];
```

### C) Enum

Full TypeScript enum.

```typescript
export enum MessageRole {
  User = "user",
  Assistant = "assistant",
  System = "system",
}
```

### D) Values Array

Useful for dropdowns, validation, iteration.

```typescript
export const MessageRoleValues = ["user", "assistant", "system"] as const;
export type MessageRole = typeof MessageRoleValues[number];
```

---

## 2. Typed Int Constants (iota)

**Go input:**

```go
type LockType int

const (
    _ LockType = iota
    LockReadonly
    LockU
    LockD
)
```

**TS output formats:**

### A) Numeric Enum

```typescript
export enum LockType {
  Readonly = 1,
  U = 2,
  D = 3,
}
```

### B) Const Object

```typescript
export const LockType = {
  Readonly: 1,
  U: 2,
  D: 3,
} as const;

export type LockType = typeof LockType[keyof typeof LockType];
```

### C) Union

```typescript
export type LockType = 1 | 2 | 3;
```

---

## 3. Iota with Expressions

**Go input:**

```go
type ByteSize int64

const (
    KB ByteSize = 1 << (10 * (iota + 1))
    MB
    GB
    TB
)
```

**TS output:**

```typescript
export const ByteSize = {
  KB: 1024,
  MB: 1048576,
  GB: 1073741824,
  TB: 1099511627776,
} as const;

export type ByteSize = typeof ByteSize[keyof typeof ByteSize];
```

Values are resolved at generation time. The TS output contains computed literals, not expressions.

---

## 4. Iota with Gaps / Explicit Values

**Go input:**

```go
type Priority int

const (
    PriorityLow    Priority = iota + 1 // 1
    PriorityMedium                      // 2
    _                                   // skip 3
    PriorityHigh                        // 4
    PriorityCritical Priority = 10
)
```

**TS output:**

```typescript
export enum Priority {
  Low = 1,
  Medium = 2,
  High = 4,
  Critical = 10,
}
```

---

## 5. Untyped String Constants (Loose Groups)

Constants without an explicit Go type. Requires config to decide grouping.

**Go input:**

```go
const (
    WorkspaceSubStatusActive   = "active"
    WorkspaceSubStatusInactive = "inactive"
)
```

**TS output formats:**

### A) Individual Exports

```typescript
export const WorkspaceSubStatusActive = "active";
export const WorkspaceSubStatusInactive = "inactive";
```

### B) Grouped Object (config provides the group name)

```typescript
export const WorkspaceSubStatus = {
  Active: "active",
  Inactive: "inactive",
} as const;
```

### C) Union (config provides the type name)

```typescript
export type WorkspaceSubStatus = "active" | "inactive";
```

---

## 6. Untyped Int Constants

**Go input:**

```go
const (
    MaxRetries     = 3
    DefaultTimeout = 30
    MaxPageSize    = 100
)
```

**TS output:**

```typescript
export const MaxRetries = 3;
export const DefaultTimeout = 30;
export const MaxPageSize = 100;
```

---

## 7. Untyped Float Constants

**Go input:**

```go
const (
    Pi      = 3.14159
    Epsilon = 0.0001
)
```

**TS output:**

```typescript
export const Pi = 3.14159;
export const Epsilon = 0.0001;
```

---

## 8. Untyped Bool Constants

**Go input:**

```go
const (
    FeatureEnabled  = true
    DebugMode       = false
)
```

**TS output:**

```typescript
export const FeatureEnabled = true;
export const DebugMode = false;
```

---

## 9. Standalone Constants

**Go input:**

```go
const DefaultAIModelID = "openai_gpt_4"
const DefaultVectorDimensions = 768
```

**TS output:**

```typescript
export const DefaultAIModelID = "openai_gpt_4";
export const DefaultVectorDimensions = 768;
```

---

## 10. Multi-line String Constants

**Go input:**

```go
const DefaultAIInstruction = `You are a helpful assistant.
You are knowledgeable, friendly, and always try to provide accurate information.`
```

**TS output:**

```typescript
export const DefaultAIInstruction = `You are a helpful assistant.
You are knowledgeable, friendly, and always try to provide accurate information.`;
```

---

## 11. Mixed Const Blocks (Typed + Untyped)

A single const block containing both typed and untyped constants.

**Go input:**

```go
const (
    PrefixPostType   = "a"
    PrefixPostStatus = "b"

    SiteCheckURL        = "url"
    SiteCheckHttpStatus = "httpstatus"

    SiteStatusActive = "active"
)
```

**TS output formats:**

### A) Individual Exports (default)

```typescript
export const PrefixPostType = "a";
export const PrefixPostStatus = "b";
export const SiteCheckURL = "url";
export const SiteCheckHttpStatus = "httpstatus";
export const SiteStatusActive = "active";
```

### B) Config-driven Grouping

Config can split them into logical groups:

```typescript
export const Prefix = {
  PostType: "a",
  PostStatus: "b",
} as const;

export const SiteCheck = {
  URL: "url",
  HttpStatus: "httpstatus",
} as const;

export const SiteStatus = {
  Active: "active",
} as const;
```

---

## 12. Structs — Basic

**Go input:**

```go
type User struct {
    ID        int64           `json:"id"`
    Name      string          `json:"name"`
    Email     *string         `json:"email,omitempty"`
    Meta      json.RawMessage `json:"meta"`
    CreatedAt time.Time       `json:"created_at"`
    password  string          // unexported
}
```

**TS output formats:**

### A) Interface

```typescript
export interface User {
  id: number;
  name: string;
  email?: string;
  meta: unknown;
  created_at: string;
}
```

### B) Type Alias

```typescript
export type User = {
  id: number;
  name: string;
  email?: string;
  meta: unknown;
  created_at: string;
};
```

**Configurable behaviors:**

- `omitempty` tag → optional field (`?`) or not
- Pointer fields → `optional` (`?`), `nullable` (`| null`), or `both` (`?: string | null`)
- Unexported fields → skip (default) or include
- Which struct tag to read for field names (`json`, `yaml`, `db`, custom)
- No struct tag → use Go field name as-is, or convert to camelCase

---

## 13. Structs — Nested / Embedded

**Go input:**

```go
type BaseModel struct {
    ID        int64     `json:"id"`
    CreatedAt time.Time `json:"created_at"`
}

type Post struct {
    BaseModel
    Title string `json:"title"`
}
```

**TS output formats:**

### A) Flattened

```typescript
export interface Post {
  id: number;
  created_at: string;
  title: string;
}
```

### B) Intersection

```typescript
export interface BaseModel {
  id: number;
  created_at: string;
}

export type Post = BaseModel & {
  title: string;
};
```

### C) Extends

```typescript
export interface Post extends BaseModel {
  title: string;
}
```

---

## 14. Structs — Multiple Embedding

**Go input:**

```go
type Timestamps struct {
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type SoftDelete struct {
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type Post struct {
    Timestamps
    SoftDelete
    ID    int64  `json:"id"`
    Title string `json:"title"`
}
```

**TS output formats:**

### A) Flattened

```typescript
export interface Post {
  created_at: string;
  updated_at: string;
  deleted_at?: string;
  id: number;
  title: string;
}
```

### B) Multiple Extends

```typescript
export interface Post extends Timestamps, SoftDelete {
  id: number;
  title: string;
}
```

### C) Multiple Intersection

```typescript
export type Post = Timestamps & SoftDelete & {
  id: number;
  title: string;
};
```

---

## 15. Structs — Nested Struct (Not Embedded)

**Go input:**

```go
type Post struct {
    ID     int64  `json:"id"`
    Title  string `json:"title"`
    Author User   `json:"author"`
}
```

**TS output:**

```typescript
export interface Post {
  id: number;
  title: string;
  author: User;
}
```

Referenced types must also be generated or already exist in the output.

---

## 16. Structs — Inline Anonymous Struct

**Go input:**

```go
type Response struct {
    Data struct {
        ID   int64  `json:"id"`
        Name string `json:"name"`
    } `json:"data"`
    Error string `json:"error"`
}
```

**TS output formats:**

### A) Inline

```typescript
export interface Response {
  data: {
    id: number;
    name: string;
  };
  error: string;
}
```

### B) Extracted

```typescript
export interface ResponseData {
  id: number;
  name: string;
}

export interface Response {
  data: ResponseData;
  error: string;
}
```

---

## 17. Structs — json:"-" (Ignored Fields)

**Go input:**

```go
type User struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    Password string `json:"-"`
    internal string
}
```

**TS output:**

```typescript
export interface User {
  id: number;
  name: string;
  // Password: skipped (json:"-")
  // internal: skipped (unexported)
}
```

Fields with `json:"-"` and unexported fields are excluded.

---

## 18. Structs — With Struct Tags Only (No json tag)

**Go input:**

```go
type Config struct {
    DatabaseURL string `yaml:"database_url" env:"DATABASE_URL"`
    Port        int    `yaml:"port" env:"PORT"`
    Debug       bool
}
```

**TS output (using yaml tag):**

```typescript
export interface Config {
  database_url: string;
  port: number;
  Debug: boolean; // no tag, uses Go field name
}
```

Config decides which struct tag to read. Falls back to Go field name if no tag found.

---

## 19. Structs — Generic (Go 1.18+)

**Go input:**

```go
type PaginatedResponse[T any] struct {
    Data       []T `json:"data"`
    Total      int `json:"total"`
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    TotalPages int `json:"total_pages"`
}

type APIResponse[T any] struct {
    Success bool   `json:"success"`
    Data    T      `json:"data"`
    Error   string `json:"error,omitempty"`
}
```

**TS output:**

```typescript
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface APIResponse<T> {
  success: boolean;
  data: T;
  error?: string;
}
```

---

## 20. Structs — Multiple Generic Parameters

**Go input:**

```go
type Result[T any, E any] struct {
    Data  T `json:"data"`
    Error E `json:"error"`
}

type KeyValue[K comparable, V any] struct {
    Key   K `json:"key"`
    Value V `json:"value"`
}
```

**TS output:**

```typescript
export interface Result<T, E> {
  data: T;
  error: E;
}

export interface KeyValue<K extends string | number, V> {
  key: K;
  value: V;
}
```

`comparable` maps to `string | number` (or `string | number | boolean` — configurable).

---

## 21. Structs — With Validation Tags

**Go input:**

```go
type CreatePostRequest struct {
    Title   string   `json:"title" validate:"required,min=3,max=200"`
    Content string   `json:"content" validate:"required,min=10"`
    Tags    []string `json:"tags" validate:"max=5,dive,max=30"`
    Email   string   `json:"email" validate:"required,email"`
    URL     string   `json:"url" validate:"omitempty,url"`
    Age     int      `json:"age" validate:"gte=0,lte=150"`
}
```

**TS output formats:**

### A) Interface Only

```typescript
export interface CreatePostRequest {
  title: string;
  content: string;
  tags: string[];
  email: string;
  url: string;
  age: number;
}
```

### B) Zod Schema

```typescript
export const CreatePostRequestSchema = z.object({
  title: z.string().min(3).max(200),
  content: z.string().min(10),
  tags: z.array(z.string().max(30)).max(5),
  email: z.string().email(),
  url: z.string().url().optional(),
  age: z.number().int().gte(0).lte(150),
});

export type CreatePostRequest = z.infer<typeof CreatePostRequestSchema>;
```

### C) Yup Schema

```typescript
export const CreatePostRequestSchema = yup.object({
  title: yup.string().required().min(3).max(200),
  content: yup.string().required().min(10),
  tags: yup.array().of(yup.string().max(30)).max(5),
  email: yup.string().required().email(),
  url: yup.string().url(),
  age: yup.number().integer().min(0).max(150),
});
```

---

## 22. Slices / Arrays

**Go input:**

```go
type Tags []string
type Matrix [][]int
type UserList []*User
```

**TS output:**

```typescript
export type Tags = string[];
export type Matrix = number[][];
export type UserList = User[];
```

Pointer elements (`*User`) collapse to the base type since TS has no pointer concept.

---

## 23. Maps

**Go input:**

```go
type Metadata map[string]any
type Scores map[string]int
type UserCache map[int64]*User
type NestedMap map[string]map[string][]int
```

**TS output:**

```typescript
export type Metadata = Record<string, unknown>;
export type Scores = Record<string, number>;
export type UserCache = Record<number, User>;
export type NestedMap = Record<string, Record<string, number[]>>;
```

---

## 24. Type Aliases / Named Primitives

**Go input:**

```go
type UserID int64
type Slug string
type Amount float64
type Flag bool
```

**TS output formats:**

### A) Plain Alias

```typescript
export type UserID = number;
export type Slug = string;
export type Amount = number;
export type Flag = boolean;
```

### B) Branded Types (prevents mixing UserID with other numbers)

```typescript
export type UserID = number & { readonly __brand: "UserID" };
export type Slug = string & { readonly __brand: "Slug" };
export type Amount = number & { readonly __brand: "Amount" };
```

---

## 25. Composite Type Aliases

**Go input:**

```go
type StringSlice []string
type IntMap map[string]int
type UserMap map[string]User
type Matrix [][]float64
```

**TS output:**

```typescript
export type StringSlice = string[];
export type IntMap = Record<string, number>;
export type UserMap = Record<string, User>;
export type Matrix = number[][];
```

---

## 26. Channel Types (Skip or Warn)

**Go input:**

```go
type EventStream chan Event
type Subscriber chan<- string
```

Channels have no TypeScript equivalent. Options:
- Skip silently
- Emit a comment: `// EventStream: skipped (Go channel type)`
- Emit as the element type: `export type EventStream = Event`

---

## 27. Function Types

**Go input:**

```go
type Handler func(ctx context.Context, req Request) (Response, error)
type Callback func(event Event)
type Predicate func(item Item) bool
type Middleware func(next Handler) Handler
```

**TS output:**

```typescript
export type Handler = (req: Request) => Response;
export type Callback = (event: Event) => void;
export type Predicate = (item: Item) => boolean;
export type Middleware = (next: Handler) => Handler;
```

- `context.Context` is stripped (no TS equivalent)
- `error` return is stripped (TS uses exceptions/Result pattern)
- `bool` return stays
- No return value → `void`

---

## 28. Function Types — Multiple Return Values

**Go input:**

```go
type Parser func(input string) (result AST, remaining string, err error)
```

**TS output formats:**

### A) First Non-Error Return

```typescript
export type Parser = (input: string) => AST;
```

### B) Tuple

```typescript
export type Parser = (input: string) => [AST, string];
```

### C) Object

```typescript
export type Parser = (input: string) => { result: AST; remaining: string };
```

---

## 29. Interface Types

**Go input:**

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Stringer interface {
    String() string
}

type ReadWriter interface {
    Reader
    Writer
}
```

**TS output:**

```typescript
export interface Reader {
  read(p: Uint8Array): number;
}

export interface Stringer {
  string(): string;
}

export interface ReadWriter extends Reader, Writer {}
```

---

## 30. Interface Types — Empty Interface

**Go input:**

```go
type Any interface{}

// or in a struct:
type Event struct {
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}
```

**TS output:**

```typescript
export type Any = unknown;

export interface Event {
  type: string;
  payload: unknown;
}
```

`interface{}` / `any` → `unknown` (safer) or `any` (configurable).

---

## 31. Pointer Types in Structs

**Go input:**

```go
type User struct {
    Name      string     `json:"name"`
    Nickname  *string    `json:"nickname"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
    Profile   *Profile   `json:"profile"`
}
```

**TS output formats:**

### A) Optional (pointer = might not be present)

```typescript
export interface User {
  name: string;
  nickname?: string;
  deleted_at?: string;
  profile?: Profile;
}
```

### B) Nullable (pointer = explicitly null)

```typescript
export interface User {
  name: string;
  nickname: string | null;
  deleted_at: string | null;
  profile: Profile | null;
}
```

### C) Both (pointer + omitempty = optional and nullable)

```typescript
export interface User {
  name: string;
  nickname: string | null;
  deleted_at?: string | null; // pointer + omitempty
  profile: Profile | null;
}
```

---

## 32. Struct Field with Custom Types

**Go input:**

```go
type Post struct {
    ID        uuid.UUID       `json:"id"`
    Amount    decimal.Decimal `json:"amount"`
    Tags      pq.StringArray  `json:"tags"`
    Meta      datatypes.JSON  `json:"meta"`
    IP        net.IP          `json:"ip"`
    Duration  time.Duration   `json:"duration"`
}
```

**TS output (with type map from config):**

```typescript
// type_map:
//   uuid.UUID: string
//   decimal.Decimal: string  (or number)
//   pq.StringArray: string[]
//   datatypes.JSON: unknown
//   net.IP: string
//   time.Duration: number

export interface Post {
  id: string;
  amount: string;
  tags: string[];
  meta: unknown;
  ip: string;
  duration: number;
}
```

---

## 33. Struct with json.Marshaler / Custom Serialization

**Go input:**

```go
type Color struct {
    R, G, B uint8
}

// Has MarshalJSON — serializes as "#RRGGBB"
func (c Color) MarshalJSON() ([]byte, error) { ... }
```

Config can override: if a type implements `json.Marshaler`, the user specifies what the JSON output actually is:

```yaml
type_map:
  Color: string  # because MarshalJSON outputs a string
```

**TS output:**

```typescript
// Color implements custom JSON marshaling — mapped via config
export type Color = string;
```

---

## 34. Struct with json.RawMessage Fields

**Go input:**

```go
type Webhook struct {
    Event   string          `json:"event"`
    Payload json.RawMessage `json:"payload"`
}
```

**TS output formats:**

### A) unknown (safe default)

```typescript
export interface Webhook {
  event: string;
  payload: unknown;
}
```

### B) Generic (if you want the caller to narrow)

```typescript
export interface Webhook<T = unknown> {
  event: string;
  payload: T;
}
```

---

## 35. Enum with Methods (Stringer)

**Go input:**

```go
type Status int

const (
    StatusActive Status = iota
    StatusInactive
    StatusBanned
)

func (s Status) String() string {
    return [...]string{"active", "inactive", "banned"}[s]
}
```

**TS output:**

If detected, use the String() return values instead of raw ints:

```typescript
export type Status = "active" | "inactive" | "banned";

// or with both:
export const Status = {
  Active: { value: 0, label: "active" },
  Inactive: { value: 1, label: "inactive" },
  Banned: { value: 2, label: "banned" },
} as const;
```

---

## 36. Discriminated Union from Interface Implementations

**Go input:**

```go
type Shape interface {
    Area() float64
}

type Circle struct {
    Type   string  `json:"type"` // "circle"
    Radius float64 `json:"radius"`
}

type Rectangle struct {
    Type   string  `json:"type"` // "rectangle"
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
}
```

**TS output:**

```typescript
export interface Circle {
  type: "circle";
  radius: number;
}

export interface Rectangle {
  type: "rectangle";
  width: number;
  height: number;
}

export type Shape = Circle | Rectangle;
```

Config specifies which structs implement the interface and which field is the discriminator.

---

## 37. Error Variables → Error Code Registry

**Go input:**

```go
var (
    ErrNotFound      = NewAppError("NOT_FOUND", 404, "Resource not found")
    ErrUnauthorized  = NewAppError("UNAUTHORIZED", 401, "Not authorized")
    ErrQuotaExceeded = NewAppError("QUOTA_EXCEEDED", 429, "Plan quota exceeded")
)
```

**TS output:**

```typescript
export type AppErrorCode = "NOT_FOUND" | "UNAUTHORIZED" | "QUOTA_EXCEEDED";

export const AppErrors = {
  NOT_FOUND: { code: "NOT_FOUND", status: 404, message: "Resource not found" },
  UNAUTHORIZED: { code: "UNAUTHORIZED", status: 401, message: "Not authorized" },
  QUOTA_EXCEEDED: { code: "QUOTA_EXCEEDED", status: 429, message: "Plan quota exceeded" },
} as const;
```

---

## 38. Variable Declarations (Non-Error)

**Go input:**

```go
var (
    DefaultHeaders = map[string]string{
        "Content-Type": "application/json",
        "Accept":       "application/json",
    }

    SupportedLanguages = []string{"en", "es", "fr", "de", "ja"}
)
```

**TS output:**

```typescript
export const DefaultHeaders = {
  "Content-Type": "application/json",
  Accept: "application/json",
} as const;

export const SupportedLanguages = ["en", "es", "fr", "de", "ja"] as const;
```

---

## 39. Map Literals / Lookup Tables

**Go input:**

```go
var StatusLabels = map[Status]string{
    StatusActive:   "Active",
    StatusInactive: "Inactive",
    StatusBanned:   "Banned",
}

var RolePermissions = map[string][]string{
    "owner":  {"read", "write", "admin"},
    "member": {"read", "write"},
    "viewer": {"read"},
}
```

**TS output:**

```typescript
export const StatusLabels: Record<Status, string> = {
  active: "Active",
  inactive: "Inactive",
  banned: "Banned",
};

export const RolePermissions = {
  owner: ["read", "write", "admin"],
  member: ["read", "write"],
  viewer: ["read"],
} as const;
```

---

## 40. State Machine / Transition Map

**Go input:**

```go
var OrderTransitions = map[OrderStatus][]OrderStatus{
    OrderPending: {OrderPaid, OrderCancelled},
    OrderPaid:    {OrderShipped, OrderCancelled},
    OrderShipped: {OrderDelivered},
}
```

**TS output:**

```typescript
export const OrderTransitions: Record<OrderStatus, OrderStatus[]> = {
  pending: ["paid", "cancelled"],
  paid: ["shipped", "cancelled"],
  shipped: ["delivered"],
  delivered: [],
  cancelled: [],
};
```

---

## 41. Struct Slice Types (API List Responses)

**Go input:**

```go
type PostListResponse struct {
    Posts      []Post `json:"posts"`
    TotalCount int    `json:"total_count"`
    NextCursor string `json:"next_cursor,omitempty"`
}
```

**TS output:**

```typescript
export interface PostListResponse {
  posts: Post[];
  total_count: number;
  next_cursor?: string;
}
```

---

## 42. Recursive / Self-Referencing Types

**Go input:**

```go
type TreeNode struct {
    ID       string     `json:"id"`
    Value    string     `json:"value"`
    Children []TreeNode `json:"children"`
    Parent   *TreeNode  `json:"parent,omitempty"`
}

type Comment struct {
    ID      int64     `json:"id"`
    Body    string    `json:"body"`
    Replies []Comment `json:"replies"`
}
```

**TS output:**

```typescript
export interface TreeNode {
  id: string;
  value: string;
  children: TreeNode[];
  parent?: TreeNode;
}

export interface Comment {
  id: number;
  body: string;
  replies: Comment[];
}
```

---

## 43. Struct with Time Fields (Multiple Format Strategies)

**Go input:**

```go
type Event struct {
    StartAt   time.Time  `json:"start_at"`
    EndAt     *time.Time `json:"end_at,omitempty"`
    Duration  time.Duration `json:"duration"`
}
```

**TS output formats:**

### A) All strings (ISO 8601 from JSON)

```typescript
export interface Event {
  start_at: string;
  end_at?: string;
  duration: number; // nanoseconds as number
}
```

### B) Date objects

```typescript
export interface Event {
  start_at: Date;
  end_at?: Date;
  duration: number;
}
```

### C) Dayjs / Luxon (configurable)

```typescript
import { Dayjs } from "dayjs";

export interface Event {
  start_at: Dayjs;
  end_at?: Dayjs;
  duration: number;
}
```

---

## 44. Struct with Byte Fields

**Go input:**

```go
type File struct {
    Name    string `json:"name"`
    Content []byte `json:"content"`
    Hash    [32]byte `json:"hash"`
}
```

**TS output:**

```typescript
export interface File {
  name: string;
  content: string;   // base64 encoded in JSON
  hash: string;      // base64 encoded in JSON
}
```

`[]byte` serializes as base64 string in JSON.

---

## 45. Constants from Multiple Packages

Different Go packages exporting related constants.

**Go:**

```go
// pkg/constants/roles.go
type Role string
const RoleAdmin Role = "admin"

// pkg/constants/permissions.go
type Permission string
const PermRead Permission = "read"

// pkg/billing/plans.go
type PlanTier string
const PlanFree PlanTier = "free"
```

**TS output (single file or multiple, config-driven):**

```typescript
// generated/constants.ts
export type Role = "admin" | "member";
export type Permission = "read" | "write";

// generated/billing.ts
export type PlanTier = "free" | "pro" | "enterprise";
```

---

## 46. Protobuf-generated Go Types

**Go input (generated by protoc):**

```go
type UserStatus int32

const (
    UserStatus_UNKNOWN  UserStatus = 0
    UserStatus_ACTIVE   UserStatus = 1
    UserStatus_INACTIVE UserStatus = 2
)
```

**TS output:**

```typescript
export enum UserStatus {
  Unknown = 0,
  Active = 1,
  Inactive = 2,
}
```

Prefix stripping removes `UserStatus_` automatically.

---

## 47. Bit Flag / Bitmask Constants

**Go input:**

```go
type Permission uint

const (
    PermRead    Permission = 1 << iota // 1
    PermWrite                           // 2
    PermDelete                          // 4
    PermAdmin                           // 8

    PermReadWrite = PermRead | PermWrite // 3
    PermAll       = PermRead | PermWrite | PermDelete | PermAdmin // 15
)
```

**TS output:**

```typescript
export const Permission = {
  Read: 1,
  Write: 2,
  Delete: 4,
  Admin: 8,
  ReadWrite: 3,
  All: 15,
} as const;

export type Permission = typeof Permission[keyof typeof Permission];
```

---

## 48. SQL Migration Enums

Scanning `.sql` files directly (not Go).

**SQL input:**

```sql
CREATE TYPE post_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE user_role AS ENUM ('admin', 'member', 'viewer');
```

**TS output:**

```typescript
export type PostStatus = "draft" | "published" | "archived";
export type UserRole = "admin" | "member" | "viewer";
```

---

## Primitive Type Mapping Reference

| Go Type | TS Type |
|---|---|
| `string` | `string` |
| `bool` | `boolean` |
| `int`, `int8`, `int16`, `int32`, `int64` | `number` |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | `number` |
| `float32`, `float64` | `number` |
| `byte` | `number` |
| `rune` | `number` |
| `complex64`, `complex128` | `{ real: number; imag: number }` or skip |
| `[]byte` | `string` (base64) |
| `interface{}` / `any` | `unknown` or `any` (configurable) |
| `error` | skip (TS uses exceptions) |
| `context.Context` | skip |
| `time.Time` | `string` or `Date` (configurable) |
| `time.Duration` | `number` |
| `*T` (pointer) | `T`, `T \| null`, or `T \| undefined` (configurable) |
| `[]T` | `T[]` |
| `[N]T` (array) | `T[]` (or `readonly [T, T, ...]` for small N) |
| `map[K]V` | `Record<K, V>` |
| `chan T` | skip |
| `func(...)` | `(...) => ...` |

---

## Config Features

| Feature | Description |
|---|---|
| Include/exclude patterns | `"Message*"`, `"*Status"`, exact names |
| Output style per type | `MessageRole: union`, `LockType: enum` |
| Global default style | `default_style: union` |
| Name transforms | Strip prefix: `MessageRoleUser` → key `User` |
| Custom type map | `time.Time: string`, `uuid.UUID: string`, `decimal.Decimal: number` |
| Multiple output files | Different sources write to different `.ts` files |
| Struct tag source | Which tag to read for field names (`json`, `db`, `graphql`) |
| Pointer handling | `optional`, `nullable`, or `both` |
| Group untyped consts | Assign a virtual type name to an untyped const block |
| Generated header | Auto `// Code generated by gots. DO NOT EDIT.` banner |
| Prefix stripping | Auto-detect or manual per type |
| Watch mode | Re-generate on file change |
| Values array | Optionally emit array of all values alongside the type |
| Validation output | `zod`, `yup`, `valibot`, or `none` |
| Empty interface mapping | `unknown` or `any` |
| Branded types | Enable per type or globally |
| Embed strategy | `flatten`, `extends`, or `intersection` |
| Generic mapping | Preserve Go generics as TS generics |
| SQL source | Scan `.sql` migration files for `CREATE TYPE ... ENUM` |

---

## Summary

### Daily use (implement first)

1. Typed string consts → union / const-object / enum
2. Typed int consts (iota) → numeric enum / const-object
3. Untyped const groups → individual exports or grouped object
4. Standalone consts → individual exports
5. Structs → interfaces / type aliases
6. Pointer / omitempty handling
7. Custom type mapping (time.Time, uuid.UUID, etc.)
8. Prefix stripping

### Common (implement second)

9. Embedded structs (flatten / extends / intersection)
10. Generic structs
11. Validation tag → Zod/Yup schemas
12. Slices, maps, composite type aliases
13. Nested/anonymous structs
14. json:"-" field skipping
15. Multi-line string consts
16. Variable declarations (maps, slices)
17. Error variable registries

### Advanced (implement later)

18. Recursive/self-referencing types
19. Discriminated unions from interface implementations
20. Bit flag / bitmask resolution
21. State machine / transition maps
22. Iota with expressions (bit shifts, arithmetic)
23. Branded types
24. Function types
25. Go interface types
26. Protobuf-generated type cleanup
27. SQL migration enum scanning
28. Multiple return value strategies
29. Custom JSON marshaler awareness
30. Channel types (skip/warn)
