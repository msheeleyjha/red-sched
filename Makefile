.PHONY: help build up down logs clean seed-assignor

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all containers
	docker-compose build

up: ## Start all services
	docker-compose up -d
	@echo ""
	@echo "✅ Services started!"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend:  http://localhost:8080"
	@echo "Health:   http://localhost:8080/health"

down: ## Stop all services
	docker-compose down

logs: ## Follow logs from all services
	docker-compose logs -f

clean: ## Stop services and remove volumes
	docker-compose down -v
	@echo "⚠️  All data has been removed!"

seed-assignor: ## Create an assignor user (requires email)
	@read -p "Enter email address: " email; \
	docker exec -it referee-scheduler-db psql -U referee_scheduler -c \
		"UPDATE users SET role = 'assignor', status = 'active' WHERE email = '$$email';"
	@echo "✅ User updated to assignor role"

seed-superadmin: ## Assign Super Admin role to user by email (RBAC V2)
	@read -p "Enter email address: " email; \
	docker exec referee-scheduler-db psql -U referee_scheduler -c " \
		WITH user_info AS ( \
		    SELECT id FROM users WHERE email = '$$email' \
		), role_info AS ( \
		    SELECT id FROM roles WHERE name = 'Super Admin' \
		) \
		INSERT INTO user_roles (user_id, role_id, assigned_by) \
		SELECT u.id, r.id, u.id \
		FROM user_info u, role_info r \
		ON CONFLICT (user_id, role_id) DO NOTHING; \
		UPDATE users SET role = 'assignor', status = 'active' WHERE email = '$$email'; \
		SELECT \
		    CASE \
		        WHEN EXISTS (SELECT 1 FROM user_roles ur \
		                     JOIN users u ON ur.user_id = u.id \
		                     JOIN roles r ON ur.role_id = r.id \
		                     WHERE u.email = '$$email' AND r.name = 'Super Admin') \
		        THEN '✓ Super Admin role assigned to ' || '$$email' \
		        ELSE '✗ User not found: ' || '$$email' \
		    END as result; \
	"
	@echo "✅ Done"

db-shell: ## Open PostgreSQL shell
	docker exec -it referee-scheduler-db psql -U referee_scheduler

backend-logs: ## Follow backend logs
	docker-compose logs -f backend

frontend-logs: ## Follow frontend logs
	docker-compose logs -f frontend
