export type AggregatedMetric = {
  apiKey: string;
  timestamp: number;
  total: number;
  rejected: number;
};

export type RateLimitConfig = {
  apiKey: string;
  capacity: number;
  refillRate: number;
};

export type TrafficPattern = "burst" | "steady" | "ramp-up";

export type SimulationResult = {
  allowed: number;
  rejected: number;
};

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    ...init,
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `Request failed with status ${response.status}`);
  }

  return (await response.json()) as T;
}

export async function getConfigs(): Promise<RateLimitConfig[]> {
  const data = await request<RateLimitConfig[] | null>("/configs");
  return Array.isArray(data) ? data : [];
}

export async function getConfig(apiKey: string): Promise<RateLimitConfig> {
  const params = new URLSearchParams({ apiKey });
  return request<RateLimitConfig>(`/config?${params.toString()}`);
}

export async function updateConfig(config: RateLimitConfig): Promise<RateLimitConfig> {
  return request<RateLimitConfig>("/config", {
    method: "PUT",
    body: JSON.stringify(config),
  });
}

export async function getMetrics(apiKey: string, from: number, to: number): Promise<AggregatedMetric[]> {
  const params = new URLSearchParams({
    apiKey,
    from: String(from),
    to: String(to),
  });

  const data = await request<AggregatedMetric[] | null>(`/metrics?${params.toString()}`);
  return Array.isArray(data) ? data : [];
}

async function sendCheckRequest(apiKey: string): Promise<boolean> {
  const response = await fetch(`${API_BASE_URL}/check`, {
    headers: {
      "x-api-key": apiKey,
    },
  });

  if (response.status === 429) {
    return false;
  }

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `Simulation failed with status ${response.status}`);
  }

  return true;
}

function wait(durationMs: number): Promise<void> {
  return new Promise((resolve) => {
    window.setTimeout(resolve, durationMs);
  });
}

async function runBatch(apiKey: string, batchSize: number): Promise<SimulationResult> {
  let allowed = 0;
  let rejected = 0;

  const results = await Promise.all(Array.from({ length: batchSize }, () => sendCheckRequest(apiKey)));

  results.forEach((isAllowed) => {
    if (isAllowed) {
      allowed += 1;
    } else {
      rejected += 1;
    }
  });

  return { allowed, rejected };
}

function mergeResults(current: SimulationResult, next: SimulationResult): SimulationResult {
  return {
    allowed: current.allowed + next.allowed,
    rejected: current.rejected + next.rejected,
  };
}

export async function simulateBurst(
  apiKey: string,
  count = 100,
  concurrency = 20,
): Promise<SimulationResult> {
  let allowed = 0;
  let rejected = 0;

  for (let index = 0; index < count; index += concurrency) {
    const batchSize = Math.min(concurrency, count - index);
    const results = await runBatch(apiKey, batchSize);

    allowed += results.allowed;
    rejected += results.rejected;
  }

  return { allowed, rejected };
}

export async function simulateSteady(apiKey: string, count = 100): Promise<SimulationResult> {
  let totals: SimulationResult = { allowed: 0, rejected: 0 };

  for (let index = 0; index < count; index += 1) {
    const result = await runBatch(apiKey, 1);
    totals = mergeResults(totals, result);

    if (index < count - 1) {
      await wait(180);
    }
  }

  return totals;
}

export async function simulateRampUp(apiKey: string, count = 100): Promise<SimulationResult> {
  const batchPlan = [2, 4, 8, 12, 16];
  let totals: SimulationResult = { allowed: 0, rejected: 0 };
  let sent = 0;

  for (let index = 0; sent < count; index += 1) {
    const targetBatchSize = batchPlan[Math.min(index, batchPlan.length - 1)];
    const batchSize = Math.min(targetBatchSize, count - sent);
    const result = await runBatch(apiKey, batchSize);

    totals = mergeResults(totals, result);
    sent += batchSize;

    if (sent < count) {
      const pauseMs = Math.max(60, 260 - index * 35);
      await wait(pauseMs);
    }
  }

  return totals;
}

export async function simulateTrafficPattern(
  apiKey: string,
  pattern: TrafficPattern,
  count = 100,
): Promise<SimulationResult> {
  switch (pattern) {
    case "steady":
      return simulateSteady(apiKey, count);
    case "ramp-up":
      return simulateRampUp(apiKey, count);
    case "burst":
    default:
      return simulateBurst(apiKey, count, 20);
  }
}
