<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  
  export let open = false;
  export let title = '';

  const dispatch = createEventDispatcher();

  function close() {
    open = false;
    dispatch('close');
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') close();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="backdrop" on:click={close}>
    <div class="modal" on:click|stopPropagation role="dialog" aria-modal="true" aria-labelledby="modal-title" tabindex="-1">
      <div class="header">
        <h2 id="modal-title">{title}</h2>
        <button class="close-btn" on:click={close} aria-label="Close">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="18 6 6 18"></polyline>
            <polyline points="6 6 18 18"></polyline>
          </svg>
        </button>
      </div>
      
      <div class="content">
        <slot></slot>
      </div>
      
      {#if $$slots.footer}
        <div class="footer">
          <slot name="footer"></slot>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 50;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(2px);
  }

  .modal {
    width: 100%;
    max-width: 32rem;
    margin: 1rem;
    background-color: var(--bg-card);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    display: flex;
    flex-direction: column;
    max-height: calc(100vh - 4rem);
    border: 1px solid var(--border);
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.25rem 1.5rem;
    border-bottom: 1px solid var(--border);
  }

  .header h2 {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 0.25rem;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.2s, color 0.2s;
  }

  .close-btn:hover {
    background-color: var(--bg-subtle);
    color: var(--text-h);
  }

  .content {
    padding: 1.5rem;
    overflow-y: auto;
  }

  .footer {
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--border);
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    background-color: var(--bg-subtle);
    border-bottom-left-radius: var(--radius-lg);
    border-bottom-right-radius: var(--radius-lg);
  }
</style>
