# Story 7.1: Weekend Date Range Filter - Complete

## Story Overview
**Epic**: 7 - Scheduling Interface Improvements (V2)  
**Story**: 7.1 - Weekend Date Range Filter  
**Status**: Complete  
**Date**: 2026-04-28  
**Story Points**: 5

## Objective
Allow assignors and referees to quickly filter matches by weekend date ranges so they can view and manage a full weekend of matches at once.

## Acceptance Criteria - All Met

- [x] Filter UI includes "Weekend" option alongside existing date range
- [x] "This Weekend" shortcut selects current Saturday-Sunday range
- [x] "Next Weekend" shortcut selects next Saturday-Sunday range
- [x] API endpoint accepts date range parameters (`date_from`, `date_to`)
- [x] Results refresh without page reload
- [x] Clear filter button resets to all matches

## Implementation Summary

### 1. Weekend Shortcut Buttons
**Files Modified**:
- `frontend/src/routes/assignor/matches/+page.svelte`
- `frontend/src/routes/referee/matches/+page.svelte`
- `frontend/src/routes/referee/assignments/+page.svelte`

**Implementation**:
Added `setWeekend()` function and helper utilities to all three scheduling pages. The function handles edge cases when called on any day of the week, including Saturday and Sunday.

```typescript
function setWeekend(which: 'this' | 'next') {
    const today = new Date();
    const day = today.getDay();
    let saturday: Date;

    if (which === 'this') {
        if (day === 6) saturday = new Date(today);
        else if (day === 0) { saturday = new Date(today); saturday.setDate(today.getDate() - 1); }
        else saturday = getNextSaturday(today);
    } else {
        // Next weekend logic...
    }

    const sunday = new Date(saturday);
    sunday.setDate(saturday.getDate() + 1);
    dateFrom = formatDateParam(saturday);
    dateTo = formatDateParam(sunday);
}
```

### 2. Quick Select UI
Added a "Quick select" row with "This Weekend" and "Next Weekend" buttons between the date filter inputs and the footer. Clicking a shortcut sets the date range and immediately triggers `applyFilters()`.

```html
<div class="weekend-shortcuts">
    <span class="shortcut-label">Quick select:</span>
    <button class="btn-shortcut" on:click={() => { setWeekend('this'); applyFilters(); }}>
        This Weekend
    </button>
    <button class="btn-shortcut" on:click={() => { setWeekend('next'); applyFilters(); }}>
        Next Weekend
    </button>
</div>
```

### 3. Server-Side Date Range Filtering (Pre-existing)
The API endpoints already accept `date_from` and `date_to` query parameters, implemented in the prior pagination work:
- `GET /api/matches` (assignor)
- `GET /api/referee/matches` (referee availability)
- `GET /api/referee/assignments` (referee assignments)

## Pages Updated
| Page | Route | Weekend Shortcuts |
|------|-------|-------------------|
| Assignor Match Schedule | `/assignor/matches` | This Weekend, Next Weekend |
| Referee Matches | `/referee/matches` | This Weekend, Next Weekend |
| Referee Assignments | `/referee/assignments` | This Weekend, Next Weekend |
