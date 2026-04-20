package grpc

import (
	"context"
	"time"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/services"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- Mappings ---

// Ensure we don't redeclare types present in user_handler.go within the same package.
// Reused types: UserIdRequest, UserProfileResponse, DailyStatsResponse, BiometricsRequest,
// UpdateHealthProfileRequest, NutritionRequest, WorkoutRequest, Ack, Empty.

// --- Core Service Interface ---

type CoreServiceServer interface {
	GetUserProfile(context.Context, *UserIdRequest) (*UserProfileResponse, error)
	GetDailyStats(context.Context, *UserIdRequest) (*DailyStatsResponse, error)
	UpdateBiometrics(context.Context, *BiometricsRequest) (*Empty, error)
	LogNutrition(context.Context, *NutritionRequest) (*Ack, error)
	LogWorkout(context.Context, *WorkoutRequest) (*Ack, error)
	UpdateHealthProfile(context.Context, *UpdateHealthProfileRequest) (*Ack, error)

	UpdateFoodPreferences(context.Context, *UpdateFoodPreferencesRequest) (*Ack, error)
	AnalyzeMeal(context.Context, *AnalyzeMealRequest) (*AnalyzeMealResponse, error)
	GetMealPlan(context.Context, *GetMealPlanRequest) (*MealPlanResponse, error)
	GetFoodPreferences(context.Context, *GetFoodPreferencesRequest) (*FoodPreferencesResponse, error)
	GetWorkoutRecommendation(context.Context, *GetWorkoutRecommendationRequest) (*WorkoutPlanResponse, error)
	GetCaloriesHistory(context.Context, *HistoryRequest) (*CaloriesHistoryResponse, error)
	GetWeightHistory(context.Context, *HistoryRequest) (*WeightHistoryResponse, error)
}

// --- Handler Implementation ---

type CoreHandler struct {
	userService      *services.UserService
	activityService  *services.ActivityService
	nutritionService *services.NutritionService
	workoutService   *services.WorkoutService
}

func NewCoreHandler(
	u *services.UserService,
	a *services.ActivityService,
	n *services.NutritionService,
	w *services.WorkoutService,
) *CoreHandler {
	return &CoreHandler{
		userService:      u,
		activityService:  a,
		nutritionService: n,
		workoutService:   w,
	}
}

// 1. GetUserProfile (Proxy to UserHandler logic)
func (h *CoreHandler) GetUserProfile(ctx context.Context, req *UserIdRequest) (*UserProfileResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	user, err := h.userService.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	isPremium := h.userService.CanAccessAdvancedFeatures(user)
	var dobStr string
	var age int32
	var goals []string
	var allergies []string

	if user.DateOfBirth != nil {
		dobStr = user.DateOfBirth.Format("2006-01-02")
		now := time.Now()
		age = int32(now.Year() - user.DateOfBirth.Year())
		if now.YearDay() < user.DateOfBirth.YearDay() {
			age--
		}
	}
	if user.HealthProfile != nil {
		goals = []string(user.HealthProfile.Goals)
		allergies = []string(user.HealthProfile.Allergies)
	}

	return &UserProfileResponse{
		Id:                 user.ID,
		Email:              user.Email,
		SubscriptionStatus: string(user.SubscriptionStatus),
		Weight:             user.Weight,
		Height:             user.Height,
		IsPremium:          isPremium,
		DateOfBirth:        dobStr,
		Age:                age,
		Goals:              goals,
		Allergies:          allergies,
	}, nil
}

// 2. GetDailyStats
func (h *CoreHandler) GetDailyStats(ctx context.Context, req *UserIdRequest) (*DailyStatsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	log, err := h.activityService.GetDailyStats(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stats")
	}
	if log == nil {
		return &DailyStatsResponse{}, nil
	}
	return &DailyStatsResponse{Calories: log.TotalCalories, Protein: log.TotalProtein, Carbs: log.TotalCarbs, Fat: log.TotalFat, WorkoutsCount: 0}, nil
}

// 3. UpdateBiometrics
func (h *CoreHandler) UpdateBiometrics(ctx context.Context, req *BiometricsRequest) (*Empty, error) {
	// Reusing logic (simplified)
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "id req")
	}
	h.userService.UpdateBiometrics(ctx, req.UserId, req.Weight, req.Height)
	return &Empty{}, nil
}

// 4. UpdateHealthProfile
func (h *CoreHandler) UpdateHealthProfile(ctx context.Context, req *UpdateHealthProfileRequest) (*Ack, error) {
	if req.UserId == "" {
		return &Ack{Success: false, Message: "id required"}, nil
	}
	var dob *time.Time
	if req.DateOfBirth != "" {
		t, _ := time.Parse("2006-01-02", req.DateOfBirth)
		dob = &t
	}
	err := h.userService.UpdateHealthProfile(ctx, req.UserId, dob, req.Goals, req.Allergies, req.Weight, req.Height)
	if err != nil {
		return &Ack{Success: false, Message: err.Error()}, nil
	}
	return &Ack{Success: true, Message: "updated"}, nil
}

