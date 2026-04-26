# V2 PRD Decomposition

## Objective
Deliver V2 enhancements through 9 epics that implement backend architectural refactoring, RBAC, audit logging, UI/UX improvements, match archival, referee reporting, CSV import enhancements, scheduling improvements, and comprehensive testing infrastructure with CI/CD. Target outcome: improved code maintainability, increased referee satisfaction, 50% reduction in assignor time spent on scheduling tasks, and automated quality assurance preventing regressions.

## Scope Summary
- **In Scope**: Backend vertical slice refactoring, RBAC implementation, audit logging, TailwindCSS UI refresh, match archival/history, referee match reporting, CSV import deduplication, scheduling UI enhancements, comprehensive unit testing with CI/CD
- **Out of Scope**: Notifications (stretch), file attachments (stretch), payment processing, mobile apps, third-party integrations

## Assumptions
1. Existing Go + SvelteKit stack remains unchanged
2. Current database can be migrated to support new schema (roles, audit logs, match reports, archived matches)
3. Existing users will be migrated to new RBAC system via automated script + manual review
4. CSV import format remains consistent with V1 (includes `reference_id` field)
5. Audit logging and archival retention defaults (2 years) are configurable via environment variables or database settings

## Key Non-Functional Requirements
- **Security**: All API endpoints enforce RBAC (Critical)
- **Accessibility**: WCAG 2.1 AA compliance (Critical)
- **Performance**: Scheduling page loads in < 3s with 1000+ matches (High)
- **Performance**: CSV import processes 500 matches in < 10s (High)
- **Data Integrity**: No duplicate matches; all state changes audited (Critical)

## Risks
- Backend refactoring introduces regressions → Mitigation: Comprehensive test coverage before refactor (Epic 9), incremental migration by feature, parallel testing against original structure
- RBAC breaks existing workflows → Mitigation: Feature flags, gradual rollout, extensive RBAC tests (Story 9.4)
- Data integrity issues during archival migration → Mitigation: Dry-run, backups, phased rollout, match reporting tests (Story 9.6)
- User migration to RBAC → Mitigation: Automated script + manual System Admin review
- TailwindCSS breaks styles → Mitigation: Visual regression testing, incremental migration, snapshot tests (Story 9.7)
- Audit logging performance → Mitigation: Async writes, indexing, retention policy
- Refactoring while building new features increases complexity → Mitigation: Complete Epic 8 first (Option A), or run parallel with dedicated developer (Option B)
- Tests add development time → Mitigation: Write tests alongside features (not after); tests prevent costly bugs and regressions later

---

## Epic 1: Role-Based Access Control (RBAC) Foundation
**Goal**: Implement permission-based RBAC system with 5-table structure; support Super Admin, Assignor, and Referee roles; enable multi-role assignment; enforce permission-based authorization across all API endpoints.

**Priority**: Critical (foundational for security)

**Dependencies**: None (must complete before other epics can enforce permission-based authorization)

**V2 Scope**: Build full data model with seeded permissions; permission management UI deferred to post-V2

### Stories

#### Story 1.1: Database Schema for RBAC (5 Tables)
**As a** system architect  
**I want** a complete RBAC database schema with roles, permissions, and assignments  
**So that** users can be assigned multiple roles with granular permissions

**Acceptance Criteria**:
- [ ] `roles` table created with fields: `id`, `name` (Super Admin, Assignor, Referee), `description`, `created_at`
- [ ] `permissions` table created with fields: `id`, `name` (technical name like `can_import_matches`), `display_name` (UI-friendly), `description`, `resource`, `action`, `created_at`
- [ ] `user_roles` junction table: `user_id`, `role_id`, `assigned_by`, `assigned_at` (multi-role support)
- [ ] `role_permissions` junction table: `role_id`, `permission_id`, `created_at`
- [ ] Constraints prevent duplicate user-role and role-permission pairs
- [ ] Database indexes on all junction table foreign keys
- [ ] Migration handles existing users (assigns appropriate roles)

**Story Points**: 5

---

#### Story 1.2: Seed Initial Roles and Permissions
**As a** system architect  
**I want** initial roles and permissions seeded in the database  
**So that** the system has a working permission set from day 1

**Acceptance Criteria**:
- [ ] Seed 3 roles: Super Admin, Assignor, Referee
- [ ] Seed initial permissions with technical and display names:
  - `can_manage_users` / "Manage Users" (Super Admin, Assignor)
  - `can_import_matches` / "Import Match Schedule" (Super Admin, Assignor)
  - `can_assign_referees` / "Assign Referees to Matches" (Super Admin, Assignor)
  - `can_request_assignments` / "Request Match Assignments" (Super Admin, Referee)
  - `can_edit_own_match_reports` / "Edit Own Match Reports" (Super Admin, Referee)
  - `can_view_audit_logs` / "View Audit Logs" (Super Admin only)
  - `can_assign_roles` / "Assign User Roles" (Super Admin only)
- [ ] Seed `role_permissions` assignments for each role
- [ ] Migration script is idempotent (safe to re-run)
- [ ] Documentation lists all seeded permissions and their assignments

**Story Points**: 3

---

#### Story 1.3: Permission-Based Authorization Middleware
**As a** developer  
**I want** middleware that enforces permission-based authorization on all endpoints  
**So that** users can only perform actions they have permissions for

**Acceptance Criteria**:
- [ ] Middleware checks user permissions before processing requests (not just roles)
- [ ] Authorization logic: user has permission if ANY of their roles includes it (most permissive wins)
- [ ] Super Admin role automatically passes all permission checks
- [ ] Endpoint-level permission requirements defined (e.g., CSV import requires `can_import_matches`)
- [ ] Returns 403 Forbidden if user lacks required permission
- [ ] Returns 401 Unauthorized if user not authenticated
- [ ] New users with no roles can only access profile edit endpoint
- [ ] All existing endpoints updated with permission requirements
- [ ] Middleware logs authorization failures to audit log
- [ ] Unit tests cover all permission combinations and multi-role scenarios

