package server

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"gosprints/internal/grpc/task/pb"
	"gosprints/internal/models"
	"gosprints/internal/repositories"
)

type TaskServer struct {
	pb.UnimplementedTaskServiceServer
	repo repositories.TaskRepository
}

func NewTaskServer(repo repositories.TaskRepository) *TaskServer {
	return &TaskServer{repo: repo}
}

func (s *TaskServer) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	task := &models.Task{
		Text:   req.GetText(),
		UserID: req.GetUserId(),
		Status: "pending",
	}

	createdTask, err := s.repo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	return &pb.CreateTaskResponse{
		Task: taskToProto(createdTask),
	}, nil
}

func (s *TaskServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	task, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.GetTaskResponse{
		Task: taskToProto(task),
	}, nil
}

func (s *TaskServer) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	filter := repositories.TaskFilter{
		UserID: req.GetUserId(),
		Status: req.GetStatus(),
		Page:   int(req.GetPage()),
		Limit:  int(req.GetPageSize()),
	}

	tasks, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	var protoTasks []*pb.Task
	for _, task := range tasks {
		protoTasks = append(protoTasks, taskToProto(&task))
	}

	return &pb.ListTasksResponse{
		Tasks:    protoTasks,
		Total:    int32(total),
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

func (s *TaskServer) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.UpdateTaskResponse, error) {
	task := &models.Task{
		ID:     req.GetId(),
		Text:   req.GetText(),
		Status: req.GetStatus(),
	}

	updatedTask, err := s.repo.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateTaskResponse{
		Task: taskToProto(updatedTask),
	}, nil
}

func (s *TaskServer) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	err := s.repo.Delete(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.DeleteTaskResponse{
		Success: true,
	}, nil
}

func (s *TaskServer) SearchTasks(ctx context.Context, req *pb.SearchTasksRequest) (*pb.SearchTasksResponse, error) {
	return &pb.SearchTasksResponse{
		Tasks:    []*pb.Task{},
		Total:    0,
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// Вспомогательная функция для преобразования модели в proto
func taskToProto(task *models.Task) *pb.Task {
	protoTask := &pb.Task{
		Id:        task.ID,
		Text:      task.Text,
		Status:    task.Status,
		UserId:    task.UserID,
		CreatedAt: timestamppb.New(task.CreatedAt),
	}

	if task.StartedAt != nil {
		protoTask.StartedAt = timestamppb.New(*task.StartedAt)
	}

	if task.EndedAt != nil {
		protoTask.EndedAt = timestamppb.New(*task.EndedAt)
	}

	return protoTask
}

// Запуск gRPC сервера
func StartServer(repo repositories.TaskRepository, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterTaskServiceServer(s, NewTaskServer(repo))
	reflection.Register(s)

	log.Printf("gRPC Task Service запущен на порту %s", port)
	return s.Serve(lis)
}