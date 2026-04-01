<script lang="ts">
  import { onMount } from 'svelte';
  import { inventoryApi } from '../../lib/api';
  import type { Item, Category } from '../../lib/types';
  import Modal from '../../lib/components/Modal.svelte';
  import ConfirmDialog from '../../lib/components/ConfirmDialog.svelte';
  import StatusBadge from '../../lib/components/StatusBadge.svelte';

  let items: Item[] = [];
  let categories: Category[] = [];
  let loading = true;
  let error = '';

  // Filters
  let filterType = '';
  let filterCategory = '';
  let filterSearch = '';

  // Item modal
  let isItemModalOpen = false;
  let isEditing = false;
  let formData: any = {};

  // Stock adjustment modal
  let isStockModalOpen = false;
  let stockData = { quantity: 0, notes: '' };
  let selectedItem: Item | null = null;

  // Category modal
  let isCategoryModalOpen = false;
  let categoryFormData = { name: '', description: '' };

  // Confirm delete
  let isConfirmOpen = false;
  let deleteTarget: Item | null = null;

  async function loadData() {
    loading = true;
    error = '';
    try {
      const params: Record<string, string> = {};
      if (filterType) params.type = filterType;
      if (filterCategory) params.category_id = filterCategory;
      if (filterSearch) params.search = filterSearch;
      
      const [itemsData, categoriesData] = await Promise.all([
        inventoryApi.listItems(params),
        inventoryApi.listCategories()
      ]);
      items = itemsData;
      categories = categoriesData;
    } catch (err: any) {
      error = err.message || 'Error al cargar inventario';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadData();
  });

  function openCreateModal() {
    isEditing = false;
    formData = {
      sku: '',
      name: '',
      item_type: 'tool',
      category_id: null,
      stock: 0,
      min_stock: 0,
      unit: 'pza',
      maintenance_status: 'ok',
      location: '',
      module_data: ''
    };
    isItemModalOpen = true;
  }

  function openEditModal(item: Item) {
    isEditing = true;
    selectedItem = item;
    formData = {
      id: item.id,
      sku: item.sku,
      name: item.name,
      item_type: item.item_type || (item as any).type,
      category_id: item.category_id || null,
      stock: item.stock,
      min_stock: item.min_stock,
      unit: item.unit,
      maintenance_status: item.maintenance_status,
      location: item.location || '',
      module_data: item.module_data || ''
    };
    isItemModalOpen = true;
  }

  function openStockModal(item: Item) {
    selectedItem = item;
    stockData = { quantity: 0, notes: '' };
    isStockModalOpen = true;
  }

  function confirmDelete(item: Item) {
    deleteTarget = item;
    isConfirmOpen = true;
  }

  async function handleItemSubmit() {
    try {
      const data = {
        ...formData,
        category_id: formData.category_id ? Number(formData.category_id) : null,
        stock: Number(formData.stock),
        min_stock: Number(formData.min_stock),
      };

      if (isEditing && formData.id) {
        await inventoryApi.updateItem(formData.id, data);
      } else {
        await inventoryApi.createItem(data);
      }
      isItemModalOpen = false;
      await loadData();
    } catch (err: any) {
      alert(err.message || 'Error al guardar el item');
    }
  }

  async function handleStockAdjust() {
    if (!selectedItem) return;
    try {
      await inventoryApi.adjustStock(selectedItem.id, {
        quantity: Number(stockData.quantity),
        notes: stockData.notes
      });
      isStockModalOpen = false;
      await loadData();
    } catch (err: any) {
      alert(err.message || 'Error al ajustar stock');
    }
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    try {
      await inventoryApi.deleteItem(deleteTarget.id);
      isConfirmOpen = false;
      deleteTarget = null;
      await loadData();
    } catch (err: any) {
      alert(err.message || 'Error al eliminar item');
    }
  }

  async function handleCategorySubmit() {
    try {
      await inventoryApi.createCategory(categoryFormData);
      isCategoryModalOpen = false;
      categoryFormData = { name: '', description: '' };
      await loadData();
    } catch (err: any) {
      alert(err.message || 'Error al crear categoría');
    }
  }

  function getTypeLabel(type: string) {
    switch(type) {
      case 'tool': return 'Herramienta';
      case 'consumable': return 'Consumible';
      case 'machine': return 'Maquinaria';
      default: return type;
    }
  }

  function getTypeClass(type: string) {
    switch(type) {
      case 'tool': return 'type-tool';
      case 'consumable': return 'type-consumable';
      case 'machine': return 'type-machine';
      default: return '';
    }
  }

  function getMaintenanceClass(status: string) {
    switch(status) {
      case 'ok': return 'status-ok';
      case 'scheduled': return 'status-scheduled';
      case 'in_maintenance': return 'status-maintenance';
      case 'critical': return 'status-critical';
      case 'out_of_service': return 'status-out';
      default: return '';
    }
  }

  function getMaintenanceLabel(status: string) {
    switch(status) {
      case 'ok': return 'OK';
      case 'scheduled': return 'Programado';
      case 'in_maintenance': return 'En Mant.';
      case 'critical': return 'Crítico';
      case 'out_of_service': return 'Fuera de Servicio';
      default: return status;
    }
  }

  $: filteredItems = items;

  function handleFilter() {
    loadData();
  }
