import { ChangeEvent, useEffect, useState } from "react";

import { RateLimitConfig } from "../services/api";

type CreateApiKeyModalProps = {
  open: boolean;
  saving: boolean;
  mode: "create" | "edit";
  initialConfig?: RateLimitConfig | null;
  onClose: () => void;
  onSubmit: (config: RateLimitConfig) => Promise<void>;
};

export function CreateApiKeyModal({
  open,
  saving,
  mode,
  initialConfig,
  onClose,
  onSubmit,
}: CreateApiKeyModalProps) {
  const [draft, setDraft] = useState<RateLimitConfig>({
    apiKey: "",
    capacity: 10,
    refillRate: 2,
  });

  useEffect(() => {
    if (open) {
      if (mode === "edit" && initialConfig) {
        setDraft(initialConfig);
        return;
      }

      setDraft({
        apiKey: "",
        capacity: 10,
        refillRate: 2,
      });
    }
  }, [open, mode, initialConfig]);

  if (!open) {
    return null;
  }

  const handleApiKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    setDraft((current) => ({
      ...current,
      apiKey: event.target.value,
    }));
  };

  const canSubmit = draft.apiKey.trim().length > 0 && draft.capacity > 0 && draft.refillRate > 0;
  const title = mode === "create" ? "Add a new configured key" : "Edit selected API key";
  const eyebrow = mode === "create" ? "Create API Key" : "Edit API Key";
  const description =
    mode === "create"
      ? "Create the key and its initial token bucket settings together so it is ready to simulate immediately."
      : "Update the token bucket settings for the selected key. The key name stays fixed while you edit its configuration.";
  const submitLabel = mode === "create" ? "Create API key" : "Save changes";

  return (
    <div className="modal-overlay" role="presentation" onClick={onClose}>
      <section
        className="modal-card"
        role="dialog"
        aria-modal="true"
        aria-labelledby="key-config-modal-title"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="panel-heading">
          <p className="eyebrow">{eyebrow}</p>
          <h2 id="key-config-modal-title">{title}</h2>
        </div>

        <p className="panel-copy">{description}</p>

        <label className="field">
          <span>API key name</span>
          <input
            type="text"
            value={draft.apiKey}
            onChange={handleApiKeyChange}
            placeholder="partner-mobile-prod"
            spellCheck={false}
            disabled={mode === "edit"}
          />
        </label>

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
                refillRate: Number(event.target.value),
              }))
            }
          />
        </label>

        <div className="modal-actions">
          <button className="secondary-button" type="button" onClick={onClose} disabled={saving}>
            Cancel
          </button>
          <button
            className="primary-button"
            type="button"
            disabled={!canSubmit || saving}
            onClick={() => onSubmit({ ...draft, apiKey: draft.apiKey.trim() })}
          >
            {saving ? (mode === "create" ? "Creating..." : "Saving...") : submitLabel}
          </button>
        </div>
      </section>
    </div>
  );
}
