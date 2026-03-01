import React from "react";
import { createRoot, type Root } from "react-dom/client";

import "editor-core/styles.css";
import {
  StateProEditor,
  type StateProEditorProps,
  type StudioChangePayload,
  type StudioLocale,
} from "editor-core";

export const STUDIO_WEB_COMPONENT_TAG = "statepro-studio";
export const STUDIO_CHANGE_EVENT = "studio-change";
export const STUDIO_LOCALE_EVENT = "studio-locale-change";

type StudioElementProps = Partial<StateProEditorProps>;

const parseBooleanAttribute = (value: string | null): boolean | undefined => {
  if (value == null) {
    return undefined;
  }
  if (value === "" || value.toLowerCase() === "true") {
    return true;
  }
  if (value.toLowerCase() === "false") {
    return false;
  }
  return undefined;
};

const parseNumberAttribute = (value: string | null): number | undefined => {
  if (value == null || value === "") {
    return undefined;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
};

export class StateProStudioElement extends HTMLElement {
  static get observedAttributes(): string[] {
    return [
      "locale",
      "default-locale",
      "show-locale-switcher",
      "persist-locale",
      "change-debounce-ms",
    ];
  }

  private root: Root | null = null;

  private mountNode: HTMLDivElement | null = null;

  private props: StudioElementProps = {};

  connectedCallback(): void {
    if (!this.mountNode) {
      this.mountNode = document.createElement("div");
      this.mountNode.style.width = "100%";
      this.mountNode.style.height = "100%";
      this.appendChild(this.mountNode);
    }

    if (!this.root && this.mountNode) {
      this.root = createRoot(this.mountNode);
    }

    this.renderReactTree();
  }

  disconnectedCallback(): void {
    if (this.root) {
      this.root.unmount();
      this.root = null;
    }
  }

  attributeChangedCallback(
    name: string,
    _oldValue: string | null,
    newValue: string | null,
  ): void {
    if (name === "locale") {
      this.props.locale = (newValue as StudioLocale | null) || undefined;
    }

    if (name === "default-locale") {
      this.props.defaultLocale = (newValue as StudioLocale | null) || undefined;
    }

    if (name === "show-locale-switcher") {
      this.props.showLocaleSwitcher = parseBooleanAttribute(newValue);
    }

    if (name === "persist-locale") {
      this.props.persistLocale = parseBooleanAttribute(newValue);
    }

    if (name === "change-debounce-ms") {
      this.props.changeDebounceMs = parseNumberAttribute(newValue);
    }

    this.renderReactTree();
  }

  get value() {
    return this.props.value;
  }

  set value(nextValue: StateProEditorProps["value"]) {
    this.props.value = nextValue;
    this.renderReactTree();
  }

  get defaultValue() {
    return this.props.defaultValue;
  }

  set defaultValue(nextValue: StateProEditorProps["defaultValue"]) {
    this.props.defaultValue = nextValue;
    this.renderReactTree();
  }

  get universeTemplates() {
    return this.props.universeTemplates;
  }

  set universeTemplates(nextValue: StateProEditorProps["universeTemplates"]) {
    this.props.universeTemplates = nextValue;
    this.renderReactTree();
  }

  get libraryBehaviors() {
    return this.props.libraryBehaviors;
  }

  set libraryBehaviors(nextValue: StateProEditorProps["libraryBehaviors"]) {
    this.props.libraryBehaviors = nextValue;
    this.renderReactTree();
  }

  get features() {
    return this.props.features;
  }

  set features(nextValue: StateProEditorProps["features"]) {
    this.props.features = nextValue;
    this.renderReactTree();
  }

  get onChange() {
    return this.props.onChange;
  }

  set onChange(nextValue: StateProEditorProps["onChange"]) {
    this.props.onChange = nextValue;
    this.renderReactTree();
  }

  get onLocaleChange() {
    return this.props.onLocaleChange;
  }

  set onLocaleChange(nextValue: StateProEditorProps["onLocaleChange"]) {
    this.props.onLocaleChange = nextValue;
    this.renderReactTree();
  }

  private renderReactTree(): void {
    if (!this.root) {
      return;
    }

    const userOnChange = this.props.onChange;
    const userOnLocaleChange = this.props.onLocaleChange;

    this.root.render(
      React.createElement(StateProEditor, {
        ...this.props,
        onChange: (payload: StudioChangePayload) => {
          this.dispatchEvent(
            new CustomEvent<StudioChangePayload>(STUDIO_CHANGE_EVENT, {
              detail: payload,
              bubbles: true,
              composed: true,
            }),
          );
          userOnChange?.(payload);
        },
        onLocaleChange: (locale: StudioLocale) => {
          this.dispatchEvent(
            new CustomEvent<{ locale: StudioLocale }>(STUDIO_LOCALE_EVENT, {
              detail: { locale },
              bubbles: true,
              composed: true,
            }),
          );
          userOnLocaleChange?.(locale);
        },
      }),
    );
  }
}

export const defineStateProStudioElement = (
  tagName: string = STUDIO_WEB_COMPONENT_TAG,
): void => {
  if (!customElements.get(tagName)) {
    customElements.define(tagName, StateProStudioElement);
  }
};

declare global {
  interface HTMLElementTagNameMap {
    "statepro-studio": StateProStudioElement;
  }
}
