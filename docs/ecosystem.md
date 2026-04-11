# Ecosystem (Brainstorm)

Everything that could stay in sync between a Go backend and a TypeScript frontend. This is a future roadmap document — features beyond core type generation.

For current capabilities, see [README.md](../README.md).

---

## 1. API Client Generation

Go routes + handlers → fully typed fetch client.

**Go:**

```go
r.POST("/api/v1/posts", handler.CreatePost)
r.GET("/api/v1/posts/:id", handler.GetPost)
r.GET("/api/v1/posts", handler.ListPosts)

func (h *Handler) CreatePost(c *gin.Context) {
    var req CreatePostRequest
    // ...
    c.JSON(200, PostResponse{...})
}
```

**TypeScript:**

```typescript
export const api = {
  posts: {
    create: (req: CreatePostRequest) =>
      post<PostResponse>("/api/v1/posts", req),
    get: (id: string) =>
      get<PostResponse>(`/api/v1/posts/${id}`),
    list: (params?: ListPostsParams) =>
      get<PostResponse[]>("/api/v1/posts", params),
  },
};
```

No hand-written fetch calls. Request/response types enforced at compile time.

---

## 2. Error Codes / Error Types

Go error definitions → typed error codes the frontend can match on.

**Go:**

```go
var (
    ErrNotFound      = NewAppError("NOT_FOUND", 404, "Resource not found")
    ErrUnauthorized  = NewAppError("UNAUTHORIZED", 401, "Not authorized")
    ErrQuotaExceeded = NewAppError("QUOTA_EXCEEDED", 429, "Plan quota exceeded")
)
```

**TypeScript:**

```typescript
export type AppErrorCode = "NOT_FOUND" | "UNAUTHORIZED" | "QUOTA_EXCEEDED";

export const AppErrors = {
  NOT_FOUND: { code: "NOT_FOUND", status: 404, message: "Resource not found" },
  UNAUTHORIZED: { code: "UNAUTHORIZED", status: 401, message: "Not authorized" },
  QUOTA_EXCEEDED: { code: "QUOTA_EXCEEDED", status: 429, message: "Plan quota exceeded" },
} as const;
```

Frontend gets autocomplete on `err.code` checks. No more magic strings.

---

## 3. Validation Rules Sync

Go struct validation tags → client-side validation schemas.

**Go:**

```go
type CreatePostRequest struct {
    Title   string   `json:"title" validate:"required,min=3,max=200"`
    Content string   `json:"content" validate:"required,min=10"`
    Tags    []string `json:"tags" validate:"max=5,dive,max=30"`
}
```

**TypeScript (Zod):**

```typescript
export const CreatePostRequestSchema = z.object({
  title: z.string().min(3).max(200),
  content: z.string().min(10),
  tags: z.array(z.string().max(30)).max(5),
});

export type CreatePostRequest = z.infer<typeof CreatePostRequestSchema>;
```

Same rules, both sides. Could also target Yup or Valibot.

---

## 4. WebSocket / Realtime Event Types

Go event constants + payload structs → type-safe event map.

**Go:**

```go
type WSEvent string

const (
    WSEventMessageNew    WSEvent = "message.new"
    WSEventMessageUpdate WSEvent = "message.update"
    WSEventUserJoined    WSEvent = "user.joined"
)

type WSMessageNewPayload struct {
    MessageID string `json:"message_id"`
    Content   string `json:"content"`
    UserID    string `json:"user_id"`
}
```

**TypeScript:**

```typescript
export type WSEvent = "message.new" | "message.update" | "user.joined";

export interface WSEventMap {
  "message.new": WSMessageNewPayload;
  "message.update": WSMessageUpdatePayload;
  "user.joined": WSUserJoinedPayload;
}

// Type-safe listener
export type WSHandler = <E extends WSEvent>(
  event: E,
  handler: (payload: WSEventMap[E]) => void,
) => void;
```

Wrong event name or wrong payload shape → compile error.

---

## 5. Webhook Payload Registry

Go webhook event definitions → discriminated type map for consumers.

**Go:**

```go
type WebhookEvent string

const (
    WebhookPostCreated   WebhookEvent = "post.created"
    WebhookPostUpdated   WebhookEvent = "post.updated"
    WebhookMemberInvited WebhookEvent = "member.invited"
)

type WebhookPostCreatedPayload struct {
    Post Post `json:"post"`
}

type WebhookPostUpdatedPayload struct {
    Post    Post     `json:"post"`
    Changes []string `json:"changes"`
}
```

**TypeScript:**

