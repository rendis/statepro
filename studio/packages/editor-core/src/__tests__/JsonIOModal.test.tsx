import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { JsonIOModal } from "../features/modals/JsonIOModal";

describe("JsonIOModal", () => {
  it("permite habilitar pestañas y reabrir sin cambiar el orden de hooks", () => {
    const props = {
      isOpen: true,
      modelJson: '{"id":"machine"}',
      layoutJson: "{}",
      modelIssues: [],
      canExportModel: true,
      onClose: vi.fn(),
      onImportModel: vi.fn(),
      onImportLayout: vi.fn(),
    };
    const { container, rerender } = render(
      <JsonIOModal {...props} allowImport={false} allowExport={false} />,
    );
    expect(container).toBeEmptyDOMElement();
    rerender(<JsonIOModal {...props} allowImport allowExport={false} />);
    expect(screen.getByRole("textbox")).toBeInTheDocument();
    rerender(<JsonIOModal {...props} isOpen={false} />);
    expect(container).toBeEmptyDOMElement();
    rerender(<JsonIOModal {...props} allowImport={false} allowExport />);
    expect(screen.getByText(props.modelJson)).toBeInTheDocument();
  });
});
