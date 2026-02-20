import React, { useEffect, useState, useCallback } from "react";
import { Button, Card, DataTable } from "../components/ui.js";
import type { Column } from "../components/ui.js";

interface Gateway {
  id: string;
  name: string;
  hostname: string;
  port: number;
  status: string;
  tls_policy: string;
  max_connections: number;
  created_at: string;
  [key: string]: unknown;
}

export const GatewaysPage: React.FC = () => {
  const [gateways, setGateways] = useState<Gateway[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchGateways = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/v1/gateways?limit=50");
      const data = await res.json();
      setGateways(data.gateways ?? []);
    } catch {
      setGateways([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchGateways();
  }, [fetchGateways]);

  const columns: Column<Gateway>[] = [
    { key: "name", header: "Name" },
    { key: "hostname", header: "Hostname" },
    { key: "port", header: "Port" },
    { key: "status", header: "Status" },
    { key: "tls_policy", header: "TLS Policy" },
    { key: "max_connections", header: "Max Conn" },
  ];

  return (
    <div className="qtn-gateways-page">
      <Card
        title="Gateway Instances"
        subtitle="Manage quantum-safe gateway instances"
        actions={
          <Button variant="primary" size="sm" onClick={fetchGateways}>
            Refresh
          </Button>
        }
      >
        <DataTable<Gateway>
          columns={columns}
          data={gateways}
          keyField="id"
          loading={loading}
          emptyMessage="No gateways configured. Create one using the API."
        />
      </Card>
    </div>
  );
};
