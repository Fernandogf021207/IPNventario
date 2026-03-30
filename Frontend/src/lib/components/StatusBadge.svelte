<script lang="ts">
  export let status: string;
  export let type: 'session' | 'assignment' | 'request' | 'maintenance' = 'session';

  let colorClass = '';
  let label = status;

  $: {
    if (type === 'session') {
      const colors: Record<string, string> = {
        'planned': 'bg-warning text-warning-fg',
        'open': 'bg-success text-success-fg',
        'closed': 'bg-gray-200 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
        'cancelled': 'bg-error text-error-fg'
      };
      colorClass = colors[status] || 'bg-gray-200 text-gray-700';
      const labels: Record<string, string> = {
        'planned': 'Programada',
        'open': 'Abierta',
        'closed': 'Cerrada',
        'cancelled': 'Cancelada'
      };
      label = labels[status] || status;
    }
    // Expand for other types later
  }
</script>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    padding: 0.125rem 0.5rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
  }
  
  /* Fallback colors for status badge using our CSS variables */
  .bg-warning { background-color: var(--warning-bg); }
  .text-warning-fg { color: var(--warning); }
  
  .bg-success { background-color: var(--success-bg); }
  .text-success-fg { color: var(--success); }
  
  .bg-error { background-color: var(--error-bg); }
  .text-error-fg { color: var(--error); }
  
  .bg-gray-200 { background-color: var(--bg-subtle); }
  .text-gray-700 { color: var(--text-muted); }
</style>

<span class="badge {colorClass}">
  {label}
</span>
