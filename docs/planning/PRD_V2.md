# PRD: Referee Scheduler V2

## Objective
V2 enhances the referee scheduling system by implementing role-based access control (RBAC), improving UI/UX with responsive design, tightening API security to prevent unauthorized data access, enabling referees to report match outcomes and view their history, and improving assignor workflows through better CSV import handling and scheduling interface enhancements. System administrators gain visibility into audit logs and system configuration. The outcome is increased referee satisfaction, reduced time spent by assignors managing schedules, and improved security posture.

## Background & Context
V1 successfully delivered core referee scheduling functionality but revealed several pain points:
- **Security gaps**: No role-based access control (RBAC); new users can access match data before assignor approval; API endpoints expose data users shouldn't see; no system admin role to manage configuration
- **UI limitations**: Non-responsive design, poor mobile experience, navigation requires excessive clicks
- **Workflow friction**: CSV re-imports create duplicates; assignors lack filtering and pagination on scheduling screen; no way to track data changes or audit user actions
- **Incomplete referee experience**: No match history view, no ability to report outcomes, played matches clutter active schedules
- **Data integrity risks**: Duplicate imports and same-match detection failures compromise data quality
- **Maintainability issues**: Flat backend structure (all .go files in single directory) becoming difficult to navigate and maintain as codebase grows

V2 addresses these gaps by introducing RBAC, audit logging, system administration capabilities, and refactoring the backend to a vertical slice architecture pattern while maintaining the existing Go + SvelteKit stack.

## Users
| Persona | Description | Primary need |
|---------|-------------|--------------|
| System Admin | System administrator with full access to audit logs, system configuration, and all assignor/referee functions | System observability, configuration management, security oversight |
| Assignor | League administrator who imports schedules, assigns referees, and manages users | Efficient bulk import and assignment workflows; security control over user access |
| Referee | Officials who receive assignments and work matches | Clear view of upcoming matches, ability to report outcomes, access to work history |

**Note**: System Admins can also function as Assignors and/or Referees. Users can hold multiple roles simultaneously.

## Scope
### In scope (V2)
**All Users:**
- TailwindCSS integration for responsive design and theming
- API security hardening (new user approval gate, endpoint authorization fixes)
- Match archival (automatic when final score submitted; history view; removal from active schedules/dashboards)
- Audit logging for all data-modifying user actions
- Navigation improvements (clickable cards/components for direct navigation)

**Referees:**
- Match outcome reporting (final score, infractions, issues/notes)
- Editable match reports (by referee and assignor)
- Match history view (all previously worked matches)
- Change notification/highlighting for updated assignments

**Assignors:**
- CSV import deduplication using `reference_id`
- CSV re-import update logic (modify existing matches instead of creating duplicates)
- Same-match detection improvements (prevent duplicate entries for same home vs. away matchup)
- CSV filtering (exclude practices, away matches; allow marking `reference_id` as "not of concern")
- Scheduling UI enhancements:
  - Advanced filters (weekend ranges, not just single days)
  - Screen state management (maintain scroll position after actions)
  - Pagination

**System Admins:**
- Role-based access control (RBAC) framework implementation
- Audit log viewer (search, filter, export capabilities)
- User role assignment and management
- Full access to all Assignor and Referee functions

**Backend Architecture:**
- Refactor backend from flat structure to vertical slice architecture
- Organize code by feature/capability (matches, assignments, users, etc.)
- Each feature slice contains its own handlers, services, repositories, and models
- Shared infrastructure (middleware, database, auth) in common packages
- Improve code maintainability and feature isolation

**Testing & Quality Assurance:**
- Unit test infrastructure for backend (Go) and frontend (SvelteKit/TypeScript)
- Tests focus on business logic and critical paths
- Mock database calls for unit tests
- Component tests for frontend user flows
- Test fixtures and seed data separate from production
- CI/CD pipeline blocks PR merge if tests fail

### Out of scope (V2)
- Real-time/email notifications (stretch goal: configurable opt-in notifications for referees)
- System configuration management UI (future enhancement)
- Payment processing for referees
- Automated referee availability tracking
- Mobile native applications
- Integration with third-party league management systems
- Referee certification/credential tracking
- SMS reminders
- File attachments for match reports (stretch goal)

