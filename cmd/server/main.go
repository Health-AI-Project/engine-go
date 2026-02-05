package main

import (
	"context"
	"fmt"
	"net"
	"os"

	usergrpc "healthai/engine/internal/adapters/handlers/grpc"
	"healthai/engine/internal/adapters/repositories/postgres"
	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/services"
	// "healthai/engine/internal/core/ports" // Not strictly needed if fx uses type inference, but good for clarity if referenced. Use implicit matching.

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	
	// AutoMigrate
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.DailyLog{},
		&domain.Workout{},
	); err != nil {
		return nil, err
	}
	
	return db, nil
}

func NewGRPCServer(lc fx.Lifecycle, userHandler *usergrpc.UserHandler) *grpc.Server {
	server := grpc.NewServer()

	// TODO: Register the UserServiceServer here once code is generated.
	// pb.RegisterUserServiceServer(server, userHandler)
	
	// Enable reflection for debugging (e.g. with grpcurl)
	reflection.Register(server)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", ":50051")
			if err != nil {
				return err
			}
			fmt.Println("Starting gRPC server on :50051")
			go server.Serve(lis)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Stopping gRPC server")
			server.GracefulStop()
			return nil
		},
	})

	return server
}

// Alias for import clarity since we named our local package 'grpc'
// In real code, use a different name for the local package or alias the system one.
// Here I actually aliased the local one in imports but let's be explicit.
// The import above is "healthai/engine/internal/adapters/handlers/grpc" -> which creates package content accessible as 'grpc'.
// The standard library one is "google.golang.org/grpc".
// Go will complain about conflict. I should alias the standard one or the local one.

// Retrying imports for clarity in the file content below.
func main() {
	fx.New(
		fx.Provide(
			NewDatabase,
			postgres.NewUserRepository, // Returns ports.UserRepository
			postgres.NewActivityRepository, // Returns ports.ActivityRepository
			services.NewUserService,
			services.NewActivityService,
			usergrpc.NewUserHandler,
			NewGRPCServer,
		),
		fx.Invoke(func(*grpc.Server) {}), // Invoke to trigger server creation
	).Run()
}
