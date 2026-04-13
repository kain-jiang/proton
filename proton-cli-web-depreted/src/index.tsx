import * as React from "react";
import DeployTools from "./components/DeployTools/component.view";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import Success from "./components/Success/component.view";
import { setFavicon, setTitle } from "./core/bootstrap";
import { createRoot } from "react-dom/client";
import "@aishutech/ui/dist/ui.min.css";

function bootstrap() {
  const container = document.getElementById("root");
  const root = createRoot(container!);

  setFavicon();
  setTitle();

  root.render(
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<DeployTools />} />
        <Route path="/success" element={<Success />} />
      </Routes>
    </BrowserRouter>,
  );
}

bootstrap();
