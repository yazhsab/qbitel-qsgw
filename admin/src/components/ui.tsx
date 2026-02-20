/**
 * Lightweight UI components for the QSGW Admin panel.
 *
 * These replace the shared @quantun/ui package used in the monorepo,
 * providing the same public API (Card, Button, DataTable, Column) so
 * that page-level code requires zero changes.
 */

import React from "react";

/* ------------------------------------------------------------------ */
/*  Card                                                               */
/* ------------------------------------------------------------------ */

interface CardProps {
  title: string;
  subtitle?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}

export const Card: React.FC<CardProps> = ({ title, subtitle, actions, children }) => (
  <div className="qtn-card" style={{ border: "1px solid #e2e8f0", borderRadius: 8, padding: "1rem", marginBottom: "1rem" }}>
    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "0.75rem" }}>
      <div>
        <h3 style={{ margin: 0 }}>{title}</h3>
        {subtitle && <p style={{ margin: 0, color: "#64748b", fontSize: "0.875rem" }}>{subtitle}</p>}
      </div>
      {actions && <div>{actions}</div>}
    </div>
    {children}
  </div>
);

/* ------------------------------------------------------------------ */
/*  Button                                                             */
/* ------------------------------------------------------------------ */

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "danger";
  size?: "sm" | "md" | "lg";
}

export const Button: React.FC<ButtonProps> = ({ variant = "primary", size = "md", children, style, ...rest }) => {
  const bgColor = variant === "primary" ? "#3b82f6" : variant === "danger" ? "#ef4444" : "#e2e8f0";
  const textColor = variant === "secondary" ? "#1e293b" : "#fff";
  const padding = size === "sm" ? "0.25rem 0.5rem" : size === "lg" ? "0.75rem 1.5rem" : "0.5rem 1rem";
  return (
    <button
      style={{ backgroundColor: bgColor, color: textColor, padding, borderRadius: 4, border: "none", cursor: "pointer", fontSize: "0.875rem", ...style }}
      {...rest}
    >
      {children}
    </button>
  );
};

/* ------------------------------------------------------------------ */
/*  DataTable                                                          */
/* ------------------------------------------------------------------ */

export interface Column<T> {
  key: keyof T & string;
  header: string;
  render?: (value: T[keyof T], row: T) => React.ReactNode;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyField: keyof T & string;
  loading?: boolean;
  emptyMessage?: string;
}

export function DataTable<T extends Record<string, unknown>>({
  columns,
  data,
  keyField,
  loading = false,
  emptyMessage = "No data.",
}: DataTableProps<T>) {
  if (loading) {
    return <div style={{ padding: "1rem", textAlign: "center", color: "#64748b" }}>Loading...</div>;
  }

  if (data.length === 0) {
    return <div style={{ padding: "1rem", textAlign: "center", color: "#64748b" }}>{emptyMessage}</div>;
  }

  return (
    <table style={{ width: "100%", borderCollapse: "collapse" }}>
      <thead>
        <tr>
          {columns.map((col) => (
            <th key={col.key} style={{ textAlign: "left", padding: "0.5rem", borderBottom: "2px solid #e2e8f0", fontSize: "0.75rem", textTransform: "uppercase", color: "#64748b" }}>
              {col.header}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {data.map((row) => (
          <tr key={String(row[keyField])}>
            {columns.map((col) => (
              <td key={col.key} style={{ padding: "0.5rem", borderBottom: "1px solid #f1f5f9" }}>
                {col.render ? col.render(row[col.key], row) : String(row[col.key] ?? "")}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
