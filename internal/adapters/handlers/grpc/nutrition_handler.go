package grpc

import (
	"context"
	"fmt"
	"time"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/services"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- Manual Protobuf Struct Definitions (Shadowing helper) ---

type UpdateFoodPreferencesRequest struct {
	UserId              string   `json:"user_id"`
	DietType            string   `json:"diet_type"`
	Allergies           []string `json:"allergies"`
	DislikedIngredients []string `json:"disliked_ingredients"`
}

func (m *UpdateFoodPreferencesRequest) Reset()         { *m = UpdateFoodPreferencesRequest{} }
func (m *UpdateFoodPreferencesRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*UpdateFoodPreferencesRequest) ProtoMessage()    {}

type AnalyzeMealRequest struct {
	UserId string               `json:"user_id"`
	Meal   *ProtoMealSuggestion `json:"meal"`
}

func (m *AnalyzeMealRequest) Reset()         { *m = AnalyzeMealRequest{} }
func (m *AnalyzeMealRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*AnalyzeMealRequest) ProtoMessage()    {}

type ProtoMealSuggestion struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Calories    float64  `json:"calories"`
	Protein     float64  `json:"protein"`
	Carbs       float64  `json:"carbs"`
	Fat         float64  `json:"fat"`
	Ingredients []string `json:"ingredients"`
	DietTags    []string `json:"diet_tags"`
}

func (m *ProtoMealSuggestion) Reset()         { *m = ProtoMealSuggestion{} }
func (m *ProtoMealSuggestion) String() string { return fmt.Sprintf("%+v", *m) }
func (*ProtoMealSuggestion) ProtoMessage()    {}

type AnalyzeMealResponse struct {
	IsBalanced     bool     `json:"is_balanced"`
	Warnings       []string `json:"warnings"`
	CriticalAlerts []string `json:"critical_alerts"`
}

func (m *AnalyzeMealResponse) Reset()         { *m = AnalyzeMealResponse{} }
func (m *AnalyzeMealResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*AnalyzeMealResponse) ProtoMessage()    {}

type GetMealPlanRequest struct {
	UserId string `json:"user_id"`
	Date   string `json:"date"` // YYYY-MM-DD
}

func (m *GetMealPlanRequest) Reset()         { *m = GetMealPlanRequest{} }
func (m *GetMealPlanRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*GetMealPlanRequest) ProtoMessage()    {}

type MealPlanResponse struct {
	PlanId         string          `json:"plan_id"`
	Date           string          `json:"date"`
	TargetCalories float64         `json:"target_calories"`
	Meals          []*MealPlanItem `json:"meals"`
}

func (m *MealPlanResponse) Reset()         { *m = MealPlanResponse{} }
func (m *MealPlanResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*MealPlanResponse) ProtoMessage()    {}

type MealPlanItem struct {
	MealType string               `json:"meal_type"`
	Meal     *ProtoMealSuggestion `json:"meal"`
	IsEaten  bool                 `json:"is_eaten"`
}

type GetFoodPreferencesRequest struct {
	UserId string `json:"user_id"`
}

func (m *GetFoodPreferencesRequest) Reset()         { *m = GetFoodPreferencesRequest{} }
func (m *GetFoodPreferencesRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*GetFoodPreferencesRequest) ProtoMessage()    {}

type FoodPreferencesResponse struct {
	Allergies           []string `json:"allergies"`
	DietType            string   `json:"diet_type"`
	DislikedIngredients []string `json:"disliked_ingredients"`
}

func (m *FoodPreferencesResponse) Reset()         { *m = FoodPreferencesResponse{} }
func (m *FoodPreferencesResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*FoodPreferencesResponse) ProtoMessage()    {}

// --- Handler Logic ---

type NutritionHandler struct {
	service *services.NutritionService
}

func NewNutritionHandler(service *services.NutritionService) *NutritionHandler {
	return &NutritionHandler{service: service}
}

func (h *NutritionHandler) GetService() *services.NutritionService {
	return h.service
}

func (h *NutritionHandler) UpdateFoodPreferences(ctx context.Context, req *UpdateFoodPreferencesRequest) (*Ack, error) {
	// Not implemented in Service yet, but repo has Upsert. Simple pass-through or todo.
	return &Ack{Success: true, Message: "Preferences updated (mock)"}, nil
}

