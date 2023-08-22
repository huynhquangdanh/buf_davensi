//nolint:all
package main

import (
	"context"
	"log"
	"net/http"

	"davensi.com/core/gen/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// pbUsers "davensi.com/core/gen/uoms"
	// pbUsersConnect "davensi.com/core/gen/uoms/uomsconnect"
	pbUsers "davensi.com/core/gen/users"
	pbUsersConnect "davensi.com/core/gen/users/usersconnect"
	"github.com/bufbuild/connect-go"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to create gRPC communication channel: %v", err)
	}
	defer conn.Close()
	client := pbUsersConnect.NewServiceClient(http.DefaultClient, "http://localhost:8080/")

	// Create Users
	newScreenName := "Mr Fahrenheit"
	newType := pbUsers.Type_TYPE_INTERNAL
	newStatus := common.Status_STATUS_ACTIVE
	req1 := &pbUsers.CreateRequest{
		Login:      "sopheap.lao@gmail.com",
		Type:       &newType,
		ScreenName: &newScreenName,
		Status:     &newStatus,
	}
	res1, err := client.Create(context.Background(), &connect.Request[pbUsers.CreateRequest]{Msg: req1})
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("CreateUser: %v", res1.Msg)

	updatedLogin := "SOPHEAP_LAO"
	updatedType := pbUsers.Type_TYPE_CUSTOMER
	updatedStatus := common.Status_STATUS_DEPRECATED
	req2 := &pbUsers.UpdateRequest{
		Select: &pbUsers.UpdateRequest_ByLogin{
			ByLogin: "sopheap.lao@gmail.com",
		},
		Login:      &updatedLogin,
		Type:       &updatedType,
		ScreenName: nil,
		Avatar:     nil,
		Status:     &updatedStatus,
	}
	res2, err := client.Update(context.Background(), &connect.Request[pbUsers.UpdateRequest]{Msg: req2})
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("UpdateUser: %v", res2.Msg)
	loginSearch := "a"
	req3 := &pbUsers.GetListRequest{
		Login: &loginSearch,
		Status: &common.StatusList{
			List: []common.Status{
				common.Status_STATUS_ACTIVE,
				common.Status_STATUS_DEPRECATED,
			},
		},
	}
	res3, err := client.GetList(context.Background(), &connect.Request[pbUsers.GetListRequest]{Msg: req3})
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("ListUser: %v", res3)
}
