import { useEffect, useMemo, useState } from "react";

import { AggregatedMetric, getMetrics } from "../services/api";

const DEFAULT_WINDOW_SECONDS = 5 * 60;

type UseMetricsResult = {
  metrics: AggregatedMetric[];
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
};

// Custom hook to fetch and manage metrics for a given API key and time window
export function useMetrics(apiKey: string, windowSeconds = DEFAULT_WINDOW_SECONDS): UseMetricsResult {
  const [metrics, setMetrics] = useState<AggregatedMetric[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState(0);

  const refresh = async () => {
    setRefreshToken((value) => value + 1);
  };

  const range = useMemo(() => {
    const to = Math.floor(Date.now() / 1000);
    return {
      from: to - windowSeconds,
      to,
    };
  }, [windowSeconds, refreshToken]);

  useEffect(() => {
    if (!apiKey.trim()) {
      setMetrics([]);
      setError(null);
      return;
    }

    let isActive = true;

    const loadMetrics = async () => {
      try {
        setLoading(true);
        const response = await getMetrics(apiKey, range.from, range.to);
        if (isActive) {
          setMetrics(response);
          setError(null);
        }
      } catch (loadError) {
        if (isActive) {
          setError(loadError instanceof Error ? loadError.message : "Unable to load metrics");
        }
      } finally {
        if (isActive) {
          setLoading(false);
        }
      }
    };

    void loadMetrics();
    const intervalId = window.setInterval(() => {
      void loadMetrics();
    }, 2000);

    return () => {
      isActive = false;
      window.clearInterval(intervalId);
    };
  }, [apiKey, range.from, range.to]);

  return { metrics, loading, error, refresh };
}
