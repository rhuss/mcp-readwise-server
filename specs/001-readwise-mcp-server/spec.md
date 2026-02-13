# Feature Specification: Readwise MCP Server

**Feature Branch**: `001-readwise-mcp-server`
**Created**: 2026-02-13
**Status**: Draft
**Input**: User description: "A Go-based MCP server providing AI assistants access to Readwise and Reader APIs with configurable tool profiles and server-side caching."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Browse Readwise Highlights (Priority: P1)

An AI assistant user wants to browse their Readwise highlights and sources. They connect their MCP client (e.g., Claude Desktop) to the server, provide their Readwise API key, and use tools to list sources, view highlights, and search across their library.

**Why this priority**: Core read functionality is the foundation. Without browsing highlights, no other feature has value. This represents the most common use case for Readwise users who want AI-assisted knowledge retrieval.

**Independent Test**: Can be fully tested by connecting to the server with a valid API key and listing sources/highlights. Delivers immediate value for knowledge retrieval.

**Acceptance Scenarios**:

1. **Given** a configured MCP client with a valid Readwise API key, **When** the user requests a list of sources, **Then** a paginated list of sources (books, articles, etc.) is returned with id, title, author, category, and highlight count.
2. **Given** a source with highlights, **When** the user requests highlights filtered by source, **Then** only highlights belonging to that source are returned.
3. **Given** a large highlight library, **When** the user searches highlights with a query, **Then** up to 50 matching highlights are returned ranked by relevance (exact matches ranked above partial matches), searching across text, notes, and source titles.
4. **Given** a valid API key, **When** the user requests today's daily review, **Then** today's review highlights are returned.
5. **Given** a source with tags, **When** the user requests tags for that source, **Then** all tags applied to the source are returned.

---

### User Story 2 - Browse Reader Documents (Priority: P1)

An AI assistant user wants to browse their Reader library. They use tools to list documents, filter by location (new, later, archive) or category (article, PDF, email), and view document details including content.

**Why this priority**: Reader is the second major reading surface and equally important for users who use Reader alongside Readwise. Together with Story 1, these cover all read-only operations.

**Independent Test**: Can be fully tested by connecting with a valid API key and listing/searching Reader documents. Delivers value for document discovery and reading management.

**Acceptance Scenarios**:

1. **Given** a valid API key with Reader documents, **When** the user requests documents filtered by location "later", **Then** only documents in the "later" list are returned.
2. **Given** a document ID, **When** the user requests that document with content included, **Then** the document details include the full content.
3. **Given** documents in Reader, **When** the user searches documents with a query, **Then** up to 50 matching documents are returned ranked by relevance (exact matches ranked above partial matches).
4. **Given** tags in Reader, **When** the user requests all tags, **Then** all tags are returned.

---

### User Story 3 - Create and Modify Content (Priority: P2)

An AI assistant user wants to save new documents, create highlights, and manage tags. They use write tools to save URLs to Reader, create highlights with notes, and organize content with tags.

**Why this priority**: Write operations extend value by allowing the assistant to help organize and capture knowledge, but read access must work first.

**Independent Test**: Can be tested by saving a URL to Reader, creating a highlight, and adding tags. Verifiable by subsequent read operations confirming the changes.

**Acceptance Scenarios**:

1. **Given** write capabilities are enabled alongside reader access, **When** the user saves a URL, **Then** the document is saved to Reader and its identifier is returned.
2. **Given** write capabilities are enabled alongside readwise access, **When** the user creates a highlight with text and a source, **Then** the highlight is created and visible in subsequent list operations.
3. **Given** a source and a tag name, **When** the user adds a tag to the source, **Then** the tag is applied and visible on subsequent source queries.
4. **Given** multiple highlights to create, **When** the user submits a batch of highlights, **Then** all highlights are created and their identifiers are returned.

---

### User Story 4 - Profile-Based Tool Configuration (Priority: P2)

A server operator wants to control which tools are exposed to MCP clients. They configure an environment variable to select specific profiles, and only tools matching active profiles are registered with the MCP server.

**Why this priority**: Profiles are the primary configuration mechanism. They gate access to write and destructive operations, and must work correctly for safe, controlled operation.

**Independent Test**: Can be tested by starting the server with different profile configurations and verifying which tools are advertised to clients.

**Acceptance Scenarios**:

