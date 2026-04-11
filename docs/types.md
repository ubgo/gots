# Go Type Coverage

All 119 Go types recognized by gots, organized by Go's type system categories. Tracks which types are parsed and which generate correct TypeScript output.

For the conversion reference with Go→TypeScript examples, see [README.md](../README.md#conversion-reference).

---

## Built-in Basic Types (17)

All 17 predeclared basic types in Go.

| # | Go Type | TS Output | Category | Tested |
|---|---|---|---|---|
| 1 | `string` | `string` | string | yes |
| 2 | `bool` | `boolean` | boolean | yes |
| 3 | `int` | `number` | integer | yes |
| 4 | `int8` | `number` | integer | yes |
| 5 | `int16` | `number` | integer | yes |
| 6 | `int32` | `number` | integer | yes |
| 7 | `int64` | `number` | integer | yes |
| 8 | `uint` | `number` | unsigned integer | yes |
| 9 | `uint8` | `number` | unsigned integer | yes |
| 10 | `uint16` | `number` | unsigned integer | yes |
| 11 | `uint32` | `number` | unsigned integer | yes |
| 12 | `uint64` | `number` | unsigned integer | yes |
| 13 | `float32` | `number` | float | yes |
| 14 | `float64` | `number` | float | yes |
| 15 | `byte` | `number` | alias for uint8 | yes |
| 16 | `rune` | `number` | alias for int32 | yes |
| 17 | `uintptr` | `number` | pointer-sized integer | yes |

## Built-in Types Not Mappable to TS

| # | Go Type | TS Output | Notes | Tested |
|---|---|---|---|---|
| 18 | `complex64` | skip or `{ real: number; imag: number }` | no JSON equivalent | yes (parsed) |
| 19 | `complex128` | skip or `{ real: number; imag: number }` | no JSON equivalent | yes (parsed) |

---

## Composite Types (8)

All composite type forms in Go.

### Pointer Types

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 20 | `*string` | `string` (optional/nullable) | yes |
| 21 | `*User` | `User` (optional/nullable) | yes |
| 22 | `*time.Time` | `string` (optional/nullable, via type map) | yes |
| 23 | `**int` | `number` (pointer to pointer) | yes (parsed) |

### Struct Types

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 24 | Basic struct | `export interface Foo { ... }` | yes |
| 25 | Struct with pointer fields | `field?: T` or `field: T \| null` | yes |
| 26 | Struct with embedded struct | flatten / extends / intersection | yes (parsed) |
| 27 | Struct with multiple embedding | multiple extends | yes (parsed) |
| 28 | Struct with unexported fields | skipped in output | yes |
| 29 | Struct with `json:"-"` fields | skipped in output | yes |
| 30 | Struct with no json tags | Go field name used as-is | yes |
| 31 | Struct with `omitempty` | `field?: T` | yes |
| 32 | Struct with `interface{}` field | `unknown` or `any` | yes |
| 33 | Struct with `any` field (Go 1.18+) | `unknown` or `any` | yes |
| 34 | Struct with map field | `Record<K, V>` | yes (parsed) |
| 35 | Struct with function field | skipped (`json:"-"` typically) | yes (parsed) |
| 36 | Struct with anonymous inline struct | nested or extracted | yes (parsed) |
| 37 | Struct with multi-name fields (`X, Y float64`) | two separate TS fields | yes |
| 38 | Empty struct | `export interface Foo {}` | yes |
| 39 | Single field struct | `export interface Foo { value: T }` | yes |

### Function Types

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 40 | `func(string) (string, error)` | `(arg: string) => string` | yes (parsed) |
| 41 | `func()` | `() => void` | yes (parsed) |
| 42 | `func(item int) bool` | `(item: number) => boolean` | yes (parsed) |
| 43 | `func(next Handler) Handler` | `(next: Handler) => Handler` | yes (parsed) |
| 44 | `func(a, b int) (int, int, error)` | multiple return strategies | yes (parsed) |

### Array Types

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 45 | `[4]float64` | `number[]` | yes |
| 46 | `[5]User` | `User[]` | yes |
| 47 | `[3]int` | `number[]` | yes |
| 48 | `[3][3]float64` | `number[][]` | yes (parsed) |
| 49 | `[4]byte` | `number[]` | yes (parsed) |

### Slice Types

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 50 | `[]string` | `string[]` | yes |
| 51 | `[]int` | `number[]` | yes |
| 52 | `[]Post` | `Post[]` | yes |
| 53 | `[]*User` | `User[]` | yes (parsed) |
| 54 | `[][]int` | `number[][]` | yes (parsed) |
| 55 | `*[]string` | `string[]` (optional) | yes (parsed) |

### Map Types

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 56 | `map[string]string` | `Record<string, string>` | yes (parsed) |
| 57 | `map[string]int` | `Record<string, number>` | yes (parsed) |
| 58 | `map[string][]int` | `Record<string, number[]>` | yes (parsed) |
| 59 | `map[string]map[string]string` | `Record<string, Record<string, string>>` | yes (parsed) |
| 60 | `[]map[string]int` | `Record<string, number>[]` | yes (parsed) |

### Channel Types

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 61 | `chan T` | skip (no TS equivalent) | yes (parsed) |
| 62 | `chan<- T` | skip (send-only) | yes (parsed) |
| 63 | `<-chan T` | skip (receive-only) | yes (parsed) |

### Interface Types

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 64 | `interface{}` | `unknown` or `any` (configurable) | yes |
| 65 | `any` (Go 1.18+) | `unknown` or `any` (configurable) | yes |
| 66 | Interface with methods | `export interface Foo { method(): T }` | yes (parsed) |
| 67 | Interface composition | `export interface Foo extends Bar, Baz` | yes (parsed) |
| 68 | Empty interface type | `unknown` | yes (parsed) |

---

## Type Declaration Forms

### Type Definitions

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 69 | `type Foo string` | union / enum / const-object (if has consts) | yes |
| 70 | `type Foo int` | union / enum (if has iota consts) | yes |
| 71 | `type Foo struct { ... }` | `export interface Foo { ... }` | yes |
| 72 | `type UserID int64` | `export type UserID = number` | yes (parsed) |
| 73 | `type Slug string` | `export type Slug = string` | yes (parsed) |
| 74 | `type Amount float64` | `export type Amount = number` | yes (parsed) |
| 75 | `type Flag bool` | `export type Flag = boolean` | yes (parsed) |
| 76 | `type Tags []string` | `export type Tags = string[]` | yes (parsed) |
| 77 | `type StringMap map[string]string` | `export type StringMap = Record<string, string>` | yes (parsed) |

### Type Aliases (Go 1.9+)

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 78 | `type Name = string` | `export type Name = string` | yes (parsed) |
| 79 | `type Age = int` | `export type Age = number` | yes (parsed) |
| 80 | `type Table = map[string]int` | `export type Table = Record<string, number>` | yes (parsed) |
| 81 | `type Bytes = []byte` | `export type Bytes = number[]` | yes (parsed) |
| 82 | `type JSONMap = map[string]interface{}` | `export type JSONMap = Record<string, unknown>` | yes (parsed) |
| 83 | `type HandlerFunc = func(string) error` | `export type HandlerFunc = (arg: string) => void` | yes (parsed) |

---

## Generic Types (Go 1.18+)

### Generic Structs

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 84 | `type Box[T any] struct { Value T }` | `export interface Box<T> { value: T }` | yes |
| 85 | `type Pair[K comparable, V any] struct { ... }` | `export interface Pair<K, V> { ... }` | yes |
| 86 | `type NumberBox[T int \| int64 \| float64] struct { ... }` | `export interface NumberBox<T> { ... }` | yes (parsed) |
| 87 | `type PaginatedResponse[T any] struct { Data []T }` | `export interface PaginatedResponse<T> { data: T[] }` | yes |
| 88 | `type Result[T any, E any] struct { Data *T; Error *E }` | `export interface Result<T, E> { data?: T; error?: E }` | yes |
| 89 | `type Cache[K comparable, V any] struct { Items map[K]V }` | `export interface Cache<K, V> { items: Record<K, V> }` | yes |

### Generic Interfaces

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 90 | `type Container[T any] interface { Get() T }` | `export interface Container<T> { get(): T }` | yes (parsed) |
| 91 | `type Ordered interface { ~int \| ~string \| ... }` | type constraint, not directly mapped | yes (parsed) |

### Generic Function Types

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 92 | `type Comparator[T any] func(a, b T) int` | `export type Comparator<T> = (a: T, b: T) => number` | yes (parsed) |

---

## Constant Patterns

### Typed Constants

| # | Go Pattern | TS Output Styles | Tested |
|---|---|---|---|
| 93 | `type Foo string` + string consts | union / enum / const-object | yes |
| 94 | Single typed const | `export type Foo = "value"` | yes |
| 95 | Many typed consts (long prefix) | all styles, prefix stripping | yes |
| 96 | `type Foo int` + iota (with `_` skip) | enum / const-object (inherited type) | yes |
| 97 | `type Foo int` + iota (no skip) | enum / const-object | yes |

### Untyped Constants

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 98 | Untyped string group | `export const Foo = "bar"` | yes |
| 99 | Untyped mixed group | individual exports | yes |
| 100 | Standalone string | `export const Foo = "bar"` | yes |
| 101 | Standalone int | `export const Foo = "123"` | yes |
| 102 | Standalone bool | `export const Foo = "true"` | yes |
| 103 | Multi-line backtick string | `` export const Foo = `...` `` (template literal) | yes |

---

## External Package Types (via type_map config)

| # | Go Type | Default TS Mapping | Tested |
|---|---|---|---|
| 104 | `time.Time` | `string` | yes |
| 105 | `*time.Time` | `string` (optional) | yes |
| 106 | `json.RawMessage` | `unknown` | yes |
| 107 | `uuid.UUID` | `string` (config) | configurable |
| 108 | `decimal.Decimal` | `string` (config) | configurable |
| 109 | `pq.StringArray` | `string[]` (config) | configurable |
| 110 | `net.IP` | `string` (config) | configurable |
| 111 | `sql.NullString` | `string \| null` (config) | configurable |
| 112 | `datatypes.JSON` | `unknown` (config) | configurable |

---

## Named Composite Types

| # | Go Pattern | TS Output | Tested |
|---|---|---|---|
| 113 | `type EventStream chan Event` | skip (channel) | yes (parsed) |
| 114 | `type SendOnly chan<- string` | skip (channel) | yes (parsed) |
| 115 | `type RecvOnly <-chan int` | skip (channel) | yes (parsed) |
| 116 | `type Reducer func(a, b int) int` | `export type Reducer = (a: number, b: number) => number` | yes (parsed) |
| 117 | `type ErrorHandler func(err error)` | `export type ErrorHandler = () => void` | yes (parsed) |
| 118 | `type IPv4 [4]byte` | `export type IPv4 = number[]` | yes (parsed) |
| 119 | `type Matrix3x3 [3][3]float64` | `export type Matrix3x3 = number[][]` | yes (parsed) |

---

## Summary

| Category | Count | Parsed | Generated |
|---|---|---|---|
| Built-in basic types | 19 | 19 | 17 (complex types skipped) |
| Pointer types | 4 | 4 | 3 |
| Struct patterns | 16 | 16 | 12 |
| Function types | 5 | 5 | 0 (future) |
| Array types | 5 | 5 | 3 |
| Slice types | 6 | 6 | 3 |
| Map types | 5 | 5 | 0 (future) |
| Channel types | 3 | 3 | 0 (skip) |
| Interface types | 5 | 5 | 2 |
| Type definitions | 9 | 9 | 3 |
| Type aliases | 6 | 6 | 0 (future) |
| Generic structs | 6 | 6 | 0 (future) |
| Generic interfaces | 2 | 2 | 0 (future) |
| Generic function types | 1 | 1 | 0 (future) |
| Typed constant patterns | 5 | 5 | 5 |
| Untyped constant patterns | 6 | 6 | 6 |
| External package types | 9 | 3 | 3 (rest via config) |
| Named composite types | 7 | 7 | 0 (future) |
| **Total** | **119** | **113** | **57** |

**Parsed** = the Go AST parser extracts it correctly.
**Generated** = the generator produces correct TypeScript output.
Items marked "future" are parsed but TypeScript generation is not yet implemented.
