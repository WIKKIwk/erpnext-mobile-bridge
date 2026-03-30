package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"mobile_server/internal/config"
	"mobile_server/internal/core"
	"mobile_server/internal/erpnext"
	"mobile_server/internal/mobileapi"
)

func main() {
	addr := strings.TrimSpace(os.Getenv("MOBILE_API_ADDR"))
	if addr == "" {
		addr = ":8081"
	}
	profileStorePath := strings.TrimSpace(os.Getenv("MOBILE_API_PROFILE_STORE_PATH"))
	if profileStorePath == "" {
		profileStorePath = "data/mobile_profile_prefs.json"
	}
	adminSupplierStorePath := strings.TrimSpace(os.Getenv("MOBILE_API_ADMIN_SUPPLIER_STORE_PATH"))
	if adminSupplierStorePath == "" {
		adminSupplierStorePath = "data/mobile_admin_suppliers.json"
	}
	sessionStorePath := strings.TrimSpace(os.Getenv("MOBILE_API_SESSION_STORE_PATH"))
	if sessionStorePath == "" {
		sessionStorePath = "data/mobile_sessions.json"
	}
	sessionTTL := 30 * 24 * time.Hour
	if raw := strings.TrimSpace(os.Getenv("MOBILE_API_SESSION_TTL_HOURS")); raw != "" {
		hours, err := strconv.Atoi(raw)
		if err != nil {
			log.Fatalf("invalid MOBILE_API_SESSION_TTL_HOURS: %v", err)
		}
		if hours < 0 {
			log.Fatalf("invalid MOBILE_API_SESSION_TTL_HOURS: must be >= 0")
		}
		sessionTTL = time.Duration(hours) * time.Hour
	}

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	erpClient := erpnext.NewClient(&http.Client{Timeout: cfg.RequestTimeout})
	service := core.NewERPAuthenticator(
		erpClient,
		cfg.DefaultERPURL,
		cfg.DefaultERPAPIKey,
		cfg.DefaultERPAPISecret,
		cfg.DefaultTargetWarehouse,
		os.Getenv("MOBILE_DEV_SUPPLIER_PREFIX"),
		os.Getenv("MOBILE_DEV_WERKA_PREFIX"),
		os.Getenv("MOBILE_DEV_WERKA_CODE"),
		cfg.WerkaPhone,
		os.Getenv("MOBILE_DEV_WERKA_NAME"),
		core.NewProfileStore(profileStorePath),
		core.NewAdminSupplierStore(adminSupplierStorePath),
	)
	service.SetAdminIdentity(
		"+998880000000",
		"Admin",
		"19621978",
		config.NewDotEnvPersister(".env"),
	)

	server := mobileapi.NewServerWithSessionManager(
		service,
		mobileapi.NewPersistentSessionManager(sessionStorePath, sessionTTL),
	)
	log.Printf("core listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatalf("core stopped: %v", err)
	}
}
