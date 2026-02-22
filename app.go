package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type GraphPayload struct {
	Nodes interface{} `json:"nodes"`
	Edges interface{} `json:"edges"`
}

func (a *App) Export(graphJSON string) (string, error) {
	var payload GraphPayload
	if err := json.Unmarshal([]byte(graphJSON), &payload); err != nil {
		return "", fmt.Errorf("invalid graph json: %w", err)
	}

	bin, err := convertGraphToBinary(payload)
	if err != nil {
		return "", fmt.Errorf("convert to binary failed: %w", err)
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export graph",
		DefaultFilename: "graph.bin",
		Filters: []runtime.FileFilter{
			{DisplayName: "Binary", Pattern: "*.bin"},
			{DisplayName: "All files", Pattern: "*"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		// user cancelled
		return "", nil
	}

	if err := os.WriteFile(path, bin, 0o644); err != nil {
		return "", fmt.Errorf("write failed: %w", err)
	}

	return path, nil
}

func convertGraphToBinary(payload GraphPayload) ([]byte, error) {
	// TODO: protobuf encoding
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return b, nil
}