**Story Points**: 8

---

#### Story 1.4: Backend Role Assignment API
**As a** System Admin  
**I want** API endpoints to assign and revoke user roles  
**So that** I can manage user access to the system

**Acceptance Criteria**:
- [ ] `POST /api/admin/users/:id/roles` assigns role to user (requires `can_assign_roles` permission)
- [ ] `DELETE /api/admin/users/:id/roles/:roleId` revokes role from user (requires `can_assign_roles` permission)
- [ ] `GET /api/admin/users/:id/roles` returns all roles for a user (requires `can_assign_roles` permission)
- [ ] `GET /api/admin/roles` returns all available roles (requires `can_assign_roles` permission)
- [ ] `GET /api/admin/permissions` returns all permissions with display names (requires `can_assign_roles` permission)
- [ ] API returns 403 Forbidden if user lacks required permission
- [ ] Audit log entry created for all role changes
- [ ] Cannot revoke own Super Admin role (prevent lockout)

**Story Points**: 5

---

#### Story 1.5: System Admin Role Management UI
**As a** System Admin  
**I want** a UI to assign/revoke roles for users  
**So that** I can manage user access without using API directly

**Acceptance Criteria**:
- [ ] Admin panel page `/admin/users` lists all users with their assigned roles
- [ ] Shows user status: "No Roles (Profile Only)" for users without roles
- [ ] "Manage Roles" button opens modal with checkboxes for each role (Super Admin, Assignor, Referee)
- [ ] Modal shows which permissions each role grants (read-only informational display)
- [ ] Saving role changes calls backend API and refreshes list
- [ ] Cannot remove own Super Admin role (UI prevents + backend validates)
- [ ] Success/error messages displayed after role changes
- [ ] Page only accessible to users with `can_assign_roles` permission
- [ ] First-time users see banner: "No roles assigned. You can only edit your profile until an admin assigns you a role."

**Story Points**: 5

---

#### Story 1.6: User Migration Script
**As a** system admin  
**I want** a migration script that assigns roles to existing users  
**So that** V1 users can transition to V2 without manual data entry

**Acceptance Criteria**:
- [ ] Script identifies existing assignors (by permission flags, email domain, or manual config file)
- [ ] Script assigns "Assignor" role to identified users
- [ ] Script assigns "Referee" role to identified referee users
- [ ] Script creates at least one Super Admin user (from config or environment variable)
- [ ] Script handles users who should have multiple roles (Assignor + Referee)
- [ ] Script is idempotent (can run multiple times safely without duplicating assignments)
- [ ] Migration report generated (users processed, roles assigned, warnings for unmapped users)
- [ ] Unmapped users left with no roles (can only edit profile until admin assigns role)

**Story Points**: 3

---

#### Story 1.7: Assignment Position Field
**As a** system architect  
**I want** assignments to track referee position (center vs assistant)  
**So that** match report editing can be restricted to center referees

**Acceptance Criteria**:
- [ ] `assignments` table adds `position` field (VARCHAR/ENUM: 'center', 'assistant_1', 'assistant_2', etc.)
- [ ] Position field is nullable for backward compatibility with existing assignments
- [ ] Migration sets existing assignments to `position = 'center'` as default
- [ ] Assignment API endpoints accept position parameter
- [ ] Position is configurable per match type (different sports may have different position counts)
- [ ] Database index on `assignments(match_id, position)` for query performance
- [ ] UI for creating assignments includes position dropdown

**Story Points**: 3

---

## Epic 2: Audit Logging & System Administration
**Goal**: Log all data-modifying actions, provide System Admins with searchable audit log viewer and export functionality.

**Priority**: Critical (security/compliance requirement)

**Dependencies**: Epic 1 (RBAC must exist to restrict audit log access to System Admins)

### Stories

#### Story 2.1: Audit Log Database Schema
**As a** system architect  
**I want** a database schema for audit logs  
**So that** all user actions are recorded with full context

**Acceptance Criteria**:
- [ ] `audit_logs` table created with fields: `id`, `user_id`, `action_type` (create/update/delete), `entity_type` (match, assignment, user, etc.), `entity_id`, `old_values` (JSON), `new_values` (JSON), `timestamp`, `ip_address`
- [ ] Retention policy field or timestamp-based partitioning for 2-year default
- [ ] Indexes on `user_id`, `entity_type`, `entity_id`, `timestamp`
- [ ] Table supports high write throughput (consider async writes)

**Story Points**: 3

---

#### Story 2.2: Audit Logging Middleware/Service
**As a** developer  
**I want** automatic audit logging for all CUD operations  
**So that** I don't have to manually log each action

**Acceptance Criteria**:
- [ ] Middleware or service function captures all create/update/delete operations
- [ ] Logs include: user, action, entity type, entity ID, old/new values, timestamp
- [ ] Writes to `audit_logs` table asynchronously (non-blocking)
- [ ] Handles JSON serialization of old/new values
- [ ] Does not log read-only operations (GET requests)
- [ ] Unit tests verify logging for all entity types

**Story Points**: 8

---

#### Story 2.3: Audit Log Viewer UI
**As a** System Admin  
**I want** a UI to view, search, and filter audit logs  
**So that** I can investigate security incidents and track changes

**Acceptance Criteria**:
- [ ] `/admin/audit-logs` page accessible only to System Admins
- [ ] Displays paginated table with columns: timestamp, user, action, entity type, entity ID
- [ ] Search by user, entity type, date range
- [ ] Filter by action type (create/update/delete)
- [ ] Click row to expand and view old/new values JSON
- [ ] Default view shows last 100 entries, sorted by timestamp descending
- [ ] Page loads in < 2 seconds with 10,000+ log entries

**Story Points**: 5

---

#### Story 2.4: Audit Log Export
**As a** System Admin  
**I want** to export audit logs as CSV or JSON  
**So that** I can perform offline analysis or provide reports to compliance

