package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Config struct {
	host       *string
	password   *string
	port       *int
	threads    *int
	numElems   *int
	packetSize *int
	interval   *int
	runCount   *int
}

var (
	cfg  = &Config{}
	pool *redis.Pool
	wg   sync.WaitGroup
)

func init() {
	cfg.host = flag.String("h", "localhost", "Redis server host")
	cfg.password = flag.String("p", "", "Redis password")
	cfg.port = flag.Int("P", 6379, "Redis port")
	cfg.threads = flag.Int("t", 3, "Number of concurrent threads")
	cfg.numElems = flag.Int("n", 10, "Number of elements to push/pop")
	cfg.packetSize = flag.Int("s", 100, "Packet size")
	cfg.interval = flag.Int("i", 10, "Interval between test runs in seconds")
	cfg.runCount = flag.Int("m", 2, "Number of times to run the tests")
}

func connectRedis() error {
	pool = &redis.Pool{
		MaxIdle:     *cfg.threads + 1,
		MaxActive:   *cfg.threads + 1,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:%d", *cfg.host, *cfg.port),
				redis.DialPassword(*cfg.password))
		},
	}

	// Test connection
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}
	return nil
}

func worker(threadID int) {
	defer wg.Done()
	conn := pool.Get()
	defer conn.Close()

	key := fmt.Sprintf("benchmark:list:%d", threadID)

	// Create test data string
	data := strings.Repeat("x", *cfg.packetSize)
	log.Printf("Starting worker %d, key: %s, data len: %d", threadID, key, len(data))
	// Push phase
	for j := 0; j < *cfg.numElems; j++ {
		_, err := conn.Do("RPUSH", key, data)
		if err != nil {
			log.Printf("Push error: %v", err)
			return
		}
	}

	// Pop phase
	for j := 0; j < *cfg.numElems; j++ {
		_, err := conn.Do("LPOP", key)
		if err != nil {
			log.Printf("Pop error: %v", err)
			return
		}
	}
	log.Printf("Worker %d completed", threadID)
}

func main() {
	// Add check for -h before flag.Parse()
	if len(os.Args) == 2 && os.Args[1] == "-h" {
		flag.Usage()
		return
	}

	flag.Parse()

	if err := connectRedis(); err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Run benchmark multiple times
	for i := 1; i <= *cfg.runCount; i++ {
		log.Printf("\nTest run %d of %d\n", i, *cfg.runCount)
		runBench()

		// Sleep between runs, but not after the last run
		if i == *cfg.runCount {
			break
		}
		fmt.Printf("Sleeping for %d seconds...\n", *cfg.interval)
		time.Sleep(time.Duration(*cfg.interval) * time.Second)

	}
}

func runBench() {

	startTime := time.Now()

	// Launch worker threads
	for i := 0; i < *cfg.threads; i++ {
		wg.Add(1)
		go worker(i)
	}
	wg.Wait()
	duration := time.Since(startTime)

	totalOps := *cfg.threads * *cfg.numElems * 2 // multiply by 2 for push + pop
	opsPerSec := float64(totalOps) / duration.Seconds()

	fmt.Printf("Completed %d operations in %.2f seconds\n", totalOps, duration.Seconds())
	fmt.Printf("Operations per second: %.2f\n", opsPerSec)
}
