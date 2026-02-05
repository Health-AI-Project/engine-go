package grpc

import (
	"context"

	"healthai/engine/internal/core/services"
	// pb "healthai/engine/proto/user" // generic import, assuming generation
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// To make this compilable without the generated code, I am defining aliases here.
// IN A REAL SCENARIO: Remove these and import the generated 'pb' package.
type UserProfileResponse struct {
	Id                 string
	Email              string
	SubscriptionStatus string
	Weight             float64
	Height             float64
	IsPremium          bool
}

type UserIdRequest struct {
	UserId string
}

type BiometricsRequest struct {
	UserId string
	Weight float64
	Height float64
}

type Empty struct{}

// Manual definitions for Feature B
type NutritionRequest struct {
	UserId   string
	Calories float64
	Protein  float64
	Carbs    float64
	Fat      float64
}

type WorkoutRequest struct {
	UserId          string
	Type            string
	DurationMinutes int32
	CaloriesBurned  float64
}

type DailyStatsResponse struct {
	Calories      float64
	Protein       float64
	Carbs         float64
	Fat           float64
	WorkoutsCount int32
}

type Ack struct {
	Success bool
	Message string
}


// UserHandler implements the gRPC service methods.
type UserHandler struct {
	service         *services.UserService
	activityService *services.ActivityService
	// pb.UnimplementedUserServiceServer 
}

func NewUserHandler(service *services.UserService, activityService *services.ActivityService) *UserHandler {
	return &UserHandler{
		service:         service,
		activityService: activityService,
	}
}

func (h *UserHandler) GetUserProfile(ctx context.Context, req *UserIdRequest) (*UserProfileResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := h.service.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	isPremium := h.service.CanAccessAdvancedFeatures(user)

	return &UserProfileResponse{
		Id:                 user.ID,
		Email:              user.Email,
		SubscriptionStatus: string(user.SubscriptionStatus),
		Weight:             user.Weight,
		Height:             user.Height,
		IsPremium:          isPremium,
	}, nil
}

func (h *UserHandler) GetDailyStats(ctx context.Context, req *UserIdRequest) (*DailyStatsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	log, err := h.activityService.GetDailyStats(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get daily stats")
	}

	if log == nil {
		return &DailyStatsResponse{}, nil
	}

	return &DailyStatsResponse{
		Calories:      log.TotalCalories,
		Protein:       log.TotalProtein,
		Carbs:         log.TotalCarbs,
		Fat:           log.TotalFat,
		WorkoutsCount: 0, // In a real scenario, we'd count today's workouts too
	}, nil
}

func (h *UserHandler) UpdateBiometrics(ctx context.Context, req *BiometricsRequest) (*Empty, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := h.service.UpdateBiometrics(ctx, req.UserId, req.Weight, req.Height)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update biometrics")
	}

	return &Empty{}, nil
}

// Feature B: Activity RPCs

func (h *UserHandler) LogNutrition(ctx context.Context, req *NutritionRequest) (*Ack, error) {
	if req.UserId == "" {
		return &Ack{Success: false, Message: "user_id is required"}, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := h.activityService.LogNutrition(ctx, req.UserId, req.Calories, req.Protein, req.Carbs, req.Fat)
	if err != nil {
		return &Ack{Success: false, Message: err.Error()}, status.Error(codes.Internal, "failed to log nutrition")
	}

	return &Ack{Success: true, Message: "nutrition logged"}, nil
}

func (h *UserHandler) LogWorkout(ctx context.Context, req *WorkoutRequest) (*Ack, error) {
	if req.UserId == "" {
		return &Ack{Success: false, Message: "user_id is required"}, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := h.activityService.LogWorkout(ctx, req.UserId, req.Type, int(req.DurationMinutes), req.CaloriesBurned)
	if err != nil {
		return &Ack{Success: false, Message: err.Error()}, status.Error(codes.Internal, "failed to log workout")
	}

	return &Ack{Success: true, Message: "workout logged"}, nil
}
