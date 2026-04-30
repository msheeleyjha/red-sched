<script lang="ts">
	import { onMount } from 'svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

	interface Role {
		id: number;
		name: string;
		description: string;
		permissions?: string[];
	}

	interface Permission {
		id: number;
		name: string;
		display_name: string;
		description: string;
		resource: string;
		action: string;
	}

	interface User {
		id: number;
		name: string;
		email: string;
		roles?: Role[];
	}

	let loading = true;
	let users: User[] = [];
	let allRoles: Role[] = [];
	let allPermissions: Permission[] = [];
	let error = '';
	let success = '';

	// Modal state
	let showRoleModal = false;
	let selectedUser: User | null = null;
	let selectedRoles: number[] = [];
	let savingRoles = false;

	$: currentUserId = data.user?.id;
	$: currentUserIsSuperAdmin = data.user?.role === 'assignor' || false; // TODO: Update when RBAC is live

	onMount(async () => {
		await Promise.all([loadUsers(), loadRoles(), loadPermissions()]);
	});

	async function loadUsers() {
		loading = true;
		error = '';

		try {
			const response = await fetch(`${API_URL}/api/referees`, {
				credentials: 'include'
			});

			if (response.ok) {
				users = await response.json();
				// Load roles for each user
				for (const user of users) {
					await loadUserRoles(user);
				}
			} else if (response.status === 403) {
				error = 'Access denied. Only System Admins can manage user roles.';
			} else {
				error = 'Failed to load users';
			}
		} catch (err) {
			error = 'Failed to load users';
		} finally {
			loading = false;
		}
	}

	async function loadUserRoles(user: User) {
		try {
			const response = await fetch(`${API_URL}/api/admin/users/${user.id}/roles`, {
				credentials: 'include'
			});

			if (response.ok) {
				const data = await response.json();
				user.roles = data.roles || [];
			}
		} catch (err) {
			console.error(`Failed to load roles for user ${user.id}`);
		}
	}

	async function loadRoles() {
		try {
			const response = await fetch(`${API_URL}/api/admin/roles`, {
				credentials: 'include'
			});

			if (response.ok) {
				allRoles = await response.json();
			}
		} catch (err) {
			console.error('Failed to load roles');
		}
	}

	async function loadPermissions() {
		try {
			const response = await fetch(`${API_URL}/api/admin/permissions`, {
				credentials: 'include'
			});

			if (response.ok) {
				allPermissions = await response.json();
			}
		} catch (err) {
			console.error('Failed to load permissions');
		}
	}

	function openRoleModal(user: User) {
		selectedUser = user;
		selectedRoles = user.roles?.map(r => r.id) || [];
		showRoleModal = true;
	}

	function closeRoleModal() {
		showRoleModal = false;
		selectedUser = null;
		selectedRoles = [];
	}

	async function saveRoles() {
		if (!selectedUser) return;

		savingRoles = true;
		error = '';
		success = '';

		try {
			const currentRoleIds = selectedUser.roles?.map(r => r.id) || [];

			// Find roles to add
			const rolesToAdd = selectedRoles.filter(roleId => !currentRoleIds.includes(roleId));

			// Find roles to remove
			const rolesToRemove = currentRoleIds.filter(roleId => !selectedRoles.includes(roleId));

			// Add new roles
			for (const roleId of rolesToAdd) {
				const response = await fetch(`${API_URL}/api/admin/users/${selectedUser.id}/roles`, {
					method: 'POST',
					credentials: 'include',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ role_id: roleId })
				});

				if (!response.ok) {
					const data = await response.json();
					throw new Error(data.error || 'Failed to assign role');
				}
			}

			// Remove roles
			for (const roleId of rolesToRemove) {
				// Prevent removing own Super Admin role
				const role = allRoles.find(r => r.id === roleId);
				if (role?.name === 'Super Admin' && selectedUser.id === currentUserId) {
					error = 'Cannot remove your own Super Admin role';
					savingRoles = false;
					return;
				}

				const response = await fetch(`${API_URL}/api/admin/users/${selectedUser.id}/roles/${roleId}`, {
					method: 'DELETE',
					credentials: 'include'
				});

				if (!response.ok) {
					const data = await response.json();
					throw new Error(data.error || 'Failed to revoke role');
				}
			}

			success = `Roles updated for ${selectedUser.name}`;
			await loadUserRoles(selectedUser);
			closeRoleModal();

			// Clear success message after 3 seconds
			setTimeout(() => { success = ''; }, 3000);
		} catch (err: any) {
			error = err.message || 'Failed to update roles';
		} finally {
			savingRoles = false;
		}
	}

	function getRoleName(roleId: number): string {
		return allRoles.find(r => r.id === roleId)?.name || 'Unknown';
	}

	function getRolePermissions(roleId: number): string[] {
		return allRoles.find(r => r.id === roleId)?.permissions || [];
	}

	function getPermissionDisplayName(permName: string): string {
		return allPermissions.find(p => p.name === permName)?.display_name || permName;
	}
</script>

<svelte:head>
	<title>User Role Management - Admin - Referee Scheduler</title>
</svelte:head>

<div class="page-header">
	<h2>User Role Management</h2>
</div>

