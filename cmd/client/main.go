package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// -- Manual Type Definitions matching Server --
type NutritionRequest struct {
	UserId   string
	Calories float64
	Protein  float64
	Carbs    float64
	Fat      float64
}

type Ack struct {
	Success bool
	Message string
}

type UserIdRequest struct {
	UserId string
}

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

type UpdateHealthProfileRequest struct {
	UserId      string   `protobuf:"bytes,1,opt,name=user_id" json:"user_id,omitempty"`
	DateOfBirth string   `protobuf:"bytes,2,opt,name=date_of_birth" json:"date_of_birth,omitempty"`
	Goals       []string `protobuf:"bytes,3,rep,name=goals" json:"goals,omitempty"`
	Allergies   []string `protobuf:"bytes,4,rep,name=allergies" json:"allergies,omitempty"`
	Weight      float64  `protobuf:"fixed64,5,opt,name=weight" json:"weight,omitempty"`
}

// -- Client Implementation --

func main() {
	fmt.Println("Starting Demo Client...")

	// 1. Connect to Server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Since we don't have generated client code, we use Invoke directly
	// This is a bit advanced but standard for dynamic clients without gen code.
	// Actually, easier to just define the method name string.

	// Simulation User ID (Must exist in DB ideally, but for now we trust the flow)
	// PRO TIP: Run 'psql' to insert a user if "User Not Found" error occurs.
	userID := "demo-user-123"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 2. Call LogNutrition
	fmt.Println("\n[1] Logging Nutrition (500 kcal)...")
	nutritionReq := &NutritionRequest{
		UserId:   userID,
		Calories: 500,
		Protein:  30,
		Carbs:    50,
		Fat:      20,
	}
	ack := &Ack{}

	// NOTE: Because we don't have the generated Client interface, we can't easily call methods without using the invoke method
	// accessible via the connection. But `Invoke` takes method name string.
	// Method path: /user.UserService/LogNutrition

	err = conn.Invoke(ctx, "/user.UserService/LogNutrition", nutritionReq, ack)
	if err != nil {
		fmt.Printf("Error Logging Nutrition: %v\n(Did you start the server? Is DB Up?)\n", err)
	} else {
		fmt.Printf("Success: %v (Msg: %s)\n", ack.Success, ack.Message)
	}

	// 2.5 Call UpdateHealthProfile
	fmt.Println("\n[1.5] Updating Health Profile (Weight: 82.5, DOB: 1990-05-15)...")
	healthReq := &UpdateHealthProfileRequest{
		UserId:      userID,
		DateOfBirth: "1990-05-15",
		Goals:       []string{"Marathon", "Better Sleep"},
		Allergies:   []string{"Peanuts"},
		Weight:      82.5,
	}
	healthAck := &Ack{}
	err = conn.Invoke(ctx, "/user.UserService/UpdateHealthProfile", healthReq, healthAck)
	if err != nil {
		fmt.Printf("Error Updating Health Profile: %v\n", err)
	} else {
		fmt.Printf("Success: %v (Msg: %s)\n", healthAck.Success, healthAck.Message)
	}

	// 3. Call GetUserProfile (Might fail if user doesn't exist in 'users' table)
	fmt.Println("\n[2] Getting User Profile...")
	profileReq := &UserIdRequest{UserId: userID}
	profileRes := &UserProfileResponse{}

	err = conn.Invoke(ctx, "/user.UserService/GetUserProfile", profileReq, profileRes)
	if err != nil {
		fmt.Printf("Fetch Profile Failed: %v\n(This is expected if user 'demo-user-123' is not in the 'users' table yet.)\n", err)
		fmt.Println("To fix: Insert a user into DB or use an existing ID.")
	} else {
		fmt.Printf("User Profile: %+v\n", profileRes)
	}
}