**Acceptance Criteria**:
- [ ] "Export" button on audit log viewer page
- [ ] Modal allows selection of format (CSV or JSON)
- [ ] Export respects current filters and search criteria
- [ ] CSV format includes all fields (flattened JSON for old/new values)
- [ ] JSON format is array of log objects
- [ ] Export limited to 10,000 records (warns if more exist)
- [ ] File downloads to user's browser

**Story Points**: 3

---

#### Story 2.5: Audit Log Retention Policy
**As a** System Admin  
**I want** audit logs older than 2 years to be automatically purged  
**So that** the database doesn't grow indefinitely

**Acceptance Criteria**:
- [ ] Retention period configurable via environment variable (default: 2 years)
- [ ] Scheduled job runs daily to delete logs older than retention period
- [ ] Deletion process handles large volumes without blocking
- [ ] Logs purge activity is itself logged (meta-audit)
- [ ] System Admin can manually trigger purge via UI or CLI

**Story Points**: 3

---

## Epic 3: UI/UX Modernization with TailwindCSS
**Goal**: Integrate TailwindCSS for responsive design, improve navigation with clickable components, meet WCAG 2.1 AA accessibility standards.

**Priority**: High (user experience improvement)

**Dependencies**: None (can proceed in parallel with other epics)

### Stories

#### Story 3.1: TailwindCSS Integration
**As a** frontend developer  
**I want** TailwindCSS integrated into the SvelteKit project  
**So that** I can use utility classes for responsive design

**Acceptance Criteria**:
- [ ] TailwindCSS installed and configured in SvelteKit project
- [ ] `tailwind.config.js` created with design tokens (colors, spacing, breakpoints)
- [ ] PostCSS configured to process Tailwind directives
- [ ] Existing custom CSS audited for conflicts
- [ ] Build process includes Tailwind purging for production
- [ ] Dev server hot-reloads Tailwind changes

**Story Points**: 3

---

#### Story 3.2: Responsive Layout Conversion
**As a** user  
**I want** the app to work seamlessly on mobile, tablet, and desktop  
**So that** I can access it from any device

**Acceptance Criteria**:
- [ ] All pages responsive from 320px (mobile) to 1920px (desktop)
- [ ] Navigation menu converts to hamburger on mobile
- [ ] Tables convert to stacked cards on mobile
- [ ] Forms adjust layout for mobile (single column vs. multi-column)
- [ ] Visual regression tests verify layouts at 320px, 768px, 1024px, 1920px
- [ ] No horizontal scrolling on any screen size

**Story Points**: 8

---

#### Story 3.3: Clickable Navigation Components
**As a** user  
**I want** cards and rows to be clickable for direct navigation  
**So that** I can access match details with one click

**Acceptance Criteria**:
- [ ] Dashboard match cards are clickable (navigate to match detail)
- [ ] Schedule table rows are clickable (navigate to match detail)
- [ ] Assignment cards are clickable (navigate to match detail)
- [ ] Hover state indicates clickability (cursor, background color)
- [ ] Keyboard navigation supports Enter key on focused card/row
- [ ] Screen reader announces links correctly

**Story Points**: 3

---

#### Story 3.4: Accessibility Improvements (WCAG 2.1 AA)
**As a** user with disabilities  
**I want** the app to be accessible via keyboard and screen reader  
**So that** I can use it independently

**Acceptance Criteria**:
- [ ] All interactive elements keyboard-accessible (tab order logical)
- [ ] Color contrast meets WCAG 2.1 AA (4.5:1 for text, 3:1 for UI components)
- [ ] Form inputs have labels and ARIA attributes
- [ ] Error messages announced by screen readers
- [ ] Skip-to-main-content link available
- [ ] Focus indicators visible on all interactive elements
- [ ] Tested with screen reader (NVDA or JAWS)

**Story Points**: 5

---

#### Story 3.5: Theming Support
**As a** user  
**I want** a consistent visual theme across the app  
**So that** it looks professional and cohesive

**Acceptance Criteria**:
- [ ] Design tokens defined in Tailwind config (primary, secondary, error, success colors)
- [ ] Buttons use consistent Tailwind classes
- [ ] Forms use consistent input styling
- [ ] Cards/panels use consistent spacing and shadows
- [ ] Typography scale defined (headings, body, small text)
- [ ] Dark mode groundwork (optional for V2, but tokens prepared)

**Story Points**: 3

---

## Epic 4: Match Archival & History
**Goal**: Automatically archive matches when referees submit final scores; provide history views for archived matches; implement 2-year retention policy.

**Priority**: High (referee experience improvement)

**Dependencies**: Epic 5 (match reporting must exist to trigger archival)

### Stories

#### Story 4.1: Match Archival Database Schema
**As a** system architect  
**I want** matches to have an `archived` status and `archived_at` timestamp  
**So that** archived matches can be tracked separately

**Acceptance Criteria**:
- [ ] `matches` table adds `archived` boolean field (default: false)
- [ ] `matches` table adds `archived_at` timestamp field (nullable)
- [ ] `matches` table adds `archived_by` user_id field (nullable)
- [ ] Migration handles existing matches (all set to `archived = false`)
- [ ] Index created on `archived` for efficient filtering

**Story Points**: 2

---

#### Story 4.2: Automatic Archival Logic
**As a** system  
**I want** matches to be automatically archived when a referee submits a final score  
**So that** completed matches are removed from active views

**Acceptance Criteria**:
- [ ] When match report submitted with final score, set `archived = true`
- [ ] Set `archived_at = current timestamp`
- [ ] Set `archived_by = user_id of referee who submitted score`
- [ ] Audit log entry created for archival action
- [ ] API endpoint `POST /api/matches/:id/archive` for manual archival (Assignor only)

**Story Points**: 3

---

#### Story 4.3: Filter Archived Matches from Active Views
**As a** user  
**I want** archived matches excluded from my dashboard and schedule  
**So that** I only see upcoming/active matches

