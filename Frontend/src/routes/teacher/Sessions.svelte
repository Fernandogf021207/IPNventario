<script lang="ts">
  import { onMount } from 'svelte';
  import { sessionsApi, assignmentsApi } from '../../lib/api';
  import type { LabSession, Assignment } from '../../lib/types';
  import Modal from '../../lib/components/Modal.svelte';
  import ConfirmDialog from '../../lib/components/ConfirmDialog.svelte';
  import StatusBadge from '../../lib/components/StatusBadge.svelte';

  let sessions: LabSession[] = [];
  let assignments: Assignment[] = [];
  let loading = true;
  let error = '';

  let isModalOpen = false;
  let isEditing = false;
  let formData: any = {};

  let isConfirmOpen = false;
  let confirmAction: 'open' | 'close' | 'cancel' | null = null;
  let selectedSession: LabSession | null = null;

  async function loadData() {
    loading = true;
    error = '';
    try {
      const [sessionsData, assignmentsData] = await Promise.all([
        sessionsApi.list(),
        assignmentsApi.list()
      ]);
      sessions = sessionsData;
      assignments = assignmentsData.filter(a => a.status === 'published');
    } catch (err: any) {
      error = err.message || 'Error al cargar sesiones';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadData();
  });

  function openCreateModal() {
    isEditing = false;
    const now = new Date();
    const later = new Date(now.getTime() + 2 * 60 * 60 * 1000);
    
    formData = {
      assignment_id: undefined,
      title: '',
      group_name: '',
      scheduled_start: now.toISOString().slice(0, 16),
      scheduled_end: later.toISOString().slice(0, 16),
      notes: ''
    };
    isModalOpen = true;
  }

  function openEditModal(session: LabSession) {
    isEditing = true;
    selectedSession = session;
    
    // Convert ISO to local datetime-local format
    const start = new Date(session.scheduled_start);
    const end = new Date(session.scheduled_end);
    
    formData = {
      id: session.id,
      assignment_id: session.assignment_id,
      title: session.title,
      group_name: session.group_name,
      scheduled_start: new Date(start.getTime() - start.getTimezoneOffset() * 60000).toISOString().slice(0, 16),
      scheduled_end: new Date(end.getTime() - end.getTimezoneOffset() * 60000).toISOString().slice(0, 16),
      notes: session.notes || ''
    };
    isModalOpen = true;
  }

  // Auto-fill title when assignment is selected (only for new sessions)
  $: if (!isEditing && formData.assignment_id) {
    const selected = assignments.find(a => a.id === Number(formData.assignment_id));
    if (selected && !formData.title) {
      formData.title = `Sesión: ${selected.title}`;
    }
  }

  function confirmStatusChange(session: LabSession, action: 'open' | 'close' | 'cancel') {
    selectedSession = session;
    confirmAction = action;
    isConfirmOpen = true;
  }

  async function handleStatusChange() {
    if (!selectedSession || !confirmAction) return;

    try {
      if (confirmAction === 'open') {
        await sessionsApi.open(selectedSession.id);
      } else if (confirmAction === 'close') {
        await sessionsApi.close(selectedSession.id);
      } else if (confirmAction === 'cancel') {
        await sessionsApi.cancel(selectedSession.id);
      }
      await loadData();
    } catch (err: any) {
      alert(err.message || `Error al cambiar estado de la sesión`);
    } finally {
      isConfirmOpen = false;
      selectedSession = null;
      confirmAction = null;
    }
  }

  async function handleSubmit() {
    try {
      const dataToSend = {
        ...formData,
        assignment_id: Number(formData.assignment_id),
        scheduled_start: new Date(formData.scheduled_start).toISOString(),
        scheduled_end: new Date(formData.scheduled_end).toISOString()
      };
      
      if (isEditing && formData.id) {
        await sessionsApi.update(formData.id, dataToSend);
      } else {
        await sessionsApi.create(dataToSend);
      }
      
      isModalOpen = false;
      await loadData();
    } catch (err: any) {
      alert(err.message || 'Error al guardar la sesión');
    }
  }

  function formatDate(isoString: string) {
    if (!isoString) return '--';
    const date = new Date(isoString);
    return date.toLocaleString('es-MX', {
      day: '2-digit', month: '2-digit', year: 'numeric',
      hour: '2-digit', minute: '2-digit'
    });
  }
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Sesiones de Laboratorio</h1>
      <p class="page-subtitle">Programación y gestión de sesiones</p>
    </div>
    <button class="btn btn-primary" on:click={openCreateModal}>
      + Programar Sesión
    </button>
  </div>

  {#if error}
    <div class="error-alert">{error}</div>
  {/if}

  <div class="card table-container">
    {#if loading}
      <div class="text-center p-8 text-muted">Cargando...</div>
    {:else if sessions.length === 0}
      <div class="text-center p-8 text-muted">No hay sesiones registradas.</div>
    {:else}
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Título / Práctica</th>
            <th>Grupo</th>
            <th>Inicio Programado</th>
            <th>Estado</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {#each sessions as session}
            <tr>
              <td class="code">#{session.id}</td>
              <td>
                <div class="font-medium">{session.title}</div>
                <div class="text-xs text-muted">({session.assignment_title || 'Práctica'})</div>
              </td>
              <td>{session.group_name}</td>
              <td>{formatDate(session.scheduled_start)}</td>
              <td>
                <StatusBadge status={session.status} type="session" />
              </td>
              <td class="actions">
                {#if session.status === 'planned'}
                  <button class="btn-icon text-success" on:click={() => confirmStatusChange(session, 'open')} title="Abrir Sesión">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"></path><path d="m12 5 7 7-7 7"></path></svg>
                  </button>
                  <button class="btn-icon" on:click={() => openEditModal(session)} title="Editar Sesión">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                  </button>
                  <button class="btn-icon text-error" on:click={() => confirmStatusChange(session, 'cancel')} title="Cancelar Sesión">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>
                  </button>
                {:else if session.status === 'open'}
                  <button class="btn-icon text-error" on:click={() => confirmStatusChange(session, 'close')} title="Cerrar Sesión">
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

<Modal bind:open={isModalOpen} title={isEditing ? 'Editar Sesión' : 'Programar Sesión'}>
  <form id="session-form" on:submit|preventDefault={handleSubmit}>
    {#if assignments.length === 0}
      <div class="error-alert mb-4">No hay prácticas publicadas para programar sesiones.</div>
    {:else}
      <div class="form-grid">
        <div class="form-group full-width">
          <label for="assignment_id">Práctica *</label>
          <select class="input form-select" id="assignment_id" bind:value={formData.assignment_id} required>
            <option value="" disabled selected>Selecciona una práctica</option>
            {#each assignments as assignment}
              <option value={assignment.id}>{assignment.title} ({assignment.group_name})</option>
            {/each}
          </select>
        </div>
        <div class="form-group full-width">
          <label for="title">Título de la Sesión *</label>
          <input class="input" id="title" bind:value={formData.title} required placeholder="Ej. Laboratorio de Motores" />
        </div>
        <div class="form-group">
          <label for="group_name">Grupo *</label>
          <input class="input" id="group_name" bind:value={formData.group_name} required placeholder="Ej. 6MM1" />
        </div>
        <div class="form-group">
          <label for="scheduled_start">Inicio *</label>
          <input type="datetime-local" class="input" id="scheduled_start" bind:value={formData.scheduled_start} required />
        </div>
        <div class="form-group">
          <label for="scheduled_end">Fin *</label>
          <input type="datetime-local" class="input" id="scheduled_end" bind:value={formData.scheduled_end} required />
        </div>
        <div class="form-group full-width">
          <label for="notes">Notas / Observaciones</label>
          <textarea class="input" id="notes" bind:value={formData.notes} rows="2"></textarea>
        </div>
      </div>
    {/if}
  </form>

  <div slot="footer">
    <button class="btn btn-secondary" on:click={() => isModalOpen = false}>Cancelar</button>
    <button class="btn btn-primary" type="submit" form="session-form" disabled={assignments.length === 0}>
      {isEditing ? 'Actualizar' : 'Programar'} Sesión
    </button>
  </div>
</Modal>

<ConfirmDialog
  bind:open={isConfirmOpen}
  title={
    confirmAction === 'open' ? 'Abrir Sesión' : 
    confirmAction === 'close' ? 'Cerrar Sesión' : 
    'Cancelar Sesión'
  }
  message={
    confirmAction === 'open' ? `¿Estás seguro de que deseas ABRIR la sesión "${selectedSession?.title}"?` :
    confirmAction === 'close' ? `¿Estás seguro de que deseas CERRAR la sesión?` :
    `¿Estás seguro de que deseas CANCELAR esta sesión?`
  }
  confirmText={confirmAction === 'open' ? 'Sí, abrir' : confirmAction === 'close' ? 'Sí, cerrar' : 'Sí, cancelar'}
  danger={confirmAction === 'close' || confirmAction === 'cancel'}
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
  .data-table th { background-color: var(--bg-subtle); color: var(--text-muted); font-size: 0.875rem; font-weight: 500; }
  .data-table tr:last-child td { border-bottom: none; }
  .code { font-family: var(--mono); color: var(--text-muted); }
  .font-medium { font-weight: 500; }
  .text-xs { font-size: 0.75rem; }
  .actions { display: flex; gap: 0.5rem; }
  .btn-icon { background: none; border: none; color: var(--text-muted); cursor: pointer; padding: 0.25rem; border-radius: var(--radius-sm); }
  .btn-icon:hover { color: var(--accent); background-color: var(--bg-subtle); }
  .text-success:hover { color: var(--success); }
  .text-error:hover { color: var(--error); }
  .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  .form-group { display: flex; flex-direction: column; gap: 0.25rem; }
  .full-width { grid-column: span 2; }
  .form-group label { font-size: 0.875rem; font-weight: 500; }
  .form-select { appearance: none; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2371717a'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E"); background-size: 1.5em 1.5em; background-repeat: no-repeat; background-position: right 0.5rem center; padding-right: 2.5rem; }
  textarea.input { resize: vertical; min-height: 80px; }
  .text-center { text-align: center; }
  .p-8 { padding: 2rem; }
  .text-muted { color: var(--text-muted); }
  .mb-4 { margin-bottom: 1rem; }
  .error-alert { background-color: var(--error-bg); color: var(--error); padding: 1rem; border-radius: var(--radius-md); border: 1px solid var(--error-border, rgba(239, 68, 68, 0.3)); }
</style>
