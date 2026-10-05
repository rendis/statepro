import React, { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { StateProEditorProps, StudioChangePayload } from "@rendis/statepro-studio-react";

const renderState = vi.hoisted(() => ({ props: null as StateProEditorProps | null }));
vi.mock("@rendis/statepro-studio-react", () => ({
  StateProEditor: (props: StateProEditorProps) => {
    renderState.props = props;
    return React.createElement("button", {
      onClick: () => props.onLocaleChange?.("es"),
    }, "Cambiar idioma");
  },
}));

import { defineStateProStudioElement, StateProStudioElement } from "../index";

(globalThis as typeof globalThis & { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true;
defineStateProStudioElement();

afterEach(async () => { await act(async () => { document.body.replaceChildren(); }); });

describe("StateProStudioElement", () => {
  it("actualiza atributos y propiedades y desmonta al desconectar", async () => {
    const element = document.createElement("statepro-studio");
    await act(async () => { document.body.append(element); });
    await act(async () => {
      element.setAttribute("locale", "es");
      element.setAttribute("show-locale-switcher", "false");
      element.setAttribute("persist-locale", "");
      element.setAttribute("change-debounce-ms", "125");
      element.setAttribute("auto-layout-worker-url", "/elk-worker.js");
      element.features = { json: { import: false, export: true } };
    });
    expect(renderState.props).toMatchObject({ locale: "es", showLocaleSwitcher: false, persistLocale: true, changeDebounceMs: 125 });
    expect(renderState.props?.features).toEqual(element.features);
    expect(element.autoLayoutWorkerUrl).toBe("/elk-worker.js");
    expect(renderState.props?.autoLayoutWorkerUrl).toBe("/elk-worker.js");
    await act(async () => {
      element.removeAttribute("locale");
      element.setAttribute("change-debounce-ms", "invalid");
      element.setAttribute("persist-locale", "invalid");
      element.removeAttribute("auto-layout-worker-url");
    });
    expect(renderState.props?.locale).toBeUndefined();
    expect(renderState.props?.changeDebounceMs).toBeUndefined();
    expect(renderState.props?.persistLocale).toBeUndefined();
    expect(renderState.props?.autoLayoutWorkerUrl).toBeUndefined();
    await act(async () => { element.autoLayoutWorkerUrl = "/other-worker.js"; });
    expect(renderState.props?.autoLayoutWorkerUrl).toBe("/other-worker.js");
    await act(async () => { element.remove(); });
    expect(element.querySelector("button")).toBeNull();
    await act(async () => { document.body.append(element); });
    expect(element.querySelectorAll("button")).toHaveLength(1);
    expect(element.children).toHaveLength(1);
  });

  it("propaga eventos DOM y callbacks una sola vez", async () => {
    const element = document.createElement("statepro-studio");
    const onChange = vi.fn(), onLocale = vi.fn();
    const domChange = vi.fn(), domLocale = vi.fn();
    element.onChange = onChange;
    element.onLocaleChange = onLocale;
    document.body.addEventListener("studio-change", domChange);
    document.body.addEventListener("studio-locale-change", domLocale);
    try {
      await act(async () => { document.body.append(element); });
      const payload = { marker: "change" } as unknown as StudioChangePayload;
      renderState.props?.onChange?.(payload);
      await act(async () => { element.querySelector("button")!.click(); });
      expect(onChange).toHaveBeenCalledExactlyOnceWith(payload);
      expect(onLocale).toHaveBeenCalledExactlyOnceWith("es");
      expect(domChange).toHaveBeenCalledTimes(1);
      expect(domLocale).toHaveBeenCalledTimes(1);
      const event = domChange.mock.calls[0][0] as CustomEvent;
      expect(event.detail).toBe(payload);
      expect(event.bubbles && event.composed).toBe(true);
      expect((domLocale.mock.calls[0][0] as CustomEvent).detail).toEqual({ locale: "es" });
    } finally {
      document.body.removeEventListener("studio-change", domChange);
      document.body.removeEventListener("studio-locale-change", domLocale);
    }
  });

  it("permite registrar varios nombres e inicializarlos de forma idempotente", async () => {
    defineStateProStudioElement("statepro-custom-one");
    defineStateProStudioElement("statepro-custom-two");
    const constructor = customElements.get("statepro-custom-one");
    defineStateProStudioElement("statepro-custom-one");
    expect(customElements.get("statepro-custom-one")).toBe(constructor);
    const element = document.createElement("statepro-custom-two");
    expect(element).toBeInstanceOf(StateProStudioElement);
    await act(async () => { document.body.append(element); });
    expect(element.querySelector("button")).not.toBeNull();
  });
});
