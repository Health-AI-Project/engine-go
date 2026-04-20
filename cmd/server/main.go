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

	"github.com/lib/pq"

	// "healthai/engine/internal/core/ports" // Not strictly needed if fx uses type inference, but good for clarity if referenced. Use implicit matching.

	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func NewDatabase() (*gorm.DB, error) {
	var dsn string
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		dsn = dbURL
	} else {
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
	}

	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		PrepareStmt: false, // Disable prepared statements to fix array binding
	})
	if err != nil {
		return nil, err
	}

	// AutoMigrate
	// db.Exec("CREATE EXTENSION IF NOT EXISTS vector;") // Disabled: pgvector not installed on shared DB
	if err := db.AutoMigrate(
		// &domain.User{}, // Migrated manually to avoid conflicts with Drizzle
		&domain.DailyLog{},
		&domain.Workout{},
		&domain.HealthProfile{},
		// Feature 1: Nutrition
		&domain.FoodPreference{},
		&domain.MealPlan{},
		&domain.MealPlanItem{},
		// Feature 2: Workout Recommendation
		&domain.WorkoutPlan{},
		&domain.WorkoutItem{},
		&domain.Exercise{},
		// &domain.MealSuggestion{}, // Disabled: require pgvector extension
	); err != nil {
		fmt.Printf("AutoMigrate warning: %v\n", err)
		// Continue anyway - tables may already exist
	}

	// Seed exercises if empty
	seedExercises(db)

	// DEBUG: Print actual connection info
	sqlDB, _ := db.DB()
	var currentDB string
	var serverAddr string
	sqlDB.QueryRow("SELECT current_database()").Scan(&currentDB)
	sqlDB.QueryRow("SELECT inet_server_addr()").Scan(&serverAddr)
	fmt.Printf("[DEBUG] Connected to Database: %s at %s (Configured Host: %s)\n", currentDB, serverAddr, os.Getenv("DB_HOST"))

	// DEBUG: Check columns in user table
	rows, err := sqlDB.Query("SELECT column_name FROM information_schema.columns WHERE table_name = 'user'")
	if err == nil {
		defer rows.Close()
		fmt.Println("[DEBUG] Columns in 'user' table:")
		for rows.Next() {
			var colName string
			rows.Scan(&colName)
			fmt.Printf(" - %s\n", colName)
		}
	} else {
		fmt.Printf("[DEBUG] Failed to list columns: %v\n", err)
	}

	return db, nil
}

func NewGRPCServer(lc fx.Lifecycle, coreHandler *usergrpc.CoreHandler, userHandler *usergrpc.UserHandler, nutritionHandler *usergrpc.NutritionHandler, workoutHandler *usergrpc.WorkoutHandler) *grpc.Server {
	server := grpc.NewServer()

	// Register the CoreService
	usergrpc.RegisterCoreService(server, coreHandler)

	// Keep existing ones for backward compatibility if needed, or remove them
	usergrpc.RegisterUserService(server, userHandler)
	usergrpc.RegisterNutritionService(server, nutritionHandler)
	usergrpc.RegisterWorkoutService(server, workoutHandler)

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
func seedExercises(db *gorm.DB) {
	var count int64
	db.Model(&domain.Exercise{}).Count(&count)
	if count > 0 {
		fmt.Printf("[SEED] %d exercises already exist, skipping seed\n", count)
		return
	}

	exercises := []domain.Exercise{
		{Name: "Pompes", Type: domain.ExerciseTypeStrength, RequiredEquipment: domain.EquipmentNone, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Pectoraux", "Triceps", "Epaules"}},
		{Name: "Squats", Type: domain.ExerciseTypeStrength, RequiredEquipment: domain.EquipmentNone, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Quadriceps", "Fessiers", "Ischio-jambiers"}},
		{Name: "Planche", Type: domain.ExerciseTypeStrength, RequiredEquipment: domain.EquipmentNone, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Abdominaux", "Dos"}},
		{Name: "Fentes", Type: domain.ExerciseTypeStrength, RequiredEquipment: domain.EquipmentNone, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Quadriceps", "Fessiers"}},
		{Name: "Burpees", Type: domain.ExerciseTypeCardio, RequiredEquipment: domain.EquipmentNone, Difficulty: domain.DifficultyIntermediate, MusclesTargeted: pq.StringArray{"Corps entier"}},
		{Name: "Mountain Climbers", Type: domain.ExerciseTypeCardio, RequiredEquipment: domain.EquipmentNone, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Abdominaux", "Epaules", "Cardio"}},
		{Name: "Crunchs", Type: domain.ExerciseTypeStrength, RequiredEquipment: domain.EquipmentNone, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Abdominaux"}},
		{Name: "Dips sur chaise", Type: domain.ExerciseTypeStrength, RequiredEquipment: domain.EquipmentNone, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Triceps", "Pectoraux"}},
		{Name: "Jumping Jacks", Type: domain.ExerciseTypeCardio, RequiredEquipment: domain.EquipmentNone, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Corps entier", "Cardio"}},
		{Name: "Superman", Type: domain.ExerciseTypeStrength, RequiredEquipment: domain.EquipmentNone, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Dos", "Fessiers"}},
		{Name: "Curl Biceps", Type: domain.ExerciseTypeStrength, RequiredEquipment: domain.EquipmentDumbbells, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Biceps"}},
		{Name: "Developpe Militaire", Type: domain.ExerciseTypeStrength, RequiredEquipment: domain.EquipmentDumbbells, Difficulty: domain.DifficultyIntermediate, MusclesTargeted: pq.StringArray{"Epaules", "Triceps"}},
		{Name: "Rowing Haltere", Type: domain.ExerciseTypeStrength, RequiredEquipment: domain.EquipmentDumbbells, Difficulty: domain.DifficultyIntermediate, MusclesTargeted: pq.StringArray{"Dos", "Biceps"}},
		{Name: "Squat Goblet", Type: domain.ExerciseTypeStrength, RequiredEquipment: domain.EquipmentDumbbells, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Quadriceps", "Fessiers"}},
		{Name: "Etirements", Type: domain.ExerciseTypeFlexibility, RequiredEquipment: domain.EquipmentNone, Difficulty: domain.DifficultyBeginner, MusclesTargeted: pq.StringArray{"Corps entier"}},
	}

	for i := range exercises {
		if err := db.Create(&exercises[i]).Error; err != nil {
			fmt.Printf("[SEED] Warning: failed to seed exercise %s: %v\n", exercises[i].Name, err)
		}
	}
	fmt.Printf("[SEED] Seeded %d exercises\n", len(exercises))
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	fx.New(
		fx.Provide(
			NewDatabase,
			postgres.NewUserRepository,     // Returns ports.UserRepository
			postgres.NewActivityRepository, // Returns ports.ActivityRepository
			postgres.NewNutritionRepository,
			postgres.NewWorkoutRepository,
			services.NewUserService,
			services.NewActivityService,
			services.NewNutritionService,
			services.NewWorkoutService,
			usergrpc.NewUserHandler,
			usergrpc.NewNutritionHandler,
			usergrpc.NewWorkoutHandler,
			usergrpc.NewCoreHandler,
			NewGRPCServer,
		),
		fx.Invoke(func(*grpc.Server) {}), // Invoke to trigger server creation
	).Run()
}
