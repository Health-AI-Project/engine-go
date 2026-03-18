package grpc

import (
	"context"
	"fmt"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/services"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- Manual Protobuf Structs ---

type GetWorkoutRecommendationRequest struct {
	UserId             string   `json:"user_id"`
	DurationMinutes    int32    `json:"duration_minutes"`
	UserInjuries       []string `json:"user_injuries"`
	AvailableEquipment []string `json:"available_equipment"`
}

func (m *GetWorkoutRecommendationRequest) Reset()         { *m = GetWorkoutRecommendationRequest{} }
func (m *GetWorkoutRecommendationRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*GetWorkoutRecommendationRequest) ProtoMessage()    {}

type WorkoutPlanResponse struct {
	PlanId          string              `json:"plan_id"`
	Date            string              `json:"date"`
	DurationMinutes int32               `json:"duration_minutes"`
	EstCaloriesBurn float64             `json:"est_calories_burn"`
	Exercises       []*ProtoWorkoutItem `json:"exercises"`
}

func (m *WorkoutPlanResponse) Reset()         { *m = WorkoutPlanResponse{} }
func (m *WorkoutPlanResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*WorkoutPlanResponse) ProtoMessage()    {}

type ProtoWorkoutItem struct {
	ExerciseId  string           `json:"exercise_id"`
	Name        string           `json:"name"`
	Order       int32            `json:"order"`
	Sets        int32            `json:"sets"`
	Reps        int32            `json:"reps"`
	DurationSec int32            `json:"duration_sec"`
	RestSec     int32            `json:"rest_sec"`
	Details     *ExerciseDetails `json:"details"`
}

type ExerciseDetails struct {
	Type          string   `json:"type"`
	Difficulty    string   `json:"difficulty"`
	Equipment     string   `json:"equipment"`
	MuscleTargets []string `json:"muscle_targets"`
	VideoUrl      string   `json:"video_url"`
}

// --- Handler Logic ---

type WorkoutHandler struct {
	service *services.WorkoutService
}

func NewWorkoutHandler(service *services.WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{service: service}
}

func (h *WorkoutHandler) GetService() *services.WorkoutService {
	return h.service
}

func (h *WorkoutHandler) GetWorkoutRecommendation(ctx context.Context, req *GetWorkoutRecommendationRequest) (*WorkoutPlanResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Map Equipment
	equip := []domain.EquipmentType{}
	for _, e := range req.AvailableEquipment {
		equip = append(equip, domain.EquipmentType(e))
	}

	constraints := services.WorkoutConstraints{
		DurationMinutes: int(req.DurationMinutes),
		Equipment:       equip,
		UserInjuries:    req.UserInjuries,
	}

	plan, err := h.service.GenerateWorkout(ctx, req.UserId, constraints)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &WorkoutPlanResponse{
		PlanId:          plan.ID,
		Date:            plan.Date.Format("2006-01-02"),
		DurationMinutes: int32(plan.DurationMinutes),
		EstCaloriesBurn: plan.EstCaloriesBurn,
		Exercises:       []*ProtoWorkoutItem{},
	}

	for _, item := range plan.Exercises {
		resp.Exercises = append(resp.Exercises, &ProtoWorkoutItem{
			ExerciseId:  item.ExerciseID,
			Name:        item.Exercise.Name,
			Order:       int32(item.Order),
			Sets:        int32(item.Sets),
			Reps:        int32(item.Reps),
			DurationSec: int32(item.DurationSec),
			RestSec:     int32(item.RestSec),
			Details: &ExerciseDetails{
				Type:          string(item.Exercise.Type),
				Difficulty:    string(item.Exercise.Difficulty),
				Equipment:     string(item.Exercise.RequiredEquipment),
				MuscleTargets: item.Exercise.MusclesTargeted,
				VideoUrl:      item.Exercise.VideoURL,
			},
		})
	}

	return resp, nil
}

// --- Registration ---

type WorkoutServiceServer interface {
	GetWorkoutRecommendation(context.Context, *GetWorkoutRecommendationRequest) (*WorkoutPlanResponse, error)
}

func RegisterWorkoutService(s *grpc.Server, srv *WorkoutHandler) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "workout.WorkoutService",
		HandlerType: (*WorkoutServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{MethodName: "GetWorkoutRecommendation", Handler: _WorkoutService_GetWorkoutRecommendation_Handler},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "workout.proto",
	}, srv)
}

func _WorkoutService_GetWorkoutRecommendation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorkoutRecommendationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*WorkoutHandler).GetWorkoutRecommendation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/workout.WorkoutService/GetWorkoutRecommendation"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*WorkoutHandler).GetWorkoutRecommendation(ctx, req.(*GetWorkoutRecommendationRequest))
	}
	return interceptor(ctx, in, info, handler)
}
