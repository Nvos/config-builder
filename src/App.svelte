<script lang="ts">
  import {
    SvelteFlow,
    Controls,
    Background,
    BackgroundVariant,
    MiniMap,
    ConnectionMode,
  } from "@xyflow/svelte";
  import "@xyflow/svelte/dist/style.css";

  import CapabilityNode from "./lib/nodes/CapabilityNode.svelte";
  import PartNode from "./lib/nodes/PartNode.svelte";
  import NodeNode from "./lib/nodes/Node.svelte";
  import NodePalette from "./lib/NodePalette.svelte";

  import ELK from "elkjs/lib/elk.bundled.js";
  import { Export } from "./wailsjs/go/main/App";

  type CapabilityData = { id: string; maxValue: number; label?: string };
  type NodeData = { id: string; from: number; to: number; reactive: boolean; label?: string };
  type PartData = { id: string; label?: string };

  type AppNode =
    | { id: string; type: "capability"; position: { x: number; y: number }; data: CapabilityData; selected?: boolean }
    | { id: string; type: "node"; position: { x: number; y: number }; data: NodeData; selected?: boolean }
    | { id: string; type: "part"; position: { x: number; y: number }; data: PartData; selected?: boolean };

  type AppEdge = {
    id: string;
    source: string;
    target: string;
    sourceHandle?: string;
    targetHandle?: string;
    class?: string;
  };

  let nodeIdCounter = $state(1);
  let showJson = $state(false);
  let jsonOutput = $state("");
  let copiedToast = $state(false);

  const nodeTypes = {
    capability: CapabilityNode,
    part: PartNode,
    node: NodeNode,
  } as const;

  let nodes = $state.raw<AppNode[]>([]);
  let edges = $state.raw<AppEdge[]>([]);

  const PROXIMITY_DISTANCE = 170;

  function nodeSizeFallback(type: AppNode["type"]): { w: number; h: number } {
    if (type === "node") return { w: 190, h: 110 };
    if (type === "capability") return { w: 180, h: 95 };
    return { w: 160, h: 90 };
  }

  function getNodeWH(n: any): { w: number; h: number } {
    const fb = nodeSizeFallback(n.type);
    const w = n.measured?.width ?? n.width ?? fb.w;
    const h = n.measured?.height ?? n.height ?? fb.h;
    return { w, h };
  }

  async function autoLayoutElk(): Promise<void> {
    // Don't let preview edge influence layout
    const layoutEdges = edges.filter((e) => e.class !== "temp");

    const elk = new ELK();

    const graph = {
      id: "root",
      layoutOptions: {
        "elk.algorithm": "layered",
        "elk.direction": "RIGHT",
        "elk.layered.spacing.nodeNodeBetweenLayers": "120",
        "elk.spacing.nodeNode": "80",
      },
      children: nodes.map((n) => {
        const { w, h } = getNodeWH(n);
        return {
          id: n.id,
          width: w,
          height: h,
        };
      }),
      edges: layoutEdges.map((e) => ({
        id: e.id,
        sources: [e.source],
        targets: [e.target],
      })),
    };

    const result = await elk.layout(graph);

    // Apply new positions
    const posById = new Map<string, { x: number; y: number }>();
    (result.children ?? []).forEach((c: any) => {
      posById.set(c.id, { x: c.x ?? 0, y: c.y ?? 0 });
    });

    nodes = nodes.map((n) => {
      const p = posById.get(n.id);
      return p ? { ...n, position: p } : n;
    });

    // Remove any temp preview edge after layout
    edges = edges.filter((e) => e.class !== "temp");
  }

  function getNodeBox(n: any): { x: number; y: number; w: number; h: number } {
    // Prefer measured sizes when available (set by the library after layout)
    const fb = nodeSizeFallback(n.type);
    const w = n.measured?.width ?? n.width ?? fb.w;
    const h = n.measured?.height ?? n.height ?? fb.h;

    // Some builds provide positionAbsolute; fall back to position
    const x = n.positionAbsolute?.x ?? n.position.x;
    const y = n.positionAbsolute?.y ?? n.position.y;

    return { x, y, w, h };
  }

  function handlePoint(n: any, handleId: string): { x: number; y: number } {
    const { x, y, w, h } = getNodeBox(n);

    // handle ids:
    // - "out" should be on the right
    // - "in" should be on the left
    if (handleId === "out") return { x: x + w, y: y + h / 2 };
    if (handleId === "in") return { x: x, y: y + h / 2 };

    // Fallback: center
    return { x: x + w / 2, y: y + h / 2 };
  }

  function dist(a: { x: number; y: number }, b: { x: number; y: number }): number {
    const dx = a.x - b.x;
    const dy = a.y - b.y;
    return Math.sqrt(dx * dx + dy * dy);
  }

  type ConnectionLike = {
    source: string | null;
    target: string | null;
    sourceHandle?: string | null;
    targetHandle?: string | null;
  };

  function getNodeTypeById(nodeId: string | null): AppNode["type"] | null {
    if (!nodeId) return null;
    const n = nodes.find((x) => x.id === nodeId);
    return n?.type ?? null;
  }

  function countPartsConnectedToNode(nodeId: string): number {
    return edges.reduce((count, e) => {
      const aType = getNodeTypeById(e.source);
      const bType = getNodeTypeById(e.target);

      const connectsPartToThisNode =
        (e.source === nodeId && bType === "part") ||
        (e.target === nodeId && aType === "part");

      return connectsPartToThisNode ? count + 1 : count;
    }, 0);
  }

  function countNodesConnectedToPart(partId: string): number {
    return edges.reduce((count, e) => {
      const aType = getNodeTypeById(e.source);
      const bType = getNodeTypeById(e.target);

      const connectsNodeToThisPart =
        (e.source === partId && bType === "node") ||
        (e.target === partId && aType === "node");

      return connectsNodeToThisPart ? count + 1 : count;
    }, 0);
  }

  function countCapabilitiesConnectedToNode(nodeId: string): number {
    return edges.reduce((count, e) => {
      const aType = getNodeTypeById(e.source);
      const bType = getNodeTypeById(e.target);

      const connectsCapabilityToThisNode =
        (e.source === nodeId && bType === "capability") ||
        (e.target === nodeId && aType === "capability");

      return connectsCapabilityToThisNode ? count + 1 : count;
    }, 0);
  }

  function isValidConnection(conn: ConnectionLike): boolean {
    const sourceType = getNodeTypeById(conn.source);
    const targetType = getNodeTypeById(conn.target);

    if (!sourceType || !targetType) return false;
    if (conn.source === conn.target) return false;

    const isPartNode =
      (sourceType === "part" && targetType === "node") ||
      (sourceType === "node" && targetType === "part");

    const isNodeCapability =
      (sourceType === "node" && targetType === "capability") ||
      (sourceType === "capability" && targetType === "node");

    if (!isPartNode && !isNodeCapability) return false;

    // If exact connection already exists, allow it (prevents "can't reconnect" weirdness)
    const alreadyExists = edges.some(
      (e) =>
        (e.source === conn.source && e.target === conn.target) ||
        (e.source === conn.target && e.target === conn.source),
    );
    if (alreadyExists) return true;

    // Part <-> Node rules
    if (isPartNode) {
      const nodeId = sourceType === "node" ? (conn.source as string) : (conn.target as string);
      const partId = sourceType === "part" ? (conn.source as string) : (conn.target as string);

      // Part can connect to only 1 Node
      if (countNodesConnectedToPart(partId) >= 1) return false;

      //  Node can have max 3 Parts
      return countPartsConnectedToNode(nodeId) < 3;
    }


    // Node <-> Capability rules
    if (isNodeCapability) {
      const nodeId = sourceType === "node" ? (conn.source as string) : (conn.target as string);

      // Node can connect to only 1 Capability
      return countCapabilitiesConnectedToNode(nodeId) < 1;
    }

    return true;
  }

  function onBeforeConnect(conn: ConnectionLike) {
    if (!isValidConnection(conn)) return null;

    return {
      ...conn,
      id: `e-${conn.source}-${conn.target}-${Date.now()}`,
      style: "stroke: #4a5568; stroke-width: 1.5px;",
    };
  }

  function makeProximityEdge(a: AppNode, b: AppNode): AppEdge | null {
    // Normalize to allowed directed pairs with stable handles:
    // part(out) -> node(in)
    // node(out) -> capability(in)
    let source: AppNode | null = null;
    let target: AppNode | null = null;
    let sourceHandle: string | undefined;
    let targetHandle: string | undefined;

    if (a.type === "part" && b.type === "node") {
      source = a; target = b; sourceHandle = "out"; targetHandle = "in";
    } else if (a.type === "node" && b.type === "part") {
      source = b; target = a; sourceHandle = "out"; targetHandle = "in";
    } else if (a.type === "node" && b.type === "capability") {
      source = a; target = b; sourceHandle = "out"; targetHandle = "in";
    } else if (a.type === "capability" && b.type === "node") {
      source = b; target = a; sourceHandle = "out"; targetHandle = "in";
    } else {
      return null;
    }

    const conn: ConnectionLike = {
      source: source.id,
      target: target.id,
      sourceHandle,
      targetHandle,
    };

    if (!isValidConnection(conn)) return null;

    return {
      id: `temp-${source.id}-${target.id}`,
      source: source.id,
      target: target.id,
      sourceHandle,
      targetHandle,
      class: "temp",
    };
  }

  function onNodeDrag(evt: any): void {
    const dragged: AppNode | undefined = evt?.detail?.node ?? evt?.targetNode ?? evt?.node;
    if (!dragged) return;

    const draggedLive = nodes.find((n) => n.id === dragged.id);
    if (!draggedLive) return;

    let best: { edge: AppEdge; d: number } | null = null;

    for (const other of nodes) {
      if (other.id === draggedLive.id) continue;

      const edge = makeProximityEdge(draggedLive, other);
      if (!edge) continue;

      const sourceNode = nodes.find((n) => n.id === edge.source);
      const targetNode = nodes.find((n) => n.id === edge.target);
      if (!sourceNode || !targetNode) continue;

      const p1 = handlePoint(sourceNode, edge.sourceHandle ?? "");
      const p2 = handlePoint(targetNode, edge.targetHandle ?? "");
      const d = dist(p1, p2);

      if (d > PROXIMITY_DISTANCE) continue;
      if (!best || d < best.d) best = { edge, d };
    }

    const withoutTemp = edges.filter((e) => e.class !== "temp");

    if (!best) {
      edges = withoutTemp;
      return;
    }

    const alreadyExists = withoutTemp.some(
      (e) =>
        (e.source === best.edge.source &&
          e.target === best.edge.target &&
          e.sourceHandle === best.edge.sourceHandle &&
          e.targetHandle === best.edge.targetHandle) ||
        (e.source === best.edge.target && e.target === best.edge.source),
    );

    edges = alreadyExists ? withoutTemp : [...withoutTemp, best.edge];
  }

  function onNodeDragStop(): void {
    const temp = edges.find((e) => e.class === "temp");
    if (!temp) return;

    const conn: ConnectionLike = {
      source: temp.source,
      target: temp.target,
      sourceHandle: temp.sourceHandle ?? null,
      targetHandle: temp.targetHandle ?? null,
    };

    // finalize temp edge only if still valid
    if (!isValidConnection(conn)) {
      edges = edges.filter((e) => e.class !== "temp");
      return;
    }

    edges = edges.map((e) =>
      e.class === "temp"
        ? { ...e, id: `e-${e.source}-${e.target}-${Date.now()}`, class: "" }
        : e,
    );
  }

  function nextId(prefix: string): string {
    return `${prefix}-${nodeIdCounter++}`;
  }

  type NodeKind = "node" | "part" | "capability";

  function addNodeAt(type: NodeKind, position: { x: number; y: number }): void {
    const id = nextId(type);
    const baseData = { id };

    const typeData =
      (
        {
          capability: { ...baseData, maxValue: 50 },
          node: { ...baseData, from: 0, to: 100, reactive: false },
          part: { ...baseData },
        } as const
      )[type];

    nodes = [
      ...nodes,
      {
        id,
        type,
        position,
        data: typeData as any,
      } as any,
    ];
  }

  function buildJson(): string {
    const capabilities = nodes
      .filter((n): n is Extract<AppNode, { type: "capability" }> => n.type === "capability")
      .map((n) => ({ id: n.data.id, maxValue: n.data.maxValue }));

    const nodeItems = nodes
      .filter((n): n is Extract<AppNode, { type: "node" }> => n.type === "node")
      .map((n) => {
        const capEdge = edges.find((e) => e.source === n.id && e.sourceHandle === "range-cap-out");
        const capNode = capEdge ? nodes.find((nd) => nd.id === capEdge.target) : null;

        return {
          id: n.data.id,
          label: n.data.label,
          from: n.data.from,
          to: n.data.to,
          reactive: n.data.reactive,
          connectedCapabilityId: (capNode as any)?.data?.id ?? null,
        };
      });

    const parts = nodes
      .filter((n): n is Extract<AppNode, { type: "part" }> => n.type === "part")
      .map((n) => {
        const rangeEdge = edges.find((e) => e.source === n.id && e.sourceHandle === "part-range-out");
        const rangeNode = rangeEdge ? nodes.find((nd) => nd.id === rangeEdge.target) : null;

        return {
          id: n.data.id,
          label: n.data.label,
          connectedRangeId: (rangeNode as any)?.data?.id ?? null,
        };
      });

    return JSON.stringify({ capabilities, nodes: nodeItems, parts }, null, 2);
  }

  function openJson(): void {
    jsonOutput = buildJson();
    showJson = true;
  }

  function downloadJson(): void {
    jsonOutput = buildJson();
    const blob = new Blob([jsonOutput], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `config-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  }

  function buildGraphPayloadJson(): string {
    return JSON.stringify(
      {
        nodes,
        edges: edges.filter((e) => e.class !== "temp"),
      },
      null,
      2,
    );
  }

  async function exportBinary(): Promise<void> {
    const graphJSON = buildGraphPayloadJson();
    await Export(graphJSON);
  }

  const legend: Array<{ type: string; color: string; desc: string }> = [
    { type: "NODE", color: "#f6ad55", desc: "connects to Capability; receives Part" },
    { type: "PART", color: "#68d391", desc: "connects to Node" },
    { type: "CAPABILITY", color: "#2a7fff", desc: "receives Node" },
  ];
</script>

<main>
  <header>
    <div class="toolbar">
      <button class="action-btn" onclick={autoLayoutElk}>
        ⇥ Auto Layout (ELK)
      </button>
    </div>

    <div class="export-bar">
      <button class="export-btn" onclick={openJson}>{"{}"} View JSON</button>
      <button class="export-btn primary" onclick={exportBinary}>↓ Export</button>
    </div>
  </header>

  <div class="canvas-wrap">
    <SvelteFlow
      bind:nodes
      bind:edges
      {nodeTypes}
      deleteKey="Delete"
      connectionMode={ConnectionMode.Strict}
      defaultEdgeOptions={{ type: "smoothstep" }}
      {isValidConnection}
      onbeforeconnect={onBeforeConnect}
      onnodedrag={onNodeDrag}
      onnodedragstop={onNodeDragStop}
      style="background: #080b12;"
      proOptions={{hideAttribution: true}}
    >
      <Background variant={BackgroundVariant.Dots} gap={24} size={1.2} />
      <Controls />
      <MiniMap
        nodeColor={(n) =>
          n.type === "capability"
            ? "#2a7fff"
            : n.type === "range"
              ? "#f6ad55"
              : n.type === "part"
                ? "#68d391"
                : "#b794f4"}
      />
      <NodePalette onCreate={addNodeAt} />
    </SvelteFlow>

    <div class="legend">
      {#each legend as item}
        <div class="legend-item">
          <span class="legend-dot" style="background:{item.color}"></span>
          <span class="legend-type" style="color:{item.color}">{item.type}</span
          >
          <span class="legend-desc">{item.desc}</span>
        </div>
      {/each}
      <div class="legend-hint">
        ← drag handles to connect · Delete key removes selected
      </div>
    </div>

    <div class="stats">
      <span>{nodes.length} nodes</span>
      <span class="sep-inline">·</span>
      <span>{edges.length} edges</span>
    </div>
  </div>

  {#if showJson}
    <div class="modal-backdrop" onclick={() => (showJson = false)}>
      <div class="modal" onclick={(e) => e.stopPropagation()}>
        <div class="modal-header">
          <span class="modal-title">CONFIGURATION JSON</span>
          <div class="modal-actions">
            <button class="export-btn" onclick={downloadJson}>↓ Download</button
            >
            <button
              class="export-btn close-btn"
              onclick={() => (showJson = false)}>✕ Close</button
            >
          </div>
        </div>
        <pre class="json-pre">{jsonOutput}</pre>
      </div>
    </div>
  {/if}

  {#if copiedToast}
    <div class="toast">Copied to clipboard</div>
  {/if}
</main>

<style>
  :global(*) {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }
  :global(body) {
    background: #080b12;
    color: #e2e8f0;
    font-family: "IBM Plex Mono", monospace;
    overflow: hidden;
    height: 100vh;
  }

  :global(.svelte-flow__edge-path) {
    stroke: #4a5568 !important;
    stroke-width: 1.5px !important;
  }
  :global(.svelte-flow__edge.selected .svelte-flow__edge-path) {
    stroke: #90cdf4 !important;
  }
  :global(.svelte-flow__handle) {
    width: 14px !important;
    height: 14px !important;
    border-radius: 50% !important;
    border: 2px solid #080b12 !important;
    background: #4a5568 !important;
    transition:
      background 0.15s,
      transform 0.15s !important;
  }
  :global(.svelte-flow__handle:hover) {
    background: #90cdf4 !important;
    box-shadow: 0 0 0 4px #90cdf433 !important;
  }
  :global(.svelte-flow__controls) {
    background: #0f1117 !important;
    border: 1px solid #1e2330 !important;
    border-radius: 6px !important;
    overflow: hidden !important;
  }
  :global(.svelte-flow__controls-button) {
    background: transparent !important;
    border-bottom: 1px solid #1e2330 !important;
    color: #718096 !important;
    fill: #718096 !important;
  }
  :global(.svelte-flow__controls-button:hover) {
    background: #1a1f2e !important;
    fill: #e2e8f0 !important;
  }
  :global(.svelte-flow__minimap) {
    border-radius: 6px !important;
    overflow: hidden !important;
  }

  main {
    display: flex;
    flex-direction: column;
    height: 100vh;
    background: #080b12;
  }

  header {
    display: flex;
    align-items: center;
    gap: 20px;
    padding: 0 20px;
    height: 52px;
    background: #0a0d15;
    border-bottom: 1px solid #1e2330;
    flex-shrink: 0;
    z-index: 100;
  }

  .toolbar {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 1;
  }

  .sep {
    width: 1px;
    height: 20px;
    background: #1e2330;
    margin: 0 4px;
  }

  .action-btn {
    background: transparent;
    border: 1px solid #1e2330;
    border-radius: 4px;
    color: #4a5568;
    cursor: pointer;
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    padding: 5px 10px;
    transition: all 0.15s;
  }

  .export-bar {
    display: flex;
    gap: 6px;
    flex-shrink: 0;
  }
  .export-btn {
    background: #0f1117;
    border: 1px solid #1e2330;
    border-radius: 4px;
    color: #718096;
    cursor: pointer;
    font-family: "IBM Plex Mono", monospace;
    font-size: 11px;
    padding: 5px 12px;
    transition: all 0.15s;
  }
  .export-btn:hover {
    background: #1a1f2e;
    color: #e2e8f0;
    border-color: #2d3748;
  }
  .export-btn.primary {
    background: #2a7fff22;
    border-color: #2a7fff66;
    color: #2a7fff;
  }
  .export-btn.primary:hover {
    background: #2a7fff44;
  }

  .canvas-wrap {
    flex: 1;
    position: relative;
    overflow: hidden;
  }

  .legend {
    position: absolute;
    bottom: 16px;
    left: 64px;
    background: #0a0d15cc;
    backdrop-filter: blur(8px);
    border: 1px solid #1e2330;
    border-radius: 6px;
    padding: 10px 14px;
    display: flex;
    flex-direction: column;
    gap: 5px;
    z-index: 10;
    pointer-events: none;
  }
  .legend-item {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .legend-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .legend-type {
    font-size: 9px;
    font-weight: 600;
    letter-spacing: 0.1em;
    width: 78px;
  }
  .legend-desc {
    font-size: 9px;
    color: #4a5568;
  }
  .legend-hint {
    font-size: 9px;
    color: #2d3748;
    margin-top: 4px;
    border-top: 1px solid #1e2330;
    padding-top: 5px;
  }

  .stats {
    position: absolute;
    top: 12px;
    right: 16px;
    background: #0a0d15cc;
    backdrop-filter: blur(8px);
    border: 1px solid #1e2330;
    border-radius: 4px;
    padding: 5px 10px;
    font-size: 10px;
    color: #4a5568;
    z-index: 10;
    pointer-events: none;
  }
  .sep-inline {
    margin: 0 6px;
  }

  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: #000000cc;
    backdrop-filter: blur(4px);
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .modal {
    background: #0a0d15;
    border: 1px solid #2d3748;
    border-radius: 8px;
    width: min(760px, 90vw);
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 24px 64px #00000099;
  }
  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid #1e2330;
    flex-shrink: 0;
  }
  .modal-title {
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.14em;
    color: #718096;
  }
  .modal-actions {
    display: flex;
    gap: 6px;
  }
  .close-btn {
    color: #fc8181 !important;
  }
  .close-btn:hover {
    border-color: #fc8181 !important;
    background: #fc818110 !important;
  }

  .json-pre {
    overflow: auto;
    padding: 16px;
    font-family: "IBM Plex Mono", monospace;
    font-size: 12px;
    line-height: 1.7;
    color: #a0aec0;
    flex: 1;
  }

  .toast {
    position: fixed;
    bottom: 24px;
    right: 24px;
    background: #0f1117;
    border: 1px
  }

  /* Proximity preview edge: dashed and a bit brighter */
  :global(.svelte-flow__edge.temp .svelte-flow__edge-path) {
    stroke: #90cdf4 !important;
    stroke-dasharray: 6 4 !important;
    stroke-width: 2px !important;
  }

  :global(.svelte-flow__minimap) {
    background: #0a0d15cc !important;
    backdrop-filter: blur(8px) !important;
    border: 1px solid #1e2330 !important;
    border-radius: 8px !important;
    overflow: hidden !important;
    box-shadow: 0 12px 32px #00000066 !important;
  }

  :global(.svelte-flow__minimap-mask) {
    fill: #00000066 !important;
  }

  :global(.svelte-flow__minimap-node) {
    stroke: #080b12 !important;
    stroke-width: 1px !important;
  }

  :global(.svelte-flow__minimap-node.selected) {
    stroke: #90cdf4 !important;
    stroke-width: 1.5px !important;
  }

  :global(.svelte-flow__minimap-viewport) {
    stroke: #90cdf4 !important;
    stroke-width: 1.5px !important;
    fill: #90cdf411 !important;
  }

  :global(.svelte-flow__attribution) {
    background: #0a0d15cc !important;
    backdrop-filter: blur(8px) !important;
    border: 1px solid #1e2330 !important;
    border-radius: 6px !important;
    padding: 6px 8px !important;
    color: #718096 !important;
    font-family: "IBM Plex Mono", monospace !important;
    font-size: 10px !important;
    letter-spacing: 0.06em !important;
    box-shadow: 0 12px 32px #00000066 !important;
  }

  :global(.svelte-flow__attribution a) {
    color: #90cdf4 !important;
    text-decoration: none !important;
  }

  :global(.svelte-flow__attribution a:hover) {
    text-decoration: underline !important;
  }
</style>
