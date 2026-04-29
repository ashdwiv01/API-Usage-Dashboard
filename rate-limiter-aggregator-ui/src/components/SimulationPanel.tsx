import { ChangeEvent, useEffect, useState } from "react";

import { RateLimitConfig, TrafficPattern } from "../services/api";

const TRAFFIC_PATTERNS: Array<{
  value: TrafficPattern;
  label: string;
  description: string;
}> = [
  {
    value: "burst",
    label: "Burst",
    description: "Fire requests in dense parallel batches to pressure the bucket immediately.",
  },
  {
    value: "steady",
    label: "Steady",
    description: "Send a continuous stream of evenly spaced requests for a smoother load profile.",
  },
  {
    value: "ramp-up",
    label: "Ramp-up",
    description: "Start gently, then increase request intensity to reveal the breaking point.",
  },
];

type SimulationPanelProps = {
  apiKey: string;
  config: RateLimitConfig | null;
  running: boolean;
  result: { allowed: number; rejected: number } | null;
  selectedPattern: TrafficPattern;
  resultPattern: TrafficPattern | null;
  requestsPerSecond: number;
  rejectionPercentage: string;
  rejectionSummary: string;
  onPatternChange: (pattern: TrafficPattern) => void;
  onRun: (count: number, pattern: TrafficPattern) => Promise<void>;
};

export function SimulationPanel({
  apiKey,
  config,
  running,
  result,
  selectedPattern,
  resultPattern,
  requestsPerSecond,
  rejectionPercentage,
  rejectionSummary,
  onPatternChange,
  onRun,
}: SimulationPanelProps) {
  const [requestCount, setRequestCount] = useState("100");

  useEffect(() => {
    if (!apiKey.trim()) {
      setRequestCount("100");
    }
  }, [apiKey]);

  const handleCountChange = (event: ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value.replace(/[^\d]/g, "");
    setRequestCount(value);
  };

  const parsedRequestCount = Number(requestCount);
  const canRun =
    apiKey.trim().length > 0 &&
    !running &&
    requestCount.trim().length > 0 &&
    Number.isInteger(parsedRequestCount) &&
    parsedRequestCount > 0;

  return (
    <section className="panel">
      <div className="panel-heading">
        <p className="eyebrow">Simulation Panel</p>
        <h2>Force a burst and watch the chart react</h2>
      </div>

      <p className="panel-copy">
        Trigger a custom burst against <strong>{apiKey || "your active key"}</strong> and watch
        the requests/sec line and 429 overlay spike in real time.
      </p>

      <div className="simulation-details">
        <article className="detail-card">
          <span className="detail-label">API key</span>
          <strong className="detail-value">{apiKey || "No key selected"}</strong>
        </article>

        <article className="detail-card">
          <span className="detail-label">Bucket capacity</span>
          <strong className="detail-value">{config?.capacity ?? "--"} tokens</strong>
        </article>

        <article className="detail-card">
          <span className="detail-label">Refill rate</span>
          <strong className="detail-value">
            {config ? `${config.refillRate.toFixed(1)} tokens/sec` : "--"}
          </strong>
        </article>
      </div>

      <div className="simulation-actions">
        <label className="field simulation-field">
          <span>Requests to send</span>
          <input
            type="text"
            inputMode="numeric"
            value={requestCount}
            onChange={handleCountChange}
            placeholder="100"
          />
        </label>

        <div className="field simulation-field">
          <span>Traffic pattern</span>
          <div className="traffic-pattern-grid" role="radiogroup" aria-label="Traffic pattern">
            {TRAFFIC_PATTERNS.map((pattern) => {
              const selected = pattern.value === selectedPattern;

              return (
                <button
                  key={pattern.value}
                  type="button"
                  role="radio"
                  aria-checked={selected}
                  className={`traffic-pattern-card${selected ? " traffic-pattern-card-selected" : ""}`}
                  onClick={() => onPatternChange(pattern.value)}
                >
                  <strong>{pattern.label}</strong>
                  <span>{pattern.description}</span>
                </button>
              );
            })}
          </div>
        </div>

        <button
          className="primary-button simulation-run-button"
          disabled={!canRun}
          onClick={() => onRun(parsedRequestCount, selectedPattern)}
        >
          {running ? "Running simulation..." : `Run ${requestCount || "..."} request simulation`}
        </button>
      </div>

      <div className="simulation-metrics">
        <article className="detail-card">
          <span className="detail-label">Requests/sec</span>
          <strong className="detail-value">{requestsPerSecond}</strong>
          <span className="detail-help">Latest per-second bucket for the selected key.</span>
        </article>

        <article className="detail-card">
          <span className="detail-label">Rejection percentage</span>
          <strong className="detail-value">{rejectionPercentage}</strong>
          <span className="detail-help">{rejectionSummary}</span>
        </article>
      </div>

      {result ? (
        <div className="simulation-result">
          <span>{TRAFFIC_PATTERNS.find((pattern) => pattern.value === resultPattern)?.label ?? "Simulation"} result:</span>
          <span>{result.allowed} allowed</span>
          <span>{result.rejected} rejected</span>
        </div>
      ) : null}
    </section>
  );
}
