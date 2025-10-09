package cmd

import (
  "bytes"
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "io"
  "net/http"
  "os"
  "strings"
  "time"
)

const aivenAPIBase = "https://api.aiven.io/v1"

type aivenClient struct {
  token string
  hc    *http.Client
}

func newAivenClientFromEnv() (*aivenClient, error) {
  tok := strings.TrimSpace(os.Getenv("AIVEN_TOKEN"))
  if tok == "" { return nil, errors.New("AIVEN_TOKEN is not set") }
  return &aivenClient{ token: tok, hc: &http.Client{ Timeout: 60 * time.Second } }, nil
}

func (c *aivenClient) do(ctx context.Context, method, path string, in any, out any) error {
  var body io.Reader
  if in != nil {
    b, err := json.Marshal(in)
    if err != nil { return err }
    body = bytes.NewReader(b)
  }
  req, err := http.NewRequestWithContext(ctx, method, aivenAPIBase+path, body)
  if err != nil { return err }
  req.Header.Set("Authorization", "aivenv1 "+c.token)
  req.Header.Set("Content-Type", "application/json")
  resp, err := c.hc.Do(req)
  if err != nil { return err }
  defer resp.Body.Close()
  if resp.StatusCode >= 300 {
    b, _ := io.ReadAll(resp.Body)
    return fmt.Errorf("aiven api %s %s: %s: %s", method, path, resp.Status, string(b))
  }
  if out != nil {
    dec := json.NewDecoder(resp.Body)
    if err := dec.Decode(out); err != nil { return err }
  }
  return nil
}

type aivenService struct {
  Name        string `json:"service_name"`
  ServiceURI  string `json:"service_uri"`
  ServiceType string `json:"service_type"`
  State       string `json:"state"`
}

type aivenGetService struct { Service aivenService `json:"service"` }

type aivenCreateServiceReq struct {
  Cloud       string `json:"cloud,omitempty"`
  Plan        string `json:"plan"`
  ServiceName string `json:"service_name"`
  ServiceType string `json:"service_type"`
}

type aivenCreateServiceResp struct { Service aivenService `json:"service"` }

func (c *aivenClient) getService(ctx context.Context, project, name string) (*aivenService, error) {
  var out aivenGetService
  if err := c.do(ctx, http.MethodGet, "/project/"+project+"/service/"+name, nil, &out); err != nil { return nil, err }
  return &out.Service, nil
}

func (c *aivenClient) createService(ctx context.Context, project string, req aivenCreateServiceReq) (*aivenService, error) {
  var out aivenCreateServiceResp
  if err := c.do(ctx, http.MethodPost, "/project/"+project+"/service", req, &out); err != nil { return nil, err }
  return &out.Service, nil
}

func (c *aivenClient) waitServiceRunning(ctx context.Context, project, name string, timeout time.Duration) error {
  deadline := time.Now().Add(timeout)
  for {
    if time.Now().After(deadline) { return errors.New("aiven service wait timeout") }
    svc, err := c.getService(ctx, project, name)
    if err != nil { return err }
    st := strings.ToUpper(strings.TrimSpace(svc.State))
    if st == "RUNNING" { return nil }
    if st == "REBUILDING" || st == "RESTARTING" || st == "POWERED_OFF" {
      // continue waiting
    }
    time.Sleep(5 * time.Second)
  }
}
