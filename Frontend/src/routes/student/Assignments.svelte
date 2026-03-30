<script lang="ts">
  import { onMount } from 'svelte';
  import { assignmentsApi } from '../../lib/api';
  import type { Assignment } from '../../lib/types';
  import Modal from '../../lib/components/Modal.svelte';

  let assignments: Assignment[] = [];
  let loading = true;
  let error = '';
  
  let selectedAssignment: Assignment | null = null;
  let isModalOpen = false;

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

  function openDetails(assignment: Assignment) {
    selectedAssignment = assignment;
    isModalOpen = true;
  }

  onMount(() => {
    loadAssignments();
  });
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Mis Prácticas</h1>
      <p class="page-subtitle">Prácticas publicadas para tu grupo</p>
    </div>
  </div>

  {#if error}
    <div class="error-alert">{error}</div>
  {/if}

  <div class="assignments-grid">
    {#if loading}
      <div class="text-center p-8 text-muted" style="grid-column: 1 / -1;">Cargando prácticas...</div>
    {:else if assignments.length === 0}
      <div class="text-center p-8 text-muted" style="grid-column: 1 / -1;">No hay prácticas disponibles para tu grupo.</div>
    {:else}
      {#each assignments as assignment}
        <div class="card practice-card">
          <div class="practice-header">
            <h3 class="practice-title">{assignment.title}</h3>
            <span class="badge badge-success">Publicada</span>
          </div>
          <p class="practice-desc">{assignment.description || 'Sin descripción disponible.'}</p>
          <div class="practice-meta">
            <span class="meta-item">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
              Grupo: {assignment.group_name}
            </span>
            <span class="meta-item">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
              Fecha: {new Date(assignment.published_at || assignment.created_at).toLocaleDateString('es-MX')}
            </span>
          </div>
          <div class="practice-actions">
            <button class="btn btn-secondary w-full" on:click={() => openDetails(assignment)}>Ver Detalles</button>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

<Modal bind:open={isModalOpen} title={selectedAssignment?.title || 'Detalles de la Práctica'}>
  {#if selectedAssignment}
    <div class="details-view">
      <div class="details-section">
        <h4>Descripción</h4>
        <p>{selectedAssignment.description || 'Sin descripción.'}</p>
      </div>

      <div class="details-section">
        <h4>Instrucciones de Laboratorio</h4>
        <div class="instructions-box">
          {selectedAssignment.instructions || 'Tu profesor no ha proporcionado instrucciones específicas.'}
        </div>
      </div>

      <div class="details-grid">
        <div class="detail-item">
          <strong>Grupo:</strong> {selectedAssignment.group_name}
        </div>
        <div class="detail-item">
          <strong>Publicada el:</strong> {new Date(selectedAssignment.published_at || selectedAssignment.created_at).toLocaleString('es-MX')}
        </div>
      </div>
    </div>
  {/if}
  <div slot="footer">
    <button class="btn btn-primary" on:click={() => isModalOpen = false}>Cerrar</button>
  </div>
</Modal>

<style>
  .page-container {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .page-header {
    margin-bottom: 0.5rem;
  }

  .page-title {
    margin: 0;
    font-size: 1.875rem;
  }

  .page-subtitle {
    margin: 0;
    color: var(--text-muted);
  }

  .assignments-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1.5rem;
  }

  .practice-card {
    display: flex;
    flex-direction: column;
    padding: 1.5rem;
  }

  .practice-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1rem;
    gap: 1rem;
  }

  .practice-title {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
  }

  .practice-desc {
    color: var(--text-muted);
    font-size: 0.875rem;
    margin-bottom: 1.5rem;
    flex: 1;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .practice-meta {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
    padding-top: 1rem;
    border-top: 1px solid var(--border);
  }

  .meta-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: var(--text-muted);
  }

  .practice-actions {
    margin-top: auto;
  }

  .w-full {
    width: 100%;
  }

  .badge {
    display: inline-block;
    padding: 0.25rem 0.5rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  .badge-success { background-color: var(--success-bg); color: var(--success); }

  .text-center { text-align: center; }
  .p-8 { padding: 2rem; }
  .text-muted { color: var(--text-muted); }

  .error-alert {
    background-color: var(--error-bg);
    color: var(--error);
    padding: 1rem;
    border-radius: var(--radius-md);
    border: 1px solid var(--error-border, rgba(239, 68, 68, 0.3));
  }

  /* Estilos para la vista de detalles */
  .details-view {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    padding: 0.5rem 0;
  }
  .details-section h4 {
    margin: 0 0 0.5rem 0;
    color: var(--text-muted);
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .details-section p {
    margin: 0;
    line-height: 1.6;
  }
  .instructions-box {
    background-color: var(--bg-subtle);
    border-radius: var(--radius-md);
    padding: 1.25rem;
    border-left: 4px solid var(--accent);
    white-space: pre-wrap;
    font-family: inherit;
    line-height: 1.6;
  }
  .details-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--border);
  }
  .detail-item {
    font-size: 0.875rem;
    color: var(--text-muted);
  }
  .detail-item strong {
    color: var(--text-color);
    margin-right: 0.25rem;
  }
</style>
