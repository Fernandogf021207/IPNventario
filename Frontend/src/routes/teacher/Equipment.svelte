<script lang="ts">
  import { onMount } from 'svelte';
  import { equipmentApi, sessionsApi, studentsApi, inventoryApi } from '../../lib/api';
  import type { EquipmentUsage } from '../../lib/types';
  import Modal from '../../lib/components/Modal.svelte';
  import ConfirmDialog from '../../lib/components/ConfirmDialog.svelte';

  let usages: EquipmentUsage[] = [];
  let loading = true;
  let error = '';

  // Data for create modal
  let sessions: any[] = [];
  let students: any[] = [];
  let machines: any[] = [];

  let isCreateModalOpen = false;
  let formData: any = {
    session_id: '',
    item_id: '',
    student_id: '',
    notes: ''
  };

  let isConfirmOpen = false;
  let selectedUsage: EquipmentUsage | null = null;

  async function loadData() {
    loading = true;
    error = '';
    try {
      usages = await equipmentApi.list();
    } catch (err: any) {
      error = err.message || 'Error al cargar uso de maquinaria';
    } finally {
      loading = false;
    }
  }

  async function loadFormData() {
    try {
      const [sessionsData, studentsData, itemsData] = await Promise.all([
        sessionsApi.list(),
        studentsApi.list(),
        inventoryApi.listItems({ type: 'machine' })
      ]);
      sessions = sessionsData.filter((s: any) => s.status === 'open');
      students = studentsData;
      machines = itemsData.filter((i: any) => 
        i.maintenance_status !== 'critical' && i.maintenance_status !== 'out_of_service'
      );
    } catch (err: any) {
      console.error('Error loading form data:', err);
    }
  }

  onMount(() => {
    loadData();
  });

  function openCreateModal() {
    formData = { session_id: '', item_id: '', student_id: '', notes: '' };
    loadFormData();
    isCreateModalOpen = true;
  }

  function confirmEnd(usage: EquipmentUsage) {
    selectedUsage = usage;
    isConfirmOpen = true;
  }

  async function handleCreate() {
    try {
      await equipmentApi.create({
        session_id: Number(formData.session_id),
        item_id: Number(formData.item_id),
        student_id: Number(formData.student_id),
        notes: formData.notes
      });
      isCreateModalOpen = false;
      await loadData();
    } catch (err: any) {
      alert(err.message || 'Error al registrar uso');
    }
  }

  async function handleEnd() {
    if (!selectedUsage) return;
    try {
      await equipmentApi.end(selectedUsage.id);
      await loadData();
    } catch (err: any) {
      alert(err.message || 'Error al finalizar uso');
    } finally {
      isConfirmOpen = false;
      selectedUsage = null;
    }
  }

  function getStatusLabel(status: string) {
    switch(status) {
      case 'active': return 'Activo';
      case 'completed': return 'Completado';
      case 'interrupted': return 'Interrumpido';
      default: return status;
    }
  }

  function getStatusClass(status: string) {
    switch(status) {
      case 'active': return 'st-active';
      case 'completed': return 'st-completed';
      case 'interrupted': return 'st-interrupted';
      default: return '';
    }
  }

  function formatDate(isoString: string) {
    if (!isoString) return '—';
    const date = new Date(isoString);
    return date.toLocaleString('es-MX', {
      day: '2-digit', month: '2-digit',
      hour: '2-digit', minute: '2-digit'
    });
  }

  function formatDuration(start: string, end?: string) {
    const startDate = new Date(start);
    const endDate = end ? new Date(end) : new Date();
    const diffMs = endDate.getTime() - startDate.getTime();
    const minutes = Math.floor(diffMs / 60000);
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    if (hours > 0) return `${hours}h ${mins}m`;
    return `${mins}m`;
  }

  $: activeCount = usages.filter(u => u.status === 'active').length;
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Uso de Maquinaria</h1>
      <p class="page-subtitle">Registro de uso de maquinaria fija por sesión</p>
    </div>
    <div class="header-actions">
      {#if activeCount > 0}
        <div class="active-badge">{activeCount} activo{activeCount > 1 ? 's' : ''}</div>
      {/if}
      <button class="btn btn-primary" on:click={openCreateModal}>
        + Registrar Uso
      </button>
    </div>
  </div>

  {#if error}
    <div class="error-alert">{error}</div>
  {/if}

  <div class="card table-container">
    {#if loading}
      <div class="text-center p-8 text-muted">Cargando...</div>
    {:else if usages.length === 0}
      <div class="text-center p-8 text-muted">No hay registros de uso de maquinaria.</div>
    {:else}
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Máquina</th>
            <th>Alumno</th>
            <th>Supervisor</th>
            <th>Inicio</th>
            <th>Duración</th>
            <th>Estado</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {#each usages as usage}
            <tr class:active-row={usage.status === 'active'}>
              <td class="code">#{usage.id}</td>
              <td class="font-medium">{usage.item_name}</td>
              <td>{usage.student_name}</td>
              <td class="text-sm">{usage.supervisor_name}</td>
              <td class="text-sm">{formatDate(usage.started_at)}</td>
              <td class="text-sm">
                {#if usage.status === 'active'}
                  <span class="duration-active">
                    {formatDuration(usage.started_at)}
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="display:inline;vertical-align:text-bottom"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                  </span>
                {:else}
                  {formatDuration(usage.started_at, usage.ended_at || undefined)}
                {/if}
              </td>
              <td>
                <span class="status-badge {getStatusClass(usage.status)}">
                  {getStatusLabel(usage.status)}
                </span>
              </td>
              <td class="actions">
                {#if usage.status === 'active'}
                  <button class="btn-sm btn-end" on:click={() => confirmEnd(usage)} title="Finalizar">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="display:inline;vertical-align:text-bottom"><rect width="18" height="18" x="3" y="3" rx="2"/></svg>
                    Finalizar
                  </button>
                {:else}
                  <span class="text-muted text-xs">—</span>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>

<!-- Create Usage Modal -->
<Modal bind:open={isCreateModalOpen} title="Registrar Uso de Maquinaria">
  <form id="usage-form" on:submit|preventDefault={handleCreate}>
    {#if sessions.length === 0}
      <div class="error-alert mb-4">No hay sesiones abiertas para registrar uso de maquinaria.</div>
    {:else}
      <div class="form-grid">
        <div class="form-group full-width">
          <label for="session_id">Sesión Abierta *</label>
          <select class="input form-select" id="session_id" bind:value={formData.session_id} required>
            <option value="" disabled>Selecciona una sesión</option>
            {#each sessions as s}
              <option value={s.id}>{s.title} ({s.group_name})</option>
            {/each}
          </select>
        </div>
        <div class="form-group full-width">
          <label for="item_id">Máquina *</label>
          <select class="input form-select" id="item_id" bind:value={formData.item_id} required>
            <option value="" disabled>Selecciona una máquina</option>
            {#each machines as m}
              <option value={m.id}>{m.name} ({m.sku}) — {m.maintenance_status}</option>
            {/each}
          </select>
        </div>
        <div class="form-group full-width">
          <label for="student_id">Alumno *</label>
          <select class="input form-select" id="student_id" bind:value={formData.student_id} required>
            <option value="" disabled>Selecciona un alumno</option>
            {#each students as st}
              <option value={st.id}>{st.full_name} ({st.student_code})</option>
            {/each}
          </select>
        </div>
        <div class="form-group full-width">
          <label for="usage-notes">Notas</label>
          <textarea class="input" id="usage-notes" bind:value={formData.notes} rows="2" placeholder="Observaciones adicionales..."></textarea>
        </div>
      </div>
    {/if}
  </form>

  <div slot="footer">
    <button class="btn btn-secondary" on:click={() => isCreateModalOpen = false}>Cancelar</button>
    <button class="btn btn-primary" type="submit" form="usage-form" disabled={sessions.length === 0}>
      Registrar Uso
    </button>
  </div>
</Modal>

<!-- Confirm End Usage -->
<ConfirmDialog
  bind:open={isConfirmOpen}
  title="Finalizar Uso"
  message={`¿Finalizar el uso de "${selectedUsage?.item_name}" por ${selectedUsage?.student_name}?`}
  confirmText="Sí, finalizar"
  danger={false}
  on:confirm={handleEnd}
/>

<style>
  .page-container { display: flex; flex-direction: column; gap: 1.5rem; }
  .page-header { display: flex; justify-content: space-between; align-items: flex-end; flex-wrap: wrap; gap: 1rem; }
  .page-title { margin: 0; font-size: 1.875rem; }
  .page-subtitle { margin: 0; color: var(--text-muted); }
  .header-actions { display: flex; gap: 0.75rem; align-items: center; }

  .active-badge { background-color: var(--success); color: white; font-weight: 600; font-size: 0.813rem; padding: 0.375rem 0.875rem; border-radius: var(--radius-full); }

  .table-container { padding: 0; overflow-x: auto; }
  .data-table { width: 100%; border-collapse: collapse; text-align: left; }
  .data-table th, .data-table td { padding: 0.75rem 1rem; border-bottom: 1px solid var(--border); }
  .data-table th { background-color: var(--bg-subtle); color: var(--text-muted); font-size: 0.8rem; font-weight: 500; white-space: nowrap; }
  .data-table tr:last-child td { border-bottom: none; }
  .data-table tr.active-row { background-color: var(--success-bg); }

  .code { font-family: var(--mono); color: var(--text-muted); font-size: 0.8rem; }
  .font-medium { font-weight: 500; }
  .text-sm { font-size: 0.813rem; }
  .text-xs { font-size: 0.75rem; }

  .duration-active { color: var(--success); font-weight: 600; }

  .status-badge { display: inline-block; padding: 0.2rem 0.6rem; border-radius: var(--radius-full); font-size: 0.75rem; font-weight: 600; }
  .st-active { background: var(--success-bg); color: var(--success); }
  .st-completed { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
  .st-interrupted { background: var(--warning-bg); color: var(--warning); }

  .actions { display: flex; gap: 0.375rem; }
  .btn-sm { border: none; padding: 0.3rem 0.7rem; border-radius: var(--radius-md); font-size: 0.75rem; font-weight: 600; cursor: pointer; transition: all 0.2s; font-family: var(--sans); }
  .btn-end { background: var(--error-bg); color: var(--error); }
  .btn-end:hover { background: var(--error); color: white; }

  .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  .form-group { display: flex; flex-direction: column; gap: 0.25rem; }
  .full-width { grid-column: span 2; }
  .form-group label { font-size: 0.875rem; font-weight: 500; }
  .form-select { appearance: none; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2371717a'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E"); background-size: 1.5em 1.5em; background-repeat: no-repeat; background-position: right 0.5rem center; padding-right: 2.5rem; }
  textarea.input { resize: vertical; min-height: 60px; }
  .text-center { text-align: center; }
  .p-8 { padding: 2rem; }
  .text-muted { color: var(--text-muted); }
  .mb-4 { margin-bottom: 1rem; }
  .error-alert { background-color: var(--error-bg); color: var(--error); padding: 1rem; border-radius: var(--radius-md); border: 1px solid rgba(239, 68, 68, 0.3); }
</style>