**Acceptance Criteria**:
- [ ] Dashboard API filters `archived = false` by default
- [ ] Schedule page API filters `archived = false` by default
- [ ] Referee "My Matches" view filters `archived = false` by default
- [ ] No changes to assignment logic (archived matches still linked to referees for history)

**Story Points**: 2

---

#### Story 4.4: Archived Match History View (All Users)
**As a** user  
**I want** to view archived matches  
**So that** I can see past match results

**Acceptance Criteria**:
- [ ] New page `/matches/history` shows archived matches
- [ ] Paginated table with columns: date, home team, away team, final score, referee
- [ ] Search by team name, date range
- [ ] Click row to view full match details and report
- [ ] Page accessible to all authenticated users
- [ ] Loads in < 3 seconds with 1000+ archived matches

**Story Points**: 5

---

#### Story 4.5: Referee Match History View
**As a** referee  
**I want** to see all matches I've worked (active and archived)  
**So that** I can track my experience

**Acceptance Criteria**:
- [ ] New page `/referee/my-history` shows all matches assigned to current user
- [ ] Includes both active and archived matches
- [ ] Sorted by date descending
- [ ] Shows: date, home vs. away, final score (if archived), status (active/archived)
- [ ] Paginated (20 matches per page)
- [ ] Accessible only to users with Referee role

**Story Points**: 3

---

#### Story 4.6: Archived Match Retention Policy
**As a** System Admin  
**I want** archived matches older than 2 years to be purged  
**So that** the database doesn't grow indefinitely

**Acceptance Criteria**:
- [ ] Retention period configurable via environment variable (default: 2 years)
- [ ] Scheduled job runs monthly to delete matches where `archived_at` > retention period
- [ ] Also deletes associated match reports, assignments, and audit logs for those matches
- [ ] Deletion process is logged to audit log
- [ ] System Admin can manually trigger purge via UI or CLI

**Story Points**: 3

---

## Epic 5: Match Reporting by Referees
**Goal**: Enable referees to submit structured match reports (final score, red/yellow cards, injuries, notes); allow editing by referees and assignors; highlight updated assignments.

**Priority**: High (referee experience improvement)

**Dependencies**: Epic 1 (RBAC to restrict access)

### Stories

#### Story 5.1: Match Report Database Schema
**As a** system architect  
**I want** a database schema for match reports  
**So that** referees can record match outcomes

**Acceptance Criteria**:
- [ ] `match_reports` table created with fields: `id`, `match_id` (FK), `submitted_by` (user_id), `final_score_home`, `final_score_away`, `red_cards`, `yellow_cards`, `injuries`, `other_notes` (text), `submitted_at`, `updated_at`
- [ ] One-to-one relationship with `matches` (each match has max 1 report)
- [ ] Indexes on `match_id` and `submitted_by`
- [ ] Foreign key constraints ensure referential integrity

**Story Points**: 2

---

#### Story 5.2: Match Report Submission API
**As a** referee  
**I want** an API to submit match reports  
**So that** I can record outcomes after working a match

**Acceptance Criteria**:
- [ ] `POST /api/matches/:id/report` creates or updates match report
- [ ] Request body includes: `final_score_home`, `final_score_away`, `red_cards`, `yellow_cards`, `injuries`, `other_notes`
- [ ] Requires `can_edit_own_match_reports` permission
- [ ] Authorization check: user must be assigned to this match as CENTER referee (position = 'center')
- [ ] Assignors with `can_manage_users` permission can submit for any match
- [ ] Returns 403 if user lacks permission or is not center referee
- [ ] Audit log entry created for submission
- [ ] Triggers match archival (calls archival logic)

**Story Points**: 5

---

#### Story 5.3: Match Report Edit API
**As a** referee or assignor  
**I want** to edit submitted match reports  
**So that** I can correct errors

**Acceptance Criteria**:
- [ ] `PUT /api/matches/:id/report` updates existing match report
- [ ] Authorization: center referee for this match OR user with `can_manage_users` permission
- [ ] Returns 403 if user is not center referee and lacks management permission
- [ ] Audit log captures old/new values (including who made the edit)
- [ ] `updated_at` timestamp refreshed
- [ ] Cannot delete report via API (no DELETE endpoint)

**Story Points**: 3

---

#### Story 5.4: Match Report Submission UI
**As a** referee  
**I want** a form to submit match reports  
**So that** I can record outcomes without using the API directly

**Acceptance Criteria**:
- [ ] Match detail page shows "Submit Report" button only if user is assigned as CENTER referee for this match OR has assignor permissions
- [ ] Assistant referees see message: "Only the center referee can submit match reports"
- [ ] Form includes fields: final score (home/away), red cards (number), yellow cards (number), injuries (text), other notes (textarea)
- [ ] All fields optional except final score
- [ ] Success message displayed after submission
- [ ] Match automatically archived upon submission
- [ ] Form validates final score is numeric

**Story Points**: 5

---

#### Story 5.5: Match Report Edit UI
**As a** referee or assignor  
**I want** to edit submitted match reports via UI  
**So that** I can correct mistakes

**Acceptance Criteria**:
- [ ] Match detail page shows "Edit Report" button if report exists AND user is center referee or has assignor permissions
- [ ] Assistant referees cannot edit reports (button not shown)
- [ ] Pre-populates form with existing report data
- [ ] Save button updates report via API
- [ ] Success/error messages displayed
- [ ] No "Delete Report" button (only edit allowed)
- [ ] Form clearly indicates who originally submitted the report and when

**Story Points**: 3

---

#### Story 5.6: Assignment Change Indicator
**As a** referee  
**I want** to see a visual indicator when my assignment is updated  
**So that** I know to check for changes

**Acceptance Criteria**:
- [ ] `assignments` table adds `updated_at` timestamp
- [ ] `assignments` table adds `viewed_by_referee` boolean (default: false)
- [ ] When match details (time, location) updated via CSV import, set `viewed_by_referee = false` for all assigned referees
- [ ] "My Matches" page shows badge/icon on updated assignments
- [ ] Clicking into match detail sets `viewed_by_referee = true`
- [ ] Badge disappears after viewing

