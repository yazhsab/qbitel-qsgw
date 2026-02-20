import React, { useEffect, useState, useCallback } from "react";
import { Button, Card, DataTable } from "../components/ui.js";
import type { Column } from "../components/ui.js";

interface ThreatEvent {
  id: string;
  gateway_id: string;
  threat_type: string;
  severity: string;
  source_ip: string | null;
  description: string;
  mitigated: boolean;
  detected_at: string;
  [key: string]: unknown;
}

export const ThreatsPage: React.FC = () => {
  const [threats, setThreats] = useState<ThreatEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [gatewayId, setGatewayId] = useState("");

  const fetchThreats = useCallback(async () => {
    if (!gatewayId) {
      setThreats([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const res = await fetch(
        `/api/v1/threats?gateway_id=${gatewayId}&limit=50`,
      );
      const data = await res.json();
      setThreats(data.events ?? []);
    } catch {
      setThreats([]);
    } finally {
      setLoading(false);
    }
  }, [gatewayId]);

  useEffect(() => {
    fetchThreats();
  }, [fetchThreats]);

  const columns: Column<ThreatEvent>[] = [
    { key: "threat_type", header: "Type" },
    { key: "severity", header: "Severity" },
    { key: "source_ip", header: "Source IP" },
    { key: "description", header: "Description" },
    { key: "mitigated", header: "Mitigated" },
    { key: "detected_at", header: "Detected" },
  ];

  return (
    <div className="qtn-threats-page">
      <Card
        title="Threat Events"
        subtitle="Security threats detected by the AI engine"
        actions={
          <Button variant="primary" size="sm" onClick={fetchThreats}>
            Refresh
          </Button>
        }
      >
        <div style={{ marginBottom: "1rem" }}>
          <label htmlFor="gateway-id">Gateway ID: </label>
          <input
            id="gateway-id"
            type="text"
            placeholder="Enter gateway UUID"
            value={gatewayId}
            onChange={(e) => setGatewayId(e.target.value)}
            style={{ padding: "0.25rem 0.5rem", width: "320px" }}
          />
        </div>
        <DataTable<ThreatEvent>
          columns={columns}
          data={threats}
          keyField="id"
          loading={loading}
          emptyMessage={
            gatewayId
              ? "No threat events for this gateway."
              : "Enter a gateway ID to view threat events."
          }
        />
      </Card>
    </div>
  );
};
