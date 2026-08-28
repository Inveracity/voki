package remote

import (
	"fmt"
	"net"
	"os"

	"github.com/inveracity/voki/internal/proto"
	"github.com/inveracity/voki/internal/targets"
)

// Conn is one dispatch session with the local voki controller daemon.
type Conn struct {
	conn net.Conn
}

// Dial connects to the controller daemon's unix socket.
func Dial(socketPath string) (*Conn, error) {
	if _, err := os.Stat(socketPath); err != nil {
		return nil, fmt.Errorf("controller socket %s unreachable, is 'voki server' running?: %w", socketPath, err)
	}
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}
	return &Conn{conn: conn}, nil
}

func (c *Conn) do(agent, op string, reqBody, respBody any) error {
	body, err := proto.Marshal(reqBody)
	if err != nil {
		return err
	}
	req := &proto.Request{Agent: agent, Op: op, Body: body}
	if err := proto.Encode(c.conn, req); err != nil {
		return err
	}
	var resp proto.Response
	if err := proto.Decode(c.conn, &resp); err != nil {
		return err
	}
	if resp.Body != nil && respBody != nil {
		if err := proto.Unmarshal(resp.Body, respBody); err != nil {
			return err
		}
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

// AgentExecutor runs steps on one named agent through the controller daemon.
type AgentExecutor struct {
	conn   *Conn
	agent  string
}

func (c *Conn) Agent(name string) *AgentExecutor {
	return &AgentExecutor{conn: c, agent: name}
}

// Exec runs a shell command on the agent and returns its output.
func (e *AgentExecutor) Exec(command string) (string, string, error) {
	var out proto.ExecResponse
	err := e.conn.do(e.agent, proto.OpExec, proto.ExecRequest{Command: command}, &out)
	return out.Stdout, out.Stderr, err
}

// Upload reads source locally and sends it to the agent via the controller.
func (e *AgentExecutor) Upload(step targets.Step) error {
	if step.Source == "" {
		return fmt.Errorf("upload step needs a source")
	}
	data, err := os.ReadFile(step.Source)
	if err != nil {
		return err
	}
	return e.conn.do(e.agent, proto.OpUpload, proto.UploadRequest{
		Destination: step.Destination,
		Mode:        step.Mode,
		Chown:       step.Chown,
		Sudo:        step.Sudo,
		Data:        data,
	}, nil)
}

func (e *AgentExecutor) Close() error { return nil }

func (c *Conn) Close() error { return c.conn.Close() }
