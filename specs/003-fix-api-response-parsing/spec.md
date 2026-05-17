# Feature Specification: Fix API Response Parsing Bugs

**Feature Branch**: `003-fix-api-response-parsing`
**Created**: 2026-05-16
**Status**: Draft
**Input**: User description: "Fix two API response parsing bugs and improve rate limiting resilience in the Readwise MCP server"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Search Documents Returns Correct Results (Priority: P1)

A user searches for documents by title or content using the `search_documents` tool. Currently, every call returns a spurious "rate limited" error regardless of actual API usage. After the fix, searches return matching documents from the user's Readwise library.

**Why this priority**: Document search is the primary discovery mechanism. A broken search makes the entire MCP server significantly less useful since users cannot find articles by title or content.

**Independent Test**: Can be fully tested by issuing a `search_documents` call with a known query and verifying that matching documents are returned instead of a rate-limit error.

**Acceptance Scenarios**:

1. **Given** the server is running and the user has documents in their Readwise library, **When** the user calls `search_documents` with a valid query, **Then** the server returns a list of matching documents (not a rate-limit error).
2. **Given** the server is running and no documents match the query, **When** the user calls `search_documents` with a non-matching query, **Then** the server returns an empty result set (not a rate-limit error).
3. **Given** the upstream Readwise API genuinely returns a 429 rate-limit response, **When** the user calls `search_documents`, **Then** the server reports a rate-limit error with an accurate retry-after duration.

---

### User Story 2 - Search Highlights Returns Parsed Results (Priority: P2)

A user searches for highlights using the `search_highlights` tool. Currently, every call fails with a JSON parsing error because the response cursor field type is mismatched. After the fix, highlight searches return matching results with proper pagination support.

**Why this priority**: Highlight search is a valuable feature for users who annotate heavily, though less commonly used than document search.

**Independent Test**: Can be fully tested by issuing a `search_highlights` call with a known query and verifying that matching highlights are returned without JSON parsing errors.

**Acceptance Scenarios**:

1. **Given** the server is running and the user has highlights in their Readwise library, **When** the user calls `search_highlights` with a valid query, **Then** the server returns matching highlights without parsing errors.
2. **Given** the upstream API returns a numeric cursor value in the pagination response, **When** the server parses the response, **Then** it handles the numeric cursor correctly and supports pagination to subsequent pages.
3. **Given** the upstream API returns a string cursor value, **When** the server parses the response, **Then** it handles the string cursor correctly (backward compatibility).

---

### User Story 3 - Resilient Rate Limit Handling (Priority: P3)

When the server encounters a genuine rate limit from the Readwise API (20 requests/minute), it automatically retries with appropriate backoff rather than immediately surfacing the error to the user. Users experience fewer interruptions during burst usage.

**Why this priority**: Improves reliability during heavy usage sessions but is an enhancement rather than a bug fix.

**Independent Test**: Can be tested by simulating a burst of requests that triggers the upstream rate limit and verifying that the server retries automatically before surfacing an error.

**Acceptance Scenarios**:

1. **Given** the server receives a genuine 429 response from the Readwise API, **When** the response includes a Retry-After header, **Then** the server waits the specified duration and retries the request automatically.
2. **Given** the server is handling multiple concurrent requests, **When** the request rate approaches the upstream limit, **Then** the server throttles outgoing requests to stay within the rate limit.
3. **Given** retries are exhausted after the maximum number of attempts, **When** the upstream API still returns 429, **Then** the server surfaces a clear rate-limit error to the user with the expected wait time.

---

### Edge Cases

- What happens when the Readwise API returns an unexpected HTTP status code (e.g., 500, 503)?
- How does the system handle a response body that is valid JSON but has an unexpected structure?
- What happens when the Retry-After header is missing from a 429 response?
- How does the system handle concurrent search requests that all trigger rate limiting?
- What happens when the cursor field is null, empty, or a different unexpected type (boolean, object)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST correctly parse non-429 HTTP responses from the Readwise Reader search endpoint without misinterpreting them as rate-limit errors.
- **FR-002**: System MUST return search results from the `search_documents` tool when the upstream API returns a successful response.
- **FR-003**: System MUST handle both numeric and string values for the `nextPageCursor` field in paginated API responses without parsing errors.
- **FR-004**: System MUST support pagination through search results when multiple pages of results exist.
- **FR-005**: System MUST automatically retry requests when the upstream API returns a genuine 429 rate-limit response, using exponential backoff with jitter up to a maximum of 3 retries. When a Retry-After header is present, the backoff duration MUST respect that value. When the header is absent, a default backoff of 60 seconds MUST be used.
- **FR-006**: System MUST enforce client-side rate limiting using a token bucket algorithm configured to the upstream API's 20 requests/minute limit.
- **FR-007**: System MUST surface a clear, accurate error message to users when all 3 retries are exhausted and the upstream API remains rate-limited.
- **FR-008**: System MUST preserve existing behavior for all other endpoints (`list_documents`, `get_document`, etc.) that currently work correctly.

### Key Entities

- **API Response**: The HTTP response from the Readwise API, including status code, headers (particularly Retry-After), and body.
- **Cursor**: A pagination token that can be either a string or a number, used to fetch subsequent pages of results.
- **Rate Limit State**: Tracks request counts and timing to enforce the 20 requests/minute upstream limit.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully search documents by title or content and receive matching results on every attempt (when not genuinely rate-limited).
- **SC-002**: Users can successfully search highlights and receive matching results, including paginated result sets.
- **SC-003**: During normal usage (fewer than 20 requests per minute), users never encounter rate-limit errors.
- **SC-004**: During burst usage, the system automatically retries rate-limited requests and the user receives results without manual intervention in at least 90% of cases.
- **SC-005**: All existing MCP tools (`list_documents`, `get_document`, etc.) continue to function without regression.

## Clarifications

### Session 2026-05-16

- Q: What retry strategy should be used for genuine 429 responses? → A: Exponential backoff with jitter, maximum 3 retries.
- Q: What client-side rate limiting algorithm should be used? → A: Token bucket, configured to the 20 requests/minute upstream limit.
- Q: What should happen when a 429 response lacks a Retry-After header? → A: Use a default backoff of 60 seconds (matching the upstream rate window).

## Assumptions

- The Readwise Reader search API uses the same base URL as other Reader endpoints.
- The upstream rate limit of 20 requests/minute applies uniformly across all endpoints.
- The `nextPageCursor` field may arrive as either a JSON string or a JSON number, and both forms are valid.
- The server's existing caching layer (when enabled) reduces the effective request rate and does not need modification for this feature.
- A maximum of 3 automatic retries with exponential backoff is sufficient for handling transient rate limits.
