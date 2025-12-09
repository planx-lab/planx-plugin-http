// Package plugin implements the HTTP sink plugin logic.
package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/planx-lab/planx-common/logger"
	planxv1 "github.com/planx-lab/planx-proto/gen/go/planx/v1"
	"github.com/planx-lab/planx-sdk-go/batch"
	"github.com/planx-lab/planx-sdk-go/session"
)

// Config holds the HTTP sink configuration.
type Config struct {
	Endpoint    string            `json:"endpoint"`
	Method      string            `json:"method"` // POST, PUT, PATCH
	Headers     map[string]string `json:"headers"`
	Timeout     string            `json:"timeout"`      // e.g., "30s"
	BatchFormat string            `json:"batch_format"` // json_array, ndjson
}

// HTTPSink implements the SinkPlugin service.
type HTTPSink struct {
	planxv1.UnimplementedSinkPluginServer
	sessions *session.Manager
}

// NewHTTPSink creates a new HTTPSink.
func NewHTTPSink() *HTTPSink {
	return &HTTPSink{
		sessions: session.NewManager(),
	}
}

// CreateSession initializes a new session.
func (s *HTTPSink) CreateSession(ctx context.Context, req *planxv1.SessionCreateRequest) (*planxv1.SessionCreateResponse, error) {
	// Validate config
	var cfg Config
	if err := json.Unmarshal(req.ConfigJson, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	sess := s.sessions.Create(req.TenantId, req.ConfigJson)

	// Create HTTP client for this session
	timeout := 30 * time.Second
	if cfg.Timeout != "" {
		if d, err := time.ParseDuration(cfg.Timeout); err == nil {
			timeout = d
		}
	}

	client := &http.Client{Timeout: timeout}
	sess.SetData("http_client", client)
	sess.SetData("config", cfg)

	logger.Info().
		Str("session_id", sess.ID).
		Str("tenant_id", req.TenantId).
		Str("endpoint", cfg.Endpoint).
		Msg("HTTP sink session created")

	return &planxv1.SessionCreateResponse{
		SessionId: sess.ID,
	}, nil
}

// Write receives batches and writes them to the HTTP endpoint.
func (s *HTTPSink) Write(stream planxv1.SinkPlugin_WriteServer) error {
	var currentSession *session.Session
	var cfg Config
	var client *http.Client

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Get session on first request
		if currentSession == nil {
			currentSession, err = s.sessions.Get(req.SessionId)
			if err != nil {
				return err
			}

			clientVal, _ := currentSession.GetData("http_client")
			client = clientVal.(*http.Client)

			cfgVal, _ := currentSession.GetData("config")
			cfg = cfgVal.(Config)
		}

		// Unpack batch
		b, err := batch.UnpackBatch(req.PackedBatch)
		if err != nil {
			if sendErr := stream.Send(&planxv1.AckResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to unpack batch: %v", err),
			}); sendErr != nil {
				return sendErr
			}
			continue
		}

		// Send to HTTP endpoint
		if err := s.sendBatch(stream.Context(), client, cfg, b); err != nil {
			logger.Error().Err(err).Str("session_id", req.SessionId).Msg("Failed to send batch")
			if sendErr := stream.Send(&planxv1.AckResponse{
				Success: false,
				Error:   err.Error(),
			}); sendErr != nil {
				return sendErr
			}
			continue
		}

		logger.Debug().
			Str("session_id", req.SessionId).
			Int("records", len(b.Records)).
			Msg("Batch sent to HTTP endpoint")

		if err := stream.Send(&planxv1.AckResponse{Success: true}); err != nil {
			return err
		}
	}
}

func (s *HTTPSink) sendBatch(ctx context.Context, client *http.Client, cfg Config, b batch.Batch) error {
	method := cfg.Method
	if method == "" {
		method = http.MethodPost
	}

	// Format batch based on config
	var body []byte
	var err error

	switch cfg.BatchFormat {
	case "ndjson":
		// Newline-delimited JSON
		var buf bytes.Buffer
		for _, r := range b.Records {
			buf.Write(r.Payload)
			buf.WriteByte('\n')
		}
		body = buf.Bytes()
	default:
		// JSON array (default)
		payloads := make([]json.RawMessage, len(b.Records))
		for i, r := range b.Records {
			payloads[i] = r.Payload
		}
		body, err = json.Marshal(payloads)
		if err != nil {
			return fmt.Errorf("failed to marshal batch: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// CloseSession terminates a session.
func (s *HTTPSink) CloseSession(ctx context.Context, req *planxv1.SessionCloseRequest) (*planxv1.Empty, error) {
	if err := s.sessions.Close(req.SessionId); err != nil {
		logger.Warn().Err(err).Str("session_id", req.SessionId).Msg("Failed to close session")
	} else {
		logger.Info().Str("session_id", req.SessionId).Msg("HTTP sink session closed")
	}
	return &planxv1.Empty{}, nil
}
