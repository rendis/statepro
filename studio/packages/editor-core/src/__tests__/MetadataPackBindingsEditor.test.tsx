import "@testing-library/jest-dom";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { MetadataPackBindingsEditor } from "../components/shared";
import type { MetadataPackBindingMap, MetadataPackRegistry } from "../types";
import { buildMachineEntityRef } from "../utils";

const MACHINE_REF = buildMachineEntityRef();

const PACK_BASE = {
  id: "base-profile",
  label: "Base Profile",
  scopes: ["machine"],
  schema: {
    type: "object",
    properties: {
      name: { type: "string" },
    },
  },
} satisfies MetadataPackRegistry[number];

const PACK_NAME_ID = {
  id: "name-id",
  label: "Name Id",
  scopes: ["machine"],
  schema: {
    type: "object",
    properties: {
      name: {
        type: "object",
        properties: {
          id: { type: "string" },
        },
      },
    },
  },
} satisfies MetadataPackRegistry[number];

const PACK_TAGS = {
  id: "tags",
  label: "Tags",
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
} satisfies MetadataPackRegistry[number];

const buildBindings = (machineBindings: MetadataPackBindingMap["machine"]): MetadataPackBindingMap => ({
  machine: machineBindings,
  universe: [],
  reality: [],
  transition: [],
});

describe("MetadataPackBindingsEditor", () => {
  it("allows attaching more than one compatible pack to the same entity", async () => {
    const user = userEvent.setup();

    const StatefulEditor = () => {
      const [bindings, setBindings] = useState<MetadataPackBindingMap>(buildBindings([]));
      return (
        <MetadataPackBindingsEditor
          scope="machine"
          entityRef={MACHINE_REF}
          packRegistry={[PACK_BASE, PACK_TAGS]}
          bindings={bindings}
          metadataRaw="{}"
          onChangeBindings={setBindings}
        />
      );
    };

    render(<StatefulEditor />);

    expect(
      screen.getByText(/This entity has no metadata packs attached/i),
    ).toBeInTheDocument();

    await user.selectOptions(screen.getByRole("combobox"), PACK_BASE.id);
    await user.click(screen.getByRole("button", { name: /attach/i }));

    await user.selectOptions(screen.getByRole("combobox"), PACK_TAGS.id);
    await user.click(screen.getByRole("button", { name: /attach/i }));

    expect(screen.queryByText(/This entity has no metadata packs attached/i)).toBeNull();
    expect(screen.getByText(/^Base Profile$/)).toBeInTheDocument();
    expect(screen.getByText(/^Tags$/)).toBeInTheDocument();
  });

  it("shows detailed block reason when attach collides with existing pack", () => {
    render(
      <MetadataPackBindingsEditor
        scope="machine"
        entityRef={MACHINE_REF}
        packRegistry={[PACK_BASE, PACK_NAME_ID]}
        bindings={
          buildBindings([
            {
              id: "binding-base",
              packId: PACK_BASE.id,
              scope: "machine",
              entityRef: MACHINE_REF,
              values: {},
            },
          ])
        }
        metadataRaw="{}"
        onChangeBindings={vi.fn()}
      />,
    );

    expect(screen.getByText(/name-id:/i)).toBeInTheDocument();
    expect(screen.getByText(/collides with pack "base-profile"/i)).toBeInTheDocument();
    expect(screen.getByText(/descendant/i)).toBeInTheDocument();
  });

  it("shows imported conflict banner and blocks new attaches while collisions remain", () => {
    render(
      <MetadataPackBindingsEditor
        scope="machine"
        entityRef={MACHINE_REF}
        packRegistry={[PACK_BASE, PACK_NAME_ID, PACK_TAGS]}
        bindings={
          buildBindings([
            {
              id: "binding-base",
              packId: PACK_BASE.id,
              scope: "machine",
              entityRef: MACHINE_REF,
              values: {},
            },
            {
              id: "binding-name-id",
              packId: PACK_NAME_ID.id,
              scope: "machine",
              entityRef: MACHINE_REF,
              values: {},
            },
          ])
        }
        metadataRaw="{}"
        onChangeBindings={vi.fn()}
      />,
    );

    expect(screen.getByText(/^Active ownership collisions$/i)).toBeInTheDocument();
    expect(screen.getByText(/state was imported with conflict/i)).toBeInTheDocument();
    expect(screen.getByText(/tags: attachment blocked due to active collisions/i)).toBeInTheDocument();
  });

  it("allows collapsing and expanding an attached pack card", async () => {
    const user = userEvent.setup();

    render(
      <MetadataPackBindingsEditor
        scope="machine"
        entityRef={MACHINE_REF}
        packRegistry={[PACK_BASE]}
        bindings={
          buildBindings([
            {
              id: "binding-base",
              packId: PACK_BASE.id,
              scope: "machine",
              entityRef: MACHINE_REF,
              values: {},
            },
          ])
        }
        metadataRaw="{}"
        onChangeBindings={vi.fn()}
      />,
    );

    expect(screen.getByText(/^name$/i)).toBeInTheDocument();

    expect(screen.queryByText(/colapsar/i)).toBeNull();
    expect(screen.queryByText(/expandir/i)).toBeNull();

    await user.click(screen.getByTestId("toggle-pack-binding-base"));
    expect(screen.getByTestId("collapsed-pack-indicator-binding-base")).toBeInTheDocument();
    expect(screen.queryByText(/^name$/i)).toBeNull();

    await user.click(screen.getByTestId("toggle-pack-binding-base"));
    expect(screen.queryByTestId("collapsed-pack-indicator-binding-base")).toBeNull();
    expect(screen.getByText(/^name$/i)).toBeInTheDocument();
  });
});
