import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { MetadataPackForm } from "../components/shared/MetadataPackForm";
import type { MetadataPackDefinition } from "../types";

describe("MetadataPackForm", () => {
  it("renders the same labels configured in constructor ui.title", () => {
    const pack: MetadataPackDefinition = {
      id: "profile-pack",
      label: "Profile Pack",
      scopes: ["machine"],
      schema: {
        type: "object",
        properties: {
          user_name: { type: "string", title: "Schema fallback" },
          age: { type: "number" },
        },
      },
      ui: {
        "/user_name": {
          title: "Nombre Visible",
        },
      },
    };

    render(<MetadataPackForm pack={pack} value={{}} onChange={vi.fn()} />);

    expect(screen.getByText(/^Nombre Visible$/)).toBeInTheDocument();
    expect(screen.getByText(/^age$/)).toBeInTheDocument();
    expect(screen.queryByText(/^user name$/i)).toBeNull();
  });
});