// 5. LogNutrition
func (h *CoreHandler) LogNutrition(ctx context.Context, req *NutritionRequest) (*Ack, error) {
	h.activityService.LogNutrition(ctx, req.UserId, req.Calories, req.Protein, req.Carbs, req.Fat)
	return &Ack{Success: true}, nil
}

// 6. LogWorkout
func (h *CoreHandler) LogWorkout(ctx context.Context, req *WorkoutRequest) (*Ack, error) {
	h.activityService.LogWorkout(ctx, req.UserId, req.Type, int(req.DurationMinutes), req.CaloriesBurned)
	return &Ack{Success: true}, nil
}

func (h *CoreHandler) UpdateFoodPreferences(ctx context.Context, req *UpdateFoodPreferencesRequest) (*Ack, error) {
	err := h.nutritionService.UpdateFoodPreferences(ctx, req.UserId, req.Allergies, domain.DietType(req.DietType), req.DislikedIngredients)
	if err != nil {
		return &Ack{Success: false, Message: err.Error()}, nil
	}
	return &Ack{Success: true, Message: "updated"}, nil
}

func (h *CoreHandler) GetFoodPreferences(ctx context.Context, req *GetFoodPreferencesRequest) (*FoodPreferencesResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	prefs, err := h.nutritionService.GetFoodPreferences(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if prefs == nil {
		return &FoodPreferencesResponse{Allergies: []string{}, DietType: "NONE", DislikedIngredients: []string{}}, nil
	}
	return &FoodPreferencesResponse{
		Allergies:           prefs.Allergies,
		DietType:            string(prefs.DietType),
		DislikedIngredients: prefs.DislikedIngredients,
	}, nil
}

// Stubs for others with Freemium Block
func (h *CoreHandler) AnalyzeMeal(ctx context.Context, req *AnalyzeMealRequest) (*AnalyzeMealResponse, error) {
	user, err := h.userService.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if !h.userService.CanAccessAdvancedFeatures(user) {
		return nil, status.Error(codes.PermissionDenied, "freemium users cannot access advanced features")
	}
	return nil, status.Error(codes.Unimplemented, "AnalyzeMeal unimplemented")
}
func (h *CoreHandler) GetMealPlan(ctx context.Context, req *GetMealPlanRequest) (*MealPlanResponse, error) {
	user, err := h.userService.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if !h.userService.CanAccessAdvancedFeatures(user) {
		return nil, status.Error(codes.PermissionDenied, "freemium users cannot access advanced features")
	}
	return nil, status.Error(codes.Unimplemented, "GetMealPlan unimplemented")
}
func (h *CoreHandler) GetWorkoutRecommendation(ctx context.Context, req *GetWorkoutRecommendationRequest) (*WorkoutPlanResponse, error) {
	user, err := h.userService.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if !h.userService.CanAccessAdvancedFeatures(user) {
		return nil, status.Error(codes.PermissionDenied, "freemium users cannot access advanced features")
	}

	equip := []domain.EquipmentType{}
	for _, e := range req.AvailableEquipment {
		equip = append(equip, domain.EquipmentType(e))
	}
	if len(equip) == 0 {
		equip = append(equip, domain.EquipmentNone)
	}

	constraints := services.WorkoutConstraints{
		DurationMinutes: int(req.DurationMinutes),
		Equipment:       equip,
		UserInjuries:    req.UserInjuries,
	}
	if constraints.DurationMinutes <= 0 {
		constraints.DurationMinutes = 30
	}

	plan, err := h.workoutService.GenerateWorkout(ctx, req.UserId, constraints)
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

// 7. GetCaloriesHistory
func (h *CoreHandler) GetCaloriesHistory(ctx context.Context, req *HistoryRequest) (*CaloriesHistoryResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	days := int(req.Days)
	if days <= 0 {
		days = 7
	}
	logs, err := h.activityService.GetCaloriesHistory(ctx, req.UserId, days)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get calories history")
	}
	entries := make([]*DailyLogEntry, len(logs))
	for i, log := range logs {
		entries[i] = &DailyLogEntry{
			Date:     log.Date.Format("2006-01-02"),
			Calories: log.TotalCalories,
			Protein:  log.TotalProtein,
			Carbs:    log.TotalCarbs,
			Fat:      log.TotalFat,
		}
	}
	return &CaloriesHistoryResponse{Entries: entries}, nil
}

// 8. GetWeightHistory - returns the user's current weight as a single point (no weight log table yet)
func (h *CoreHandler) GetWeightHistory(ctx context.Context, req *HistoryRequest) (*WeightHistoryResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	user, err := h.userService.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	// Since there is no weight log table, return the current weight as a single entry
	entries := []*WeightEntry{}
	if user.Weight > 0 {
		entries = append(entries, &WeightEntry{
			Date:   time.Now().Format("2006-01-02"),
			Weight: user.Weight,
		})
	}
	return &WeightHistoryResponse{Entries: entries}, nil
}

// --- Registration ---

func RegisterCoreService(s *grpc.Server, srv *CoreHandler) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "core.CoreService",
		HandlerType: (*CoreServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{MethodName: "GetUserProfile", Handler: _Core_GetUserProfile_Handler},
			{MethodName: "GetDailyStats", Handler: _Core_GetDailyStats_Handler},
			{MethodName: "UpdateBiometrics", Handler: _Core_UpdateBiometrics_Handler},
			{MethodName: "UpdateHealthProfile", Handler: _Core_UpdateHealthProfile_Handler},
			{MethodName: "LogNutrition", Handler: _Core_LogNutrition_Handler},
			{MethodName: "LogWorkout", Handler: _Core_LogWorkout_Handler},
			{MethodName: "UpdateFoodPreferences", Handler: _Core_UpdateFoodPreferences_Handler},
			{MethodName: "AnalyzeMeal", Handler: _Core_AnalyzeMeal_Handler},
			{MethodName: "GetMealPlan", Handler: _Core_GetMealPlan_Handler},
			{MethodName: "GetFoodPreferences", Handler: _Core_GetFoodPreferences_Handler},
			{MethodName: "GetWorkoutRecommendation", Handler: _Core_GetWorkoutRecommendation_Handler},
			{MethodName: "GetCaloriesHistory", Handler: _Core_GetCaloriesHistory_Handler},
			{MethodName: "GetWeightHistory", Handler: _Core_GetWeightHistory_Handler},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "core.proto",
	}, srv)
}

