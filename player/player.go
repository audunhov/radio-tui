package player

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type AudioService struct {
	cmd        *exec.Cmd
	cancel     context.CancelFunc
	socketPath string
	Updates    chan string
	isPaused   bool
}

func NewAudioService() *AudioService {
	tempDir := os.TempDir()
	socketPath := filepath.Join(tempDir, fmt.Sprintf("radio-tui-%d.sock", os.Getpid()))
	return &AudioService{
		socketPath: socketPath,
		Updates:    make(chan string, 100),
	}
}

func (p *AudioService) Play(streamURL string) error {
	p.Stop()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.isPaused = false

	p.cmd = exec.CommandContext(ctx, "mpv",
		"--no-video",
		"--input-ipc-server="+p.socketPath,
		"--idle=yes",
		streamURL,
	)

	err := p.cmd.Start()
	if err != nil {
		cancel()
		return err
	}

	go p.listenForMetadata(ctx)

	return nil
}

func (p *AudioService) TogglePause() error {
	conn, err := net.Dial("unix", p.socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	p.isPaused = !p.isPaused
	val := "false"
	if p.isPaused {
		val = "true"
	}

	fmt.Fprintf(conn, `{"command": ["set_property", "pause", %s]}`+"\n", val)
	return nil
}

func (p *AudioService) IsPaused() bool {
	return p.isPaused
}

func (p *AudioService) Stop() {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		_ = p.cmd.Wait()
		p.cmd = nil
	}
	_ = os.Remove(p.socketPath)
	p.isPaused = false
}

func (p *AudioService) listenForMetadata(ctx context.Context) {
	var conn net.Conn
	var err error
	for range 20 {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err = net.Dial("unix", p.socketPath)
			if err == nil {
				goto connected
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
	return

connected:
	defer conn.Close()

	fmt.Fprintf(conn, `{"command": ["observe_property", 1, "media-title"]}`+"\n")

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var resp struct {
			Event string `json:"event"`
			Name  string `json:"name"`
			Data  any    `json:"data"`
		}

		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			continue
		}

		if resp.Event == "property-change" && resp.Name == "media-title" {
			if title, ok := resp.Data.(string); ok && title != "" {
				select {
				case p.Updates <- title:
				case <-ctx.Done():
					return
				default:
				}
			}
		}
	}
}
