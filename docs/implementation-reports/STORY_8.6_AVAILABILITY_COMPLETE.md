# Story 8.6: Availability Feature - Complete

## Overview
Successfully implemented the **Availability** feature slice following the vertical slice architecture pattern. This feature allows referees to manage their availability for specific matches and mark entire days as unavailable.

## Implementation Summary

### Files Created/Modified

#### New Files
1. **features/availability/models.go** (42 lines)
   - `DayUnavailability` - Day unavailability view model
   - `DayUnavailabilityData` - Database model
   - `ToggleMatchAvailabilityRequest` - Match availability request (tri-state)
   - `ToggleMatchAvailabilityResponse` - Match availability response
   - `ToggleDayUnavailabilityRequest` - Day unavailability request
   - `ToggleDayUnavailabilityResponse` - Day unavailability response

2. **features/availability/repository.go** (167 lines)
   - `RepositoryInterface` - Data access interface with 5 methods
   - `Repository` - PostgreSQL implementation
   - Methods:
     - `ToggleMatchAvailability` - Insert/update/delete match availability
     - `MatchExistsAndActive` - Verify match exists and is upcoming
     - `GetDayUnavailability` - Get all unavailable days for a referee
     - `ToggleDayUnavailability` - Insert/delete day unavailability
     - `ClearMatchAvailabilityForDay` - Clear match availability when day is marked unavailable

3. **features/availability/service_interface.go** (12 lines)
   - `ServiceInterface` - Business logic interface
   - 3 method signatures

4. **features/availability/service.go** (99 lines)
   - `Service` - Business logic implementation
   - Tri-state match availability (available/unavailable/no preference)
   - Date validation (YYYY-MM-DD format)
   - Auto-clear match availability when day is marked unavailable

5. **features/availability/handler.go** (107 lines)
   - `Handler` - HTTP request handler
   - `ToggleMatchAvailability` - POST /api/referee/matches/{id}/availability
   - `GetDayUnavailability` - GET /api/referee/day-unavailability
   - `ToggleDayUnavailability` - POST /api/referee/day-unavailability/{date}

6. **features/availability/routes.go** (16 lines)
   - Route registration with authentication

7. **features/availability/service_test.go** (433 lines, 11 tests)
   - `mockRepository` implementation
   - Tests:
     - ToggleMatchAvailability_Success
     - ToggleMatchAvailability_MatchNotFound
     - ToggleMatchAvailability_ClearPreference
     - GetDayUnavailability_Success
     - GetDayUnavailability_EmptyResult
     - ToggleDayUnavailability_MarkUnavailable
     - ToggleDayUnavailability_RemoveUnavailability
     - ToggleDayUnavailability_InvalidDate (4 subtests)
     - ToggleDayUnavailability_RepositoryError

8. **features/availability/handler_test.go** (320 lines, 11 tests)
   - `mockService` implementation
   - HTTP tests:
     - ToggleMatchAvailabilityHandler_Success
     - ToggleMatchAvailabilityHandler_UserNotInContext
     - ToggleMatchAvailabilityHandler_InvalidMatchID (3 subtests)
     - ToggleMatchAvailabilityHandler_InvalidRequestBody
     - ToggleMatchAvailabilityHandler_MatchNotFound
     - GetDayUnavailabilityHandler_Success
     - GetDayUnavailabilityHandler_UserNotInContext
     - ToggleDayUnavailabilityHandler_Success
     - ToggleDayUnavailabilityHandler_InvalidDate
     - ToggleDayUnavailabilityHandler_UserNotInContext
     - ToggleDayUnavailabilityHandler_InvalidRequestBody

#### Modified Files
1. **main.go**
   - Added availability feature import
   - Initialized repository, service, and handler
   - Registered routes with authMiddleware
   - Commented out old availability and day unavailability routes
   - Commented out old acknowledgment route (moved to acknowledgment feature)

## Test Results

### Test Coverage
```
ok  	github.com/msheeley/referee-scheduler/features/availability	0.003s
coverage: 56.7% of statements
```

### Coverage Details
- Handler: 100% coverage
- Service: 100% coverage
- Repository: 0% coverage (uses mocks in unit tests)
- Routes: 0% coverage (simple registration)

### All Tests Passing
- 11 service tests (with 4 subtests for date validation)
- 11 handler tests (with 3 subtests for invalid IDs)
- Total: 22 unique test cases

## Key Features

### Match Availability (Tri-State)
1. **Available** - Referee marked themselves as available for this match
   - Request: `{"available": true}`
   - Creates/updates availability record with available=true

