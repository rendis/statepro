import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ExportModal } from "../features/modals";

describe("ExportModal strict mode", () => {
  it("bloquea copy cuando hay errores", () => {
    render(
      <ExportModal
        isOpen
        json="{}"
        canExport={false}
        issues={[
          {
            code: "SEMANTIC_ERROR",
            severity: "error",
            field: "transition:1.targets[0]",
            message: "Invalid target",
          },
        ]}
        onClose={vi.fn()}
      />,
    );

    expect(screen.getByText(/export blocked/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /copy json/i })).toBeDisabled();
  });
});