## Functional Requirements
1. **FR-1**: System SHALL prevent new users from performing any actions (except editing their own profile) until a System Admin assigns them appropriate roles
2. **FR-2**: System SHALL enforce permission-based authorization on all API endpoints to prevent users from accessing data or actions they don't have permissions for
3. **FR-3**: System SHALL automatically archive matches when a referee submits a final score
4. **FR-4**: System SHALL provide a history view showing archived matches with configurable retention (default: 2 years)
5. **FR-5**: System SHALL remove archived matches from active schedule and dashboard views
6. **FR-6**: Referees SHALL be able to submit match reports containing structured fields: final score, red cards, yellow cards, injuries, and other notes
7. **FR-7**: Referees and assignors SHALL be able to edit submitted match reports; referees SHALL NOT be able to delete match reports
8. **FR-8**: Referees SHALL be able to view a history of all matches they have worked
9. **FR-9**: System SHALL log all user actions that create, update, or delete data (audit trail)
10. **FR-10**: Audit logs SHALL capture: user, action type, timestamp, affected entity, old/new values with configurable retention (default: 2 years)
11. **FR-11**: CSV import SHALL reject files containing duplicate `reference_id` values
12. **FR-12**: CSV re-import SHALL update existing matches (by `reference_id`) instead of creating duplicates
13. **FR-13**: CSV import SHALL detect and prevent duplicate entries when same home/away teams play each other
14. **FR-14**: CSV import SHALL allow filtering out practices and away matches
15. **FR-15**: Assignors SHALL be able to mark `reference_id` values as "not of concern" globally to exclude them from future imports
16. **FR-16**: Scheduling interface SHALL support filtering by weekend date ranges (not just single days)
17. **FR-17**: Scheduling interface SHALL maintain scroll position after user actions
18. **FR-18**: Scheduling interface SHALL support pagination for large match lists
19. **FR-19**: Dashboard and list views SHALL have clickable cards/rows for direct navigation to match details
20. **FR-20**: When a match is updated via CSV re-import, assigned referees SHALL be notified or shown a clear visual indicator of the change
21. **FR-21**: System SHALL implement role-based access control (RBAC) with a 5-table structure: users, roles, permissions, user_roles, role_permissions
22. **FR-22**: System SHALL support three initial roles: Super Admin, Assignor, Referee
23. **FR-23**: Users SHALL be assignable to multiple roles simultaneously
24. **FR-24**: Each role SHALL have a set of permissions that define allowed actions (e.g., can_import_matches, can_assign_referees, can_submit_match_reports)
25. **FR-25**: User authorization SHALL follow "most permissive wins" - user has permission if ANY of their roles includes it
26. **FR-26**: Super Admin role SHALL automatically pass all permission checks
27. **FR-27**: Permissions SHALL have technical names in backend/database and display-friendly names in UI
28. **FR-28**: System Admins SHALL be able to view, search, and filter audit logs through a dedicated UI
29. **FR-29**: System Admins SHALL be able to export audit logs (CSV/JSON format)
30. **FR-30**: System Admins SHALL be able to assign and revoke roles for any user
31. **FR-31**: API endpoints SHALL enforce permission-based authorization (e.g., can_import_matches permission required for CSV import)
32. **FR-32**: Assignment records SHALL include a position field (center, assistant_1, assistant_2, etc.) to track referee role per match
33. **FR-33**: Match report editing SHALL be restricted to referees assigned as CENTER referee for that match
34. **FR-34**: Backend code SHALL be organized using vertical slice architecture with feature-based modules
35. **FR-35**: Each feature slice SHALL contain its own handlers, services, repositories, and models
36. **FR-36**: Shared infrastructure (middleware, database, auth) SHALL be in common/shared packages
37. **FR-37**: Backend SHALL have unit tests for business logic in handlers, services, repositories, and middleware using Go's testing framework
38. **FR-38**: Frontend SHALL have unit tests for utilities and component tests for user flows using Vitest and Testing Library
39. **FR-39**: Tests SHALL use mocked database calls (not real DB connections) for isolation and speed
40. **FR-40**: Test fixtures and seed data SHALL be separate from production seed data
41. **FR-41**: CI/CD pipeline SHALL run all tests on pull requests and block merge to main branch if any tests fail
42. **FR-42**: Critical paths SHALL be tested first: authentication, RBAC authorization, match assignment, CSV import, match reporting
43. **FR-43 (Stretch)**: System SHALL support configurable opt-in notifications for referees (new assignment, schedule change, match reminder events)
44. **FR-44 (Stretch)**: Match reports SHALL support file attachments (images, PDFs)

