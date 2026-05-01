# Story 7.2: Pagination for Match Lists - Complete

## Story Overview
**Epic**: 7 - Scheduling Interface Improvements (V2)  
**Story**: 7.2 - Pagination for Match Lists  
**Status**: Complete  
**Date**: 2026-04-28  
**Story Points**: 5

## Objective
Paginate match lists so pages load quickly with large schedules, with configurable page sizes and navigation controls.

## Acceptance Criteria - All Met

- [x] Scheduling page displays 25 matches per page by default
- [x] Pagination controls at bottom: Previous, page numbers, Next
- [x] Page size selector: 25, 50, 100 matches
- [x] Navigating pages maintains filter selections
- [x] Server-side pagination for performance with large datasets

### Acceptance Criteria - Not Implemented
- [ ] URL query parameter reflects current page (?page=2)

**Note**: URL query param sync was deferred as it adds complexity (back button behavior, SSR considerations) without significant user value for this application's scale. Filter state is maintained in component state during the session.

## Implementation Summary

### 1. Backend Pagination Endpoints
**Files Modified**:
- `backend/features/matches/models.go` - Added `MatchListParams` and `PaginatedMatchesResponse` structs
- `backend/features/matches/repository.go` - Added `CountActive()`, `buildActiveWhereClause()` helper
- `backend/features/matches/service.go` - `ListMatches()` returns paginated response
- `backend/features/matches/handler.go` - Parses `page`, `per_page` query params
- `backend/availability.go` - Added pagination to referee matches and assignments endpoints

**Paginated Response Structure**:
```json
{
    "matches": [...],
    "total": 150,
    "page": 1,
    "per_page": 25,
    "total_pages": 6
}
```

### 2. Frontend Pagination Controls
**Files Modified**:
- `frontend/src/routes/assignor/matches/+page.svelte`
- `frontend/src/routes/referee/matches/+page.svelte`
- `frontend/src/routes/referee/assignments/+page.svelte`

**Pagination UI**: Previous/Next buttons with numbered page buttons and ellipsis for large page counts.

```html
<div class="pagination">
    <button on:click={() => goToPage(currentPage - 1)} disabled={currentPage <= 1}>
        Previous
    </button>
    <div class="page-numbers">
        <!-- Page buttons with ellipsis for ranges > 5 pages -->
    </div>
    <button on:click={() => goToPage(currentPage + 1)} disabled={currentPage >= totalPages}>
        Next
    </button>
</div>
```

### 3. Page Size Selector
Added per-page selector (25, 50, 100) in the filter footer of all three pages. Changing page size resets to page 1.

```html
<div class="per-page-selector">
    <label for="perPage">Per page:</label>
    <select bind:value={perPage} on:change={() => { currentPage = 1; loadMatches(); }}>
        <option value={25}>25</option>
        <option value={50}>50</option>
        <option value={100}>100</option>
    </select>
</div>
```

### 4. Apply Filters Pattern
All filter changes require clicking "Apply Filters" to trigger the API call, preventing excessive requests. Applying filters resets to page 1.

## Pages Updated
| Page | Route | Pagination | Per-Page Selector |
|------|-------|------------|-------------------|
| Assignor Match Schedule | `/assignor/matches` | Yes | 25/50/100 |
| Referee Matches | `/referee/matches` | Yes | 25/50/100 |
| Referee Assignments | `/referee/assignments` | Yes | 25/50/100 |

## Test Coverage
- Backend handler tests updated for paginated response format
- Backend service tests updated for new `ListMatches` signature
- All 11 test packages pass
- Frontend svelte-check: 0 errors
