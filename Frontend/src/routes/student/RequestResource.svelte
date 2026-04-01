<script lang="ts">
  import { onMount } from 'svelte';
  import { requestsApi, inventoryApi, sessionsApi } from '../../lib/api';
  import { auth } from '../../lib/stores/auth';
  import Modal from '../../lib/components/Modal.svelte';

  let myRequests: any[] = [];
  let items: any[] = [];
  let openSessions: any[] = [];
  let loading = true;
  let error = '';

  let isRequestModalOpen = false;
  let formData = {
    session_id: '',
    item_id: '',
    request_type: 'loan',
    quantity: 1,
    notes: ''
  };

  async function loadData() {
    loading = true;
    error = '';
    try {
      const [requestsData, sessionsData] = await Promise.all([
        requestsApi.list(),
        sessionsApi.list()
      ]);
      myRequests = requestsData;
      openSessions = sessionsData.filter((s: any) => s.status === 'open');
    } catch (err: any) {
      error = err.message || 'Error al cargar datos';
    } finally {
      loading = false;
    }
  }

  async function openRequestModal() {
    try {
      items = await inventoryApi.listItems({ active_only: 'true' });
    } catch (err: any) {
      console.error('Error loading items:', err);
    }
    formData = { session_id: '', item_id: '', request_type: 'loan', quantity: 1, notes: '' };
    isRequestModalOpen = true;
  }

  async function handleSubmit() {
    try {
      await requestsApi.create({
        session_id: Number(formData.session_id),
        item_id: Number(formData.item_id),
        request_type: formData.request_type,
        quantity: Number(formData.quantity),
        notes: formData.notes
      });
      isRequestModalOpen = false;
      await loadData();
    } catch (err: any) {
      alert(err.message || 'Error al crear solicitud');
    }
  }

  onMount(() => {
    loadData();
  });

  function getStatusLabel(status: string) {
    switch(status) {
      case 'pending': return 'Pendiente';
      case 'approved': return 'Aprobado';
      case 'rejected': return 'Rechazado';
      case 'returned': return 'Devuelto';
      default: return status;
    }
  }

  function getStatusClass(status: string) {
    switch(status) {
      case 'pending': return 'st-pending';
      case 'approved': return 'st-approved';
      case 'rejected': return 'st-rejected';
      case 'returned': return 'st-returned';
      default: return '';
    }
  }

  function getTypeLabel(type: string) {
    switch(type) {
      case 'loan': return 'Préstamo';
      case 'consumption': return 'Consumo';
      case 'machine_access': return 'Maquinaria';
      default: return type;
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

  // Filter items based on request type
  $: filteredItems = items.filter(item => {
    const type = item.item_type || item.type;
    if (formData.request_type === 'loan') return type === 'tool';
    if (formData.request_type === 'consumption') return type === 'consumable';
    if (formData.request_type === 'machine_access') return type === 'machine';
    return true;
  });

  $: pendingCount = myRequests.filter(r => r.status === 'pending').length;
  $: approvedCount = myRequests.filter(r => r.status === 'approved').length;
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Solicitar Recurso</h1>
      <p class="page-subtitle">Solicita herramientas, materiales o acceso a maquinaria</p>
    </div>
    <button class="btn btn-primary" on:click={openRequestModal} disabled={openSessions.length === 0}>
      + Nueva Solicitud
    </button>
  </div>

  {#if openSessions.length === 0 && !loading}
    <div class="info-alert">
      <strong>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="display:inline;vertical-align:text-bottom"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
        Sin sesiones abiertas
      </strong>
      <p>No hay sesiones de laboratorio abiertas actualmente. Las solicitudes solo pueden crearse durante una sesión abierta.</p>
    </div>
  {/if}

  {#if error}
    <div class="error-alert">{error}</div>
  {/if}

  <!-- Summary -->
  {#if !loading}
    <div class="status-cards">
      <div class="status-card pending">
        <div class="sc-value">{pendingCount}</div>
        <div class="sc-label">Pendientes</div>
      </div>
      <div class="status-card approved">
        <div class="sc-value">{approvedCount}</div>
        <div class="sc-label">Aprobadas</div>
      </div>
      <div class="status-card total">
        <div class="sc-value">{myRequests.length}</div>
        <div class="sc-label">Total</div>
      </div>
    </div>
  {/if}

  <!-- Requests list -->
  <div class="card">
    <h3 class="section-title">Mis Solicitudes</h3>
    {#if loading}
      <div class="text-center p-4 text-muted">Cargando...</div>
    {:else if myRequests.length === 0}
      <div class="text-center p-4 text-muted">No tienes solicitudes registradas.</div>
    {:else}
      <div class="requests-list">
        {#each myRequests as req}
          <div class="request-card">
            <div class="request-main">
              <div class="request-item">{req.item_name}</div>
              <div class="request-meta">
                <span class="request-type">{getTypeLabel(req.request_type || req.type)}</span>
                <span>×{req.quantity}</span>
                <span class="text-muted">{formatDate(req.requested_at || req.created_at)}</span>
              </div>
            </div>
            <span class="status-badge {getStatusClass(req.status)}">
              {getStatusLabel(req.status)}
            </span>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Request Modal -->
<Modal bind:open={isRequestModalOpen} title="Nueva Solicitud de Recurso">
  <form id="request-form" on:submit|preventDefault={handleSubmit}>
    <div class="form-grid">
      <div class="form-group full-width">
        <label for="req-session">Sesión Abierta *</label>
        <select class="input form-select" id="req-session" bind:value={formData.session_id} required>
          <option value="" disabled>Selecciona la sesión</option>
          {#each openSessions as s}
            <option value={s.id}>{s.title} ({s.group_name})</option>
          {/each}
        </select>
      </div>
      <div class="form-group full-width">
        <label for="req-type">Tipo de solicitud *</label>
        <select class="input form-select" id="req-type" bind:value={formData.request_type} required>
          <option value="loan">Préstamo (herramienta)</option>
          <option value="consumption">Consumo (material)</option>
          <option value="machine_access">Acceso a maquinaria</option>
        </select>
      </div>
      <div class="form-group full-width">
        <label for="req-item">Recurso *</label>
        <select class="input form-select" id="req-item" bind:value={formData.item_id} required>
          <option value="" disabled>Selecciona un recurso</option>
          {#each filteredItems as item}
            <option value={item.id}>
              {item.name} ({item.sku})
              {#if item.item_type !== 'machine' && (item as any).type !== 'machine'}
                — Stock: {item.stock} {item.unit}
              {/if}
            </option>
          {/each}
        </select>
      </div>
      <div class="form-group">
        <label for="req-qty">Cantidad</label>
        <input type="number" step="0.01" min="0.01" class="input" id="req-qty" bind:value={formData.quantity} />
      </div>
      <div class="form-group full-width">
        <label for="req-notes">Notas</label>
        <textarea class="input" id="req-notes" bind:value={formData.notes} rows="2" placeholder="Observaciones opcionales..."></textarea>
      </div>
    </div>
  </form>

  <div slot="footer">
    <button class="btn btn-secondary" on:click={() => isRequestModalOpen = false}>Cancelar</button>
    <button class="btn btn-primary" type="submit" form="request-form">Enviar Solicitud</button>
  </div>
</Modal>

<style>
  .page-container { display: flex; flex-direction: column; gap: 1.5rem; }
  .page-header { display: flex; justify-content: space-between; align-items: flex-end; flex-wrap: wrap; gap: 1rem; }
  .page-title { margin: 0; font-size: 1.875rem; }
  .page-subtitle { margin: 0; color: var(--text-muted); }

  .info-alert { background-color: rgba(59, 130, 246, 0.1); color: #3b82f6; padding: 1rem; border-radius: var(--radius-md); border: 1px solid rgba(59, 130, 246, 0.3); }
  .info-alert p { margin: 0.25rem 0 0; font-size: 0.875rem; opacity: 0.8; }

  .status-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; }
  .status-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 1.25rem; text-align: center; }
  .status-card.pending { border-left: 3px solid var(--warning); }
  .status-card.approved { border-left: 3px solid var(--success); }
  .status-card.total { border-left: 3px solid var(--accent); }
  .sc-value { font-size: 1.75rem; font-weight: 700; color: var(--text-h); }
  .sc-label { font-size: 0.75rem; color: var(--text-muted); margin-top: 0.25rem; }

  .section-title { font-size: 1.125rem; margin-bottom: 1rem; }

  .requests-list { display: flex; flex-direction: column; gap: 0.5rem; }
  .request-card { display: flex; justify-content: space-between; align-items: center; padding: 0.875rem; border: 1px solid var(--border); border-radius: var(--radius-md); transition: background-color 0.2s; }
  .request-card:hover { background-color: var(--bg-subtle); }
  .request-main { display: flex; flex-direction: column; gap: 0.25rem; }
  .request-item { font-weight: 600; color: var(--text-h); }
  .request-meta { display: flex; gap: 0.75rem; font-size: 0.8rem; color: var(--text-muted); align-items: center; }
  .request-type { background: var(--bg-subtle); padding: 0.1rem 0.5rem; border-radius: var(--radius-full); font-size: 0.7rem; font-weight: 500; }

  .status-badge { display: inline-block; padding: 0.25rem 0.75rem; border-radius: var(--radius-full); font-size: 0.75rem; font-weight: 600; white-space: nowrap; }
  .st-pending { background: var(--warning-bg); color: var(--warning); }
  .st-approved { background: var(--success-bg); color: var(--success); }
  .st-rejected { background: var(--error-bg); color: var(--error); }
  .st-returned { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }

  .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  .form-group { display: flex; flex-direction: column; gap: 0.25rem; }
  .full-width { grid-column: span 2; }
  .form-group label { font-size: 0.875rem; font-weight: 500; }
  .form-select { appearance: none; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2371717a'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E"); background-size: 1.5em 1.5em; background-repeat: no-repeat; background-position: right 0.5rem center; padding-right: 2.5rem; }
  textarea.input { resize: vertical; min-height: 60px; }
  .text-center { text-align: center; }
  .p-4 { padding: 1rem; }
  .text-muted { color: var(--text-muted); }
  .error-alert { background-color: var(--error-bg); color: var(--error); padding: 1rem; border-radius: var(--radius-md); border: 1px solid rgba(239, 68, 68, 0.3); }
</style>
