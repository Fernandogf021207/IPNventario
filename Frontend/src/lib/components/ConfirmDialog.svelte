<script lang="ts">
  import Modal from './Modal.svelte';
  import { createEventDispatcher } from 'svelte';

  export let open = false;
  export let title = 'Confirmar acción';
  export let message = '¿Estás seguro de que deseas continuar?';
  export let confirmText = 'Confirmar';
  export let cancelText = 'Cancelar';
  export let danger = false;

  const dispatch = createEventDispatcher();

  function onConfirm() {
    dispatch('confirm');
    open = false;
  }

  function onCancel() {
    dispatch('cancel');
    open = false;
  }
</script>

<Modal bind:open {title} on:close={onCancel}>
  <p>{message}</p>
  
  <div slot="footer">
    <button class="btn btn-secondary" on:click={onCancel}>
      {cancelText}
    </button>
    <button 
      class="btn" 
      class:btn-primary={!danger}
      style={danger ? 'background-color: var(--error); color: white;' : ''}
      on:click={onConfirm}
    >
      {confirmText}
    </button>
  </div>
</Modal>
