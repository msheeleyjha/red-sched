package audit

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all audit log routes
func (h *Handler) RegisterRoutes(
	r *mux.Router,
	requirePermission func(string, http.HandlerFunc) http.HandlerFunc,
) {
	r.HandleFunc("/api/admin/audit-logs", requirePermission("can_view_audit_logs", h.GetAuditLogs)).Methods("GET")
	r.HandleFunc("/api/admin/audit-logs/export", requirePermission("can_view_audit_logs", h.ExportAuditLogs)).Methods("GET")
	r.HandleFunc("/api/admin/audit-logs/purge", requirePermission("can_view_audit_logs", h.PurgeAuditLogs)).Methods("POST")
}