```typescript
export interface WebhookEventMap {
  "post.created": WebhookPostCreatedPayload;
  "post.updated": WebhookPostUpdatedPayload;
  "member.invited": WebhookMemberInvitedPayload;
}

export type WebhookEvent = keyof WebhookEventMap;

// Discriminated handler
export function onWebhook<E extends WebhookEvent>(
  event: E,
  handler: (payload: WebhookEventMap[E]) => void,
): void;
```

---

## 6. Route / URL Builder

Go router definitions → type-safe URL constructors.

**Go:**

```go
r.GET("/api/v1/workspaces/:workspaceId/posts/:postId", handler.GetPost)
r.GET("/api/v1/workspaces/:workspaceId/members", handler.ListMembers)
r.DELETE("/api/v1/posts/:postId/comments/:commentId", handler.DeleteComment)
```

**TypeScript:**

```typescript
export const routes = {
  getPost: (p: { workspaceId: string; postId: string }) =>
    `/api/v1/workspaces/${p.workspaceId}/posts/${p.postId}`,
  listMembers: (p: { workspaceId: string }) =>
    `/api/v1/workspaces/${p.workspaceId}/members`,
  deleteComment: (p: { postId: string; commentId: string }) =>
    `/api/v1/posts/${p.postId}/comments/${p.commentId}`,
};
```

Forget a param → compile error. Typo in path → impossible.

---

## 7. Permission / Authorization Definitions

Go role+permission constants → frontend can gate UI without network calls.

**Go:**

```go
type Permission string

const (
    PermCanAdmin       Permission = "canAdmin"
    PermCanApp         Permission = "canApp"
    PermCanManageUsers Permission = "canManageUsers"
    PermCanBilling     Permission = "canBilling"
)

var RoleOwner = Role{
    Name:        "owner",
    Permissions: []Permission{PermCanAdmin, PermCanApp, PermCanManageUsers, PermCanBilling},
}

var RoleMember = Role{
    Name:        "member",
    Permissions: []Permission{PermCanApp},
}
```

**TypeScript:**

```typescript
export type Permission = "canAdmin" | "canApp" | "canManageUsers" | "canBilling";

export const Roles = {
  owner: {
    name: "owner",
    permissions: ["canAdmin", "canApp", "canManageUsers", "canBilling"],
  },
  member: {
    name: "member",
    permissions: ["canApp"],
  },
} as const;

export type RoleName = keyof typeof Roles;
```

Show/hide a billing button without hitting an API. Permission check logic lives in one place.

---

## 8. Feature Flags

Go feature flag keys → typed flags the frontend can check.

**Go:**

```go
const (
    FeatureNewEditor   = "new_editor"
    FeatureDarkMode    = "dark_mode"
    FeatureAIAssistant = "ai_assistant"
)
```

**TypeScript:**

```typescript
export type FeatureFlag = "new_editor" | "dark_mode" | "ai_assistant";

export const FeatureFlags = {
  NewEditor: "new_editor",
  DarkMode: "dark_mode",
  AIAssistant: "ai_assistant",
} as const;
```

Typo in a flag name → compile error. Remove a flag from Go → TS breaks at build time, not at runtime in production.

---

## 9. i18n / Translation Keys

Go translation key constants → typed `t()` function.

**Go:**

```go
const (
    MsgWelcome       = "msg.welcome"
    MsgPostCreated   = "msg.post_created"
    MsgQuotaExceeded = "msg.quota_exceeded"
    ErrInvalidEmail  = "err.invalid_email"
)
```

**TypeScript:**

```typescript
export type TranslationKey =
  | "msg.welcome"
  | "msg.post_created"
  | "msg.quota_exceeded"
  | "err.invalid_email";

declare function t(key: TranslationKey, params?: Record<string, string>): string;
```

Use a key that doesn't exist → compile error. Delete a key from Go → find every broken callsite instantly.

---

## 10. State Machine / Transitions

Go status types + transition maps → frontend knows valid next states.

**Go:**

```go
type OrderStatus string

const (
    OrderPending   OrderStatus = "pending"
    OrderPaid      OrderStatus = "paid"
    OrderShipped   OrderStatus = "shipped"
    OrderDelivered OrderStatus = "delivered"
    OrderCancelled OrderStatus = "cancelled"
)

var OrderTransitions = map[OrderStatus][]OrderStatus{
    OrderPending: {OrderPaid, OrderCancelled},
    OrderPaid:    {OrderShipped, OrderCancelled},
    OrderShipped: {OrderDelivered},
}
```

**TypeScript:**

