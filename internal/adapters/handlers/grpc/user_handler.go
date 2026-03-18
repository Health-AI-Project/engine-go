package grpc

import (
	"context"
	"fmt"
	"time"

	"healthai/engine/internal/core/services"
	// pb "healthai/engine/proto/user" // generic import, assuming generation
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// To make this compilable without the generated code, I am defining aliases here.
// IN A REAL SCENARIO: Remove these and import the generated 'pb' package.
type UserProfileResponse struct {
	Id                 string   `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Email              string   `protobuf:"bytes,2,opt,name=email" json:"email,omitempty"`
	SubscriptionStatus string   `protobuf:"bytes,3,opt,name=subscription_status" json:"subscription_status,omitempty"`
	Weight             float64  `protobuf:"fixed64,4,opt,name=weight" json:"weight,omitempty"`
	Height             float64  `protobuf:"fixed64,5,opt,name=height" json:"height,omitempty"`
	IsPremium          bool     `protobuf:"varint,6,opt,name=is_premium" json:"is_premium,omitempty"`
	DateOfBirth        string   `protobuf:"bytes,7,opt,name=date_of_birth" json:"date_of_birth,omitempty"`
	Age                int32    `protobuf:"varint,8,opt,name=age" json:"age,omitempty"`
	Goals              []string `protobuf:"bytes,9,rep,name=goals" json:"goals,omitempty"`
	Allergies          []string `protobuf:"bytes,10,rep,name=allergies" json:"allergies,omitempty"`
}

func (m *UserProfileResponse) Reset()         { *m = UserProfileResponse{} }
func (m *UserProfileResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*UserProfileResponse) ProtoMessage()    {}

type UserIdRequest struct {
	UserId string `protobuf:"bytes,1,opt,name=user_id" json:"user_id,omitempty"`
}

func (m *UserIdRequest) Reset()         { *m = UserIdRequest{} }
func (m *UserIdRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*UserIdRequest) ProtoMessage()    {}

type BiometricsRequest struct {
	UserId string  `protobuf:"bytes,1,opt,name=user_id" json:"user_id,omitempty"`
	Weight float64 `protobuf:"fixed64,2,opt,name=weight" json:"weight,omitempty"`
	Height float64 `protobuf:"fixed64,3,opt,name=height" json:"height,omitempty"`
}

func (m *BiometricsRequest) Reset()         { *m = BiometricsRequest{} }
func (m *BiometricsRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*BiometricsRequest) ProtoMessage()    {}

type UpdateHealthProfileRequest struct {
	UserId      string   `protobuf:"bytes,1,opt,name=user_id" json:"user_id,omitempty"`
	DateOfBirth string   `protobuf:"bytes,2,opt,name=date_of_birth" json:"date_of_birth,omitempty"`
	Goals       []string `protobuf:"bytes,3,rep,name=goals" json:"goals,omitempty"`
	Allergies   []string `protobuf:"bytes,4,rep,name=allergies" json:"allergies,omitempty"`
	Weight      float64  `protobuf:"fixed64,5,opt,name=weight" json:"weight,omitempty"`
}

func (m *UpdateHealthProfileRequest) Reset()         { *m = UpdateHealthProfileRequest{} }
func (m *UpdateHealthProfileRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*UpdateHealthProfileRequest) ProtoMessage()    {}

type Empty struct{}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return "{}" }
func (*Empty) ProtoMessage()    {}

// Manual definitions for Feature B
type NutritionRequest struct {
	UserId   string  `protobuf:"bytes,1,opt,name=user_id" json:"user_id,omitempty"`
	Calories float64 `protobuf:"fixed64,2,opt,name=calories" json:"calories,omitempty"`
	Protein  float64 `protobuf:"fixed64,3,opt,name=protein" json:"protein,omitempty"`
	Carbs    float64 `protobuf:"fixed64,4,opt,name=carbs" json:"carbs,omitempty"`
	Fat      float64 `protobuf:"fixed64,5,opt,name=fat" json:"fat,omitempty"`
}

func (m *NutritionRequest) Reset()         { *m = NutritionRequest{} }
func (m *NutritionRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*NutritionRequest) ProtoMessage()    {}

type WorkoutRequest struct {
	UserId          string  `protobuf:"bytes,1,opt,name=user_id" json:"user_id,omitempty"`
	Type            string  `protobuf:"bytes,2,opt,name=type" json:"type,omitempty"`
	DurationMinutes int32   `protobuf:"varint,3,opt,name=duration_minutes" json:"duration_minutes,omitempty"`
	CaloriesBurned  float64 `protobuf:"fixed64,4,opt,name=calories_burned" json:"calories_burned,omitempty"`
}

func (m *WorkoutRequest) Reset()         { *m = WorkoutRequest{} }
func (m *WorkoutRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*WorkoutRequest) ProtoMessage()    {}

type DailyStatsResponse struct {
	Calories      float64 `protobuf:"fixed64,1,opt,name=calories" json:"calories,omitempty"`
	Protein       float64 `protobuf:"fixed64,2,opt,name=protein" json:"protein,omitempty"`
	Carbs         float64 `protobuf:"fixed64,3,opt,name=carbs" json:"carbs,omitempty"`
	Fat           float64 `protobuf:"fixed64,4,opt,name=fat" json:"fat,omitempty"`
	WorkoutsCount int32   `protobuf:"varint,5,opt,name=workouts_count" json:"workouts_count,omitempty"`
}

func (m *DailyStatsResponse) Reset()         { *m = DailyStatsResponse{} }
func (m *DailyStatsResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*DailyStatsResponse) ProtoMessage()    {}

type Ack struct {
	Success bool   `protobuf:"varint,1,opt,name=success" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
}

func (m *Ack) Reset()         { *m = Ack{} }
func (m *Ack) String() string { return fmt.Sprintf("%+v", *m) }
func (*Ack) ProtoMessage()    {}

// UserServiceServer is the interface that defines the gRPC service methods.
type UserServiceServer interface {
	GetUserProfile(context.Context, *UserIdRequest) (*UserProfileResponse, error)
	GetDailyStats(context.Context, *UserIdRequest) (*DailyStatsResponse, error)
	UpdateBiometrics(context.Context, *BiometricsRequest) (*Empty, error)
	UpdateHealthProfile(context.Context, *UpdateHealthProfileRequest) (*Ack, error)
	LogNutrition(context.Context, *NutritionRequest) (*Ack, error)
	LogWorkout(context.Context, *WorkoutRequest) (*Ack, error)
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

func (h *UserHandler) GetService() *services.UserService {
	return h.service
}

func (h *UserHandler) GetActivityService() *services.ActivityService {
	return h.activityService
}

func (h *UserHandler) GetUserProfile(ctx context.Context, req *UserIdRequest) (*UserProfileResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := h.service.GetUser(ctx, req.UserId)
	if err != nil {
		fmt.Printf("[GRPC] GetUserProfile Error: %v for User %s\n", err, req.UserId)
		return nil, status.Error(codes.NotFound, "user not found")
	}

	fmt.Printf("[GRPC] Fetched User: %+v\n", user)
	if user.HealthProfile != nil {
		fmt.Printf("[GRPC] Health Profile: %+v\n", user.HealthProfile)
	} else {
		fmt.Println("[GRPC] No Health Profile found for user")
	}

	isPremium := h.service.CanAccessAdvancedFeatures(user)

	var dobStr string
	var age int32
	var goals []string
	var allergies []string

	if user.DateOfBirth != nil {
		dobStr = user.DateOfBirth.Format("2006-01-02")
		// Calculate age
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

func (h *UserHandler) UpdateHealthProfile(ctx context.Context, req *UpdateHealthProfileRequest) (*Ack, error) {
	fmt.Printf("[GRPC] Received UpdateHealthProfile for User: %s, Weight: %f, Goals: %v\n", req.UserId, req.Weight, req.Goals)
	if req.UserId == "" {
		return &Ack{Success: false, Message: "user_id is required"}, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var dob *time.Time
	if req.DateOfBirth != "" {
		t, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return &Ack{Success: false, Message: "invalid date_of_birth format, use YYYY-MM-DD"}, status.Error(codes.InvalidArgument, "invalid date_of_birth format")
		}
		dob = &t
	}

	err := h.service.UpdateHealthProfile(ctx, req.UserId, dob, req.Goals, req.Allergies, req.Weight)
	if err != nil {
		fmt.Printf("[GRPC] UpdateHealthProfile ERROR: %v\n", err)
		return &Ack{Success: false, Message: err.Error()}, status.Error(codes.Internal, "failed to update health profile")
	}

	return &Ack{Success: true, Message: "health profile updated"}, nil
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

// --- Manual gRPC Registration (since we don't have protoc) ---

func RegisterUserService(s *grpc.Server, srv *UserHandler) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "user.UserService",
		HandlerType: (*UserServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "GetUserProfile",
				Handler:    _UserService_GetUserProfile_Handler,
			},
			{
				MethodName: "GetDailyStats",
				Handler:    _UserService_GetDailyStats_Handler,
			},
			{
				MethodName: "UpdateBiometrics",
				Handler:    _UserService_UpdateBiometrics_Handler,
			},
			{
				MethodName: "UpdateHealthProfile",
				Handler:    _UserService_UpdateHealthProfile_Handler,
			},
			{
				MethodName: "LogNutrition",
				Handler:    _UserService_LogNutrition_Handler,
			},
			{
				MethodName: "LogWorkout",
				Handler:    _UserService_LogWorkout_Handler,
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "user.proto",
	}, srv)
}

func _UserService_GetUserProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*UserHandler).GetUserProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.UserService/GetUserProfile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*UserHandler).GetUserProfile(ctx, req.(*UserIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_GetDailyStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*UserHandler).GetDailyStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.UserService/GetDailyStats",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*UserHandler).GetDailyStats(ctx, req.(*UserIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_UpdateBiometrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BiometricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*UserHandler).UpdateBiometrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.UserService/UpdateBiometrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*UserHandler).UpdateBiometrics(ctx, req.(*BiometricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_UpdateHealthProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateHealthProfileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*UserHandler).UpdateHealthProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.UserService/UpdateHealthProfile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*UserHandler).UpdateHealthProfile(ctx, req.(*UpdateHealthProfileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_LogNutrition_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NutritionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*UserHandler).LogNutrition(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.UserService/LogNutrition",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*UserHandler).LogNutrition(ctx, req.(*NutritionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_LogWorkout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkoutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*UserHandler).LogWorkout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.UserService/LogWorkout",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*UserHandler).LogWorkout(ctx, req.(*WorkoutRequest))
	}
	return interceptor(ctx, in, info, handler)
}