func (h *NutritionHandler) GetFoodPreferences(ctx context.Context, req *GetFoodPreferencesRequest) (*FoodPreferencesResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	prefs, err := h.service.GetFoodPreferences(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if prefs == nil {
		// Return empty default if not set
		return &FoodPreferencesResponse{
			Allergies:           []string{},
			DietType:            "NONE",
			DislikedIngredients: []string{},
		}, nil
	}

	return &FoodPreferencesResponse{
		Allergies:           prefs.Allergies,
		DietType:            string(prefs.DietType),
		DislikedIngredients: prefs.DislikedIngredients,
	}, nil
}

func (h *NutritionHandler) AnalyzeMeal(ctx context.Context, req *AnalyzeMealRequest) (*AnalyzeMealResponse, error) {
	if req.UserId == "" || req.Meal == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid arguments")
	}

	domainMeal := domain.MealSuggestion{
		ID:          req.Meal.Id,
		Name:        req.Meal.Name,
		Calories:    req.Meal.Calories,
		Protein:     req.Meal.Protein,
		Carbs:       req.Meal.Carbs,
		Fat:         req.Meal.Fat,
		Ingredients: req.Meal.Ingredients,
		DietTags:    req.Meal.DietTags,
	}

	report, err := h.service.AnalyzeMeal(ctx, req.UserId, domainMeal)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &AnalyzeMealResponse{
		IsBalanced:     report.IsBalanced,
		Warnings:       report.Warnings,
		CriticalAlerts: report.CriticalAlerts,
	}, nil
}

func (h *NutritionHandler) GetMealPlan(ctx context.Context, req *GetMealPlanRequest) (*MealPlanResponse, error) {
	targetDate := time.Now()
	if req.Date != "" {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err == nil {
			targetDate = parsed
		}
	}

	plan, err := h.service.GenerateDailyPlan(ctx, req.UserId, targetDate)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &MealPlanResponse{
		PlanId:         plan.ID,
		Date:           plan.Date.Format("2006-01-02"),
		TargetCalories: plan.TargetCalories,
		Meals:          []*MealPlanItem{},
	}

	for _, item := range plan.Meals {
		resp.Meals = append(resp.Meals, &MealPlanItem{
			MealType: string(item.MealType),
			IsEaten:  item.IsEatened,
			Meal: &ProtoMealSuggestion{
				Id:          item.MealSuggestion.ID,
				Name:        item.MealSuggestion.Name,
				Calories:    item.MealSuggestion.Calories,
				Protein:     item.MealSuggestion.Protein,
				Carbs:       item.MealSuggestion.Carbs,
				Fat:         item.MealSuggestion.Fat,
				Ingredients: item.MealSuggestion.Ingredients,
				DietTags:    item.MealSuggestion.DietTags,
			},
		})
	}

	return resp, nil
}

// --- Registration ---

type NutritionServiceServer interface {
	UpdateFoodPreferences(context.Context, *UpdateFoodPreferencesRequest) (*Ack, error)
	AnalyzeMeal(context.Context, *AnalyzeMealRequest) (*AnalyzeMealResponse, error)
	GetMealPlan(context.Context, *GetMealPlanRequest) (*MealPlanResponse, error)
	GetFoodPreferences(context.Context, *GetFoodPreferencesRequest) (*FoodPreferencesResponse, error)
}

func RegisterNutritionService(s *grpc.Server, srv *NutritionHandler) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "nutrition.NutritionService",
		HandlerType: (*NutritionServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{MethodName: "UpdateFoodPreferences", Handler: _NutritionService_UpdateFoodPreferences_Handler},
			{MethodName: "AnalyzeMeal", Handler: _NutritionService_AnalyzeMeal_Handler},
			{MethodName: "GetMealPlan", Handler: _NutritionService_GetMealPlan_Handler},
			{MethodName: "GetFoodPreferences", Handler: _NutritionService_GetFoodPreferences_Handler},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "nutrition.proto",
	}, srv)
}

func _NutritionService_UpdateFoodPreferences_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateFoodPreferencesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*NutritionHandler).UpdateFoodPreferences(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nutrition.NutritionService/UpdateFoodPreferences"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*NutritionHandler).UpdateFoodPreferences(ctx, req.(*UpdateFoodPreferencesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NutritionService_AnalyzeMeal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AnalyzeMealRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*NutritionHandler).AnalyzeMeal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nutrition.NutritionService/AnalyzeMeal"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*NutritionHandler).AnalyzeMeal(ctx, req.(*AnalyzeMealRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NutritionService_GetMealPlan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMealPlanRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*NutritionHandler).GetMealPlan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nutrition.NutritionService/GetMealPlan"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*NutritionHandler).GetMealPlan(ctx, req.(*GetMealPlanRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NutritionService_GetFoodPreferences_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFoodPreferencesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*NutritionHandler).GetFoodPreferences(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nutrition.NutritionService/GetFoodPreferences"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*NutritionHandler).GetFoodPreferences(ctx, req.(*GetFoodPreferencesRequest))
	}
	return interceptor(ctx, in, info, handler)
}