## Non-Functional Requirements
| Category | Requirement | Priority |
|----------|-------------|----------|
| Security | All API endpoints enforce role-based access control; new users locked until approved | Critical |
| Accessibility | UI meets WCAG 2.1 AA standards; keyboard navigation; screen reader support | Critical |
| Performance | Scheduling page with 1000+ matches loads in < 3 seconds | High |
| Performance | CSV import processes 500 matches in < 10 seconds | High |
| Responsiveness | UI is fully functional on mobile (320px width), tablet, and desktop viewports | High |
| Data Integrity | No duplicate matches created via import; all state changes audited | Critical |
| Availability | 99% uptime during peak assignment windows | Medium |
| Usability | Assignor can complete bulk import and assignment in 50% less time than V1 | High |
| Maintainability | Backend code organized by feature with clear separation of concerns; new features can be added without modifying existing slices | High |
| Testability | Critical business logic paths covered by automated unit tests; CI/CD blocks breaking changes | High |

## Integrations & Dependencies
- **CSV Import**: Continue to accept CSV files with fields including `reference_id`, home team, away team, date, time, location
- No external API integrations required for V2
- **Stretch**: Email/SMS provider for notification delivery (TBD if implemented)

## Success Metrics
| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| Referee Satisfaction | 80%+ satisfaction rating | Post-season survey |
| Assignor Time Reduction | 50% reduction in time to import and assign matches | User interviews + session analytics |
| Security Incidents | Zero unauthorized data access incidents | Audit log review |
| Import Errors | < 1% duplicate match creation rate | Error logs + database queries |
| Mobile Usage | 40%+ of referee traffic from mobile devices | Analytics |

## Risks & Assumptions
| Risk/Assumption | Likelihood | Impact | Mitigation |
|-----------------|------------|--------|------------|
| Backend refactoring introduces regressions or breaks existing functionality | High | High | Comprehensive test coverage before refactor; incremental migration by feature; parallel testing against V1 endpoints |
| RBAC implementation breaks existing user workflows | Medium | High | Feature flags; gradual rollout; comprehensive authorization testing |
| Data integrity issues during migration to archival system | Medium | High | Dry-run migration script; backup before deploy; phased rollout |
| Existing users need to be migrated to new role system | High | Medium | Migration script assigns appropriate roles based on current permissions; manual review by System Admin |
| TailwindCSS migration breaks existing styles | Medium | Medium | Visual regression testing; incremental component migration |
| CSV reference_id collisions across different leagues/seasons | Low | High | Validation during import; ensure reference_id is globally unique or add league/season prefix |
| Performance degradation with audit logging | Medium | Medium | Async audit writes; database indexing; retention policy |
| User approval workflow blocks legitimate referees | Low | Medium | Admin dashboard with pending approval queue; email notification to assignors |
| Stretch goals (notifications, file attachments) add scope creep | High | Low | Clearly separate as phase 2; deprioritize in favor of core features |

## Open Questions
None - all questions resolved during requirements gathering.

## Out of Scope (V1)
- Payment processing or invoicing for referees
- Automated conflict detection (referee working overlapping matches)
- Mobile native applications (iOS/Android)
- Integration with third-party scheduling platforms
- Referee skill-based auto-assignment
- Public-facing match schedule (parent/coach view)
- Multi-language support
- Calendar export (iCal/Google Calendar integration)
