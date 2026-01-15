package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/controller"
	clusterRegistered "github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/events/clusterRegisteredEvent"
	clusterStatusEvent "github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/events/clusterStatusEvent"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/heartbeat"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/install"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/kube"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/store/clusterRegistry"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/store/clusterStatus"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	controlPlaneURL := mustEnv("CONTROL_PLANE_URL", false)
	postgresURL := mustEnv("AGENT_DB_URL", false)

	redisURL := mustEnv("REDIS_URL", false)
	redisPassword := mustEnv("REDIS_PASSWORD", true)
	redisStream := mustEnv("REDIS_STREAM", false)
	redisGroup := mustEnv("REDIS_GROUP", false)
	redisConsumer := mustEnv("REDIS_CONSUMER", false)

	redisDB := 0
	if v := mustEnv("REDIS_DB", false); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			redisDB = parsed
		}
	}

	redisClient := clusterStatusEvent.NewRedisClient(clusterStatusEvent.RedisConfig{
		Addr:     redisURL,
		Password: redisPassword,
		DB:       redisDB,
	})

	if err := clusterStatusEvent.PingRedis(ctx, redisClient); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	statusPublisher := clusterStatusEvent.NewRedisPublisher(redisClient, redisStream)

	db, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer db.Close()

	if err := clusterStatus.RunMigrations(ctx, db); err != nil {
		log.Fatalf("cluster status migrations failed: %v", err)
	}

	if err := clusterRegistry.RunMigrations(ctx, db); err != nil {
		log.Fatalf("cluster registry migrations failed: %v", err)
	}

	clusterStatusStore := clusterStatus.NewStatusStore(db)
	clusterRegistryStore := clusterRegistry.NewPostgresStore(db)

	kubeClient, err := kube.NewKubeClient()
	if err != nil {
		log.Fatalf("failed to init kube client: %v", err)
	}

	reconcileInterval := envDuration("RECONCILE_INTERVAL", 10*time.Second)

	installer := install.NewHelmInstaller(
		"sentinel-agent",
		"https://ayuspoudel.github.io/sentinel-sre",
		"sentinel-sre",
	)

	controller := controller.NewController(
		kubeClient,
		clusterStatusStore,
		clusterRegistryStore,
		reconcileInterval,
		controlPlaneURL,
		installer,
		statusPublisher,
	)

	clusterConsumer := clusterRegistered.NewConsumer(
		redisClient,
		redisStream,
		redisGroup,
		redisConsumer,
		clusterRegistryStore,
	)

	heartbeatAddr := os.Getenv("HEARTBEAT_BIND_ADDR")
	if heartbeatAddr == "" {
		heartbeatAddr = ":9000"
	}

	heartbeatServer := heartbeat.NewServer(
		heartbeatAddr,
		clusterStatusStore,
	)

	go func() {
		if err := heartbeatServer.Start(ctx); err != nil {
			log.Printf("heartbeat server exited: %v", err)
		}
	}()

	go clusterConsumer.Run(ctx)
	go controller.Run(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received")
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("sentinel agent controller exited cleanly")
}

func mustEnv(name string, omitempty bool) string {
	value := os.Getenv(name)
	if value == "" && omitempty == false {
		log.Fatalf("environment variable %s is required", name)
	}
	return value
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
