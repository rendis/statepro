import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";

import { StateProEditor } from "../StateProEditor";
import type { StudioExternalValue } from "../types";

class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}

beforeAll(() => {
  vi.stubGlobal("ResizeObserver", ResizeObserverMock);
});

afterAll(() => {
  vi.unstubAllGlobals();
});

afterEach(() => {
  vi.restoreAllMocks();
});

const validValue = (): StudioExternalValue => ({
  definition: {
    id: "external-machine",
    canonicalName: "external-machine",
    version: "1.0.0",
    initials: ["U:main"],
    universes: {
      main: {
        id: "main",
        canonicalName: "main",
        version: "1.0.0",
        initial: "idle",
        realities: {
          idle: { id: "idle", type: "transition", on: { GO: [{ targets: ["done"] }] } },
          done: { id: "done", type: "final" },
        },
      },
    },
  },
});

const malformedDefinitions: unknown[] = [
  null,
  42,
  { universes: { main: null } },
  { universes: { main: { realities: { idle: null } } } },
  { universes: { main: { realities: { idle: { always: {}, on: { GO: {} } } } } } },
];

const asExternalValue = (definition: unknown): StudioExternalValue =>
  ({ definition }) as unknown as StudioExternalValue;

const countRealityNodes = (): number => screen.queryAllByTestId(/reality-node-wrapper-/).length;

describe("StateProEditor con valores externos malformados", () => {
  it.each(malformedDefinitions.map((definition, index) => [index, definition]))(
    "monta el editor vacio en lugar de fallar con defaultValue malformado %#",
    (_index, definition) => {
      const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
      expect(() => render(<StateProEditor defaultValue={asExternalValue(definition)} />)).not.toThrow();
      expect(screen.queryAllByTestId(/universe-node-/).length).toBeGreaterThan(0);
      expect(consoleError).toHaveBeenCalledWith(
        "Ignoring invalid StatePro Studio external value",
        expect.anything(),
      );
    },
  );

  it.each(malformedDefinitions.map((definition, index) => [index, definition]))(
    "conserva el estado actual cuando un value controlado cambia a uno malformado %#",
    (_index, definition) => {
      const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
      const { rerender } = render(<StateProEditor value={validValue()} />);
      const realitiesBefore = countRealityNodes();
      expect(realitiesBefore).toBe(2);

      expect(() => rerender(<StateProEditor value={asExternalValue(definition)} />)).not.toThrow();
      expect(countRealityNodes()).toBe(realitiesBefore);
      expect(consoleError).toHaveBeenCalledWith(
        "Ignoring invalid StatePro Studio external value",
        expect.anything(),
      );
    },
  );
});
