// Internal to editor-core and intentionally not re-exported from the package.
// Public checkpoints stay caller-owned and are always copied on commit; only
// private detached copies created by the editor itself are registered here, so
// the history can take ownership of them instead of cloning the graph again.
const ownedSnapshots = new WeakSet<object>();

export const markHistoryOwned = <T extends object>(snapshot: T): T => {
  ownedSnapshots.add(snapshot);
  return snapshot;
};

/** Returns true once for a registered snapshot, transferring it to the history. */
export const takeHistoryOwnership = (snapshot: object): boolean => ownedSnapshots.delete(snapshot);
