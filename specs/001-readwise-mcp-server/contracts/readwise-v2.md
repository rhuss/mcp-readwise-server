# Readwise v2 API Contract

**Branch**: `001-readwise-mcp-server` | **Date**: 2026-02-13

Base URL: `https://readwise.io/api/v2/`
Auth: `Authorization: Token <access_token>`
Validation: `GET /api/v2/auth/` returns HTTP 204 on success

## Rate Limits

| Endpoint | Rate |
| -------- | ---- |
| Default | 240 req/min |
| `GET /books/` | 20 req/min |
| `GET /highlights/` | 20 req/min |

429 response includes `Retry-After` header (seconds).

## Endpoints

### GET /books/

List sources. Page-number pagination.

**Query params**: `page`, `page_size` (max 1000), `category`, `last_highlight_at__gt`, `last_highlight_at__lt`, `updated__gt`, `updated__lt`

**Response**:
```json
{
  "count": 42,
  "next": "https://readwise.io/api/v2/books/?page=2",
  "previous": null,
  "results": [Source]
}
```

### GET /books/{id}/

Get single source.

### GET /books/{id}/tags

List tags on source. Returns `[Tag]`.

### POST /books/{id}/tags/

Add tag. Body: `{ "name": "tag_name" }`. Returns `Tag`.

### DELETE /books/{id}/tags/{tag_id}

Remove tag. Returns HTTP 204.

### GET /highlights/

List highlights. Page-number pagination.

**Query params**: `page`, `page_size` (max 1000), `book_id`, `updated__gt`, `updated__lt`

**Response**: Same structure as /books/ with `[Highlight]` results.

### GET /highlights/{id}/

Get single highlight.

### POST /highlights/

Create highlight(s). Accepts single or batch format.

**Body (batch)**:
```json
{
  "highlights": [
    {
      "text": "required, max 8191 chars",
      "title": "source title",
      "author": "optional",
      "source_url": "optional",
      "note": "optional",
      "location": 42,
      "location_type": "page",
      "highlighted_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

Dedup: `title` + `author` + `text` + `source_url`.

### PATCH /highlights/{id}/

Update highlight. Body: partial Highlight fields.

### DELETE /highlights/{id}/

Delete highlight. Returns HTTP 204.

### GET /highlights/{id}/tags

List tags on highlight. Returns `[Tag]`.

### POST /highlights/{id}/tags/

Add tag to highlight. Body: `{ "name": "tag_name" }`. Returns `Tag`.

### DELETE /highlights/{id}/tags/{tag_id}

Remove tag from highlight. Returns HTTP 204.

### GET /export/

Bulk export. Cursor-based pagination.

**Query params**: `pageCursor`, `updatedAfter` (ISO 8601)

**Response**:
```json
{
  "count": 100,
  "nextPageCursor": "string_or_null",
  "results": [
    {
      "user_book_id": 123,
      "title": "...",
      "author": "...",
      "category": "articles",
      "highlights": [Highlight]
    }
  ]
}
```

Loop until `nextPageCursor` is null.

### GET /review/

Get daily review.

**Response**:
```json
{
  "review_id": 789,
  "review_url": "https://readwise.io/review/...",
  "review_completed": false,
  "highlights": [ReviewHighlight]
}
```
