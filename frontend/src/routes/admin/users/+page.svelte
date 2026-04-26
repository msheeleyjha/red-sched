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

<div class="container mx-auto p-6 max-w-7xl">
	<div class="flex justify-between items-center mb-6">
		<h1 class="text-3xl font-bold">User Role Management</h1>
	</div>

	{#if error}
		<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4" role="alert">
			{error}
		</div>
	{/if}

	{#if success}
		<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded mb-4" role="alert">
			{success}
		</div>
	{/if}

	{#if !currentUserIsSuperAdmin}
		<div class="bg-yellow-100 border border-yellow-400 text-yellow-700 px-4 py-3 rounded mb-4" role="alert">
			<p class="font-bold">Access Restricted</p>
			<p>You need System Admin privileges to manage user roles.</p>
		</div>
	{/if}

	{#if loading}
		<div class="text-center py-8">
			<p>Loading users...</p>
		</div>
	{:else}
		<div class="bg-white shadow-md rounded-lg overflow-hidden">
			<table class="min-w-full">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">User</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Email</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Roles</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each users as user}
						<tr>
							<td class="px-6 py-4 whitespace-nowrap">{user.name}</td>
							<td class="px-6 py-4 whitespace-nowrap">{user.email}</td>
							<td class="px-6 py-4">
								{#if user.roles && user.roles.length > 0}
									<div class="flex flex-wrap gap-1">
										{#each user.roles as role}
											<span class="px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
												{role.name}
											</span>
										{/each}
									</div>
								{:else}
									<span class="text-gray-400 text-sm">No Roles (Profile Only)</span>
								{/if}
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<button
									on:click={() => openRoleModal(user)}
									class="text-blue-600 hover:text-blue-900 font-medium"
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
			<p class="text-center text-gray-500 py-8">No users found</p>
		{/if}
	{/if}
</div>

<!-- Role Management Modal -->
{#if showRoleModal && selectedUser}
	<div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50" on:click={closeRoleModal}>
		<div class="relative top-20 mx-auto p-5 border w-full max-w-2xl shadow-lg rounded-md bg-white" on:click|stopPropagation>
			<div class="mt-3">
				<h3 class="text-lg font-medium leading-6 text-gray-900 mb-4">
					Manage Roles for {selectedUser.name}
				</h3>

				{#if selectedUser.id === currentUserId && !selectedRoles.some(id => allRoles.find(r => r.id === id)?.name === 'Super Admin')}
					<div class="bg-yellow-100 border border-yellow-400 text-yellow-700 px-4 py-3 rounded mb-4" role="alert">
						<p class="text-sm">Warning: If you remove your own Super Admin role, you will lose access to this page.</p>
					</div>
				{/if}

				<div class="mb-4">
					<p class="text-sm text-gray-600 mb-3">Select roles to assign to this user:</p>

					{#each allRoles as role}
						<div class="mb-3 p-3 border rounded">
							<label class="flex items-start cursor-pointer">
								<input
									type="checkbox"
									value={role.id}
									bind:group={selectedRoles}
									class="mt-1 mr-3"
								/>
								<div class="flex-1">
									<div class="font-medium">{role.name}</div>
									<div class="text-sm text-gray-600 mb-2">{role.description}</div>

									{#if role.permissions && role.permissions.length > 0}
										<div class="text-xs text-gray-500">
											<strong>Permissions:</strong>
											<ul class="list-disc list-inside mt-1">
												{#each role.permissions as perm}
													<li>{getPermissionDisplayName(perm)}</li>
												{/each}
											</ul>
										</div>
									{/if}
								</div>
							</label>
						</div>
					{/each}
				</div>

				{#if selectedRoles.length === 0 && selectedUser.id !== currentUserId}
					<div class="bg-blue-100 border border-blue-400 text-blue-700 px-4 py-3 rounded mb-4" role="alert">
						<p class="text-sm">No roles assigned. User will only be able to edit their profile until roles are assigned.</p>
					</div>
				{/if}

				<div class="flex justify-end gap-3 mt-4">
					<button
						on:click={closeRoleModal}
						class="px-4 py-2 bg-gray-300 text-gray-700 rounded hover:bg-gray-400"
						disabled={savingRoles}
					>
						Cancel
					</button>
					<button
						on:click={saveRoles}
						class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400"
						disabled={savingRoles}
					>
						{savingRoles ? 'Saving...' : 'Save Changes'}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	/* Add any additional styles here if needed */
</style>
