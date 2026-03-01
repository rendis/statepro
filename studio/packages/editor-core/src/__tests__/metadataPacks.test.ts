import { describe, expect, it } from "vitest";

import {
  createInitialEditorState,
  deserializeStatePro,
  serializeStatePro,
} from "../model";
import type {
  MetadataPackBinding,
  MetadataPackDefinition,
  MetadataPackRegistry,
} from "../types";
import {
  buildMachineEntityRef,
  classifyPointerCollision,
  deepMergeJsonObjects,
  findBindingOwnershipCollisionsDetailed,
  findBindingOwnershipCollisions,
  findFieldPathCollisionsDetailed,
  mergePackBindingsToMetadata,
  replacePackIdInBindings,
} from "../utils";

const PACK_A: MetadataPackDefinition = {
  id: "pack-a",
  label: "Pack A",
  scopes: ["machine"],
  schema: {
    type: "object",
    properties: {
      profile: {
        type: "object",
        properties: {
          email: { type: "string" },
        },
      },
    },
  },
};

const PACK_B: MetadataPackDefinition = {
  id: "pack-b",
  label: "Pack B",
  scopes: ["machine"],
  schema: {
    type: "object",
    properties: {
      profile: {
        type: "object",
      },
    },
  },
};

const PACK_C: MetadataPackDefinition = {
  id: "pack-c",
  label: "Pack C",
  scopes: ["machine"],
  schema: {
    type: "object",
    properties: {
      tags: {
        type: "array",
        items: { type: "string" },
      },
    },
  },
};

const PACK_CONSTANT: MetadataPackDefinition = {
  id: "pack-constant",
  label: "Pack Constant",
  scopes: ["machine"],
  schema: {
    type: "object",
    properties: {
      profile: {
        type: "object",
        properties: {
          readonlyId: { type: "string" },
        },
      },
    },
  },
  ui: {
    "/profile/readonlyId": {
      widget: "text",
      constant: true,
      constantValue: "fixed-id",
    },
  },
};

const binding = (id: string, packId: string, values: Record<string, unknown>): MetadataPackBinding => ({
  id,
  packId,
  scope: "machine",
  entityRef: buildMachineEntityRef(),
  values: values as MetadataPackBinding["values"],
});

