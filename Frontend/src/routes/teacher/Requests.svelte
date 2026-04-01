<script lang="ts">
  import { onMount } from 'svelte';
  import { requestsApi } from '../../lib/api';
  import type { ResourceRequest } from '../../lib/types';
  import StatusBadge from '../../lib/components/StatusBadge.svelte';
  import ConfirmDialog from '../../lib/components/ConfirmDialog.svelte';

  let requests: ResourceRequest[] = [];
  let loading = true;
  let error = '';
  let filterStatus = '';

  let isConfirmOpen = false;
  let confirmAction: 'approve' | 'reject' | 'return' | null = null;
  let selectedRequest: ResourceRequest | null = null;

  async function loadData() {
    loading = true;
    error = '';
    try {
      requests = await requestsApi.list();
      if (filterStatus) {
        requests = requests.filter(r => r.status === filterStatus);
      }
    } catch (err: any) {
      error = err.message || 'Error al cargar solicitudes';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadData();
  });

  function confirmStatusChange(req: ResourceRequest, action: 'approve' | 'reject' | 'return') {
    selectedRequest = req;
    confirmAction = action;
    isConfirmOpen = true;
  }

  async function handleStatusChange() {
    if (!selectedRequest || !confirmAction) return;
    try {
      if (confirmAction === 'approve') {
        await requestsApi.approve(selectedRequest.id);
      } else if (confirmAction === 'reject') {
        await requestsApi.reject(selectedRequest.id);
      } else if (confirmAction === 'return') {
        await requestsApi.return(selectedRequest.id);
      }
      await loadData();
    } catch (err: any) {
      alert(err.message || 'Error al cambiar estado de la solicitud');
    } finally {
      isConfirmOpen = false;
      selectedRequest = null;
      confirmAction = null;
    }
  }

  function getRequestTypeLabel(type: string) {
    switch(type) {
      case 'loan': return 'Préstamo';
      case 'consumption': return 'Consumo';
      case 'machine_access': return 'Maquinaria';
      default: return type;
    }
  }

  function getRequestTypeClass(type: string) {
    switch(type) {
      case 'loan': return 'rtype-loan';
      case 'consumption': return 'rtype-consumption';
      case 'machine_access': return 'rtype-machine';
      default: return '';
    }
  }

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

  function formatDate(isoString: string) {
    if (!isoString) return '—';
    const date = new Date(isoString);
    return date.toLocaleString('es-MX', {
      day: '2-digit', month: '2-digit', year: 'numeric',
      hour: '2-digit', minute: '2-digit'
    });
  }

  $: pendingCount = requests.filter(r => r.status === 'pending').length;
  $: approvedLoans = requests.filter(r => r.status === 'approved' && (r.request_type || r.type) === 'loan').length;
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Solicitudes de Recursos</h1>
      <p class="page-subtitle">Gestión de préstamos, consumibles y acceso a maquinaria</p>
    </div>
    {#if pendingCount > 0}
      <div class="pending-badge">{pendingCount} pendiente{pendingCount > 1 ? 's' : ''}</div>
    {/if}
  </div>

  <!-- Quick stats -->
  {#if !loading}
    <div class="stats-bar">
      <button class="stat-chip" class:active={filterStatus === ''} on:click={() => { filterStatus = ''; loadData(); }}>
        Todas ({requests.length})
      </button>
      <button class="stat-chip pending" class:active={filterStatus === 'pending'} on:click={() => { filterStatus = 'pending'; loadData(); }}>
        Pendientes ({requests.filter(r => r.status === 'pending').length})
      </button>
      <button class="stat-chip approved" class:active={filterStatus === 'approved'} on:click={() => { filterStatus = 'approved'; loadData(); }}>
        Aprobadas ({requests.filter(r => r.status === 'approved').length})
      </button>
      <button class="stat-chip returned" class:active={filterStatus === 'returned'} on:click={() => { filterStatus = 'returned'; loadData(); }}>
        Devueltas ({requests.filter(r => r.status === 'returned').length})
      </button>
    </div>
  {/if}

  {#if error}
    <div class="error-alert">{error}</div>
  {/if}

  <div class="card table-container">
    {#if loading}
      <div class="text-center p-8 text-muted">Cargando...</div>
    {:else if requests.length === 0}
      <div class="text-center p-8 text-muted">No hay solicitudes registradas.</div>
    {:else}
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Alumno</th>
            <th>Item</th>
            <th>Tipo</th>
            <th>Cant.</th>
            <th>Estado</th>
            <th>Fecha</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {#each requests as req}
            <tr class:highlight-pending={req.status === 'pending'}>
              <td class="code">#{req.id}</td>
              <td>
                <div class="font-medium">{req.student_name}</div>
                <div class="text-xs text-muted">{req.student_code}</div>
              </td>
              <td>
                <div class="font-medium">{req.item_name}</div>
                <div class="text-xs text-muted">{req.item_sku}</div>
              </td>
              <td>
                <span class="rtype-badge {getRequestTypeClass(req.request_type || req.type)}">
                  {getRequestTypeLabel(req.request_type || req.type)}
                </span>
              </td>
              <td class="font-medium">{req.quantity}</td>
              <td>
                <span class="status-badge {getStatusClass(req.status)}">
                  {getStatusLabel(req.status)}
                </span>
              </td>
              <td class="text-sm">{formatDate(req.requested_at)}</td>
              <td class="actions">
                {#if req.status === 'pending'}
                  <button class="btn-sm btn-approve" on:click={() => confirmStatusChange(req, 'approve')} title="Aprobar">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="display:inline;vertical-align:text-bottom"><polyline points="20 6 9 17 4 12"/></svg>
                    Aprobar
                  </button>
                  <button class="btn-sm btn-reject" on:click={() => confirmStatusChange(req, 'reject')} title="Rechazar">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="display:inline;vertical-align:text-bottom"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                    Rechazar
                  </button>
                {:else if req.status === 'approved' && (req.request_type || req.type) === 'loan'}
                  <button class="btn-sm btn-return" on:click={() => confirmStatusChange(req, 'return')} title="Registrar Devolución">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="display:inline;vertical-align:text-bottom"><polyline points="9 14 4 9 9 4"/><path d="M20 20v-7a4 4 0 0 0-4-4H4"/></svg>
                    Devolver
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

<ConfirmDialog
  bind:open={isConfirmOpen}
  title={
    confirmAction === 'approve' ? 'Aprobar Solicitud' :
    confirmAction === 'reject' ? 'Rechazar Solicitud' :
    'Registrar Devolución'
  }
  message={
    confirmAction === 'approve' ? `¿Aprobar la solicitud de ${selectedRequest?.student_name} por ${selectedRequest?.quantity} "${selectedRequest?.item_name}"? Se descontará el stock correspondiente.` :
    confirmAction === 'reject' ? `¿Rechazar la solicitud de ${selectedRequest?.student_name}?` :
    `¿Registrar la devolución de ${selectedRequest?.quantity} "${selectedRequest?.item_name}"? Se restaurará el stock.`
  }
  confirmText={confirmAction === 'approve' ? 'Sí, aprobar' : confirmAction === 'reject' ? 'Sí, rechazar' : 'Sí, devolver'}
  danger={confirmAction === 'reject'}
  on:confirm={handleStatusChange}
/>

<style>
  .page-container { display: flex; flex-direction: column; gap: 1.5rem; }
  .page-header { display: flex; justify-content: space-between; align-items: flex-end; }
  .page-title { margin: 0; font-size: 1.875rem; }
  .page-subtitle { margin: 0; color: var(--text-muted); }

  .pending-badge { background-color: var(--warning); color: #000; font-weight: 600; font-size: 0.875rem; padding: 0.375rem 0.875rem; border-radius: var(--radius-full); animation: pulse-subtle 2s ease-in-out infinite; }
  @keyframes pulse-subtle { 0%, 100% { opacity: 1; } 50% { opacity: 0.7; } }

  .stats-bar { display: flex; gap: 0.5rem; flex-wrap: wrap; }
  .stat-chip { background: var(--bg-card); border: 1px solid var(--border); padding: 0.5rem 1rem; border-radius: var(--radius-full); font-size: 0.813rem; font-weight: 500; cursor: pointer; transition: all 0.2s; color: var(--text); font-family: var(--sans); }
  .stat-chip:hover { background: var(--bg-subtle); }
  .stat-chip.active { background: var(--accent-bg); border-color: var(--accent); color: var(--accent); }
  .stat-chip.pending.active { border-color: var(--warning); color: var(--warning); background: var(--warning-bg); }
  .stat-chip.approved.active { border-color: var(--success); color: var(--success); background: var(--success-bg); }
  .stat-chip.returned.active { border-color: #3b82f6; color: #3b82f6; background: rgba(59, 130, 246, 0.1); }

  .table-container { padding: 0; overflow-x: auto; }
  .data-table { width: 100%; border-collapse: collapse; text-align: left; }
  .data-table th, .data-table td { padding: 0.75rem 1rem; border-bottom: 1px solid var(--border); }
  .data-table th { background-color: var(--bg-subtle); color: var(--text-muted); font-size: 0.8rem; font-weight: 500; white-space: nowrap; }
  .data-table tr:last-child td { border-bottom: none; }
  .data-table tr.highlight-pending { background-color: var(--warning-bg); }

  .code { font-family: var(--mono); color: var(--text-muted); font-size: 0.8rem; }
  .font-medium { font-weight: 500; }
  .text-xs { font-size: 0.75rem; }
  .text-sm { font-size: 0.813rem; }

  .rtype-badge { display: inline-block; padding: 0.15rem 0.5rem; border-radius: var(--radius-full); font-size: 0.7rem; font-weight: 500; }
  .rtype-loan { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
  .rtype-consumption { background: rgba(16, 185, 129, 0.1); color: #10b981; }
  .rtype-machine { background: rgba(168, 85, 247, 0.1); color: #a855f7; }

  .status-badge { display: inline-block; padding: 0.2rem 0.6rem; border-radius: var(--radius-full); font-size: 0.75rem; font-weight: 600; }
  .st-pending { background: var(--warning-bg); color: var(--warning); }
  .st-approved { background: var(--success-bg); color: var(--success); }
  .st-rejected { background: var(--error-bg); color: var(--error); }
  .st-returned { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }

  .actions { display: flex; gap: 0.375rem; flex-wrap: wrap; }
  .btn-sm { border: none; padding: 0.3rem 0.6rem; border-radius: var(--radius-md); font-size: 0.75rem; font-weight: 600; cursor: pointer; transition: all 0.2s; font-family: var(--sans); }
  .btn-approve { background: var(--success-bg); color: var(--success); }
  .btn-approve:hover { background: var(--success); color: white; }
  .btn-reject { background: var(--error-bg); color: var(--error); }
  .btn-reject:hover { background: var(--error); color: white; }
  .btn-return { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
  .btn-return:hover { background: #3b82f6; color: white; }

  .text-center { text-align: center; }
  .p-8 { padding: 2rem; }
  .text-muted { color: var(--text-muted); }
  .error-alert { background-color: var(--error-bg); color: var(--error); padding: 1rem; border-radius: var(--radius-md); border: 1px solid rgba(239, 68, 68, 0.3); }
</style>
