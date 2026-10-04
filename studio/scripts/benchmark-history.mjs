// Build editor-core first, then run: node scripts/benchmark-history.mjs
// Reports local timings, not a CI threshold. Run on an idle machine for comparisons.
import { createRequire } from "node:module";
import { performance } from "node:perf_hooks";

const require = createRequire(import.meta.url);
const editor = require("../packages/editor-core/dist/index.cjs");
const results = [];

for (const count of [100, 1000, 5000]) {
  const initial = editor.createInitialEditorState();
  const reality = initial.nodes.find(node => node.type === "reality");
  initial.nodes = [initial.nodes[0], ...Array.from({ length: count }, (_, index) => ({
    ...reality,
    id: `r-${index}`,
    data: {
      ...reality.data,
      id: `state${index}`,
      isInitial: index === 0,
      metadata: JSON.stringify({ payload: "x".repeat(256) }),
    },
  }))];

  for (const operation of ["checkpoint", "record", "coalesce", "undo-redo"]) {
    const samples = [];
    for (let run = 0; run < 6; run++) {
      let state = editor.createInitialEditorHistoryState(initial);
      const edit = index => {
        state = editor.editorHistoryReducer(state, {
          type: "apply-editor-action",
          mode: operation === "coalesce" ? "coalesce" : "record",
          group: "machine-id",
          now: 1000 + index,
          action: {
            type: "set-machine-config",
            payload: { ...state.present.machineConfig, id: `machine-${index}` },
          },
        });
      };
      if (operation === "undo-redo") edit(0);
      const start = performance.now();
      for (let index = 0; index < 20; index++) {
        if (operation === "checkpoint") editor.createHistorySnapshot(initial);
        else if (operation === "undo-redo") {
          state = editor.editorHistoryReducer(state, { type: "undo" });
          state = editor.editorHistoryReducer(state, { type: "redo" });
        } else edit(index);
      }
      if (run > 0) samples.push(performance.now() - start);
    }
    samples.sort((a, b) => a - b);
    results.push({
      realities: count,
      metadataPayloadBytes: 256,
      operation,
      iterations: 20,
      medianMs: Number(samples[2].toFixed(2)),
      maxMs: Number(samples[4].toFixed(2)),
    });
  }
}

console.log(JSON.stringify({ node: process.version, results }, null, 2));
