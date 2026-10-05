import { StateProEditor } from "@rendis/statepro-studio-react";
import logo from "../assets/logo.svg";
import layoutWorkerUrl from "elkjs/lib/elk-worker.min.js?url";

export default function App() {
  return <StateProEditor logoSrc={logo} logoAlt="@studio logo" autoLayoutWorkerUrl={layoutWorkerUrl} />;
}