```typescript
export type OrderStatus = "pending" | "paid" | "shipped" | "delivered" | "cancelled";

export const OrderTransitions: Record<OrderStatus, OrderStatus[]> = {
  pending: ["paid", "cancelled"],
  paid: ["shipped", "cancelled"],
  shipped: ["delivered"],
  delivered: [],
  cancelled: [],
};
```

Frontend renders only valid action buttons per state. Logic defined once in Go, consumed everywhere.

---

## 11. Config / Settings Schema

Go config structs → typed settings objects for admin panels.

**Go:**

```go
type WorkspaceSettings struct {
    AllowPublicAccess bool   `json:"allow_public_access"`
    MaxUploadSizeMB   int    `json:"max_upload_size_mb"`
    DefaultLanguage   string `json:"default_language"`
    CustomDomain      string `json:"custom_domain,omitempty"`
}
```

**TypeScript:**

```typescript
export interface WorkspaceSettings {
  allow_public_access: boolean;
  max_upload_size_mb: number;
  default_language: string;
  custom_domain?: string;
}
```

Admin settings form is type-checked against the backend struct.

---

## 12. Pagination / Query Parameter Types

Go list request params → typed query builders.

**Go:**

```go
type ListPostsParams struct {
    Page     int    `query:"page" validate:"min=1"`
    PageSize int    `query:"page_size" validate:"min=1,max=100"`
    Sort     string `query:"sort" validate:"oneof=created_at updated_at title"`
    Order    string `query:"order" validate:"oneof=asc desc"`
    Search   string `query:"search"`
}

type PaginatedResponse[T any] struct {
    Data       []T `json:"data"`
    Total      int `json:"total"`
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    TotalPages int `json:"total_pages"`
}
```

**TypeScript:**

```typescript
export interface ListPostsParams {
  page?: number;
  page_size?: number;
  sort?: "created_at" | "updated_at" | "title";
  order?: "asc" | "desc";
  search?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}
```

Sort/order fields are unions, not open strings. Invalid query params caught at compile time.

---

## 13. Middleware / Header Constants

Go header and metadata keys → typed header helpers.

**Go:**

```go
const (
    HeaderWorkspaceId = "x-workspace-id"
    HeaderRequestId   = "x-request-id"
    HeaderAPIVersion  = "x-api-version"
)
```

**TypeScript:**

```typescript
export const Headers = {
  WorkspaceId: "x-workspace-id",
  RequestId: "x-request-id",
  APIVersion: "x-api-version",
} as const;

export type AppHeader = typeof Headers[keyof typeof Headers];
```

No more mistyped header names in fetch calls.

---

## 14. Database / SQL Enum Sync

Scan `.sql` migration files directly as a source.

**SQL:**

```sql
CREATE TYPE post_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE user_role AS ENUM ('admin', 'member', 'viewer');
```

**TypeScript:**

```typescript
export type PostStatus = "draft" | "published" | "archived";
export type UserRole = "admin" | "member" | "viewer";
```

Single source of truth is the migration file. Go and TS both derive from it.

---

## 15. GraphQL Enum / Type Sync

Go gqlgen models → TS types matching the GraphQL schema.

**Go:**

```go
type PostType string // gqlgen enum

const (
    PostTypeArticle PostType = "ARTICLE"
    PostTypeVideo   PostType = "VIDEO"
    PostTypePodcast PostType = "PODCAST"
)
```

**TypeScript:**

```typescript
export type PostType = "ARTICLE" | "VIDEO" | "PODCAST";
```

If you use gqlgen, these stay in sync with your `.graphql` schema automatically.

---

## Summary

| # | Use Case | Source | What it solves |
|---|---|---|---|
| 1 | API client | Router + handlers | Hand-written fetch calls |
| 2 | Error codes | Error variables | Untyped error matching |
| 3 | Validation | Struct validate tags | Duplicate validation logic |
| 4 | WebSocket events | Event consts + payload structs | Untyped realtime handlers |
| 5 | Webhook payloads | Webhook event registry | Untyped webhook consumers |
| 6 | URL builder | Router definitions | URL string typos |
| 7 | Permissions | Role/permission consts | Frontend permission guessing |
| 8 | Feature flags | Flag key consts | Mistyped flag names at runtime |
| 9 | i18n keys | Translation key consts | Missing/wrong translation keys |
| 10 | State machines | Status types + transition maps | Frontend state logic drift |
| 11 | Config schema | Settings structs | Untyped admin forms |
| 12 | Query params | List request structs | Invalid sort/filter values |
| 13 | Header constants | Header key consts | Mistyped header names |
| 14 | SQL enums | Migration files | DB ↔ frontend enum drift |
| 15 | GraphQL enums | gqlgen model consts | Schema ↔ frontend type drift |
