
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	
	"gosprints/internal/grpc/task/pb"
)

func main() {
	fmt.Println("=== Тестирование gRPC подключения ===")
	
	// ждем запуска сервера
	time.Sleep(3 * time.Second)
	
	// Подключаемся к gRPC серверу
	conn, err := grpc.Dial("localhost:50051", 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second))
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()
	
	fmt.Println("✓ Подключение к gRPC серверу успешно")
	
	client := pb.NewTaskServiceClient(conn)
	
	// Проверяем методы
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// попробуем простой запрос
	fmt.Println("\n=== Тестирование ListTasks ===")
	resp, err := client.ListTasks(ctx, &pb.ListTasksRequest{
		Page:     1,
		PageSize: 10,
	})
	
	if err != nil {
		log.Printf("Ошибка ListTasks: %v", err)
	} else {
		fmt.Printf("✓ ListTasks успешно. Найдено задач: %d\n", resp.GetTotal())
	}
	
	fmt.Println("\n=== Тестирование CreateTask ===")
	createResp, err := client.CreateTask(ctx, &pb.CreateTaskRequest{
		Text:   "Тестовая задача gRPC",
		UserId: "test-user",
	})
	
	if err != nil {
		log.Printf("Ошибка CreateTask: %v", err)
	} else {
		fmt.Printf("✓ CreateTask успешно. ID задачи: %s\n", createResp.GetTask().GetId())
	}
	
	fmt.Println("\n=== Все тесты завершены ===")
}