1. **Given** only the "readwise" profile is enabled, **When** the server starts, **Then** only the 9 readwise profile tools are registered.
2. **Given** the "basic" shortcut profile is selected, **When** the server starts, **Then** the shortcut expands to "reader" and "write" profiles, the internal dependency ("write" requires "reader") is satisfied by the expansion, and the corresponding tools are registered.
3. **Given** the "write" profile is selected without a corresponding read profile, **When** the server starts, **Then** validation fails because "write" requires "readwise" or "reader" as a dependency.
4. **Given** multiple profiles are selected ("readwise", "reader", "write"), **When** the server starts, **Then** tools from all three profiles are registered with no duplicates.
5. **Given** the "destructive" profile is selected without a corresponding read profile, **When** the server starts, **Then** validation fails because "destructive" requires "readwise" or "reader" as a dependency.

---

### User Story 5 - Video Document Management (Priority: P3)

An AI assistant user wants to browse video documents in Reader, view transcripts, and create timestamped highlights.

**Why this priority**: Video is a specialized subset of Reader functionality. Valuable for video-heavy users but not a core requirement for the initial release.

**Independent Test**: Can be tested by listing videos, getting a video with transcript, and creating a timestamped highlight.

**Acceptance Scenarios**:

1. **Given** the video profile is enabled with reader access, **When** the user lists videos, **Then** only video-category documents are returned (read-only, no write profile needed).
2. **Given** a video document ID, **When** the user requests that video, **Then** video details including transcript (if available) are returned (read-only, no write profile needed).
3. **Given** the video profile is enabled with reader access but without write access, **When** the user requests playback position, **Then** the current position is returned (read-only, no write profile needed).
4. **Given** a video ID and both video and write profiles enabled, **When** the user updates the playback position, **Then** the position is updated (requires write profile).
5. **Given** a video ID and both video and write profiles enabled, **When** the user creates a highlight with text and a timestamp, **Then** a highlight is created with timestamp metadata attached (requires write profile).
6. **Given** the video profile is enabled without write access, **When** the client attempts to update a video position or create a video highlight, **Then** those tools are not available.

---

### User Story 6 - Destructive Operations (Priority: P3)

An AI assistant user wants to delete highlights, documents, or tags. These operations require explicit opt-in via the "destructive" profile to prevent accidental data loss.

**Why this priority**: Delete operations are important for library management but carry the highest risk. Explicit opt-in prevents accidental data loss.

**Independent Test**: Can be tested by creating and then deleting a highlight, verifying it no longer appears in list operations.

**Acceptance Scenarios**:

1. **Given** the destructive profile is NOT enabled, **When** the client attempts to call a delete operation, **Then** the tool is not available.
2. **Given** the destructive profile IS enabled with readwise access, **When** the user deletes a highlight with a valid ID, **Then** the highlight is deleted and cached data is invalidated.
3. **Given** the destructive profile IS enabled with reader access, **When** the user deletes a document with a valid ID, **Then** the document is deleted and cached data is invalidated.

---

### User Story 7 - Efficient Caching (Priority: P2)

The server caches read-heavy responses to reduce external API calls and speed up search operations. Write and destructive operations invalidate relevant caches. Cache memory usage is bounded to prevent unbounded growth.

**Why this priority**: Caching is essential for the search tools (which operate over cached data) and for staying within upstream API rate limits.

**Independent Test**: Can be tested by making repeated read calls and verifying reduced external API calls, then making a write call and verifying cache invalidation.

**Acceptance Scenarios**:

1. **Given** a previous export request, **When** the same request is made within the cache validity window, **Then** the cached result is returned without making an external API call.
2. **Given** a cached export, **When** the user creates a highlight, **Then** the entire export cache for that user is invalidated (all cached entries for that endpoint and user, not just the affected source).
3. **Given** cache memory exceeds the configured limit, **When** a new entry is added, **Then** the least recently used entries are evicted until memory is under the limit.
4. **Given** two different API keys, **When** both request exports, **Then** each gets their own cache entry (cache entries are isolated per user).

---

### Edge Cases