**Story Points**: 5

---

## Epic 6: CSV Import Enhancements
**Goal**: Implement deduplication using `reference_id`, update-in-place logic for re-imports, same-match detection, and filtering for practices/away matches.

**Priority**: High (assignor efficiency improvement)

**Dependencies**: Epic 1 (Assignor role required), Epic 2 (audit logging for import actions)

### Stories

#### Story 6.1: Reference ID Deduplication
**As an** assignor  
**I want** the CSV import to reject files with duplicate `reference_id` values  
**So that** I don't accidentally import the same match twice

**Acceptance Criteria**:
- [ ] CSV parser checks for duplicate `reference_id` in uploaded file
- [ ] Returns error message listing duplicate `reference_id` values
- [ ] Import aborts without creating any matches if duplicates found
- [ ] Error message suggests removing duplicates and re-uploading

**Story Points**: 3

---

#### Story 6.2: Update-in-Place for Re-Imports
**As an** assignor  
**I want** re-importing a CSV to update existing matches instead of creating duplicates  
**So that** I can correct schedule changes

**Acceptance Criteria**:
- [ ] CSV import checks if `reference_id` already exists in database
- [ ] If match exists, update fields: date, time, location, home team, away team
- [ ] If match does not exist, create new match
- [ ] Audit log records update action with old/new values
- [ ] Updated matches trigger assignment change indicator (Epic 5.6)
- [ ] Import summary shows: X created, Y updated, Z skipped

**Story Points**: 8

---

#### Story 6.3: Same-Match Detection
**As an** assignor  
**I want** the import to prevent creating duplicate entries when the same home/away teams play each other  
**So that** I don't have two records for the same match

