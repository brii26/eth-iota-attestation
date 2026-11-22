import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import { loadRegistryDeployment } from "./ethereum";
import "./styles.css";

await loadRegistryDeployment();

createRoot(document.getElementById("root")!).render(<StrictMode><App /></StrictMode>);
