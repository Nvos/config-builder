<script lang="ts">
  import { Handle, Position, useSvelteFlow } from "@xyflow/svelte";

  type NodeData = {
    id: string;
    label?: string;
    from?: number;
    to?: number;
    reactive?: boolean;
  };

  type NodeProps = {
    id: string;
    data: NodeData;
    selected?: boolean;
  };

  let { id, data, selected }: NodeProps = $props();

  let editing = $state(false);

  const { updateNodeData } = useSvelteFlow();

  function onLabelInput(evt: Event): void {
    const el = evt.currentTarget as HTMLInputElement | null;
    updateNodeData(id, { label: el?.value ?? "" });
  }

  function onFromInput(evt: Event): void {
    const el = evt.currentTarget as HTMLInputElement | null;
    updateNodeData(id, { from: Number(el?.value ?? 0) });
  }

  function onToInput(evt: Event): void {
    const el = evt.currentTarget as HTMLInputElement | null;
    updateNodeData(id, { to: Number(el?.value ?? 0) });
  }

  function onReactiveChange(evt: Event): void {
    const el = evt.currentTarget as HTMLInputElement | null;
    updateNodeData(id, { reactive: Boolean(el?.checked) });
  }
</script>

<div class="node range" class:selected>
  <div class="node-header">
    <span class="node-type-tag">NODE</span>
    <button class="edit-btn" onclick={() => (editing = !editing)}>
      {editing ? "✕" : "✎"}
    </button>
  </div>

  {#if editing}
    <div class="node-form">
      <label>
        <span>Label</span>
        <input value={data.label ?? ""} oninput={onLabelInput} placeholder="Range name" />
      </label>

      <div class="row">
        <label>
          <span>From</span>
          <input type="number" value={data.from ?? 0} oninput={onFromInput} placeholder="0" />
        </label>
        <label>
          <span>To</span>
          <input type="number" value={data.to ?? 100} oninput={onToInput} placeholder="100" />
        </label>
      </div>

      <label class="checkbox-row">
        <input type="checkbox" checked={data.reactive ?? false} onchange={onReactiveChange} />
        <span>Reactive</span>
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
        <span class="field-val">{data.label || "—"}</span>
      </div>
      <div class="field">
        <span class="field-key">range</span>
        <span class="field-val accent">{data.from ?? "?"} → {data.to ?? "?"}</span>
      </div>
      <div class="field">
        <span class="field-key">reactive</span>
        <span class="field-val" class:green={data.reactive} class:muted={!data.reactive}>
          {data.reactive ? "true" : "false"}
        </span>
      </div>
    </div>
  {/if}

  <Handle id="out" type="source" position={Position.Right} />
  <Handle id="in" type="target" position={Position.Left} />
</div>

<style>
  .node {
    background: #0f1117;
    border: 1.5px solid #f6ad5544;
    border-radius: 6px;
    min-width: 190px;
    font-family: "IBM Plex Mono", monospace;
    box-shadow: 0 0 20px #f6ad5518;
    transition:
      border-color 0.15s,
      box-shadow 0.15s;
  }
  .node.selected {
    border-color: #f6ad55;
    box-shadow: 0 0 24px #f6ad5544;
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
    color: #f6ad55;
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
  .edit-btn:hover {
    color: #f6ad55;
  }
  .node-body {
    padding: 8px 10px 10px;
  }
  .field {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 3px 0;
    gap: 8px;
  }
  .field-key {
    font-size: 10px;
    color: #4a5568;
  }
  .field-val {
    font-size: 11px;
    color: #cbd5e0;
  }
  .field-val.accent {
    color: #fbd38d;
  }
  .field-val.green {
    color: #68d391;
  }
  .field-val.muted {
    color: #4a5568;
  }
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
    flex: 1;
  }
  .node-form label span {
    font-size: 9px;
    color: #4a5568;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }
  .row {
    display: flex;
    gap: 8px;
  }
  .node-form input[type="text"],
  .node-form input[type="number"] {
    background: #1a1f2e;
    border: 1px solid #2d3748;
    border-radius: 3px;
    color: #e2e8f0;
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    padding: 4px 6px;
    outline: none;
    width: 100%;
    box-sizing: border-box;
  }
  input[type="text"]:focus,
  input[type="number"]:focus {
    border-color: #f6ad55;
  }
  .checkbox-row {
    flex-direction: row !important;
    align-items: center;
    gap: 6px;
  }
  .checkbox-row span {
    font-size: 10px !important;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #718096 !important;
  }
  .checkbox-row input[type="checkbox"] {
    accent-color: #f6ad55;
    width: 14px;
    height: 14px;
    cursor: pointer;
  }
</style>
