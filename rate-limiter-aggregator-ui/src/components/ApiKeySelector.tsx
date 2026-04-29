import { useEffect, useRef, useState } from "react";

type ApiKeySelectorProps = {
  apiKeys: string[];
  value: string;
  onChange: (apiKey: string) => void;
  onCreateNew: () => void;
  onEditSelected: () => void;
  canEditSelected: boolean;
};

export function ApiKeySelector({
  apiKeys,
  value,
  onChange,
  onCreateNew,
  onEditSelected,
  canEditSelected,
}: ApiKeySelectorProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const handlePointerDown = (event: MouseEvent) => {
      if (!containerRef.current?.contains(event.target as Node)) {
        setMenuOpen(false);
      }
    };

    document.addEventListener("mousedown", handlePointerDown);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
    };
  }, []);

  useEffect(() => {
    setMenuOpen(false);
  }, [value]);

  const selectedLabel = apiKeys.includes(value) ? value : "Select a configured API key";

  return (
    <section className="panel api-key-panel">
      <div className="panel-heading">
        <p className="eyebrow">API Key Selector</p>
        <h2>Pick a live key or create a new one</h2>
      </div>

      <p className="panel-copy">
        Use the dropdown to switch the active key. When you need a new key, open the creation modal
        and configure it in one step.
      </p>

      <div className="selector-actions">
        <label className="field">
          <span>Active key</span>
          <div className="custom-select" ref={containerRef}>
            <button
              className={`api-key-select ${menuOpen ? "api-key-select-open" : ""}`}
              type="button"
              aria-haspopup="listbox"
              aria-expanded={menuOpen}
              onClick={() => setMenuOpen((current) => !current)}
            >
              <span className={apiKeys.includes(value) ? "api-key-select-value" : "api-key-select-placeholder"}>
                {selectedLabel}
              </span>
              <span className="api-key-select-chevron" aria-hidden="true">
                <svg viewBox="0 0 24 24" focusable="false">
                  <path d="M6 9l6 6 6-6" />
                </svg>
              </span>
            </button>

            {menuOpen ? (
              <div className="api-key-select-menu" role="listbox" aria-label="API keys">
                {apiKeys.length === 0 ? (
                  <div className="api-key-option api-key-option-empty">No configured API keys yet</div>
                ) : (
                  apiKeys.map((apiKey) => (
                    <button
                      key={apiKey}
                      type="button"
                      role="option"
                      aria-selected={apiKey === value}
                      className={`api-key-option ${apiKey === value ? "api-key-option-active" : ""}`}
                      onPointerDown={(event) => {
                        event.preventDefault();
                        onChange(apiKey);
                        setMenuOpen(false);
                      }}
                    >
                      {apiKey}
                    </button>
                  ))
                )}
              </div>
            ) : null}
          </div>
        </label>

        <button
          className="secondary-button create-key-button"
          type="button"
          onClick={onEditSelected}
          disabled={!canEditSelected}
          title="Edit the configuration for the selected API key"
        >
          <span className="button-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" focusable="false">
              <path d="M4 20l4.5-1 9-9-3.5-3.5-9 9L4 20z" />
              <path d="M13.5 6.5l3.5 3.5" />
            </svg>
          </span>
          Edit selected API key
        </button>

        <button
          className="secondary-button create-key-button"
          type="button"
          onClick={onCreateNew}
          title="Create a new API key and configure its rate limits"
        >
          <span className="button-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" focusable="false">
              <path d="M12 5v14" />
              <path d="M5 12h14" />
            </svg>
          </span>
          New API Key
        </button>
      </div>
    </section>
  );
}