- What happens when the upstream API returns a rate limit error? The server passes it through to the client with retry timing information.
- What happens when the API key is missing or invalid? The server returns an authentication error with a clear message.
- What happens when a write profile is enabled without a corresponding read profile? The server rejects the configuration at startup with a clear validation error.
- What happens when the destructive profile is enabled without a corresponding read profile? The server rejects the configuration at startup, same as for write.
- What happens when the cache memory limit is set to 0? Caching is effectively disabled, but the server still functions normally.
- What happens when a client provides invalid parameters (e.g., negative page size)? The server returns a validation error with a description of the problem.
- What happens when the upstream API is unreachable? The server returns an error and does not cache the failure.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST expose Readwise operations as MCP tools: list sources, get source details, list highlights, get highlight details, export highlights, get daily review, list source tags, list highlight tags, and search highlights
- **FR-002**: System MUST expose Reader operations as MCP tools: list documents, get document (with optional content), list tags, and search documents
- **FR-003**: System MUST expose write operations as MCP tools: save document, update document, create highlight, update highlight, add source tag, add highlight tag, and bulk create highlights
- **FR-004**: System MUST expose video operations as MCP tools: list videos, get video with transcript, and get playback position (read-only, require reader profile); update playback position and create timestamped highlights (require both reader and write profiles)
- **FR-005**: System MUST expose destructive operations as MCP tools: delete highlight, delete highlight tag, delete source tag, and delete document
- **FR-006**: System MUST support configurable tool profiles that control which tools are exposed to clients
- **FR-007**: System MUST resolve profile shortcuts ("basic" expands to reader and write; "all" expands to all profiles)
- **FR-008**: System MUST validate profile dependencies at startup (write requires a read profile; video requires reader; destructive requires a read profile corresponding to its tools, i.e., readwise for highlight/tag deletion and reader for document deletion)
- **FR-009**: System MUST extract the user's API key from client request headers without storing credentials server-side
- **FR-010**: System MUST cache responses for export, source listing, document listing, and tag listing operations with configurable validity periods
- **FR-011**: System MUST invalidate all cached entries for the affected endpoint and user when a write or destructive operation succeeds (full per-user-per-endpoint invalidation, not partial)
- **FR-012**: System MUST enforce configurable memory limits on the cache with least-recently-used eviction
- **FR-013**: System MUST pass through upstream API rate limit responses to the client with retry timing information
- **FR-014**: System MUST return structured error responses with error type, code, and descriptive message
- **FR-015**: System MUST provide health and readiness check endpoints
- **FR-016**: System MUST support highlight search via local search over cached export data (searching text, notes, and source titles) with a configurable result limit (default 50, maximum 200)
- **FR-017**: System MUST support document search via local search over cached document list data with a configurable result limit (default 50, maximum 200)
- **FR-018**: System MUST isolate cache entries per user using a hash of the API key to prevent cross-user data leakage

### Key Entities

- **Source**: A book, article, PDF, tweet, or podcast in the user's Readwise library. Has title, author, category, tags, and a collection of highlights.
- **Highlight**: A text excerpt from a source with optional note, location, color, and tags. The core unit of captured knowledge.
- **Document**: A Reader item (article, email, RSS entry, PDF, epub, tweet, video) with title, author, URL, location, reading progress, and optional full content.
- **Tag**: A label applied to sources, highlights, or documents for organization.
- **Profile**: A named group of MCP tools that can be enabled or disabled as a unit. Profiles have a type (read or modifier) and optional dependencies on other profiles.
- **Cache Entry**: A stored response keyed by user identity and request parameters, with a time-to-live and size tracking for memory management.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 29 MCP tools (9 readwise + 4 reader + 7 write + 5 video + 4 destructive) function correctly when tested against a live API
- **SC-002**: Profile system correctly enables and disables tools for all valid profile combinations, with no tools leaking between profiles
- **SC-003**: Cached search operations return relevant results to the user within 1 second for libraries containing up to 10,000 highlights
- **SC-004**: Cache memory stays within configured limits under sustained usage with no unbounded growth
- **SC-005**: Server starts and responds to health checks within 5 seconds of launch
- **SC-006**: Users with invalid or missing API keys receive clear, actionable error messages within 1 second
- **SC-007**: 100% of functional requirements have corresponding automated tests that pass
- **SC-008**: Server handles at least 10 concurrent client connections without degraded response times

## Assumptions

- Users have an active Readwise account with a valid API key
- The Readwise v2 and Reader v3 APIs remain stable and backward-compatible during development
- MCP clients support API key passthrough via request headers
- The server will be deployed in a single-instance configuration (no horizontal scaling required initially)
- Cache validity periods of 5 minutes for frequently changing data and 10 minutes for rarely changing data (tags) are acceptable defaults
- Default cache memory limit of 128 MB is sufficient for typical usage patterns
- The "basic" shortcut profile expanding to "reader" and "write" covers the most common multi-profile configuration
- Rate limit responses from the upstream API include retry timing information that can be forwarded to clients
