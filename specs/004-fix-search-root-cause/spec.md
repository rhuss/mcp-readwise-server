# Feature Specification: Fix Search Root Cause and Highlights ID Mismatch

**Feature Branch**: `004-fix-search-root-cause`
**Created**: 2026-05-17
**Status**: Draft
**Input**: User description: "Fix the root cause of search_documents rate limiting (fetches all documents without caching) and fix list_highlights ULID ID mismatch"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Search Documents Works Without Rate Limiting (Priority: P1)

A user searches for documents using the `search_documents` tool. Currently, even after adding retry and rate limiting, every search fetches ALL documents from the upstream API without using the cache, causing rate limiting on large libraries. After the fix, search uses cached document data when available, making searches fast and reliable.

**Why this priority**: Document search is the primary discovery mechanism. The current approach of fetching the entire library on every search is fundamentally broken for any library with more than a few hundred documents.

**Independent Test**: Issue a `search_documents` call twice in succession. The first call may take time to populate the cache, but the second call should return results immediately without making upstream API requests.

**Acceptance Scenarios**:

1. **Given** the cache is warm (documents were recently fetched), **When** the user calls `search_documents` with a query, **Then** the server returns matching results from cache without making upstream API requests.
2. **Given** the cache is cold (no cached data), **When** the user calls `search_documents`, **Then** the server fetches documents from the upstream API, caches them, and returns matching results.
3. **Given** the cache TTL has expired, **When** the user calls `search_documents`, **Then** the server refreshes the cache from the upstream API and returns matching results.
4. **Given** the user calls `list_documents` or `get_document`, **When** the user subsequently calls `search_documents`, **Then** the search can leverage any document data already cached by those tools.

---

### User Story 2 - List Highlights for Reader Documents (Priority: P2)

A user views a Reader document (which has a ULID-style ID like `01krk43zgd57fcsj3tg7xn3skz`) and wants to list its highlights using `list_highlights`. Currently, this fails with a 400 error because the classic Readwise API expects numeric `book_id` values, not ULIDs. After the fix, the system detects ULID-style IDs and routes the request to the appropriate API endpoint.

**Why this priority**: Users frequently discover documents via Reader and then want to see their highlights. The current behavior returns a confusing error message that provides no workaround.

**Independent Test**: Call `list_highlights` with a Reader document ULID and verify that highlights are returned without errors.

**Acceptance Scenarios**:

1. **Given** a Reader document with a ULID-style ID, **When** the user calls `list_highlights` with that ID as `source_id`, **Then** the server returns highlights for that document without errors.
2. **Given** a classic Readwise source with a numeric ID, **When** the user calls `list_highlights` with that numeric ID, **Then** the server continues to work as before (backward compatible).
3. **Given** a Reader document ID that has no highlights, **When** the user calls `list_highlights`, **Then** the server returns an empty result set (not an error).

---

### User Story 3 - Search Highlights Uses Cache (Priority: P3)

Similar to document search, `search_highlights` currently fetches ALL export data from the upstream API on every call. After the fix, search uses cached export data when available.

**Why this priority**: The same root cause as US1, but for highlight search. Lower priority because highlight search is less commonly used.

**Independent Test**: Call `search_highlights` twice in succession. The second call should return results immediately from cache.

**Acceptance Scenarios**:

1. **Given** the cache is warm with export data, **When** the user calls `search_highlights`, **Then** results are returned from cache without upstream API calls.
2. **Given** the cache is cold, **When** the user calls `search_highlights`, **Then** export data is fetched, cached, and results returned.

---

### Edge Cases

- What happens when a `source_id` looks like a ULID but is actually invalid?
- How does the system handle a cache entry that is partially populated (e.g., document list was fetched mid-pagination when an error occurred)?
- What happens when the user creates or modifies a highlight, then immediately searches? Does the cache serve stale data?
- What happens when multiple concurrent search requests hit a cold cache? Do they all fetch from the API or does one fetch while others wait?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `search_documents` tool MUST use cached document data when the cache contains valid (non-expired) entries, avoiding upstream API calls.
- **FR-002**: The `search_documents` tool MUST populate the cache when fetching from the upstream API, so subsequent calls benefit from cached data.
- **FR-003**: The `search_highlights` tool MUST use cached export data when the cache contains valid (non-expired) entries.
- **FR-004**: The `search_highlights` tool MUST populate the cache when fetching from the upstream API.
- **FR-005**: The `list_highlights` tool MUST detect ULID-style source IDs and route to the appropriate API to retrieve highlights for Reader documents.
- **FR-006**: The `list_highlights` tool MUST continue to work with numeric source IDs for classic Readwise sources (backward compatible).
- **FR-007**: Cache entries used by search MUST share the same cache keys as equivalent direct tool calls (e.g., `list_documents` and `search_documents` share document cache).
- **FR-008**: Write operations (create, update, delete) MUST invalidate relevant cache entries so searches reflect current data.

### Key Entities

- **Document Cache**: Cached list of user's Reader documents, keyed by API key, with configurable TTL.
- **Export Cache**: Cached export data (highlights grouped by source), keyed by API key, with configurable TTL.
- **Source ID**: Either a numeric ID (classic Readwise) or a ULID string (Reader), determining which API endpoint to use.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Repeated document searches complete in under 1 second when cache is warm (no upstream API calls needed).
- **SC-002**: Repeated highlight searches complete in under 1 second when cache is warm.
- **SC-003**: Users can list highlights for Reader documents using ULID IDs without errors, on every attempt.
- **SC-004**: All existing tools continue to function without regression.
- **SC-005**: During normal usage, search operations do not trigger rate-limit errors from the upstream API.

## Assumptions

- The existing cache infrastructure (`internal/cache/Manager`) supports the caching patterns needed for search (keyed by API key and endpoint).
- ULID-style IDs can be reliably distinguished from numeric IDs by checking whether the string is purely numeric.
- The Reader API provides an endpoint to list highlights for a document by its ULID (or the export API can be filtered by document).
- Cache TTL of 300 seconds (current default) is acceptable for search freshness.
- Write operation cache invalidation already exists in the cache manager and can be extended to cover search-relevant entries.
