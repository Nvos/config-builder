<script lang="ts">
  import { Handle, Position, useSvelteFlow } from "@xyflow/svelte";

  type PartData = {
    id: string;
    label?: string;
  };

  type NodeProps = {
    id: string;
    data: PartData;
    selected?: boolean;
  };

  let { id, data, selected }: NodeProps = $props();

  let editing = $state(false);

  const { updateNodeData } = useSvelteFlow();

  function onLabelInput(evt: Event): void {
    const el = evt.currentTarget as HTMLInputElement | null;
    updateNodeData(id, { label: el?.value ?? "" });
  }
</script>

<div class="node part" class:selected>
  <div class="node-header">
    <span class="node-type-tag">PART</span>
    <button class="edit-btn" onclick={() => editing = !editing}>
      {editing ? '✕' : '✎'}
    </button>
  </div>

  {#if editing}
    <div class="node-form">
      <label>
        <span>Label</span>
        <input value={data.label ?? ""} oninput={onLabelInput} placeholder="Part name" />
      </label>
    </div>
  {:else}
    <div class="node-body">
      <div class="field">
        <span class="field-key">id</span>
        <span class="field-val">{data.id}</span>
      </div>
      <div class="field">
        <span class="field-key">label</span>
        <span class="field-val">{data.label || '—'}</span>
      </div>
    </div>
  {/if}

  <!-- Part connects to Node (right handle out) -->
  <Handle id="out" type="source" position={Position.Right} />
</div>

<style>
  .node {
    background: #0f1117;
    border: 1.5px solid #68d39144;
    border-radius: 6px;
    min-width: 160px;
    font-family: 'IBM Plex Mono', monospace;
    box-shadow: 0 0 20px #68d39118;
    transition: border-color 0.15s, box-shadow 0.15s;
  }
  .node.selected {
    border-color: #68d391;
    box-shadow: 0 0 24px #68d39144;
  }
  .node-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 10px 6px;
    border-bottom: 1px solid #1e2330;
  }
  .node-type-tag {
    font-size: 9px;
    font-weight: 600;
    letter-spacing: 0.12em;
    color: #68d391;
  }
  .edit-btn {
    background: none;
    border: none;
    color: #4a5568;
    cursor: pointer;
    font-size: 13px;
    padding: 0 2px;
    line-height: 1;
    transition: color 0.15s;
  }
  .edit-btn:hover { color: #68d391; }
  .node-body { padding: 8px 10px 10px; }
  .field {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 3px 0;
    gap: 8px;
  }
  .field-key { font-size: 10px; color: #4a5568; }
  .field-val { font-size: 11px; color: #cbd5e0; }
  .node-form {
    padding: 8px 10px 10px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .node-form label {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }
  .node-form label span {
    font-size: 9px;
    color: #4a5568;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }
  .node-form input {
    background: #1a1f2e;
    border: 1px solid #2d3748;
    border-radius: 3px;
    color: #e2e8f0;
    font-family: 'IBM Plex Mono', monospace;
    font-size: 11px;
    padding: 4px 6px;
    outline: none;
    width: 100%;
    box-sizing: border-box;
  }
  .node-form input:focus { border-color: #68d391; }
</style>
