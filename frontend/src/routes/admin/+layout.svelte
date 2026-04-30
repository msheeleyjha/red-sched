<script lang="ts">
	import { page } from '$app/stores';
	import { goto, invalidateAll } from '$app/navigation';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

	async function handleLogout() {
		try {
			await fetch(`${API_URL}/api/auth/logout`, {
				method: 'POST',
				credentials: 'include'
			});
			await invalidateAll();
			goto('/');
		} catch (error) {
			console.error('Logout error:', error);
		}
	}

	$: currentPath = $page.url.pathname;
</script>

<div class="admin-layout">
	<header class="admin-header">
		<div class="header-left">
			<a href="/dashboard" class="back-link">
				<img src="/logo.svg" alt="Logo" class="header-logo" />
			</a>
			<h1>Admin</h1>
		</div>
		<nav class="admin-nav">
			<a
				href="/admin/users"
				class="nav-link"
				class:active={currentPath === '/admin/users'}
			>
				Users & Roles
			</a>
			<a
				href="/admin/audit-logs"
				class="nav-link"
				class:active={currentPath === '/admin/audit-logs'}
			>
				Audit Logs
			</a>
		</nav>
		<div class="header-right">
			<a href="/dashboard" class="btn btn-secondary">Dashboard</a>
			<button on:click={handleLogout} class="btn btn-secondary">Sign Out</button>
		</div>
	</header>

	<main class="admin-content">
		<slot />
	</main>
</div>

<style>
	.admin-layout {
		min-height: 100vh;
	}

	.admin-header {
		display: flex;
		align-items: center;
		gap: 1.5rem;
		max-width: 1200px;
		margin: 0 auto;
		padding: 1rem;
		flex-wrap: wrap;
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.back-link {
		display: flex;
		align-items: center;
		text-decoration: none;
	}

	.header-logo {
		height: 36px;
		width: auto;
	}

	h1 {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.admin-nav {
		display: flex;
		gap: 0.25rem;
		flex: 1;
	}

	.nav-link {
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text-secondary);
		text-decoration: none;
		transition: all 0.2s;
	}

	.nav-link:hover {
		background-color: var(--bg-primary);
		color: var(--text-primary);
		text-decoration: none;
	}

	.nav-link.active {
		background-color: var(--primary-color);
		color: white;
	}

	.header-right {
		display: flex;
		gap: 0.5rem;
	}

	.admin-content {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 1rem 2rem;
	}

	@media (max-width: 768px) {
		.admin-header {
			padding: 0.75rem 0.5rem;
			gap: 0.75rem;
		}

		h1 {
			font-size: 1.25rem;
		}

		.admin-nav {
			order: 3;
			width: 100%;
		}

		.nav-link {
			flex: 1;
			text-align: center;
		}

		.admin-content {
			padding: 0 0.5rem 1.5rem;
		}
	}
</style>
