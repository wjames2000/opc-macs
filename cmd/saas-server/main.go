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

type wrappedHandler struct {
	h   func(http.ResponseWriter, *http.Request)
	mws []func(http.Handler) http.Handler
}

func chain(h func(http.ResponseWriter, *http.Request), mws ...func(http.Handler) http.Handler) http.HandlerFunc {
	var next http.Handler = http.HandlerFunc(h)
	for i := len(mws) - 1; i >= 0; i-- {
		next = mws[i](next)
	}
	return next.ServeHTTP
}

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

	pub := saas.NewPublishHandler()

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

	// Marketplace routes
	mkt := srv.Marketplace
	mux.HandleFunc("/api/v1/marketplace/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			mkt.HandleListMyTasks(w, r)
		case "POST":
			mkt.HandleCreateTask(w, r)
		default:
			http.Error(w, "method not allowed", 405)
		}
	})
	mux.HandleFunc("/api/v1/marketplace/open-tasks", mkt.HandleListOpenTasks)
	mux.HandleFunc("/api/v1/marketplace/my-orders", mkt.HandleListMyOrders)
	mux.HandleFunc("/api/v1/marketplace/tasks/{id}", mkt.HandleGetTask)
	mux.HandleFunc("/api/v1/marketplace/tasks/{id}/apply", mkt.HandleApplyTask)

	wal := srv.Wallet
	mux.HandleFunc("GET /api/v1/wallet", wal.HandleGetWallet)
	mux.HandleFunc("GET /api/v1/wallet/transactions", wal.HandleGetTransactions)
	mux.HandleFunc("GET /api/v1/wallet/withdrawals", wal.HandleListWithdrawals)
	mux.HandleFunc("POST /api/v1/wallet/withdrawals", wal.HandleCreateWithdrawal)
	mux.HandleFunc("POST /api/v1/settlement/calculate", srv.Settlement.HandleCalculate)
	mux.HandleFunc("GET /api/v1/settlement/orders", srv.Settlement.HandleList)
	mux.HandleFunc("GET /api/v1/settlement/orders/{id}", srv.Settlement.HandleGet)
	mux.HandleFunc("GET /api/v1/settlement/invoices/{id}", srv.Settlement.HandleInvoiceDownload)

	cre := srv.Credit
	mux.HandleFunc("POST /api/v1/credit/reviews", cre.HandleAddReview)
	mux.HandleFunc("GET /api/v1/credit/score", cre.HandleGetMyScore)
	mux.HandleFunc("GET /api/v1/credit/score/{user_id}", cre.HandleGetScore)

	mux.HandleFunc("GET /api/v1/admin/withdrawals", chain(
		wal.HandleAdminListPendingWithdrawals,
		saas.RequireRole("admin", "superadmin"),
	))
	mux.HandleFunc("POST /api/v1/admin/withdrawals/{id}/process", chain(
		wal.HandleAdminProcessWithdrawal,
		saas.RequireRole("admin", "superadmin"),
	))

	notif := srv.Notify
	mux.HandleFunc("GET /api/v1/notifications", notif.ListNotifications)
	mux.HandleFunc("GET /api/v1/notifications/unread-count", notif.CountUnread)
	mux.HandleFunc("POST /api/v1/notifications/read/{id}", notif.MarkRead)
	mux.HandleFunc("POST /api/v1/notifications", notif.CreateNotification)
	mux.HandleFunc("DELETE /api/v1/notifications/{id}", notif.DeleteNotification)
	mux.HandleFunc("POST /api/v1/lifecycle/complete-order", notif.CompleteOrder)
	mux.HandleFunc("POST /api/v1/lifecycle/cancel-order", notif.CancelOrder)

	mux.HandleFunc("POST /api/v1/publish", pub.HandlePublish)
	mux.HandleFunc("GET /api/v1/publish/schedules", pub.HandleListSchedules)

	// Phase A — content marketing agent routes
	vs := srv.VideoScript
	mux.HandleFunc("POST /api/v1/video-script/generate", vs.HandleGenerate)
	mux.HandleFunc("GET /api/v1/video-script/info", vs.HandleInfo)

	eng := srv.Engage
	mux.HandleFunc("POST /api/v1/engage/reply", eng.HandleReply)
	mux.HandleFunc("POST /api/v1/engage/analyze", eng.HandleAnalyze)
	mux.HandleFunc("GET /api/v1/engage/info", eng.HandleInfo)

	tg := srv.TagGenerator
	mux.HandleFunc("POST /api/v1/tags/generate", tg.HandleGenerate)
	mux.HandleFunc("GET /api/v1/tags/info", tg.HandleInfo)

	tr := srv.TrendRadar
	mux.HandleFunc("POST /api/v1/trend-radar/report", tr.HandleReport)
	mux.HandleFunc("GET /api/v1/trend-radar/info", tr.HandleInfo)

	// Wave 3 — delivery + review routes
	dlv := srv.Delivery
	mux.HandleFunc("POST /api/v1/marketplace/orders/{id}/deliver", dlv.HandleDeliver)
	mux.HandleFunc("GET /api/v1/marketplace/orders/{id}", dlv.HandleGetOrder)

	rvw := srv.Review
	mux.HandleFunc("POST /api/v1/marketplace/orders/{id}/approve", rvw.HandleApprove)
	mux.HandleFunc("POST /api/v1/marketplace/orders/{id}/reject", rvw.HandleReject)

	// Cancel task route
	mux.HandleFunc("POST /api/v1/marketplace/tasks/{id}/cancel", mkt.HandleCancelTask)

	// Invite routes
	inv := srv.Invite
	mux.HandleFunc("POST /api/v1/marketplace/invite", inv.HandleInvite)
	mux.HandleFunc("GET /api/v1/marketplace/invitations", inv.HandleListInvitations)
	mux.HandleFunc("POST /api/v1/marketplace/invitations/{id}/accept", inv.HandleAcceptInvite)
	mux.HandleFunc("POST /api/v1/marketplace/invitations/{id}/decline", inv.HandleDeclineInvite)

	// Recommendation routes
	rec := srv.Recommend
	mux.HandleFunc("GET /api/v1/marketplace/recommendations", rec.HandleRecommendations)
	mux.HandleFunc("GET /api/v1/marketplace/recommendations/info", rec.HandleInfo)

	// Invoice routes
	invh := srv.Invoice
	mux.HandleFunc("POST /api/v1/invoices", invh.HandleCreate)
	mux.HandleFunc("GET /api/v1/invoices", invh.HandleList)
	mux.HandleFunc("GET /api/v1/invoices/{id}", invh.HandleGet)
	mux.HandleFunc("POST /api/v1/invoices/{id}/pay", invh.HandlePay)
	mux.HandleFunc("POST /api/v1/invoices/{id}/cancel", invh.HandleCancel)

	// Payment method routes
	pmh := srv.PaymentMethod
	mux.HandleFunc("GET /api/v1/payment-methods", pmh.HandleList)
	mux.HandleFunc("POST /api/v1/payment-methods", pmh.HandleUpsert)
	mux.HandleFunc("DELETE /api/v1/payment-methods/{id}", pmh.HandleDelete)
	mux.HandleFunc("POST /api/v1/payment-methods/{id}/default", pmh.HandleSetDefault)

	// Report export routes
	reph := srv.ReportExport
	mux.HandleFunc("GET /api/v1/reports/export/orders", reph.HandleExportOrders)
	mux.HandleFunc("GET /api/v1/reports/export/tasks", reph.HandleExportTasks)

	// Fraud detection routes
	fdh := srv.FraudDetect
	mux.HandleFunc("GET /api/v1/fraud/rules", fdh.HandleListRules)
	mux.HandleFunc("POST /api/v1/fraud/rules", fdh.HandleCreateRule)
	mux.HandleFunc("POST /api/v1/fraud/rules/{id}/toggle", fdh.HandleToggleRule)
	mux.HandleFunc("GET /api/v1/fraud/flags", fdh.HandleListFlags)

	// Fraud account routes
	fah := srv.FraudAccount
	mux.HandleFunc("GET /api/v1/fraud/accounts/check", fah.HandleCheckDuplicate)
	mux.HandleFunc("POST /api/v1/fraud/accounts/report", fah.HandleReportActivity)
	mux.HandleFunc("GET /api/v1/fraud/accounts/activities", fah.HandleListActivities)

	// Fraud content routes
	fch := srv.FraudContent
	mux.HandleFunc("POST /api/v1/fraud/content/check", fch.HandleCheckContent)
	mux.HandleFunc("GET /api/v1/fraud/content/checks", fch.HandleListChecks)

	// Fraud alert routes
	falh := srv.FraudAlert
	mux.HandleFunc("GET /api/v1/fraud/alerts", falh.HandleListAlerts)
	mux.HandleFunc("POST /api/v1/fraud/alerts", falh.HandleCreateAlert)
	mux.HandleFunc("POST /api/v1/fraud/alerts/{id}/status", falh.HandleUpdateStatus)
	mux.HandleFunc("GET /api/v1/fraud/alerts/stats", falh.HandleGetStats)

	// Withdraw payment routes (vendor-initiated)
	wp := srv.WithdrawPayment
	mux.HandleFunc("POST /api/v1/withdraw-payments", wp.HandleCreate)
	mux.HandleFunc("GET /api/v1/withdraw-payments", wp.HandleList)
	mux.HandleFunc("GET /api/v1/withdraw-payments/{id}", wp.HandleGet)
	mux.HandleFunc("POST /api/v1/withdraw-payments/{id}/process", wp.HandleProcess)

	// Report routes
	ar := srv.AdvertiserReport
	mux.HandleFunc("GET /api/v1/reports/advertiser/summary", ar.HandleSummary)

	cr := srv.CreatorReport
	mux.HandleFunc("GET /api/v1/reports/creator/summary", cr.HandleSummary)
	mux.HandleFunc("GET /api/v1/reports/creator/orders", cr.HandleOrders)
	mux.HandleFunc("GET /api/v1/reports/creator/monthly", cr.HandleMonthly)

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