**Acceptance Criteria**:
- [ ] CSV import checks if match with same home team, away team, and date already exists
- [ ] If detected, log warning and skip row (don't create duplicate)
- [ ] Import summary shows: "X duplicates detected and skipped"
- [ ] List skipped matches in import report with reason

**Story Points**: 5

---

#### Story 6.4: Filter Practices and Away Matches
**As an** assignor  
**I want** to exclude practices and away matches during import  
**So that** only relevant matches are created

**Acceptance Criteria**:
- [ ] CSV import identifies practices (e.g., home team or away team contains "Practice")
- [ ] CSV import identifies away matches (e.g., location is outside configured home region)
- [ ] Checkbox in import UI: "Filter out practices and away matches"
- [ ] If checked, skips rows matching filter criteria
- [ ] Import summary shows: "X practices filtered, Y away matches filtered"

**Story Points**: 5

---

#### Story 6.5: Mark Reference IDs as "Not of Concern"
**As an** assignor  
**I want** to mark certain `reference_id` values as "not of concern"  
**So that** they are automatically skipped on future imports

**Acceptance Criteria**:
- [ ] New table `excluded_reference_ids` with fields: `reference_id`, `reason`, `created_by`, `created_at`
- [ ] Import UI shows "Exclude from Future Imports" button on import preview
- [ ] Selecting rows and clicking button adds `reference_id` to exclusion table
- [ ] Future imports automatically skip excluded `reference_id` values
- [ ] Assignors can view/manage exclusion list in settings page
- [ ] Exclusions stored globally (not per-assignor)

**Story Points**: 5

---

#### Story 6.6: Import Summary Report
**As an** assignor  
**I want** a detailed summary after CSV import  
**So that** I know exactly what happened

**Acceptance Criteria**:
- [ ] Import completion page shows: total rows, created, updated, skipped, errors
- [ ] Lists skipped matches with reasons (duplicate, same-match, filtered, excluded)
- [ ] Lists errors with row numbers and error messages
- [ ] Download full report as CSV
- [ ] Report saved to audit log

**Story Points**: 3

---

## Epic 7: Scheduling Interface Improvements
**Goal**: Add weekend date range filtering, pagination, and scroll position retention to scheduling page.

**Priority**: Medium (assignor efficiency improvement)

**Dependencies**: Epic 3 (TailwindCSS for UI components)

### Stories

#### Story 7.1: Weekend Date Range Filter
**As an** assignor  
**I want** to filter matches by weekend date ranges  
**So that** I can assign referees for a full weekend at once

**Acceptance Criteria**:
- [ ] Filter UI includes "Weekend" option alongside existing "Single Day"
- [ ] Selecting "Weekend" shows date picker for weekend start date (Saturday)
- [ ] Filter includes Saturday and Sunday of selected weekend
- [ ] API endpoint accepts date range parameters
- [ ] Results refresh without page reload
- [ ] Clear filter button resets to all matches

**Story Points**: 5

---

#### Story 7.2: Pagination for Match Lists
**As an** assignor  
**I want** match lists to be paginated  
**So that** the page loads quickly with large schedules

**Acceptance Criteria**:
- [ ] Scheduling page displays 50 matches per page by default
- [ ] Pagination controls at bottom: Previous, page numbers, Next
- [ ] Page size selector: 25, 50, 100 matches
- [ ] URL query parameter reflects current page (?page=2)
- [ ] Navigating pages maintains filter selections
- [ ] Page loads in < 3 seconds with 1000+ total matches

**Story Points**: 5

---

#### Story 7.3: Scroll Position Retention
**As an** assignor  
**I want** the page to maintain my scroll position after assigning a referee  
**So that** I don't lose my place

**Acceptance Criteria**:
- [ ] After assignment action, page scrolls back to previous position
- [ ] Uses browser scroll restoration API or manual scroll tracking
- [ ] Works across pagination (scroll to top of new page is acceptable)
- [ ] Does not interfere with keyboard navigation
- [ ] Tested in Chrome, Firefox, Safari

**Story Points**: 3

---

## Epic 8: Backend Refactoring to Vertical Slice Architecture
**Goal**: Refactor backend from flat file structure to vertical slice architecture organized by feature/capability to improve maintainability and code organization.

**Priority**: High (foundational improvement enabling easier feature development)

**Dependencies**: Should be done early, ideally in parallel with Epic 1 or before other feature work begins

### Stories

#### Story 8.1: Define Vertical Slice Architecture & Project Structure
**As a** developer  
**I want** a clear project structure for vertical slices  
**So that** I know where to place code for each feature

**Acceptance Criteria**:
- [ ] Document target structure with feature slices: `features/matches/`, `features/assignments/`, `features/users/`, `features/auth/`, `features/audit/`, `features/reports/`
- [ ] Define shared packages: `shared/database/`, `shared/middleware/`, `shared/models/`, `shared/utils/`
- [ ] Each feature slice structure: `handler.go`, `service.go`, `repository.go`, `models.go`, `routes.go`
- [ ] Document naming conventions and separation of concerns
- [ ] Create ADR (Architecture Decision Record) documenting rationale
- [ ] Get team/stakeholder approval on structure

**Story Points**: 2

---

#### Story 8.2: Set Up Shared Infrastructure Packages
**As a** developer  
**I want** shared infrastructure extracted into common packages  
**So that** feature slices can depend on stable shared code

**Acceptance Criteria**:
- [ ] Create `shared/database/` package with DB connection, migrations runner
- [ ] Create `shared/middleware/` package for auth, logging, CORS, etc.
- [ ] Create `shared/config/` package for environment configuration
- [ ] Create `shared/errors/` package for standard error handling
- [ ] All shared packages have unit tests
- [ ] Existing code can compile against new shared packages

**Story Points**: 5

---

#### Story 8.3: Refactor Users Feature Slice
**As a** developer  
**I want** user-related code organized into a feature slice  
**So that** it's easier to find and maintain user functionality

**Acceptance Criteria**:
- [ ] Create `features/users/` directory
- [ ] Move user handlers from `user.go` to `features/users/handler.go`
- [ ] Create `features/users/service.go` for business logic
- [ ] Create `features/users/repository.go` for data access
- [ ] Create `features/users/models.go` for user-specific types
- [ ] Create `features/users/routes.go` to register routes
- [ ] All existing user endpoints still work (regression testing)
- [ ] Unit tests updated for new structure

**Story Points**: 8

---

#### Story 8.4: Refactor Matches Feature Slice
**As a** developer  
**I want** match-related code organized into a feature slice  
**So that** it's easier to find and maintain match functionality

**Acceptance Criteria**:
- [ ] Create `features/matches/` directory
- [ ] Move match handlers from `matches.go` to `features/matches/handler.go`
- [ ] Create `features/matches/service.go` for business logic
- [ ] Create `features/matches/repository.go` for data access
- [ ] Create `features/matches/models.go` for match-specific types
- [ ] Create `features/matches/routes.go` to register routes
- [ ] All existing match endpoints still work (regression testing)
- [ ] Unit tests updated for new structure

**Story Points**: 8

---

#### Story 8.5: Refactor Assignments Feature Slice
**As a** developer  
**I want** assignment-related code organized into a feature slice  
**So that** it's easier to find and maintain assignment functionality

**Acceptance Criteria**:
- [ ] Create `features/assignments/` directory
- [ ] Move assignment handlers from `assignments.go` to `features/assignments/handler.go`
- [ ] Create `features/assignments/service.go` for business logic
- [ ] Create `features/assignments/repository.go` for data access
- [ ] Create `features/assignments/models.go` for assignment-specific types
- [ ] Create `features/assignments/routes.go` to register routes
- [ ] All existing assignment endpoints still work (regression testing)
- [ ] Unit tests updated for new structure

**Story Points**: 8

---

#### Story 8.6: Refactor Remaining Feature Slices
**As a** developer  
**I want** all remaining features organized into slices  
**So that** the entire backend follows consistent structure

**Acceptance Criteria**:
- [ ] Create `features/referees/` slice (from `referees.go`)
- [ ] Create `features/availability/` slice (from `availability.go`, `day_unavailability.go`)
- [ ] Create `features/eligibility/` slice (from `eligibility.go`)
- [ ] Create `features/acknowledgment/` slice (from `acknowledgment.go`)
- [ ] Create `features/profile/` slice (from `profile.go`)
- [ ] Each slice has handler, service, repository, models, routes
- [ ] All existing endpoints still work (regression testing)
- [ ] Unit tests updated for new structure

**Story Points**: 13

---

#### Story 8.7: Update Main Entry Point & Router
**As a** developer  
**I want** `main.go` to be clean and register all feature routes  
**So that** the application structure is clear at a glance

**Acceptance Criteria**:
- [ ] `main.go` simplified to: config, DB setup, route registration, server start
- [ ] Each feature slice registers its own routes via `RegisterRoutes()` function
- [ ] Middleware applied at application level in `main.go`
- [ ] No business logic in `main.go`
- [ ] Clear initialization order documented
- [ ] Server starts successfully and all routes accessible

**Story Points**: 5

---

#### Story 8.8: Update Documentation & Developer Guide
**As a** new developer  
**I want** documentation explaining the vertical slice architecture  
**So that** I can contribute code following the pattern

**Acceptance Criteria**:
- [ ] Update README with architecture overview
- [ ] Create `ARCHITECTURE.md` documenting vertical slice pattern
- [ ] Document how to add a new feature slice
- [ ] Document separation of handler/service/repository responsibilities
- [ ] Include examples of typical workflows across layers
- [ ] Update CONTRIBUTING.md with code organization guidelines

**Story Points**: 3

---

#### Story 8.9: Clean Up & Remove Old Files
**As a** developer  
**I want** old flat structure files removed  
**So that** there's no confusion about where code lives

**Acceptance Criteria**:
- [ ] Delete old `.go` files from backend root (user.go, matches.go, etc.)
- [ ] Remove any unused imports or dead code discovered during refactor
- [ ] Ensure no duplicate code between old and new structure
- [ ] Git history preserved (use `git mv` where appropriate)
- [ ] Clean build with no warnings
- [ ] All tests pass

**Story Points**: 2

---

## Epic 9: Testing Infrastructure & CI/CD
**Goal**: Establish comprehensive unit testing for backend and frontend with CI/CD pipeline that blocks breaking changes; focus on business logic and critical paths.

**Priority**: High (quality assurance and regression prevention)

**Dependencies**: Should run in parallel with feature development; Epic 8 (backend refactoring) makes testing easier

**Philosophy**: Tests should validate business logic and critical user flows, not just achieve code coverage metrics.

### Stories

#### Story 9.1: Backend Test Infrastructure Setup
**As a** developer  
**I want** a testing framework configured for the Go backend  
**So that** I can write and run unit tests easily

**Acceptance Criteria**:
- [ ] Go's built-in `testing` package configured
- [ ] Test runner script created (`make test` or `go test ./...`)
- [ ] Mock framework installed for database mocking (e.g., `gomock` or `testify/mock`)
- [ ] Test database utilities for mocking DB calls (no real DB connections in unit tests)
- [ ] Directory structure for tests: `_test.go` files alongside source files
- [ ] Example table-driven test written as reference
- [ ] Documentation on writing tests added to CONTRIBUTING.md

**Story Points**: 3

---

#### Story 9.2: Frontend Test Infrastructure Setup
**As a** developer  
**I want** a testing framework configured for the SvelteKit frontend  
**So that** I can write unit and component tests

**Acceptance Criteria**:
- [ ] Vitest installed and configured in frontend project
- [ ] Testing Library for Svelte installed (`@testing-library/svelte`)
- [ ] Snapshot testing configured
- [ ] Test runner script created (`npm test`)
- [ ] Example component test written as reference
- [ ] Example unit test for utility functions written as reference
- [ ] Mock setup for API calls (MSW or fetch mocks)
- [ ] Documentation on writing tests added to frontend README

**Story Points**: 3

---

#### Story 9.3: Test Fixtures & Seed Data
**As a** developer  
**I want** reusable test fixtures and seed data  
**So that** I can easily set up test scenarios

**Acceptance Criteria**:
- [ ] Backend test fixtures: factory functions for users, matches, assignments, roles, permissions
- [ ] Frontend test fixtures: mock API responses for common entities
- [ ] Test seed data separate from production seed data
- [ ] Fixtures stored in `backend/tests/fixtures/` and `frontend/tests/fixtures/`
- [ ] Helper functions to create test data with sensible defaults
- [ ] Documentation on using fixtures in tests

**Story Points**: 3

---

#### Story 9.4: Backend Critical Path Tests - Authentication & RBAC
**As a** developer  
**I want** tests covering authentication and RBAC authorization  
**So that** security bugs are caught early

**Acceptance Criteria**:
- [ ] Tests for authentication middleware (valid/invalid tokens, expired tokens)
- [ ] Tests for permission-based authorization middleware (all permission checks)
- [ ] Tests for "most permissive wins" multi-role logic
- [ ] Tests for Super Admin auto-pass behavior
- [ ] Tests for role assignment API endpoints
- [ ] Tests verify 401/403 responses for unauthorized access
- [ ] Table-driven tests cover all permission combinations
- [ ] All tests use mocked database

**Story Points**: 8

---

#### Story 9.5: Backend Critical Path Tests - Match Assignment & CSV Import
**As a** developer  
**I want** tests covering match assignment and CSV import logic  
**So that** core assignor workflows are protected from regressions

**Acceptance Criteria**:
- [ ] Tests for CSV import deduplication (reference_id checks)
- [ ] Tests for CSV re-import update-in-place logic
- [ ] Tests for same-match detection
- [ ] Tests for practice/away match filtering
- [ ] Tests for assignment creation (including position field)
- [ ] Tests for assignment business rules (assignors only, etc.)
- [ ] Tests verify audit log entries are created
- [ ] All tests use mocked database

**Story Points**: 8

---

#### Story 9.6: Backend Critical Path Tests - Match Reporting & Archival
**As a** developer  
**I want** tests covering match reporting and archival logic  
**So that** referee workflows are protected from regressions

**Acceptance Criteria**:
- [ ] Tests for match report submission (center referee only)
- [ ] Tests for match report editing (center referee or assignor)
- [ ] Tests verify assistant referees cannot submit/edit reports
- [ ] Tests for automatic match archival on report submission
- [ ] Tests for archived match filtering from active views
- [ ] Tests for match history retrieval
- [ ] All tests use mocked database

**Story Points**: 5

---

#### Story 9.7: Frontend Critical Path Tests - User Flows
**As a** developer  
**I want** component tests covering critical user flows  
**So that** UI regressions are caught early

**Acceptance Criteria**:
- [ ] Component tests for login flow
- [ ] Component tests for role assignment UI (System Admin)
- [ ] Component tests for match assignment workflow (Assignor)
- [ ] Component tests for match report submission (Referee)
- [ ] Component tests for CSV import UI
- [ ] Tests verify correct API calls are made (mocked)
- [ ] Tests verify error handling and user feedback
- [ ] Snapshot tests for key UI components

**Story Points**: 8

---

#### Story 9.8: CI/CD Pipeline - GitHub Actions
**As a** developer  
**I want** automated tests running on pull requests  
**So that** breaking changes are caught before merge

**Acceptance Criteria**:
- [ ] GitHub Actions workflow created (`.github/workflows/test.yml`)
- [ ] Workflow triggers on pull requests to main branch
- [ ] Backend tests run: `cd backend && go test ./...`
- [ ] Frontend tests run: `cd frontend && npm test`
- [ ] Workflow fails if any tests fail (exit code non-zero)
- [ ] Branch protection rule: require passing tests before merge to main
- [ ] Test results displayed in PR checks
- [ ] Workflow runs in reasonable time (< 5 minutes)

**Story Points**: 5

---

#### Story 9.9: Test Documentation & Developer Guide
**As a** new developer  
**I want** clear documentation on writing and running tests  
**So that** I can contribute quality code with tests

**Acceptance Criteria**:
- [ ] TESTING.md document created with:
  - Philosophy: test business logic, not code coverage
  - How to run tests (backend, frontend, both)
  - How to write new tests (examples for each layer)
  - How to use test fixtures
  - Table-driven test patterns
  - Component test patterns
- [ ] CONTRIBUTING.md updated: require tests for new features
- [ ] README updated: mention test commands
- [ ] Examples of good tests highlighted in codebase

**Story Points**: 2

---

## Dependency Map

```
Epic 8 (Backend Refactoring) - foundational, should run early/parallel with Epic 1
  └─> Makes Epic 9 (testing) easier; affects all other epics

Epic 9 (Testing Infrastructure) - can run in parallel with feature development
  └─> Tests written alongside feature implementation

Epic 1 (RBAC)
  ├─> Epic 2 (Audit Logging) - requires RBAC for access control
  ├─> Epic 5 (Match Reporting) - requires RBAC for authorization
  ├─> Epic 6 (CSV Import) - requires Assignor role
  └─> Epic 9 (Testing) - RBAC tests validate permission system

Epic 3 (UI/UX) - independent, can run in parallel
  └─> Epic 7 (Scheduling UI) - benefits from TailwindCSS

Epic 5 (Match Reporting)
  └─> Epic 4 (Match Archival) - archival triggered by report submission
```

## Recommended Build Order

### Phase 0: Foundation (Weeks 1-4)
**Option A (Recommended)**: Complete Epic 8 first for cleaner feature development
1. **Epic 8: Backend Refactoring** (Stories 8.1 → 8.9) - Complete vertical slice migration
2. **Epic 9: Testing Infrastructure** (Stories 9.1 → 9.3, 9.8, 9.9) - Set up test frameworks and CI/CD

**Option B**: Run Epic 8 in parallel with Phase 1 if two developers available

### Phase 1: Security & Infrastructure (Weeks 5-8 or 2-5 depending on Option)
3. **Epic 1: RBAC** (Stories 1.1 → 1.7)
4. **Epic 2: Audit Logging** (Stories 2.1 → 2.2)
5. **Epic 9: RBAC Tests** (Story 9.4) - Run parallel with Epic 1 implementation

### Phase 2: Core Features (Weeks 9-13 or 6-10)
6. **Epic 3: UI/UX** (Stories 3.1 → 3.5) - can run parallel with Epic 1 if resources allow
7. **Epic 5: Match Reporting** (Stories 5.1 → 5.6)
8. **Epic 4: Match Archival** (Stories 4.1 → 4.6)
9. **Epic 9: Match Reporting & Archival Tests** (Story 9.6) - Run parallel with Epics 4-5

### Phase 3: Assignor Tools (Weeks 14-16 or 11-13)
10. **Epic 6: CSV Import** (Stories 6.1 → 6.6)
11. **Epic 7: Scheduling UI** (Stories 7.1 → 7.3)
12. **Epic 9: CSV Import Tests** (Story 9.5) - Run parallel with Epic 6

### Phase 4: Frontend Testing & Polish (Weeks 17-19 or 14-16)
13. **Epic 2 Completion**: Audit Log Viewer & Export (Stories 2.3 → 2.5)
14. **Epic 9: Frontend Tests** (Story 9.7) - Test critical user flows
15. Integration testing, bug fixes, performance optimization
16. User acceptance testing (UAT) with assignors and referees

**Recommendation**: 
- Follow Option A (complete Epic 8 first) to avoid refactoring while building new features
- Write tests alongside feature implementation (Stories 9.4-9.7 run parallel with their respective epics)
- Set up CI/CD early (Stories 9.1-9.3, 9.8-9.9 in Phase 0) so all subsequent code is automatically tested

## Story Point Summary
- **Epic 1**: 30 points (RBAC - 7 stories: 5+3+8+5+5+3+1)
- **Epic 2**: 22 points (Audit Logging)
- **Epic 3**: 22 points (UI/UX)
- **Epic 4**: 18 points (Match Archival)
- **Epic 5**: 23 points (Match Reporting)
- **Epic 6**: 29 points (CSV Import)
- **Epic 7**: 13 points (Scheduling UI)
- **Epic 8**: 54 points (Backend Refactoring)
- **Epic 9**: 45 points (Testing Infrastructure - 9 stories: 3+3+3+8+8+5+8+5+2)

**Total**: 256 story points

**Estimated Timeline**: 
- **Option A (Sequential)**: 19-22 weeks (13 points/week for 1 developer)
- **Option B (Parallel)**: 14-16 weeks (with 2 developers: one on Epic 8, one on Epics 1-2; tests written alongside features)
- **Aggressive (2 developers)**: 13-15 weeks (both developers working across all epics; tests integrated into feature work)

## Success Criteria
- [ ] Zero unauthorized data access incidents (Security NFR)
- [ ] All API endpoints enforce RBAC (Security NFR)
- [ ] WCAG 2.1 AA compliance verified (Accessibility NFR)
- [ ] Scheduling page loads in < 3s with 1000+ matches (Performance NFR)
- [ ] CSV import processes 500 matches in < 10s (Performance NFR)
- [ ] < 1% duplicate match creation rate (Data Integrity NFR)
- [ ] 50% reduction in assignor time for import/assignment workflows (Usability NFR)
- [ ] 80%+ referee satisfaction in post-season survey (User Satisfaction)
- [ ] Backend organized by feature with clear separation of concerns (Maintainability NFR)
- [ ] New features can be added without modifying existing feature slices (Maintainability NFR)
- [ ] Critical business logic paths covered by automated unit tests (Testability NFR)
- [ ] CI/CD pipeline blocks PRs with failing tests (Quality NFR)
- [ ] All new features include tests (Development Practice)