// --- Wrappers ---

func _Core_GetUserProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).GetUserProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/GetUserProfile"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).GetUserProfile(ctx, req.(*UserIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Repeating generic wrapper for brevity on others, just renaming types...
// I will implement a few critical ones.

func _Core_GetDailyStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).GetDailyStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/GetDailyStats"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).GetDailyStats(ctx, req.(*UserIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_GetFoodPreferences_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFoodPreferencesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).GetFoodPreferences(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/GetFoodPreferences"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).GetFoodPreferences(ctx, req.(*GetFoodPreferencesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_UpdateHealthProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateHealthProfileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).UpdateHealthProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/UpdateHealthProfile"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).UpdateHealthProfile(ctx, req.(*UpdateHealthProfileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_UpdateFoodPreferences_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateFoodPreferencesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).UpdateFoodPreferences(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/UpdateFoodPreferences"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).UpdateFoodPreferences(ctx, req.(*UpdateFoodPreferencesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_UpdateBiometrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BiometricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).UpdateBiometrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/UpdateBiometrics"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).UpdateBiometrics(ctx, req.(*BiometricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}
func _Core_LogNutrition_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NutritionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).LogNutrition(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/LogNutrition"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).LogNutrition(ctx, req.(*NutritionRequest))
	}
	return interceptor(ctx, in, info, handler)
}
func _Core_LogWorkout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkoutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).LogWorkout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/LogWorkout"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).LogWorkout(ctx, req.(*WorkoutRequest))
	}
	return interceptor(ctx, in, info, handler)
}
func _Core_AnalyzeMeal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AnalyzeMealRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).AnalyzeMeal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/AnalyzeMeal"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).AnalyzeMeal(ctx, req.(*AnalyzeMealRequest))
	}
	return interceptor(ctx, in, info, handler)
}
func _Core_GetMealPlan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMealPlanRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).GetMealPlan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/GetMealPlan"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).GetMealPlan(ctx, req.(*GetMealPlanRequest))
	}
	return interceptor(ctx, in, info, handler)
}
func _Core_GetWorkoutRecommendation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorkoutRecommendationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).GetWorkoutRecommendation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/GetWorkoutRecommendation"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).GetWorkoutRecommendation(ctx, req.(*GetWorkoutRecommendationRequest))
	}
	return interceptor(ctx, in, info, handler)
}
func _Core_GetCaloriesHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).GetCaloriesHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/GetCaloriesHistory"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).GetCaloriesHistory(ctx, req.(*HistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}
func _Core_GetWeightHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*CoreHandler).GetWeightHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/core.CoreService/GetWeightHistory"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*CoreHandler).GetWeightHistory(ctx, req.(*HistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}
