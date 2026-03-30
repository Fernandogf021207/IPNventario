<script lang="ts">
  import { onMount } from 'svelte';
  import { assignmentsApi } from '../../lib/api';
  import type { Assignment } from '../../lib/types';

  let assignments: Assignment[] = [];
  let loading = true;
  let error = '';

  async function loadAssignments() {
    loading = true;
    error = '';
    try {
      // API call to list returns published assignments relevant for the student's group automatically handled in the backend
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
            <!-- For later: view manual, view details, etc. -->
            <button class="btn btn-secondary w-full" disabled>Ver Detalles</button>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

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
</style>
