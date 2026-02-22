<script lang="ts">
    import { useSvelteFlow } from "@xyflow/svelte";

    type NodeKind = "node" | "part" | "capability";

    type Props = {
        onCreate: (type: NodeKind, position: { x: number; y: number }) => void;
    };

    let { onCreate }: Props = $props();

    const { screenToFlowPosition } = useSvelteFlow();

  let dragging = $state(false);

  function onDragStart(evt: DragEvent, type: NodeKind): void {
    dragging = true;

    if (!evt.dataTransfer) return;
    evt.dataTransfer.setData("application/x-node-type", type);
    evt.dataTransfer.effectAllowed = "move";
  }

  function onDragEnd(): void {
    dragging = false;
  }

  function onDragOver(evt: DragEvent): void {
    evt.preventDefault();
    if (evt.dataTransfer) evt.dataTransfer.dropEffect = "move";
  }

  function onDrop(evt: DragEvent): void {
    evt.preventDefault();
    dragging = false;

    const type = evt.dataTransfer?.getData("application/x-node-type") as NodeKind | "";
    if (type !== "node" && type !== "part" && type !== "capability") return;

    const position = screenToFlowPosition({ x: evt.clientX, y: evt.clientY });
    onCreate(type, position);
  }
</script>

<div
  class="dnd-overlay"
  class:dragging
  ondragover={onDragOver}
  ondrop={onDrop}
>
  <div class="palette">
    <div class="palette-title">DRAG TO CANVAS</div>

    <button
      class="palette-item node"
      draggable="true"
      ondragstart={(e) => onDragStart(e, "node")}
      ondragend={onDragEnd}
    >
      <span class="dot"></span> Node
    </button>

    <button
      class="palette-item part"
      draggable="true"
      ondragstart={(e) => onDragStart(e, "part")}
      ondragend={onDragEnd}
    >
      <span class="dot"></span> Part
    </button>

    <button
      class="palette-item capability"
      draggable="true"
      ondragstart={(e) => onDragStart(e, "capability")}
      ondragend={onDragEnd}
    >
      <span class="dot"></span> Capability
    </button>
  </div>
</div>

<style>
  .dnd-overlay {
    position: absolute;
    inset: 0;
    z-index: 20;

    /* Default: don't block interacting with the canvas */
    pointer-events: none;
  }

  /* While dragging: the overlay becomes the drop target for the whole canvas */
  .dnd-overlay.dragging {
    pointer-events: auto;
  }

    .palette {
        pointer-events: auto;
        position: absolute;
        top: 12px;
        left: 12px;
        width: 180px;
        background: #0a0d15cc;
        backdrop-filter: blur(8px);
        border: 1px solid #1e2330;
        border-radius: 6px;
        padding: 10px;
    }

    .palette-title {
        font-size: 9px;
        letter-spacing: 0.14em;
        color: #4a5568;
        margin-bottom: 8px;
    }

    .palette-item {
        width: 100%;
        display: flex;
        align-items: center;
        gap: 8px;
        background: #0f1117;
        border: 1px solid #1e2330;
        border-radius: 4px;
        color: #a0aec0;
        cursor: grab;
        font-family: "IBM Plex Mono", monospace;
        font-size: 11px;
        padding: 7px 10px;
        margin-bottom: 6px;
    }

    .palette-item:active {
        cursor: grabbing;
    }

    .dot {
        width: 7px;
        height: 7px;
        border-radius: 50%;
        background: #b794f4;
        flex-shrink: 0;
    }

    .palette-item.node .dot { background: #f6ad55; }
    .palette-item.part .dot { background: #68d391; }
    .palette-item.capability .dot { background: #2a7fff; }
</style>