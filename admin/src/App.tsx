import React from "react";
import { BrowserRouter, Routes, Route, NavLink } from "react-router-dom";
import { DashboardPage } from "./pages/DashboardPage.js";
import { GatewaysPage } from "./pages/GatewaysPage.js";
import { UpstreamsPage } from "./pages/UpstreamsPage.js";
import { ThreatsPage } from "./pages/ThreatsPage.js";

export const App: React.FC = () => {
  return (
    <BrowserRouter>
      <div className="qtn-app">
        <nav className="qtn-nav">
          <div className="qtn-nav__brand">QSGW Admin</div>
          <div className="qtn-nav__links">
            <NavLink to="/" end>
              Dashboard
            </NavLink>
            <NavLink to="/gateways">Gateways</NavLink>
            <NavLink to="/upstreams">Upstreams</NavLink>
            <NavLink to="/threats">Threats</NavLink>
          </div>
        </nav>
        <main className="qtn-main">
          <Routes>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/gateways" element={<GatewaysPage />} />
            <Route path="/upstreams" element={<UpstreamsPage />} />
            <Route path="/threats" element={<ThreatsPage />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  );
};