</script>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Inventario</h1>
      <p class="page-subtitle">Gestión de herramientas, consumibles y maquinaria</p>
    </div>
    <div class="header-actions">
      <button class="btn btn-secondary" on:click={() => { categoryFormData = { name: '', description: '' }; isCategoryModalOpen = true; }}>
        + Categoría
      </button>
      <button class="btn btn-primary" on:click={openCreateModal}>
        + Nuevo Item
      </button>
    </div>
  </div>

  <!-- Filters -->
  <div class="card filter-bar">
    <div class="filter-group">
      <label for="filter-search">Buscar</label>
      <input class="input" id="filter-search" placeholder="Nombre o SKU..." bind:value={filterSearch} on:input={handleFilter} />
    </div>
    <div class="filter-group">
      <label for="filter-type">Tipo</label>
      <select class="input form-select" id="filter-type" bind:value={filterType} on:change={handleFilter}>
        <option value="">Todos</option>
        <option value="tool">Herramienta</option>
        <option value="consumable">Consumible</option>
        <option value="machine">Maquinaria</option>
      </select>
    </div>
    <div class="filter-group">
      <label for="filter-category">Categoría</label>
      <select class="input form-select" id="filter-category" bind:value={filterCategory} on:change={handleFilter}>
        <option value="">Todas</option>
        {#each categories as cat}
          <option value={cat.id}>{cat.name}</option>
        {/each}
      </select>
    </div>
  </div>

  {#if error}
    <div class="error-alert">{error}</div>
  {/if}

  <div class="card table-container">
    {#if loading}
      <div class="text-center p-8 text-muted">Cargando...</div>
    {:else if items.length === 0}
      <div class="text-center p-8 text-muted">No hay items registrados.</div>
    {:else}
      <table class="data-table">
        <thead>
          <tr>
            <th>SKU</th>
            <th>Nombre</th>
            <th>Tipo</th>
            <th>Categoría</th>
            <th>Stock</th>
            <th>Mant.</th>
            <th>Ubicación</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {#each items as item}
            <tr class:low-stock={item.stock <= (item.min_stock || 0) && item.item_type !== 'machine' && item.type !== 'machine'}>
              <td class="code">{item.sku}</td>
              <td>
                <div class="font-medium">{item.name}</div>
              </td>
              <td>
                <span class="type-badge {getTypeClass(item.item_type || (item as any).type)}">{getTypeLabel(item.item_type || (item as any).type)}</span>
              </td>
              <td>{item.category_name || '—'}</td>
              <td>
                {#if item.item_type !== 'machine' && (item as any).type !== 'machine'}
                  <span class="stock-value" class:stock-low={item.stock <= (item.min_stock || 0)}>
                    {item.stock} {item.unit}
                  </span>
                  {#if item.stock <= (item.min_stock || 0)}
                    <span class="stock-alert">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="display:inline;vertical-align:text-bottom"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
                      Bajo
                    </span>
                  {/if}
                {:else}
                  <span class="text-muted">N/A</span>
                {/if}
              </td>
              <td>
                <span class="maintenance-badge {getMaintenanceClass(item.maintenance_status)}">
                  {getMaintenanceLabel(item.maintenance_status)}
                </span>
              </td>
              <td class="text-sm">{item.location || '—'}</td>
              <td class="actions">
                {#if item.item_type !== 'machine' && (item as any).type !== 'machine'}
                  <button class="btn-icon" on:click={() => openStockModal(item)} title="Ajustar Stock">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14"/><path d="M5 12h14"/></svg>
                  </button>
                {/if}
                <button class="btn-icon" on:click={() => openEditModal(item)} title="Editar">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
                </button>
                <button class="btn-icon text-error" on:click={() => confirmDelete(item)} title="Desactivar">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>

  <!-- Summary cards -->
  {#if !loading && items.length > 0}
    <div class="summary-grid">
      <div class="summary-card">
        <div class="summary-value">{items.length}</div>
        <div class="summary-label">Items Totales</div>
      </div>
      <div class="summary-card">
        <div class="summary-value">{items.filter(i => i.item_type === 'tool' || i.type === 'tool').length}</div>
        <div class="summary-label">Herramientas</div>
      </div>
      <div class="summary-card">
        <div class="summary-value">{items.filter(i => i.item_type === 'consumable' || i.type === 'consumable').length}</div>
        <div class="summary-label">Consumibles</div>
      </div>
      <div class="summary-card">
        <div class="summary-value">{items.filter(i => i.item_type === 'machine' || i.type === 'machine').length}</div>
        <div class="summary-label">Maquinaria</div>
      </div>
      <div class="summary-card warning">
        <div class="summary-value">{items.filter(i => i.stock <= (i.min_stock || 0) && i.item_type !== 'machine' && i.type !== 'machine').length}</div>
        <div class="summary-label">Stock Bajo</div>
      </div>
    </div>
  {/if}
</div>

<!-- Create/Edit Item Modal -->
<Modal bind:open={isItemModalOpen} title={isEditing ? 'Editar Item' : 'Nuevo Item'}>
  <form id="item-form" on:submit|preventDefault={handleItemSubmit}>
    <div class="form-grid">
      <div class="form-group">
        <label for="sku">SKU *</label>
        <input class="input" id="sku" bind:value={formData.sku} required placeholder="Ej. HER-001" />
      </div>
      <div class="form-group">
        <label for="name">Nombre *</label>
        <input class="input" id="name" bind:value={formData.name} required placeholder="Ej. Fresa de 10mm" />
      </div>
      <div class="form-group">
        <label for="item_type">Tipo *</label>
        <select class="input form-select" id="item_type" bind:value={formData.item_type} required>
          <option value="tool">Herramienta</option>
          <option value="consumable">Consumible</option>
          <option value="machine">Maquinaria</option>
        </select>
      </div>
      <div class="form-group">
        <label for="category_id">Categoría</label>
        <select class="input form-select" id="category_id" bind:value={formData.category_id}>
          <option value={null}>Sin categoría</option>
          {#each categories as cat}
            <option value={cat.id}>{cat.name}</option>
          {/each}
        </select>
      </div>
      {#if !isEditing}
        <div class="form-group">
          <label for="stock">Stock Inicial</label>
          <input type="number" step="0.01" class="input" id="stock" bind:value={formData.stock} />
        </div>
      {/if}
      <div class="form-group">
        <label for="min_stock">Stock Mínimo</label>
        <input type="number" step="0.01" class="input" id="min_stock" bind:value={formData.min_stock} />
      </div>
      <div class="form-group">
        <label for="unit">Unidad</label>
        <select class="input form-select" id="unit" bind:value={formData.unit}>
          <option value="pza">Pieza (pza)</option>
          <option value="m">Metro (m)</option>
          <option value="kg">Kilogramo (kg)</option>
          <option value="lt">Litro (lt)</option>
          <option value="set">Set</option>
          <option value="par">Par</option>
        </select>
      </div>
      <div class="form-group">
        <label for="location">Ubicación</label>
        <input class="input" id="location" bind:value={formData.location} placeholder="Ej. Estante A-3" />
      </div>
    </div>
  </form>

  <div slot="footer">
    <button class="btn btn-secondary" on:click={() => isItemModalOpen = false}>Cancelar</button>
    <button class="btn btn-primary" type="submit" form="item-form">
      {isEditing ? 'Actualizar' : 'Crear'} Item
    </button>
  </div>
</Modal>

<!-- Stock Adjustment Modal -->
<Modal bind:open={isStockModalOpen} title="Ajustar Stock — {selectedItem?.name}">
  <form id="stock-form" on:submit|preventDefault={handleStockAdjust}>
    <div class="stock-info">
      <span>Stock actual: <strong>{selectedItem?.stock} {selectedItem?.unit}</strong></span>
    </div>
    <div class="form-grid">
      <div class="form-group full-width">
        <label for="stock-qty">Cantidad (positivo = entrada, negativo = salida)</label>
        <input type="number" step="0.01" class="input" id="stock-qty" bind:value={stockData.quantity} required />
      </div>
      <div class="form-group full-width">
        <label for="stock-notes">Notas / Razón del ajuste</label>
        <textarea class="input" id="stock-notes" bind:value={stockData.notes} rows="2" placeholder="Ej. Compra de material, ajuste por inventario físico..."></textarea>
      </div>
    </div>
    {#if stockData.quantity !== 0 && selectedItem}
      <div class="stock-preview" class:stock-low={(selectedItem.stock + stockData.quantity) < 0}>
        Stock resultante: <strong>{(selectedItem.stock + stockData.quantity).toFixed(2)} {selectedItem.unit}</strong>
        {#if (selectedItem.stock + stockData.quantity) < 0}
          <span class="stock-alert">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="display:inline;vertical-align:text-bottom"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
            No se puede tener stock negativo
          </span>
        {/if}
      </div>
    {/if}
  </form>

  <div slot="footer">
    <button class="btn btn-secondary" on:click={() => isStockModalOpen = false}>Cancelar</button>
    <button class="btn btn-primary" type="submit" form="stock-form" 
      disabled={stockData.quantity === 0 || (selectedItem && (selectedItem.stock + stockData.quantity) < 0)}>
      Ajustar Stock
    </button>
  </div>
</Modal>

<!-- Category Modal -->
<Modal bind:open={isCategoryModalOpen} title="Nueva Categoría">
  <form id="category-form" on:submit|preventDefault={handleCategorySubmit}>
    <div class="form-grid">
      <div class="form-group full-width">
        <label for="cat-name">Nombre *</label>
        <input class="input" id="cat-name" bind:value={categoryFormData.name} required placeholder="Ej. Herramienta de corte" />
      </div>
      <div class="form-group full-width">
        <label for="cat-desc">Descripción</label>
        <textarea class="input" id="cat-desc" bind:value={categoryFormData.description} rows="2"></textarea>
      </div>
    </div>
  </form>

  <div slot="footer">
    <button class="btn btn-secondary" on:click={() => isCategoryModalOpen = false}>Cancelar</button>
    <button class="btn btn-primary" type="submit" form="category-form">Crear Categoría</button>
  </div>
</Modal>

<!-- Confirm Delete -->
<ConfirmDialog
  bind:open={isConfirmOpen}
  title="Desactivar Item"
  message={`¿Estás seguro de que deseas desactivar "${deleteTarget?.name}"? El item no será eliminado, pero dejará de aparecer en el inventario activo.`}
  confirmText="Sí, desactivar"
  danger={true}
  on:confirm={handleDelete}
/>

<style>
  .page-container { display: flex; flex-direction: column; gap: 1.5rem; }
  .page-header { display: flex; justify-content: space-between; align-items: flex-end; flex-wrap: wrap; gap: 1rem; }
  .page-title { margin: 0; font-size: 1.875rem; }
  .page-subtitle { margin: 0; color: var(--text-muted); }
  .header-actions { display: flex; gap: 0.5rem; }

  .filter-bar { display: flex; gap: 1rem; align-items: flex-end; flex-wrap: wrap; }
  .filter-group { display: flex; flex-direction: column; gap: 0.25rem; flex: 1; min-width: 150px; }
  .filter-group label { font-size: 0.75rem; font-weight: 500; color: var(--text-muted); }

  .table-container { padding: 0; overflow-x: auto; }
  .data-table { width: 100%; border-collapse: collapse; text-align: left; }
  .data-table th, .data-table td { padding: 0.75rem 1rem; border-bottom: 1px solid var(--border); }
  .data-table th { background-color: var(--bg-subtle); color: var(--text-muted); font-size: 0.8rem; font-weight: 500; white-space: nowrap; }
  .data-table tr:last-child td { border-bottom: none; }
  .data-table tr.low-stock { background-color: var(--warning-bg); }
  .code { font-family: var(--mono); color: var(--text-muted); font-size: 0.8rem; }
  .font-medium { font-weight: 500; }
  .text-sm { font-size: 0.8rem; }

  .type-badge { display: inline-block; padding: 0.2rem 0.6rem; border-radius: var(--radius-full); font-size: 0.75rem; font-weight: 500; }
  .type-tool { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
  .type-consumable { background: rgba(16, 185, 129, 0.1); color: #10b981; }
  .type-machine { background: rgba(168, 85, 247, 0.1); color: #a855f7; }

  .stock-value { font-weight: 600; }
  .stock-low { color: var(--error); }
  .stock-alert { font-size: 0.7rem; color: var(--warning); font-weight: 600; display: block; }

  .maintenance-badge { display: inline-block; padding: 0.2rem 0.5rem; border-radius: var(--radius-full); font-size: 0.7rem; font-weight: 500; }
  .status-ok { background: var(--success-bg); color: var(--success); }
  .status-scheduled { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
  .status-maintenance { background: var(--warning-bg); color: var(--warning); }
  .status-critical { background: var(--error-bg); color: var(--error); }
  .status-out { background: rgba(107, 114, 128, 0.1); color: #6b7280; }

  .actions { display: flex; gap: 0.25rem; }
  .btn-icon { background: none; border: none; color: var(--text-muted); cursor: pointer; padding: 0.25rem; border-radius: var(--radius-sm); transition: all 0.2s; }
  .btn-icon:hover { color: var(--accent); background-color: var(--bg-subtle); }
  .text-error:hover { color: var(--error); }

  .summary-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); gap: 1rem; }
  .summary-card { background-color: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 1.25rem; text-align: center; }
  .summary-card.warning { border-color: var(--warning); }
  .summary-value { font-size: 2rem; font-weight: 700; color: var(--text-h); }
  .summary-card.warning .summary-value { color: var(--warning); }
  .summary-label { font-size: 0.813rem; color: var(--text-muted); margin-top: 0.25rem; }

  .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  .form-group { display: flex; flex-direction: column; gap: 0.25rem; }
  .full-width { grid-column: span 2; }
  .form-group label { font-size: 0.875rem; font-weight: 500; }
  .form-select { appearance: none; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2371717a'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E"); background-size: 1.5em 1.5em; background-repeat: no-repeat; background-position: right 0.5rem center; padding-right: 2.5rem; }
  textarea.input { resize: vertical; min-height: 60px; }
  .text-center { text-align: center; }
  .p-8 { padding: 2rem; }
  .text-muted { color: var(--text-muted); }
  .error-alert { background-color: var(--error-bg); color: var(--error); padding: 1rem; border-radius: var(--radius-md); border: 1px solid rgba(239, 68, 68, 0.3); }

  .stock-info { padding: 1rem; background-color: var(--bg-subtle); border-radius: var(--radius-md); margin-bottom: 1rem; }
  .stock-preview { margin-top: 1rem; padding: 0.75rem; background-color: var(--bg-subtle); border-radius: var(--radius-md); font-size: 0.9rem; }
  .stock-preview.stock-low { background-color: var(--error-bg); color: var(--error); }
</style>
