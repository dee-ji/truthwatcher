package queue

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type JobType string

const (
	JobTypeCompile   JobType = "compile"
	JobTypeDeploy    JobType = "deploy"
	JobTypeReconcile JobType = "reconcile"
)

var ErrEmpty = errors.New("queue empty")

type Job struct {
	ID        string         `json:"id"`
	Type      JobType        `json:"type"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type Queue interface {
	Enqueue(context.Context, Job) error
	Dequeue(context.Context, []JobType, time.Duration) (Job, error)
	Depth(context.Context, JobType) (int64, error)
}

type RedisQueue struct {
	addr   string
	prefix string
}

func NewRedisQueue(addr, prefix string) *RedisQueue {
	if prefix == "" {
		prefix = "truthwatcher"
	}
	return &RedisQueue{addr: addr, prefix: prefix}
}

func (q *RedisQueue) key(t JobType) string { return fmt.Sprintf("%s:queue:%s", q.prefix, t) }

func (q *RedisQueue) Enqueue(_ context.Context, job Job) error {
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = q.exec("LPUSH", q.key(job.Type), string(b))
	return err
}

func (q *RedisQueue) Dequeue(_ context.Context, types []JobType, wait time.Duration) (Job, error) {
	args := []string{"BRPOP"}
	for _, t := range types {
		args = append(args, q.key(t))
	}
	args = append(args, strconv.Itoa(int(wait.Seconds())))
	res, err := q.exec(args...)
	if err != nil {
		if strings.Contains(err.Error(), "nil") {
			return Job{}, ErrEmpty
		}
		return Job{}, err
	}
	parts := strings.SplitN(res, "\n", 2)
	if len(parts) != 2 {
		return Job{}, ErrEmpty
	}
	var job Job
	if err := json.Unmarshal([]byte(parts[1]), &job); err != nil {
		return Job{}, err
	}
	return job, nil
}

func (q *RedisQueue) Depth(_ context.Context, t JobType) (int64, error) {
	res, err := q.exec("LLEN", q.key(t))
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(res), 10, 64)
}

func (q *RedisQueue) Ping() error {
	_, err := q.exec("PING")
	return err
}

func (q *RedisQueue) exec(args ...string) (string, error) {
	conn, err := net.DialTimeout("tcp", q.addr, 2*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return "", err
	}
	cmd := fmt.Sprintf("*%d\r\n", len(args))
	for _, a := range args {
		cmd += fmt.Sprintf("$%d\r\n%s\r\n", len(a), a)
	}
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return "", err
	}
	return readRESP(bufio.NewReader(conn))
}

func readRESP(r *bufio.Reader) (string, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return "", err
	}
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	switch prefix {
	case '+', ':':
		return line, nil
	case '-':
		return "", errors.New(line)
	case '$':
		n, _ := strconv.Atoi(line)
		if n < 0 {
			return "", errors.New("nil")
		}
		buf := make([]byte, n+2)
		if _, err := r.Read(buf); err != nil {
			return "", err
		}
		return string(buf[:n]), nil
	case '*':
		n, _ := strconv.Atoi(line)
		if n <= 0 {
			return "", errors.New("nil")
		}
		values := make([]string, 0, n)
		for i := 0; i < n; i++ {
			v, err := readRESP(r)
			if err != nil {
				return "", err
			}
			values = append(values, v)
		}
		return strings.Join(values, "\n"), nil
	default:
		return "", fmt.Errorf("unknown redis response prefix %q", prefix)
	}
}

type InMemoryQueue struct {
	mu    sync.Mutex
	jobs  map[JobType][]Job
	notif chan struct{}
}

func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{jobs: map[JobType][]Job{}, notif: make(chan struct{}, 1)}
}

func (q *InMemoryQueue) Enqueue(_ context.Context, job Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs[job.Type] = append(q.jobs[job.Type], job)
	select {
	case q.notif <- struct{}{}:
	default:
	}
	return nil
}

func (q *InMemoryQueue) Dequeue(ctx context.Context, types []JobType, wait time.Duration) (Job, error) {
	if wait <= 0 {
		wait = time.Millisecond
	}
	deadline := time.NewTimer(wait)
	defer deadline.Stop()
	for {
		q.mu.Lock()
		for _, t := range types {
			if len(q.jobs[t]) > 0 {
				job := q.jobs[t][0]
				q.jobs[t] = q.jobs[t][1:]
				q.mu.Unlock()
				return job, nil
			}
		}
		q.mu.Unlock()

		select {
		case <-ctx.Done():
			return Job{}, ctx.Err()
		case <-deadline.C:
			return Job{}, ErrEmpty
		case <-q.notif:
		}
	}
}

func (q *InMemoryQueue) Depth(_ context.Context, t JobType) (int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return int64(len(q.jobs[t])), nil
}
