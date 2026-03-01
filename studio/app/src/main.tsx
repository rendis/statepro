import React from "react";
import ReactDOM from "react-dom/client";

import "@rendis/statepro-studio-react/styles.css";
import App from "./App";
import "./styles.css";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
