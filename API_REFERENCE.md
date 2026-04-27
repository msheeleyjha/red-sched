# API Reference

Complete reference for all API endpoints in the Referee Scheduler application.

**Base URL**: `http://localhost:8080` (development)

**Authentication**: Most endpoints require authentication via Google OAuth2 session cookie.

**Permissions**: Some endpoints require specific RBAC permissions.

---

## Table of Contents

1. [Authentication](#authentication)
2. [Users & Profiles](#users--profiles)
3. [Matches](#matches)
4. [Referees](#referees)
5. [Eligibility](#eligibility)
6. [Assignments](#assignments)
7. [Availability](#availability)
8. [Acknowledgment](#acknowledgment)
9. [RBAC Administration](#rbac-administration)
10. [Audit Logging](#audit-logging)

---

## Authentication

### Health Check
```
GET /health
```

**Description**: Server health check

**Authentication**: None

**Response** (200 OK):
```json
{
  "status": "ok",
  "time": "2026-04-27T10:30:00Z"
}
```

---

### Initiate Google OAuth
```
GET /api/auth/google
```

**Description**: Redirects to Google OAuth2 consent screen

**Authentication**: None

**Response**: 307 Redirect to Google

---

### OAuth Callback
```
GET /api/auth/google/callback
```

**Description**: Handles OAuth callback from Google

**Authentication**: None (OAuth state validated)

**Query Parameters**:
- `code` (string, required) - OAuth authorization code
- `state` (string, required) - CSRF protection token

**Response**: 307 Redirect to frontend `/auth/callback`

---

### Logout
```
POST /api/auth/logout
```

**Description**: Clears session and logs out user

**Authentication**: None

**Response** (200 OK):
```json
{
  "status": "logged out"
}
```

---

### Get Current User
```
GET /api/auth/me
```

**Description**: Returns currently authenticated user

**Authentication**: Required

**Response** (200 OK):
```json
{
  "id": 1,
  "google_id": "1234567890",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "referee",
  "status": "active",
  "first_name": "John",
  "last_name": "Doe",
  "date_of_birth": "1990-01-01T00:00:00Z",
  "certified": true,
  "cert_expiry": "2028-12-31T00:00:00Z",
  "grade": "Senior",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-04-01T00:00:00Z"
}
```

---

## Users & Profiles

### Get Profile
```
GET /api/profile
```

**Description**: Get current user's full profile

**Authentication**: Required

**Response** (200 OK): Same as `/api/auth/me`

---

### Update Profile
```
PUT /api/profile
```

**Description**: Update current user's profile

**Authentication**: Required

**Request Body**:
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "date_of_birth": "1990-01-01",
  "certified": true,
  "cert_expiry": "2028-12-31"
}
```

**Validation**:
- `first_name`, `last_name`: Required
- `date_of_birth`: YYYY-MM-DD format, cannot be in future
- `cert_expiry`: Required if `certified` is true, must be in future

**Response** (200 OK):
```json
{
  "id": 1,
  "first_name": "John",
  "last_name": "Doe",
  "date_of_birth": "1990-01-01T00:00:00Z",
  "certified": true,
  "cert_expiry": "2028-12-31T00:00:00Z",
  ...
}
```

**Error Responses**:
- `400 Bad Request`: Validation error
- `401 Unauthorized`: Not authenticated
- `500 Internal Server Error`: Database error

---

## Matches

All match endpoints require **`can_assign_referees`** permission.

### Parse CSV
```
POST /api/matches/import/parse
```

**Description**: Parse and validate CSV file for match import preview

**Permission**: `can_assign_referees`

**Request**: Multipart form data
- `file`: CSV file (Stack Team App export format)

**Response** (200 OK):
```json
{
  "valid_rows": [
    {
      "reference_id": "12345",
      "match_date": "2026-05-15",
      "match_time": "10:00",
      "age_group": "U12",
      "home_team": "Dragons",
      "away_team": "Tigers",
      "location": "Field 1"
    }
  ],
  "duplicates": [
    {
      "reference_id": "12346",
      "existing_match_id": 42,
      "row_data": { ... }
    }
  ],
  "errors": [
    "Row 5: Missing required field 'match_date'"
  ]
}
```

**CSV Format**:
- Columns: Date, Time, Age Group, Home Team, Away Team, Field/Location, Reference
- Eastern timezone assumed

---

### Import Matches
```
POST /api/matches/import/confirm
```

**Description**: Import matches after CSV parse

**Permission**: `can_assign_referees`

**Request Body**:
```json
{
  "rows": [
    {
      "reference_id": "12345",
      "match_date": "2026-05-15",
      "match_time": "10:00",
      "age_group": "U12",
      "home_team": "Dragons",
      "away_team": "Tigers",
      "location": "Field 1"
    }
  ],
  "skip_duplicates": true
}
```

**Response** (200 OK):
```json
{
  "imported_count": 10,
  "skipped_count": 2,
  "failed_count": 0
}
```

---

### List Matches
```
GET /api/matches?status={status}&start_date={date}&end_date={date}
```

**Description**: List all matches with filters

**Permission**: `can_assign_referees`

**Query Parameters**:
- `status` (string, optional): Filter by status (active, cancelled)
- `start_date` (string, optional): YYYY-MM-DD format
- `end_date` (string, optional): YYYY-MM-DD format

**Response** (200 OK):
```json
[
  {
    "id": 1,
    "reference_id": "12345",
    "match_date": "2026-05-15T14:00:00Z",
    "age_group": "U12",
    "home_team": "Dragons",
    "away_team": "Tigers",
    "location": "Field 1",
    "status": "active",
    "assignment_status": "partial",
    "overdue_acknowledgment": false,
    "roles": [
      {
        "id": 1,
        "role_type": "center",
        "referee_id": 5,
        "referee_name": "John Doe",
        "acknowledged": true,
        "acknowledged_at": "2026-05-01T10:00:00Z"
      }
    ]
  }
]
```

---

### Update Match
```
PUT /api/matches/{id}
```

**Description**: Update match details

**Permission**: `can_assign_referees`

**Request Body**:
```json
{
  "match_date": "2026-05-15T14:00:00Z",
  "home_team": "Dragons",
  "away_team": "Tigers",
  "location": "Field 2",
  "age_group": "U14",
  "status": "active"
}
```

**Response** (200 OK): Updated match object

**Notes**:
- Changing `age_group` reconfigures role slots automatically
- U6/U8: center only
- U10: center + optional assistants
- U12+: center + 2 assistants

---

### Add Role Slot
```
POST /api/matches/{match_id}/roles/{role_type}/add
```

**Description**: Add an extra role slot to match

**Permission**: `can_assign_referees`

**URL Parameters**:
- `match_id` (int64): Match ID
- `role_type` (string): Role type (center, assistant_1, assistant_2)

**Response** (201 Created):
```json
{
  "id": 42,
  "match_id": 1,
  "role_type": "assistant_1",
  "referee_id": null
}
```

---

## Referees

All referee endpoints require **`can_assign_referees`** permission.

### List Referees
```
GET /api/referees?status={status}&role={role}&certified={bool}
```

**Description**: List all referees with filtering

**Permission**: `can_assign_referees`

**Query Parameters**:
- `status` (string, optional): active, pending, inactive, removed
- `role` (string, optional): pending_referee, referee, assignor
- `certified` (bool, optional): Filter by certification status

**Response** (200 OK):
```json
[
  {
    "id": 5,
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "referee",
    "status": "active",
    "date_of_birth": "1990-01-01T00:00:00Z",
    "certified": true,
    "cert_expiry": "2028-12-31T00:00:00Z",
    "certification_status": "valid",
    "grade": "Senior",
    "upcoming_matches_count": 3
  }
]
```

**Certification Status Values**:
- `none`: Not certified
- `valid`: Certified, expiry > 30 days away
- `expiring_soon`: Certified, expiry within 30 days
- `expired`: Certified but expired

---

### Update Referee
```
PUT /api/referees/{id}
```

**Description**: Update referee status, grade, or role

**Permission**: `can_assign_referees`

**Request Body**:
```json
{
  "status": "active",
  "grade": "Senior"
}
```

**Auto-Promotion Rules**:
- `pending_referee` → `referee` when profile is complete
- `referee` → `assignor` when assignor role is assigned

**Protection Rules**:
- Cannot modify other assignors (only self)
- Cannot deactivate self
- Cannot deactivate referee with upcoming assignments

**Response** (200 OK): Updated referee object

---

## Eligibility

All eligibility endpoints require **`can_assign_referees`** permission.

### Get Eligible Referees
```
GET /api/matches/{id}/eligible-referees?role={role_type}
```

**Description**: Get list of eligible referees for a match/role

**Permission**: `can_assign_referees`

**URL Parameters**:
- `id` (int64): Match ID

**Query Parameters**:
- `role` (string, optional): Role type (default: center)

**Response** (200 OK):
```json
[
  {
    "id": 5,
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "grade": "Senior",
    "date_of_birth": "1990-01-01",
    "certified": true,
    "cert_expiry": "2028-12-31",
    "age_at_match": 37,
    "is_eligible": true,
    "ineligible_reason": null,
    "is_available": true
  },
  {
    "id": 6,
    "first_name": "Jane",
    "last_name": "Smith",
    "email": "jane@example.com",
    "grade": null,
    "date_of_birth": "1995-06-15",
    "certified": false,
    "cert_expiry": null,
    "age_at_match": 31,
    "is_eligible": false,
    "ineligible_reason": "Certification required for center referee role on U12+ matches",
    "is_available": false
  }
]
```

**Eligibility Rules**:
1. **U10 and younger**: Age-based (age group + 1 year), all roles
2. **U12+ center**: Certification required, valid expiry
3. **U12+ assistants**: No restrictions

---

## Assignments

All assignment endpoints require **`can_assign_referees`** permission.

### Assign Referee
```
POST /api/matches/{match_id}/roles/{role_type}/assign
```

**Description**: Assign, reassign, or remove a referee

**Permission**: `can_assign_referees`

**Request Body**:
```json
{
  "referee_id": 5
}
```

**To remove assignment**: Send `referee_id: null`

**Response** (200 OK):
```json
{
  "id": 42,
  "match_id": 1,
  "role_type": "center",
  "referee_id": 5,
  "referee_name": "John Doe",
  "acknowledged": false
}
```

**Side Effects**:
- Clears acknowledgment when referee changes
- Creates assignment history entry
- Checks for duplicate assignments

---

### Check Conflicts
```
GET /api/matches/{match_id}/conflicts?referee_id={id}&role_type={type}
```

**Description**: Check for assignment conflicts

**Permission**: `can_assign_referees`

**Query Parameters**:
- `referee_id` (int64, required): Referee to check
- `role_type` (string, required): Role type

**Response** (200 OK):
```json
{
  "has_conflicts": true,
  "conflicting_matches": [
    {
      "match_id": 42,
      "match_date": "2026-05-15T14:00:00Z",
      "home_team": "Dragons",
      "away_team": "Tigers",
      "location": "Field 3",
      "role_type": "assistant_1"
    }
  ]
}
```

**Conflict Detection**: Uses PostgreSQL `OVERLAPS` for time window conflicts

---

## Availability

All availability endpoints require **authentication**.

### Get Eligible Matches
```
GET /api/referee/matches?start_date={date}&end_date={date}
```

**Description**: Get matches the current referee is eligible for

**Authentication**: Required

**Query Parameters**:
- `start_date` (string, optional): YYYY-MM-DD
- `end_date` (string, optional): YYYY-MM-DD

**Response** (200 OK):
```json
[
  {
    "id": 1,
    "match_date": "2026-05-15T14:00:00Z",
    "age_group": "U12",
    "home_team": "Dragons",
    "away_team": "Tigers",
    "location": "Field 1",
    "available": true,
    "eligible_roles": ["center", "assistant_1", "assistant_2"]
  }
]
```

---

### Toggle Match Availability
```
POST /api/referee/matches/{id}/availability
```

**Description**: Set availability for a specific match

**Authentication**: Required

**Request Body**:
```json
{
  "available": true
}
```

**Tri-State Logic**:
- `{"available": true}` - Mark as available
- `{"available": false}` - Mark as unavailable
- `{"available": null}` - Clear preference (no preference)

**Response** (200 OK):
```json
{
  "match_id": 1,
  "referee_id": 5,
  "available": true
}
```

---

### Get Day Unavailability
```
GET /api/referee/day-unavailability
```

**Description**: Get all dates marked as unavailable

**Authentication**: Required

**Response** (200 OK):
```json
[
  {
    "date": "2026-05-20",
    "reason": "Family vacation"
  },
  {
    "date": "2026-06-15",
    "reason": null
  }
]
```

---

### Toggle Day Unavailability
```
POST /api/referee/day-unavailability/{date}
```

**Description**: Mark entire day as available/unavailable

**Authentication**: Required

**URL Parameters**:
- `date` (string): YYYY-MM-DD format

**Request Body**:
```json
{
  "unavailable": true,
  "reason": "Family vacation"
}
```

**Tri-State Logic**:
- `{"unavailable": true, "reason": "..."}` - Mark day unavailable
- `{"unavailable": false}` - Mark day available (clear)
- `{"unavailable": null}` - Clear preference

**Side Effect**: Marking day unavailable clears all match availability for that date

**Response** (200 OK):
```json
{
  "date": "2026-05-20",
  "referee_id": 5,
  "unavailable": true,
  "reason": "Family vacation"
}
```

---

## Acknowledgment

All acknowledgment endpoints require **authentication**.

### Acknowledge Assignment
```
POST /api/referee/matches/{match_id}/acknowledge
```

**Description**: Acknowledge an assignment for a match

**Authentication**: Required

**Response** (200 OK):
```json
{
  "match_id": 1,
  "referee_id": 5,
  "acknowledged": true,
  "acknowledged_at": "2026-05-01T10:30:00Z"
}
```

**Notes**:
- Must have an active assignment on the match
- Idempotent (can call multiple times)
- Acknowledgment overdue if >24 hours after assignment

**Error Responses**:
- `404 Not Found`: Match not found or no assignment for this referee

---

## RBAC Administration

All RBAC endpoints require **`can_assign_roles`** permission.

### List Roles
```
GET /api/admin/roles
```

**Description**: Get all available roles

**Permission**: `can_assign_roles`

**Response** (200 OK):
```json
[
  {
    "id": 1,
    "name": "Assignor",
    "description": "Can assign referees to matches",
    "created_at": "2026-01-01T00:00:00Z"
  }
]
```

---

### List Permissions
```
GET /api/admin/permissions
```

**Description**: Get all available permissions

**Permission**: `can_assign_roles`

**Response** (200 OK):
```json
[
  {
    "id": 1,
    "name": "can_assign_referees",
    "description": "Manage matches and assignments",
    "created_at": "2026-01-01T00:00:00Z"
  }
]
```

---

### Get User Roles
```
GET /api/admin/users/{id}/roles
```

**Description**: Get roles assigned to a user

**Permission**: `can_assign_roles`

**Response** (200 OK):
```json
[
  {
    "role_id": 1,
    "role_name": "Assignor",
    "assigned_at": "2026-01-01T00:00:00Z",
    "assigned_by": 2
  }
]
```

---

### Assign Role
```
POST /api/admin/users/{id}/roles
```

**Description**: Assign a role to a user

**Permission**: `can_assign_roles`

**Request Body**:
```json
{
  "role_id": 1
}
```

**Response** (200 OK):
```json
{
  "user_id": 5,
  "role_id": 1,
  "assigned_at": "2026-05-01T10:00:00Z"
}
```

---

### Revoke Role
```
DELETE /api/admin/users/{id}/roles/{roleId}
```

**Description**: Revoke a role from a user

**Permission**: `can_assign_roles`

**Response** (200 OK):
```json
{
  "message": "Role revoked successfully"
}
```

---

## Audit Logging

All audit endpoints require **`can_view_audit_logs`** permission.

### Query Audit Logs
```
GET /api/admin/audit-logs?action={action}&user_id={id}&start={date}&end={date}&limit={n}&offset={n}
```

**Description**: Query audit logs with filters

**Permission**: `can_view_audit_logs`

**Query Parameters**:
- `action` (string, optional): Filter by action type
- `user_id` (int64, optional): Filter by user
- `start` (string, optional): Start timestamp (ISO 8601)
- `end` (string, optional): End timestamp (ISO 8601)
- `limit` (int, optional): Results per page (default: 100, max: 1000)
- `offset` (int, optional): Pagination offset

**Response** (200 OK):
```json
{
  "logs": [
    {
      "id": 1234,
      "timestamp": "2026-05-01T10:30:00Z",
      "user_id": 5,
      "user_email": "john@example.com",
      "action": "match.assign",
      "resource_type": "assignment",
      "resource_id": 42,
      "ip_address": "192.168.1.100",
      "details": {
        "match_id": 1,
        "referee_id": 5,
        "role_type": "center"
      }
    }
  ],
  "total": 150,
  "limit": 100,
  "offset": 0
}
```

---

### Export Audit Logs
```
GET /api/admin/audit-logs/export?action={action}&user_id={id}&start={date}&end={date}
```

**Description**: Export audit logs as CSV

**Permission**: `can_view_audit_logs`

**Query Parameters**: Same as Query Audit Logs

**Response** (200 OK):
```csv
id,timestamp,user_id,user_email,action,resource_type,resource_id,ip_address,details
1234,2026-05-01T10:30:00Z,5,john@example.com,match.assign,assignment,42,192.168.1.100,"{""match_id"":1}"
```

---

### Purge Audit Logs
```
POST /api/admin/audit-logs/purge
```

**Description**: Manually purge old audit logs

**Permission**: `can_view_audit_logs`

**Request Body**:
```json
{
  "older_than_days": 90
}
```

**Response** (200 OK):
```json
{
  "purged_count": 1234
}
```

**Note**: Automatic purging happens daily (configured via `AUDIT_RETENTION_DAYS`)

---

## Error Responses

All endpoints use standard HTTP status codes and return errors in this format:

```json
{
  "error": "Error message here",
  "status": 400
}
```

### Common Status Codes

- **200 OK**: Successful request
- **201 Created**: Resource created successfully
- **400 Bad Request**: Invalid input or validation error
- **401 Unauthorized**: Not authenticated
- **403 Forbidden**: Missing required permission
- **404 Not Found**: Resource not found
- **409 Conflict**: Resource conflict (e.g., duplicate)
- **500 Internal Server Error**: Server error

---

## Rate Limiting

Currently no rate limiting is implemented. This may be added in the future.

---

## Versioning

The API is currently unversioned. Breaking changes will be documented in release notes.

---

**Last Updated**: 2026-04-27  
**Architecture**: Vertical Slice Architecture  
**Version**: Epic 8 (v2)
