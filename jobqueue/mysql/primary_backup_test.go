package mysql

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/model"

	"github.com/go-sql-driver/mysql"
	"github.com/phayes/freeport"
)

func TestNodeInCluster(t *testing.T) {
	q := "jobqueue_mysql_node_test"

	jq1 := NewPrimaryBackup(&model.Queue{Name: q, MaxWorkers: 30}, Dsn())
	jq1.Start()
	defer func() { <-jq1.Stop() }()

	time.Sleep(500 * time.Millisecond) // wait for up

	jq2 := NewPrimaryBackup(&model.Queue{Name: q, MaxWorkers: 30}, Dsn())
	jq2.Start()
	defer func() { <-jq2.Stop() }()

	time.Sleep(500 * time.Millisecond) // wait for up

	var nodeID string

	{
		if !jq1.IsActive() {
			t.Error("Must be active")
		}

		hasNodeInfo, ok := jq1.(jobqueue.HasNodeInfo)
		if !ok {
			t.Error("Must have Node() method")
		}
		node, err := hasNodeInfo.Node()
		if err != nil {
			t.Error(err)
		}
		if node == nil {
			t.Error("Must return an active node")
		}
		if len(node.ID) <= 0 {
			t.Error("Must return an ID")
		}
		if len(node.Host) <= 0 {
			t.Error("Must return a host name")
		}
		nodeID = node.ID
	}

	{
		if jq2.IsActive() {
			t.Error("Must be inactive")
		}

		hasNodeInfo, ok := jq2.(jobqueue.HasNodeInfo)
		if !ok {
			t.Error("Must have Node() method")
		}
		node, err := hasNodeInfo.Node()
		if err != nil {
			t.Error(err)
		}
		if node == nil {
			t.Error("Must return an active node")
		}
		if len(node.ID) <= 0 {
			t.Error("Must return an ID")
		}
		if len(node.Host) <= 0 {
			t.Error("Must return a host name")
		}
		if node.ID != nodeID {
			t.Error("Must return an active node ID")
		}
	}
}

func TestLockTimeout(t *testing.T) {
	dsn := Dsn()

	withShortLockWaitTimeout(func() {
		q := "jobqueue_mysql_lock_timeout_test"

		jq1 := NewPrimaryBackup(&model.Queue{Name: q, MaxWorkers: 30}, dsn)
		jq1.Start()

		time.Sleep(500 * time.Millisecond) // wait for up

		jq2 := NewPrimaryBackup(&model.Queue{Name: q, MaxWorkers: 30}, dsn)
		jq2.Start()

		time.Sleep(500 * time.Millisecond) // wait for up

		if !jq1.IsActive() {
			t.Error("The primary jobqueue should be active")
		}
		if jq2.IsActive() {
			t.Error("A backup jobqueue should be inactive")
		}

		time.Sleep(2 * time.Second)

		if !jq1.IsActive() {
			t.Error("The primary jobqueue should be active")
		}
		if jq2.IsActive() {
			t.Error("A backup jobqueue should be inactive")
		}

		<-jq1.Stop()
		time.Sleep(2 * time.Second)

		if !jq2.IsActive() {
			t.Error("A backup jobqueue should be active after failing over")
		}

		<-jq2.Stop()
	})
}

func withShortLockWaitTimeout(block func()) {
	timeout := activatorLockWaitTimeout
	activatorLockWaitTimeout = 1 * time.Second
	defer func() { activatorLockWaitTimeout = timeout }()
	block()
}

func TestLostConnection(t *testing.T) {
	dsn := Dsn()

	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Error(cfg)
	}

	// Make a proxy server to emulate server-side disconnection.
	proxy := startProxy(cfg.Addr)
	cfg.Addr = proxy.from
	dsn = cfg.FormatDSN()

	config.Locally("queue_mysql_dsn", dsn, func() {
		q := "jobqueue_mysql_lost_connection_test"

		jq1 := NewPrimaryBackup(&model.Queue{Name: q, MaxWorkers: 30}, dsn)
		jq1.Start()

		time.Sleep(500 * time.Millisecond) // wait for up

		jq2 := NewPrimaryBackup(&model.Queue{Name: q, MaxWorkers: 30}, dsn)
		jq2.Start()

		jq3 := NewPrimaryBackup(&model.Queue{Name: q, MaxWorkers: 30}, dsn)
		jq3.Start()

		time.Sleep(500 * time.Millisecond) // wait for up

		if !jq1.IsActive() {
			t.Error("The primary jobqueue should be active")
		}
		if jq2.IsActive() {
			t.Error("A backup jobqueue should be inactive")
		}
		if jq3.IsActive() {
			t.Error("A backup jobqueue should be inactive")
		}

		// Disconnect and connect again.
		proxy.Restart()
		defer proxy.Stop()

		time.Sleep(1500 * time.Millisecond) // wait for up

		if countActive([]jobqueue.Impl{jq1, jq2, jq3}) != 1 {
			t.Error("A jobqueue should reconnect after connection lost")
		}

		<-jq1.Stop()

		// Wait for stopping jq1 and reconnection of jq2, jq3
		time.Sleep(1500 * time.Millisecond)

		if countActive([]jobqueue.Impl{jq2, jq3}) != 1 {
			t.Error("A backup jobqueue should be active after failing over")
		}

		<-jq2.Stop()

		// Wait for stopping jq2
		time.Sleep(1500 * time.Millisecond)

		if !jq3.IsActive() {
			t.Error("A backup jobqueue should be active after failing over")
		}

		<-jq3.Stop()
	})
}

func countActive(qs []jobqueue.Impl) int {
	c := 0
	for _, q := range qs {
		if q.IsActive() {
			c++
		}
	}
	return c
}

type proxy struct {
	cmd  *exec.Cmd
	from string
	to   string
}

func startProxy(to string) *proxy {
	proxy := &proxy{
		from: fmt.Sprintf("127.0.0.1:%d", freeport.GetPort()),
		to:   to,
	}
	proxy.Start()
	return proxy
}

func (p *proxy) Start() {
	p.cmd = exec.Command("tcp-proxy", "-l", p.from, "-r", p.to)
	p.cmd.Start()

	// Wait for starting listening on `p.from`
	<-time.After(500 * time.Millisecond)
}

func (p *proxy) Stop() {
	p.cmd.Process.Kill()
	p.cmd.Wait()
}

func (p *proxy) Restart() {
	p.Stop()
	p.Start()
}