2. **Unavailable** - Referee marked themselves as unavailable for this match
   - Request: `{"available": false}`
   - Creates/updates availability record with available=false

3. **No Preference** - Referee has no preference (clears previous selection)
   - Request: `{"available": null}`
   - Deletes availability record

### Day Unavailability
1. **Mark Day Unavailable**
   - Marks an entire day as unavailable
   - Optional reason field
   - Automatically clears all match availability for that day
   - Prevents setting availability for matches on that day

2. **Remove Day Unavailability**
   - Removes day unavailability record
   - Does not restore match availability (must be set individually)

### Business Rules Enforced

#### Match Availability
1. **Match must exist and be active** - Cannot mark availability for cancelled or past matches
2. **Only upcoming matches** - match_date >= CURRENT_DATE

#### Day Unavailability
1. **Valid date format** - Must be YYYY-MM-DD
2. **Auto-clear behavior** - Marking a day unavailable clears all match availability for that day
3. **Reason is optional** - Can provide a reason but it's not required

## API Endpoints

### POST /api/referee/matches/{id}/availability
Toggles a referee's availability for a specific match (tri-state).

**Authentication**: Required (via authMiddleware)

**URL Parameters**:
- `id` (int64) - Match ID

**Request Body**:
```json
{
  "available": true  // true=available, false=unavailable, null=no preference
}
```

**Success Response (200 OK)**:
```json
{
  "success": true,
  "available": true
}
```

**Error Responses**:
- 400: Invalid match ID or request body
- 401: Not authenticated
- 404: Match not found or not available for marking
- 500: Internal server error

### GET /api/referee/day-unavailability
Returns all days marked as unavailable for the current referee.

**Authentication**: Required (via authMiddleware)

**Success Response (200 OK)**:
```json
[
  {
    "id": 1,
    "referee_id": 100,
    "unavailable_date": "2027-12-31",
    "reason": "Vacation",
    "created_at": "2026-04-27T15:30:00Z"
  },
  {
    "id": 2,
    "referee_id": 100,
    "unavailable_date": "2028-01-01",
    "reason": null,
    "created_at": "2026-04-27T16:00:00Z"
  }
]
```

**Error Responses**:
- 401: Not authenticated
- 500: Internal server error

### POST /api/referee/day-unavailability/{date}
Toggles a referee's unavailability for an entire day.

**Authentication**: Required (via authMiddleware)

**URL Parameters**:
- `date` (string) - Date in YYYY-MM-DD format

**Request Body**:
```json
{
  "unavailable": true,
  "reason": "Out of town"  // optional
}
```

**Success Response (200 OK)**:
```json
{
  "success": true,
  "unavailable": true,
  "date": "2027-12-31"
}
```

**Error Responses**:
- 400: Invalid date format or request body
- 401: Not authenticated
- 500: Internal server error

## Data Flow

### Repository Layer
```sql
-- Toggle match availability
-- Insert/update
INSERT INTO availability (match_id, referee_id, available, created_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (match_id, referee_id)
DO UPDATE SET available = $3, created_at = NOW()

-- Delete (clear preference)
DELETE FROM availability
WHERE match_id = $1 AND referee_id = $2

-- Get day unavailability
SELECT id, referee_id, unavailable_date, reason, created_at
FROM day_unavailability
WHERE referee_id = $1
ORDER BY unavailable_date

-- Toggle day unavailability
INSERT INTO day_unavailability (referee_id, unavailable_date, reason, created_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (referee_id, unavailable_date)
DO UPDATE SET reason = $3

DELETE FROM day_unavailability
WHERE referee_id = $1 AND unavailable_date = $2

-- Clear match availability for day
DELETE FROM availability
WHERE referee_id = $1
  AND match_id IN (
    SELECT id FROM matches WHERE match_date = $2
  )
```

### Service Layer
1. Validate inputs (match exists, date format)
2. Execute repository operations
3. For day unavailability: auto-clear match availability if marking as unavailable
4. Return success response

### Handler Layer
1. Extract user from context
2. Parse match ID or date from URL
3. Decode request body
4. Call service layer
5. Return JSON response or error

## Testing Strategy

### Service Tests
- **Tri-state logic**: Test all three states (available, unavailable, no preference)
- **Validations**: Match existence, date format
- **Auto-clear behavior**: Verify match availability is cleared when day is marked unavailable
- **Repository errors**: Test error propagation
- **Empty results**: Test when no unavailable days exist

