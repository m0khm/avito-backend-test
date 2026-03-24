package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "room-booking-service/docs"
	"room-booking-service/internal/config"
	"room-booking-service/internal/http/handlers"
	authmw "room-booking-service/internal/http/middleware"
	"room-booking-service/internal/repo"
	"room-booking-service/internal/service"
	"room-booking-service/internal/tx"
)

type App struct {
	cfg    config.Config
	server *http.Server
	pool   *pgxpool.Pool
}

type migrationTx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type migrationExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context) (migrationTx, error)
}

type migrationPool struct {
	pool *pgxpool.Pool
}

func (m migrationPool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return m.pool.Exec(ctx, sql, arguments...)
}

func (m migrationPool) BeginTx(ctx context.Context) (migrationTx, error) {
	return m.pool.BeginTx(ctx, pgx.TxOptions{})
}

func New(cfg config.Config) (*App, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	if err := applyMigrations(ctx, migrationPool{pool: pool}); err != nil {
		return nil, err
	}

	userRepo := repo.NewUserRepository(pool)
	roomRepo := repo.NewRoomRepository(pool)
	scheduleRepo := repo.NewScheduleRepository(pool)
	slotRepo := repo.NewSlotRepository(pool)
	bookingRepo := repo.NewBookingRepository(pool)
	txManager := tx.NewManager(pool)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.DummyAdminUserID, cfg.DummyUserUserID)
	if err := authService.EnsureDummyUsers(ctx); err != nil {
		return nil, fmt.Errorf("ensure dummy users: %w", err)
	}
	conferenceClient := service.NewConferenceClient(cfg.ConferenceServiceURL, cfg.ConferenceTimeout)
	roomService := service.NewRoomService(roomRepo)
	slotService := service.NewSlotService(roomRepo, scheduleRepo, slotRepo, cfg.SlotHorizonDays)
	scheduleService := service.NewScheduleService(roomRepo, scheduleRepo, slotService, cfg.SlotHorizonDays)
	bookingService := service.NewBookingService(bookingRepo, slotRepo, txManager, conferenceClient)
	if err := slotService.ExtendAll(ctx); err != nil {
		return nil, fmt.Errorf("prime slots: %w", err)
	}

	h := handlers.New(authService, roomService, scheduleService, slotService, bookingService)
	r := newRouter(h, cfg.JWTSecret)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{cfg: cfg, server: server, pool: pool}, nil
}

func newRouter(h *handlers.Handler, jwtSecret string) http.Handler {
	authMiddleware := authmw.NewAuthMiddleware(jwtSecret)
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Get("/", h.Root)
	r.Get("/_info", h.Info)
	r.Post("/dummyLogin", h.DummyLogin)
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware.RequireAuth)
		protected.Get("/rooms/list", h.ListRooms)
		protected.Get("/rooms/{roomId}/slots/list", h.ListSlots)

		protected.With(authmw.RequireRole("admin")).Post("/rooms/create", h.CreateRoom)
		protected.With(authmw.RequireRole("admin")).Post("/rooms/{roomId}/schedule/create", h.CreateSchedule)
		protected.With(authmw.RequireRole("admin")).Get("/bookings/list", h.ListBookings)

		protected.With(authmw.RequireRole("user")).Post("/bookings/create", h.CreateBooking)
		protected.With(authmw.RequireRole("user")).Get("/bookings/my", h.ListMyBookings)
		protected.With(authmw.RequireRole("user")).Post("/bookings/{bookingId}/cancel", h.CancelBooking)
	})

	return r
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go a.runSlotExtensionLoop(ctx)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on :%s", a.cfg.HTTPPort)
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = a.server.Shutdown(shutdownCtx)
		a.pool.Close()
		return nil
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			a.pool.Close()
			return err
		}
		return nil
	}
}

func (a *App) runSlotExtensionLoop(ctx context.Context) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			roomRepo := repo.NewRoomRepository(a.pool)
			scheduleRepo := repo.NewScheduleRepository(a.pool)
			slotRepo := repo.NewSlotRepository(a.pool)
			slotService := service.NewSlotService(roomRepo, scheduleRepo, slotRepo, a.cfg.SlotHorizonDays)
			if err := slotService.ExtendAll(context.Background()); err != nil {
				log.Printf("slot extension failed: %v", err)
			}
		}
	}
}

func applyMigrations(ctx context.Context, db migrationExecutor) error {
	return applyMigrationsFromDir(ctx, db, filepath.Join("db", "migrations"))
}

func applyMigrationsFromDir(ctx context.Context, db migrationExecutor, migrationsDir string) error {
	files, err := migrationFiles(migrationsDir)
	if err != nil {
		return err
	}

	if _, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	for _, file := range files {
		content, readErr := os.ReadFile(file)
		if readErr != nil {
			return fmt.Errorf("read migration %s: %w", file, readErr)
		}

		tx, beginErr := db.BeginTx(ctx)
		if beginErr != nil {
			return fmt.Errorf("begin migration tx %s: %w", file, beginErr)
		}

		shouldRollback := true
		func() {
			defer func() {
				if shouldRollback {
					_ = tx.Rollback(ctx)
				}
			}()

			result, execErr := tx.Exec(ctx, `
				INSERT INTO schema_migrations (filename)
				VALUES ($1)
				ON CONFLICT (filename) DO NOTHING
			`, filepath.Base(file))
			if execErr != nil {
				err = fmt.Errorf("mark migration %s: %w", file, execErr)
				return
			}

			if result.RowsAffected() == 0 {
				if commitErr := tx.Commit(ctx); commitErr != nil {
					err = fmt.Errorf("commit skipped migration %s: %w", file, commitErr)
				}
				shouldRollback = false
				return
			}

			if len(content) > 0 {
				if _, execErr = tx.Exec(ctx, string(content)); execErr != nil {
					err = fmt.Errorf("apply migration %s: %w", file, execErr)
					return
				}
			}

			if commitErr := tx.Commit(ctx); commitErr != nil {
				err = fmt.Errorf("commit migration %s: %w", file, commitErr)
				return
			}
			shouldRollback = false
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

func migrationFiles(migrationsDir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return nil, fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)
	return files, nil
}