describe("metadata packs", () => {
  it("classifies exact, ancestor and descendant pointer collisions", () => {
    expect(classifyPointerCollision("/name", "/name")).toBe("exact");
    expect(classifyPointerCollision("/name", "/name/id")).toBe("ancestor");
    expect(classifyPointerCollision("/name/id", "/name")).toBe("descendant");
    expect(classifyPointerCollision("/name", "/profile/id")).toBeNull();
  });

  it("detects ownership collision for nested paths", () => {
    const collisions = findBindingOwnershipCollisions(
      [binding("b-1", PACK_A.id, {}), binding("b-2", PACK_B.id, {})],
      new Map([
        [PACK_A.id, PACK_A],
        [PACK_B.id, PACK_B],
      ]),
    );

    expect(collisions.length).toBeGreaterThan(0);
  });

  it("does not collide for disjoint paths", () => {
    const collisions = findBindingOwnershipCollisions(
      [binding("b-1", PACK_A.id, {}), binding("b-3", PACK_C.id, {})],
      new Map([
        [PACK_A.id, PACK_A],
        [PACK_C.id, PACK_C],
      ]),
    );

    expect(collisions).toHaveLength(0);
  });

  it("returns detailed ownership collisions with relation and pack ids", () => {
    const collisions = findBindingOwnershipCollisionsDetailed(
      [binding("b-1", PACK_A.id, {}), binding("b-2", PACK_B.id, {})],
      new Map([
        [PACK_A.id, PACK_A],
        [PACK_B.id, PACK_B],
      ]),
    );

    expect(collisions.length).toBeGreaterThan(0);
    expect(collisions[0]).toMatchObject({
      leftBindingId: "b-1",
      rightBindingId: "b-2",
      leftPackId: "pack-a",
      rightPackId: "pack-b",
    });
    expect(["exact", "ancestor", "descendant"]).toContain(collisions[0]?.relation);
  });

  it("detects hierarchical collisions in visual field paths", () => {
    const collisions = findFieldPathCollisionsDetailed(["name", "name.id", "profile.id"]);
    const hasNameCollision = collisions.some(
      (collision) =>
        collision.leftPath === "name" &&
        collision.rightPath === "name.id" &&
        collision.relation === "ancestor",
    );
    expect(hasNameCollision).toBe(true);
  });

  it("merges manual metadata and pack metadata with pack precedence", () => {
    const packRegistry: MetadataPackRegistry = [PACK_A];
    const registryIndex = new Map(packRegistry.map((pack) => [pack.id, pack]));

    const manual = {
      profile: {
        email: "manual@example.com",
        locale: "es-CL",
      },
    };

    const packMetadata = mergePackBindingsToMetadata(
      [binding("b-1", PACK_A.id, { profile: { email: "pack@example.com" } })],
      registryIndex,
    );
    const merged = deepMergeJsonObjects(manual, packMetadata);

    expect(merged.profile).toEqual({
      email: "pack@example.com",
      locale: "es-CL",
    });
  });

  it("applies pack constants during merge even when binding value is empty", () => {
    const registryIndex = new Map([[PACK_CONSTANT.id, PACK_CONSTANT]]);
    const packMetadata = mergePackBindingsToMetadata(
      [binding("binding-const", PACK_CONSTANT.id, {})],
      registryIndex,
    );

    expect((packMetadata.profile as Record<string, unknown>)?.readonlyId).toBe("fixed-id");
  });

  it("serializes merged metadata without __studio and resets packs on model-only deserialize", () => {
    const state = createInitialEditorState();

    state.machineConfig.metadata = JSON.stringify(
      {
        author: "manual-author",
        profile: {
          email: "manual@example.com",
        },
      },
      null,
      2,
    );

    state.metadataPackRegistry = [PACK_A];
    state.metadataPackBindings = {
      machine: [
        binding("binding-machine", PACK_A.id, {
          profile: {
            email: "pack@example.com",
          },
        }),
      ],
      universe: [],
      reality: [],
      transition: [],
    };

    const serialized = serializeStatePro(state);
    const machineMetadata = serialized.machine.metadata || {};

    expect((machineMetadata.profile as Record<string, unknown>)?.email).toBe("pack@example.com");
    expect((machineMetadata.__studio as Record<string, unknown>)?.packRegistry).toBeUndefined();

    const rehydrated = deserializeStatePro(serialized.machine);
    expect(rehydrated.metadataPackRegistry).toHaveLength(0);
    expect(rehydrated.metadataPackBindings.machine).toHaveLength(0);

    const parsedManual = JSON.parse(rehydrated.machineConfig.metadata) as Record<string, unknown>;
    expect((parsedManual.profile as Record<string, unknown>)?.email).toBe("pack@example.com");
  });

  it("rewrites binding pack ids when a pack id changes", () => {
    const renamed = replacePackIdInBindings(
      {
        machine: [
          {
            id: "binding-machine",
            packId: "pack-a",
            scope: "machine",
            entityRef: buildMachineEntityRef(),
            values: {},
          },
        ],
        universe: [],
        reality: [],
        transition: [],
      },
      "pack-a",
      "pack-a-renamed",
    );

    expect(renamed.machine[0]?.packId).toBe("pack-a-renamed");
  });

  it("reports detailed collision diagnostics on serialize", () => {
    const state = createInitialEditorState();
    state.metadataPackRegistry = [PACK_A, PACK_B];
    state.metadataPackBindings = {
      machine: [
        binding("binding-a", PACK_A.id, {}),
        binding("binding-b", PACK_B.id, {}),
      ],
      universe: [],
      reality: [],
      transition: [],
    };

    const serialized = serializeStatePro(state);
    const serializeCollision = serialized.issues.find(
      (issue) =>
        issue.field === "machine.metadataPacks" &&
        issue.message.includes("binding 'binding-a'") &&
        issue.message.includes("pack 'pack-a'") &&
        issue.message.includes("pack 'pack-b'") &&
        issue.message.includes("/profile"),
    );
    expect(serializeCollision).toBeDefined();
  });
});