{#if error}
	<div class="alert alert-error" role="alert">{error}</div>
{/if}

{#if success}
	<div class="alert alert-success" role="alert">{success}</div>
{/if}

{#if !currentUserIsSuperAdmin}
	<div class="alert alert-warning" role="alert">
		<strong>Access Restricted</strong>
		<p>You need System Admin privileges to manage user roles.</p>
	</div>
{/if}

{#if loading}
	<div class="loading-text">Loading users...</div>
{:else}
	<div class="card table-card">
		<table class="data-table">
			<thead>
				<tr>
					<th>User</th>
					<th>Email</th>
					<th>Roles</th>
					<th>Actions</th>
				</tr>
			</thead>
			<tbody>
				{#each users as user}
					<tr>
						<td>{user.name}</td>
						<td>{user.email}</td>
						<td>
							{#if user.roles && user.roles.length > 0}
								<div class="role-badges">
									{#each user.roles as role}
										<span class="badge">{role.name}</span>
									{/each}
								</div>
							{:else}
								<span class="text-muted">No Roles (Profile Only)</span>
							{/if}
						</td>
						<td>
							<button
								on:click={() => openRoleModal(user)}
								class="link-btn"
								disabled={!currentUserIsSuperAdmin}
							>
								Manage Roles
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	{#if users.length === 0}
		<div class="empty-state">No users found</div>
	{/if}
{/if}

{#if showRoleModal && selectedUser}
	<div class="modal-overlay" on:click={closeRoleModal}>
		<div class="modal" on:click|stopPropagation>
			<h3>Manage Roles for {selectedUser.name}</h3>

			{#if selectedUser.id === currentUserId && !selectedRoles.some(id => allRoles.find(r => r.id === id)?.name === 'Super Admin')}
				<div class="alert alert-warning">
					<p>Warning: If you remove your own Super Admin role, you will lose access to this page.</p>
				</div>
			{/if}

			<p class="modal-description">Select roles to assign to this user:</p>

			<div class="role-list">
				{#each allRoles as role}
					<label class="role-option">
						<input
							type="checkbox"
							value={role.id}
							bind:group={selectedRoles}
						/>
						<div class="role-details">
							<div class="role-name">{role.name}</div>
							<div class="role-description">{role.description}</div>

							{#if role.permissions && role.permissions.length > 0}
								<div class="permissions-list">
									<strong>Permissions:</strong>
									<ul>
										{#each role.permissions as perm}
											<li>{getPermissionDisplayName(perm)}</li>
										{/each}
									</ul>
								</div>
							{/if}
						</div>
					</label>
				{/each}
			</div>

			{#if selectedRoles.length === 0 && selectedUser.id !== currentUserId}
				<div class="alert alert-info">
					<p>No roles assigned. User will only be able to edit their profile until roles are assigned.</p>
				</div>
			{/if}

			<div class="modal-actions">
				<button
					on:click={closeRoleModal}
					class="btn btn-secondary"
					disabled={savingRoles}
				>
					Cancel
				</button>
				<button
					on:click={saveRoles}
					class="btn btn-primary"
					disabled={savingRoles}
				>
					{savingRoles ? 'Saving...' : 'Save Changes'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.page-header {
		margin-bottom: 1.5rem;
	}

	h2 {
		font-size: 1.75rem;
		font-weight: 700;
		color: var(--text-primary);
	}

	/* Loading & Empty */
	.loading-text {
		text-align: center;
		padding: 3rem;
		color: var(--text-secondary);
	}

	.empty-state {
		text-align: center;
		padding: 3rem;
		color: var(--text-secondary);
	}

	/* Table */
	.table-card {
		overflow: hidden;
		padding: 0;
	}

	/* Badges */
	.role-badges {
		display: flex;
		flex-wrap: wrap;
		gap: 0.375rem;
	}

	.badge {
		display: inline-block;
		padding: 0.25rem 0.625rem;
		border-radius: 1rem;
		font-size: 0.75rem;
		font-weight: 600;
		background-color: #dbeafe;
		color: #1e40af;
	}

	.text-muted {
		color: var(--text-secondary);
		font-size: 0.875rem;
	}

	.link-btn {
		background: none;
		border: none;
		color: var(--primary-color);
		font-weight: 500;
		cursor: pointer;
		padding: 0;
		font-size: 0.875rem;
	}

	.link-btn:hover {
		text-decoration: underline;
	}

	.link-btn:disabled {
		color: var(--text-secondary);
		cursor: not-allowed;
	}

	.link-btn:disabled:hover {
		text-decoration: none;
	}

	/* Modal */
	.modal {
		max-width: 40rem;
	}

	.modal h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.modal-description {
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin-bottom: 1rem;
	}

	.role-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.role-option {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 0.75rem;
		border: 1px solid var(--border-color);
		border-radius: 0.375rem;
		cursor: pointer;
		transition: border-color 0.2s;
	}

	.role-option:hover {
		border-color: var(--primary-color);
	}

	.role-option input[type="checkbox"] {
		margin-top: 0.25rem;
	}

	.role-details {
		flex: 1;
	}

	.role-name {
		font-weight: 600;
		color: var(--text-primary);
	}

	.role-description {
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin-bottom: 0.5rem;
	}

	.permissions-list {
		font-size: 0.75rem;
		color: var(--text-secondary);
	}

	.permissions-list strong {
		display: block;
		margin-bottom: 0.25rem;
	}

	.permissions-list ul {
		list-style: disc;
		padding-left: 1.25rem;
	}

	@media (max-width: 768px) {
		.data-table th,
		.data-table td {
			padding: 0.75rem;
		}

		.modal {
			padding: 1rem;
		}
	}
</style>
