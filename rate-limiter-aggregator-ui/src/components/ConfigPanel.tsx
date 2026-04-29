import { useEffect, useState } from "react";

import { RateLimitConfig } from "../services/api";

type ConfigPanelProps = {
  apiKey: string;
  config: RateLimitConfig | null;
  saving: boolean;
  onSave: (config: RateLimitConfig) => Promise<void>;
};

export function ConfigPanel({ apiKey, config, saving, onSave }: ConfigPanelProps) {
  const [draft, setDraft] = useState<RateLimitConfig>({
    apiKey,
    capacity: 10,
    refillRate: 2,
  });

  useEffect(() => {
    if (config) {
      setDraft(config);
      return;
    }

    setDraft({
      apiKey,
      capacity: 10,
      refillRate: 2,
    });
  }, [apiKey, config]);

  return (
    <section className="panel">
      <div className="panel-heading">
        <p className="eyebrow">Config Panel</p>
        <h2>Shape the token bucket</h2>
      </div>

      <p className="panel-copy">
        Tune the current key and save changes to update its token bucket behavior.
      </p>

      <label className="field">
        <span>Capacity: {draft.capacity}</span>
        <input
          type="range"
          min="1"
          max="200"
          value={draft.capacity}
          onChange={(event) =>
            setDraft((current) => ({
              ...current,
              apiKey,
              capacity: Number(event.target.value),
            }))
          }
        />
      </label>

      <label className="field">
        <span>Refill rate: {draft.refillRate.toFixed(1)} tokens/sec</span>
        <input
          type="range"
          min="0.5"
          max="50"
          step="0.5"
          value={draft.refillRate}
          onChange={(event) =>
            setDraft((current) => ({
              ...current,
              apiKey,
              refillRate: Number(event.target.value),
            }))
          }
        />
      </label>

      <button
        className="primary-button"
        disabled={!apiKey.trim() || saving}
        onClick={() => onSave({ ...draft, apiKey })}
      >
        {saving ? "Saving..." : "Save config"}
      </button>
    </section>
  );
}
