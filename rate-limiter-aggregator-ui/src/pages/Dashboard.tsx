import { useEffect, useMemo, useState } from "react";

import { ApiKeySelector } from "../components/ApiKeySelector";
import { CreateApiKeyModal } from "../components/CreateApiKeyModal";
import { MetricsChart } from "../components/MetricsChart";
import { SimulationPanel } from "../components/SimulationPanel";
import { useMetrics } from "../hooks/useMetrics";
import {
  RateLimitConfig,
  getConfig,
  getConfigs,
  simulateTrafficPattern,
  TrafficPattern,
  updateConfig,
} from "../services/api";

export function Dashboard() {
  const [apiKeys, setApiKeys] = useState<string[]>([]);
  const [selectedApiKey, setSelectedApiKey] = useState("test123");
  const [config, setConfig] = useState<RateLimitConfig | null>(null);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [configLoading, setConfigLoading] = useState(false);
  const [configSaving, setConfigSaving] = useState(false);
  const [simulationRunning, setSimulationRunning] = useState(false);
  const [simulationResult, setSimulationResult] = useState<{ allowed: number; rejected: number } | null>(null);
  const [selectedPattern, setSelectedPattern] = useState<TrafficPattern>("burst");
  const [resultPattern, setResultPattern] = useState<TrafficPattern | null>(null);
  const [uiError, setUiError] = useState<string | null>(null);

  const { metrics, loading: metricsLoading, error: metricsError, refresh } = useMetrics(selectedApiKey);

  // Load API keys on mount and set initial selection
  useEffect(() => {
    const loadConfigs = async () => {
      try {
        const configs = await getConfigs();
        setApiKeys(configs.map((item) => item.apiKey));
        if (!selectedApiKey && configs[0]) {
          setSelectedApiKey(configs[0].apiKey);
        }
      } catch (error) {
        setUiError(error instanceof Error ? error.message : "Unable to load API keys");
      }
    };

    void loadConfigs();
  }, []);

  // Load config whenever selected API key changes
  useEffect(() => {
    if (!selectedApiKey.trim()) {
      setConfig(null);
      return;
    }

    const loadConfig = async () => {
      try {
        setConfigLoading(true);
        const nextConfig = await getConfig(selectedApiKey);
        setConfig(nextConfig);
        setUiError(null);
      } catch (error) {
        setConfig(null);
        setUiError(error instanceof Error ? error.message : "Unable to load config");
      } finally {
        setConfigLoading(false);
      }
    };

    void loadConfig();
  }, [selectedApiKey]);

  const simulationContext = useMemo(() => {
    const latest = metrics[metrics.length - 1];
    const totals = metrics.reduce(
      (accumulator, metric) => {
        accumulator.total += metric.total;
        accumulator.rejected += metric.rejected;
        return accumulator;
      },
      { total: 0, rejected: 0 },
    );

    const rejectionRate = totals.total === 0 ? 0 : (totals.rejected / totals.total) * 100;

    return {
      requestsPerSecond: latest?.total ?? 0,
      rejectionPercentage: `${rejectionRate.toFixed(1)}%`,
      rejectionSummary: `${totals.rejected} rejected of ${totals.total} total requests in the visible range`,
    };
  }, [metrics]);

  const handleCreateApiKey = async (draft: RateLimitConfig) => {
    try {
      setConfigSaving(true);
      const saved = await updateConfig(draft);
      setApiKeys((current) => Array.from(new Set([...current, saved.apiKey])));
      setSelectedApiKey(saved.apiKey);
      setConfig(saved);
      setCreateModalOpen(false);
      setUiError(null);
    } catch (error) {
      setUiError(error instanceof Error ? error.message : "Unable to create API key");
    } finally {
      setConfigSaving(false);
    }
  };

  const handleEditApiKey = async (draft: RateLimitConfig) => {
    try {
      setConfigSaving(true);
      const saved = await updateConfig(draft);
      setConfig(saved);
      setEditModalOpen(false);
      setUiError(null);
    } catch (error) {
      setUiError(error instanceof Error ? error.message : "Unable to update API key");
    } finally {
      setConfigSaving(false);
    }
  };

  const handleRunSimulation = async (requestCount: number, pattern: TrafficPattern) => {
    try {
      setSimulationRunning(true);
      const result = await simulateTrafficPattern(selectedApiKey, pattern, requestCount);
      setSimulationResult(result);
      setResultPattern(pattern);
      setUiError(null);
      window.setTimeout(() => {
        void refresh();
      }, 1200);
    } catch (error) {
      setUiError(error instanceof Error ? error.message : "Unable to run simulation");
    } finally {
      setSimulationRunning(false);
    }
  };

  const statusMessage = uiError ?? metricsError;
  const configStatus = configLoading ? "Loading config..." : "Config synced from backend";

  return (
    <main className="dashboard-shell">
      <section className="hero">
        <p className="eyebrow">API Usage Dashboard</p>
        <h1>Watch the token bucket bend before it breaks.</h1>
        <p className="hero-copy">
          Pick an API key, tune bucket capacity and refill rate, then hammer the limiter to surface
          request spikes and 429 pressure instantly.
        </p>
      </section>

      <ApiKeySelector
        apiKeys={apiKeys}
        value={selectedApiKey}
        onChange={setSelectedApiKey}
        onCreateNew={() => setCreateModalOpen(true)}
        onEditSelected={() => setEditModalOpen(true)}
        canEditSelected={Boolean(selectedApiKey.trim() && config)}
      />

      <MetricsChart metrics={metrics} apiKey={selectedApiKey} />

      <section className="control-grid single-panel-grid">
        <SimulationPanel
          apiKey={selectedApiKey}
          config={config}
          running={simulationRunning}
          result={simulationResult}
          selectedPattern={selectedPattern}
          resultPattern={resultPattern}
          requestsPerSecond={simulationContext.requestsPerSecond}
          rejectionPercentage={simulationContext.rejectionPercentage}
          rejectionSummary={simulationContext.rejectionSummary}
          onPatternChange={setSelectedPattern}
          onRun={handleRunSimulation}
        />
      </section>

      <section className="footer-strip">
        <span>{metricsLoading ? "Refreshing metrics..." : "Metrics auto-refresh every 2s"}</span>
        <span>{configStatus}</span>
        {statusMessage ? <span className="error-text">{statusMessage}</span> : null}
      </section>

      <CreateApiKeyModal
        open={createModalOpen}
        saving={configSaving}
        mode="create"
        onClose={() => setCreateModalOpen(false)}
        onSubmit={handleCreateApiKey}
      />

      <CreateApiKeyModal
        open={editModalOpen}
        saving={configSaving}
        mode="edit"
        initialConfig={config}
        onClose={() => setEditModalOpen(false)}
        onSubmit={handleEditApiKey}
      />
    </main>
  );
}
