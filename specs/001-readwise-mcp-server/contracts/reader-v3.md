# Reader v3 API Contract

**Branch**: `001-readwise-mcp-server` | **Date**: 2026-02-13

Base URL: `https://readwise.io/api/v3/`
Auth: `Authorization: Token <access_token>` (same token as v2)
Validation: `GET /api/v2/auth/` returns HTTP 204 (shared with v2)

## Rate Limits

| Endpoint | Rate |
| -------- | ---- |
| `GET /list/` | 20 req/min |
| `POST /save/` | 50 req/min |
| `PATCH /update/{id}/` | 50 req/min |
| `DELETE /delete/{id}/` | 20 req/min |
| `GET /tags/` | 20 req/min |

429 response includes `Retry-After` header (seconds).

## Endpoints

### GET /list/

List documents. Cursor-based pagination.

**Query params**: `id` (single doc), `updatedAfter` (ISO 8601), `location` (new, later, shortlist, archive, feed), `category` (article, email, rss, highlight, note, pdf, epub, tweet, video), `tag` (up to 5), `limit` (1-100, default 100), `pageCursor`, `withHtmlContent` (bool), `withRawSourceUrl` (bool)

**Response**:
```json
{
  "count": 100,
  "nextPageCursor": "string_or_null",
  "results": [Document]
}
```

Loop until `nextPageCursor` is null.

**Single document fetch**: Pass `id` parameter. Returns single-item result set.

### POST /save/

Save a document.

**Body**:
```json
{
  "url": "required",
  "html": "optional raw HTML",
  "should_clean_html": false,
  "title": "optional override",
  "author": "optional",
  "summary": "optional",
  "published_date": "optional",
  "image_url": "optional",
  "location": "new|later|archive|feed",
  "category": "optional",
  "saved_using": "optional attribution",
  "tags": ["tag1", "tag2"],
  "notes": "optional"
}
```

**Response**: HTTP 201 (new) or 200 (existing). Body: `{ "id": "...", "url": "..." }`

### PATCH /update/{id}/

Update document metadata.

**Body**: Partial fields from save, plus `seen` (bool).

**Response**: Updated `Document`.

### DELETE /delete/{id}/

Delete document.

**Response**: HTTP 204.

### GET /tags/

List all tags.

**Response**: Tag objects (key-value pairs).
