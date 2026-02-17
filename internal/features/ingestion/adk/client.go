package adk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	boarddomain "github.com/shanehughes1990/agentic-worktrees/internal/features/board/domain"
	"github.com/shanehughes1990/agentic-worktrees/internal/features/ingestion/pipeline"
	sharederrors "github.com/shanehughes1990/agentic-worktrees/internal/shared/errors"
)

type Client struct {
	url        string
	authToken  string
	httpClient *http.Client
}

type planRequest struct {
	SchemaVersion int                  `json:"schema_version"`
	SourceScope   string               `json:"source_scope"`
	Files         []pipeline.ScopeFile `json:"files"`
}

func NewClient(url string, authToken string, timeout time.Duration) (*Client, error) {
	trimmedURL := strings.TrimSpace(url)
	if trimmedURL == "" {
		return nil, fmt.Errorf("adk url cannot be empty")
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return &Client{
		url:       trimmedURL,
		authToken: strings.TrimSpace(authToken),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) PlanBoard(ctx context.Context, sourceScope string, files []pipeline.ScopeFile) (boarddomain.Board, error) {
	requestBody, err := json.Marshal(planRequest{
		SchemaVersion: 1,
		SourceScope:   sourceScope,
		Files:         files,
	})
	if err != nil {
		return boarddomain.Board{}, sharederrors.Terminal("marshal_plan_request", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(requestBody))
	if err != nil {
		return boarddomain.Board{}, sharederrors.Terminal("build_http_request", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if c.authToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return boarddomain.Board{}, sharederrors.Transient("execute_http_request", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(httpResp.Body, 10*1024*1024))
	if err != nil {
		return boarddomain.Board{}, sharederrors.Transient("read_http_response", err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		err = fmt.Errorf("adk endpoint returned status %d", httpResp.StatusCode)
		if httpResp.StatusCode >= 500 {
			return boarddomain.Board{}, sharederrors.Transient("adk_status", err)
		}
		return boarddomain.Board{}, sharederrors.Terminal("adk_status", err)
	}

	var board boarddomain.Board
	if err := json.Unmarshal(body, &board); err != nil {
		return boarddomain.Board{}, sharederrors.Terminal("unmarshal_adk_response", err)
	}

	if board.GeneratedAt.IsZero() {
		board.GeneratedAt = time.Now().UTC()
	}
	if board.SchemaVersion == 0 {
		board.SchemaVersion = 1
	}
	if strings.TrimSpace(board.SourceScope) == "" {
		board.SourceScope = sourceScope
	}

	if err := board.Validate(); err != nil {
		return boarddomain.Board{}, sharederrors.Terminal("validate_board", err)
	}
	return board, nil
}
