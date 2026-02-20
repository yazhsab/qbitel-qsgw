import React, { useEffect, useState } from "react";
import { Card } from "../components/ui.js";

interface Stats {
  totalGateways: number;
  activeGateways: number;
  totalUpstreams: number;
  recentThreats: number;
}

export const DashboardPage: React.FC = () => {
  const [stats, setStats] = useState<Stats>({
    totalGateways: 0,
    activeGateways: 0,
    totalUpstreams: 0,
    recentThreats: 0,
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const [gwRes, upRes] = await Promise.all([
          fetch("/api/v1/gateways?limit=1"),
          fetch("/api/v1/upstreams?limit=1"),
        ]);

        const gwData = await gwRes.json();
        const upData = await upRes.json();

        setStats({
          totalGateways: gwData.total_count ?? 0,
          activeGateways: 0,
          totalUpstreams: upData.total_count ?? 0,
          recentThreats: 0,
        });
      } catch {
        // API not available
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  if (loading) {
    return <div className="qtn-loading">Loading dashboard...</div>;
  }

  return (
    <div className="qtn-dashboard">
      <h1>Gateway Dashboard</h1>
      <div className="qtn-dashboard__grid">
        <Card title="Total Gateways">
          <div className="qtn-stat">{stats.totalGateways}</div>
        </Card>
        <Card title="Active Gateways">
          <div className="qtn-stat">{stats.activeGateways}</div>
        </Card>
        <Card title="Upstreams">
          <div className="qtn-stat">{stats.totalUpstreams}</div>
        </Card>
        <Card title="Recent Threats">
          <div className="qtn-stat">{stats.recentThreats}</div>
        </Card>
      </div>
    </div>
  );
};
