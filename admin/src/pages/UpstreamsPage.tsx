import React, { useEffect, useState, useCallback } from "react";
import { Button, Card, DataTable } from "../components/ui.js";
import type { Column } from "../components/ui.js";

interface Upstream {
  id: string;
  name: string;
  host: string;
  port: number;
  protocol: string;
  is_healthy: boolean;
  health_check_path: string;
  [key: string]: unknown;
}

export const UpstreamsPage: React.FC = () => {
  const [upstreams, setUpstreams] = useState<Upstream[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchUpstreams = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/v1/upstreams?limit=50");
      const data = await res.json();
      setUpstreams(data.upstreams ?? []);
    } catch {
      setUpstreams([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchUpstreams();
  }, [fetchUpstreams]);

  const columns: Column<Upstream>[] = [
    { key: "name", header: "Name" },
    { key: "host", header: "Host" },
    { key: "port", header: "Port" },
    { key: "protocol", header: "Protocol" },
    { key: "is_healthy", header: "Healthy" },
    { key: "health_check_path", header: "Health Path" },
  ];

  return (
    <div className="qtn-upstreams-page">
      <Card
        title="Upstream Services"
        subtitle="Backend services proxied through the gateway"
        actions={
          <Button variant="primary" size="sm" onClick={fetchUpstreams}>
            Refresh
          </Button>
        }
      >
        <DataTable<Upstream>
          columns={columns}
          data={upstreams}
          keyField="id"
          loading={loading}
          emptyMessage="No upstreams configured."
        />
      </Card>
    </div>
  );
};
