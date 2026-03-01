import React from "react";
import ReactDOM from "react-dom/client";

import "@rendis/statepro-studio-react/styles.css";
import App from "./App";
import "./styles.css";
import favicon from "../assets/favicon.svg";

const upsertFavicon = (href: string): void => {
  const existing = document.querySelector<HTMLLinkElement>("link[rel='icon']");
  if (existing) {
    existing.href = href;
    existing.type = "image/svg+xml";
    return;
  }

  const link = document.createElement("link");
  link.rel = "icon";
  link.type = "image/svg+xml";
  link.href = href;
  document.head.appendChild(link);
};

upsertFavicon(favicon);

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
