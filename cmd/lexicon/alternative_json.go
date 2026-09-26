package main

// alternativeJSON is one ner.Alternative in a spanJSON line.
type alternativeJSON struct {
	Type     string            `json:"type"`
	Normal   []string          `json:"normal,omitempty"`
	Refs     []string          `json:"refs,omitempty"`
	Attrs    map[string]string `json:"attrs,omitempty"`
	Score    float32           `json:"score"`
	Evidence []string          `json:"evidence,omitempty"`
}
