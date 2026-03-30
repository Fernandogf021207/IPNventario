<script lang="ts">
  import { onMount } from 'svelte';
  import { assignmentsApi } from '../../lib/api';
  import type { Assignment, AssignmentStatus } from '../../lib/types';
  import Modal from '../../lib/components/Modal.svelte';
  import ConfirmDialog from '../../lib/components/ConfirmDialog.svelte';

  let assignments: Assignment[] = [];
  let loading = true;
  let error = '';

  let isModalOpen = false;
  let isEditing = false;
  let formData: Partial<Assignment> = {};

  let isConfirmOpen = false;
  let confirmAction: 'publish' | 'close' | 'delete' | null = null;
  let selectedAssignment: Assignment | null = null;

  async function loadAssignments() {
    loading = true;
    error = '';
    try {
      assignments = await assignmentsApi.list();
    } catch (err: any) {
      error = err.message || 'Error al cargar prácticas';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadAssignments();
  });

  function openCreateModal() {
    isEditing = false;
    formData = {
      title: '',
      description: '',
      instructions: '',
      group_name: '',
      status: 'draft'
    };
    isModalOpen = true;
  }

  function openEditModal(assignment: Assignment) {
    if (assignment.status !== 'draft') {
      alert('Solo se pueden editar prácticas en estado de borrador (draft).');
      return;
    }
    isEditing = true;
    formData = { ...assignment };
    isModalOpen = true;
  }

  function confirmStatusChange(assignment: Assignment, action: 'publish' | 'close' | 'delete') {
    selectedAssignment = assignment;
    confirmAction = action;
    isConfirmOpen = true;
  }

  async function handleStatusChange() {
    if (!selectedAssignment || !confirmAction) return;

    try {
      if (confirmAction === 'publish') {
        await assignmentsApi.publish(selectedAssignment.id);
      } else if (confirmAction === 'close') {
        await assignmentsApi.close(selectedAssignment.id);
      } else if (confirmAction === 'delete') {
        await assignmentsApi.delete(selectedAssignment.id);
      }
      await loadAssignments();
    } catch (err: any) {
      alert(err.message || `Error al ${confirmAction} la práctica`);
    } finally {
      isConfirmOpen = false;
      selectedAssignment = null;
      confirmAction = null;
    }
  }

  async function handleSubmit() {
    try {
      if (isEditing && formData.id) {
        await assignmentsApi.update(formData.id, formData);
      } else {
        await assignmentsApi.create(formData);
      }
      isModalOpen = false;
      await loadAssignments();
    } catch (err: any) {
      alert(err.message || 'Error al guardar la práctica');
    }
  }
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Prácticas</h1>
      <p class="page-subtitle">Gestión de guías de laboratorio y sesiones</p>
    </div>
    <button class="btn btn-primary" on:click={openCreateModal}>
      + Nueva Práctica
    </button>
  </div>

  {#if error}
    <div class="error-alert">{error}</div>
  {/if}

  <div class="card table-container">
    {#if loading}
      <div class="text-center p-8 text-muted">Cargando...</div>
    {:else if assignments.length === 0}
      <div class="text-center p-8 text-muted">No hay prácticas registradas.</div>
    {:else}
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Título</th>
            <th>Grupo</th>
            <th>Estado</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {#each assignments as assignment}
            <tr>
              <td class="code">#{assignment.id}</td>
              <td class="font-medium">{assignment.title}</td>
              <td>{assignment.group_name}</td>
              <td>
                <span class="badge {assignment.status === 'published' ? 'badge-success' : assignment.status === 'draft' ? 'badge-warning' : 'badge-neutral'}">
                  {assignment.status === 'published' ? 'Publicada' : assignment.status === 'draft' ? 'Borrador' : 'Cerrada'}
                </span>
              </td>
              <td class="actions">
                {#if assignment.status === 'draft'}
                  <button class="btn-icon" on:click={() => openEditModal(assignment)} title="Editar">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"></path><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path></svg>
                  </button>
                  <button class="btn-icon text-success" on:click={() => confirmStatusChange(assignment, 'publish')} title="Publicar">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
                  </button>
                  <button class="btn-icon text-error" on:click={() => confirmStatusChange(assignment, 'delete')} title="Eliminar">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"></path><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                  </button>
                {:else if assignment.status === 'published'}
                  <button class="btn-icon text-error" on:click={() => confirmStatusChange(assignment, 'close')} title="Cerrar Práctica">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect></svg>
                  </button>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>

<Modal bind:open={isModalOpen} title={isEditing ? 'Editar Práctica' : 'Nueva Práctica'}>
  <form id="assignment-form" on:submit|preventDefault={handleSubmit}>
    <div class="form-grid">
      <div class="form-group full-width">
        <label for="title">Título de la Práctica *</label>
        <input class="input" id="title" bind:value={formData.title} required />
      </div>
      <div class="form-group full-width">
        <label for="description">Descripción</label>
        <textarea class="input" id="description" bind:value={formData.description} rows="3"></textarea>
      </div>
      <div class="form-group full-width">
        <label for="instructions">Instrucciones</label>
        <textarea class="input" id="instructions" bind:value={formData.instructions} rows="4" placeholder="Pasos a seguir en el laboratorio..."></textarea>
      </div>
      <div class="form-group">
        <label for="group_name">Grupo *</label>
        <input class="input" id="group_name" bind:value={formData.group_name} required placeholder="Ej. 3TV4" />
      </div>
    </div>
  </form>

  <div slot="footer">
    <button class="btn btn-secondary" on:click={() => isModalOpen = false}>Cancelar</button>
    <button class="btn btn-primary" type="submit" form="assignment-form">
      {isEditing ? 'Actualizar' : 'Guardar Borrador'}
    </button>
  </div>
</Modal>

<ConfirmDialog
  bind:open={isConfirmOpen}
  title={
    confirmAction === 'publish' ? 'Publicar Práctica' : 
    confirmAction === 'close' ? 'Cerrar Práctica' : 
    'Eliminar Práctica'
  }
  message={
    confirmAction === 'publish' ? `¿Estás seguro de que deseas publicar la práctica "${selectedAssignment?.title}"? Los alumnos podrán verla.` :
    confirmAction === 'close' ? `¿Estás seguro de que deseas cerrar la práctica "${selectedAssignment?.title}"? Ya no se podrán abrir sesiones.` :
    `¿Estás seguro de que deseas eliminar permanentemente la práctica "${selectedAssignment?.title}"?`
  }
  confirmText={
    confirmAction === 'publish' ? 'Sí, publicar' : 
    confirmAction === 'close' ? 'Sí, cerrar' : 
    'Sí, eliminar'
  }
  danger={confirmAction === 'close' || confirmAction === 'delete'}
  on:confirm={handleStatusChange}
/>

<style>
  .page-container { display: flex; flex-direction: column; gap: 1.5rem; }
  .page-header { display: flex; justify-content: space-between; align-items: flex-end; }
  .page-title { margin: 0; font-size: 1.875rem; }
  .page-subtitle { margin: 0; color: var(--text-muted); }
  .table-container { padding: 0; overflow-x: auto; }
  .data-table { width: 100%; border-collapse: collapse; text-align: left; }
  .data-table th, .data-table td { padding: 1rem 1.5rem; border-bottom: 1px solid var(--border); }
  .data-table th { background-color: var(--bg-subtle); font-weight: 500; color: var(--text-muted); font-size: 0.875rem; }
  .data-table tr:last-child td { border-bottom: none; }
  .code { font-family: var(--mono); color: var(--text-muted); }
  .font-medium { font-weight: 500; }
  .badge { display: inline-block; padding: 0.25rem 0.5rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 500; }
  .badge-success { background-color: var(--success-bg); color: var(--success); }
  .badge-warning { background-color: var(--warning-bg); color: var(--warning); }
  .badge-neutral { background-color: var(--bg-subtle); color: var(--text-muted); }
  .actions { display: flex; gap: 0.5rem; }
  .btn-icon { background: none; border: none; color: var(--text-muted); cursor: pointer; padding: 0.25rem; border-radius: var(--radius-sm); }
  .btn-icon:hover { color: var(--accent); background-color: var(--bg-subtle); }
  .text-success:hover { color: var(--success); }
  .text-error:hover { color: var(--error); }
  .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  .form-group { display: flex; flex-direction: column; gap: 0.25rem; }
  .full-width { grid-column: span 2; }
  .form-group label { font-size: 0.875rem; font-weight: 500; }
  textarea.input { resize: vertical; min-height: 80px; }
  .text-center { text-align: center; }
  .p-8 { padding: 2rem; }
  .text-muted { color: var(--text-muted); }
  .error-alert { background-color: var(--error-bg); color: var(--error); padding: 1rem; border-radius: var(--radius-md); border: 1px solid var(--error-border, rgba(239, 68, 68, 0.3)); }
</style>