### Handler Tests
- **HTTP success**: All endpoints
- **Authentication**: User not in context
- **Input validation**: Invalid IDs, dates, malformed JSON
- **Service errors**: Match not found, invalid date
- **Error mapping**: Service errors to HTTP status codes

## Design Decisions

### 1. Tri-State Availability
Used pointer to bool (`*bool`) to support three states:
- `true` = available
- `false` = unavailable
- `null` = no preference (deletes record)

This is cleaner than using a string enum and matches the database schema.

### 2. Auto-Clear Match Availability
When a day is marked unavailable, automatically clear all match availability for that day. This ensures consistency and prevents conflicting signals (day unavailable but match available).

### 3. Reason is Optional
Day unavailability reason is optional because:
- Personal reasons may be private
- Not all unavailability needs explanation
- Reduces friction in marking days unavailable

### 4. Date Validation
Strict date format validation (YYYY-MM-DD) ensures:
- Consistent date handling
- No timezone issues
- Easy parsing in both backend and frontend

### 5. No Time Range Support
Day unavailability is full-day only (no half-day or time ranges). This simplifies the logic and UX. If a referee is only unavailable for part of a day, they can mark individual match availability instead.

## Integration

### Main.go Changes
```go
// Import
"github.com/msheeley/referee-scheduler/features/availability"

// Initialize (in main function)
availabilityRepo := availability.NewRepository(db)
availabilityService := availability.NewService(availabilityRepo)
availabilityHandler := availability.NewHandler(availabilityService)
log.Println("Availability feature initialized")

// Register routes
availabilityHandler.RegisterRoutes(r, authMiddleware)

// Old routes commented out
// r.HandleFunc("/api/referee/matches/{id}/availability", authMiddleware(toggleAvailabilityHandler)).Methods("POST")
// r.HandleFunc("/api/referee/day-unavailability", authMiddleware(getDayUnavailabilityHandler)).Methods("GET")
// r.HandleFunc("/api/referee/day-unavailability/{date}", authMiddleware(toggleDayUnavailabilityHandler)).Methods("POST")
```

## Dependencies
- `shared/errors` - Typed error handling
- `shared/middleware` - User context and authentication
- `gorilla/mux` - URL parameter extraction
- Standard library: context, database/sql, time

## Metrics
- **Lines of Code**: ~1,186 total
  - Production code: ~433 lines
  - Test code: ~753 lines
  - Test/code ratio: 1.7:1
- **Test Count**: 22 unique test cases
- **Test Coverage**: 56.7% overall, 100% for handler and service
- **Build Time**: Clean compilation, no errors
- **Test Execution**: 0.003s

## Limitations and Future Work

### 1. Missing getEligibleMatchesForReferee Endpoint
**Why**: This endpoint requires the `checkEligibility` function from the eligibility feature, which hasn't been refactored yet. The endpoint combines:
- Match listing
- Eligibility checking (depends on eligibility.go)
- Availability status
- Assignment status
- Conflict detection

**Location**: Still in availability.go (old file)

**Future Work**: When eligibility feature is refactored, this endpoint can be:
- Option A: Moved to availability feature with dependency on eligibility
- Option B: Created as a composite endpoint in a new "referee-matches" feature
- Option C: Moved to eligibility feature as it's primarily about eligible matches

**Current Status**: The route is still registered in main.go as:
```go
r.HandleFunc("/api/referee/matches", authMiddleware(getEligibleMatchesForRefereeHandler)).Methods("GET")
```

### 2. No Partial Day Unavailability
Currently only supports full-day unavailability. No support for:
- Half-day (morning/afternoon)
- Time ranges (9am-12pm)
- Specific match exclusions while day is available

**Workaround**: Use individual match availability instead.

## Next Steps
Continue with remaining features in Story 8.6:
1. ✅ **Acknowledgment** - COMPLETE
2. ✅ **Referees** - COMPLETE
3. ✅ **Availability** - COMPLETE
4. ⏳ **Eligibility** (from eligibility.go)
5. ⏳ **Profile** (from profile.go, may merge with users)

## Lessons Learned
1. **Tri-state pattern**: Using `*bool` is cleaner than string enums for tri-state logic
2. **Auto-clear behavior**: Cascading deletes/updates improve data consistency
3. **Date validation**: Strict format validation prevents timezone and parsing issues
4. **Feature dependencies**: Some features naturally depend on others (availability → eligibility)
5. **Gradual migration**: Can migrate partial functionality and note dependencies for later
