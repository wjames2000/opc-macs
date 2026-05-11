package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"github.com/wjames2000/opc-macs/internal/plugins"
	"github.com/wjames2000/opc-macs/internal/runtime"
	"github.com/wjames2000/opc-macs/internal/saas"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	dbURL := flag.String("db", "postgres://localhost:5432/opc_saas?sslmode=disable", "PostgreSQL connection string")
	flag.Parse()

	// Connect to database
	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("数据库连接失败：%v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库 Ping 失败：%v\n请确保 PostgreSQL 已启动且数据库存在", err)
	}

	// Load embedded agents
	pluginLoader := runtime.NewLoader("./build/plugins")
	if err := plugins.RegisterAll(pluginLoader); err != nil {
		log.Fatalf("插件注册失败：%v", err)
	}
	log.Printf("已加载 %d 个 Agent 插件", pluginLoader.Count())

	// Create SaaS server
	srv, err := saas.NewServer(db, pluginLoader)
	if err != nil {
		log.Fatalf("SaaS 服务初始化失败：%v", err)
	}

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/healthz", srv.HandleHealth)
	mux.HandleFunc("/api/v1/auth/register", srv.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", srv.HandleLogin)
	mux.HandleFunc("/api/v1/agents", srv.HandleListAgents)
	mux.HandleFunc("/api/v1/tenants", srv.HandleListTenants)
	mux.HandleFunc("/api/v1/usage/record", srv.HandleRecordUsage)
	mux.HandleFunc("/api/v1/usage/summary", srv.HandleUsageSummary)
	mux.HandleFunc("/api/v1/agents/execute", srv.HandleExecuteAgent)

	// Wrap with CORS
	handler := corsMiddleware(mux)

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("OPC-Agent SaaS 服务启动在 %s", addr)

	go func() {
		if err := http.ListenAndServe(addr, handler); err != nil {
			log.Fatalf("服务启动失败：%v", err)
		}
	}()

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("正在关闭服务...")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
