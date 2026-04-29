package model

type RateLimitConfig struct {
	APIKey     string  `json:"apiKey"`
	Capacity   int     `json:"capacity"`   // max tokens
	RefillRate float64 `json:"refillRate"` // tokens per second
}
