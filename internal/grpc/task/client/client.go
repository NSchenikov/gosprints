
package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gosprints/internal/grpc/task/pb"
)

type TaskClient struct {
	client pb.TaskServiceClient
	conn   *grpc.ClientConn
}

func NewTaskClient(addr string) (*TaskClient, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	client := pb.NewTaskServiceClient(conn)
	return &TaskClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *TaskClient) Close() error {
	return c.conn.Close()
}

func (c *TaskClient) CreateTask(ctx context.Context, text, userID string) (*pb.Task, error) {
	resp, err := c.client.CreateTask(ctx, &pb.CreateTaskRequest{
		Text:   text,
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetTask(), nil
}

func (c *TaskClient) GetTask(ctx context.Context, id int32) (*pb.Task, error) {
	resp, err := c.client.GetTask(ctx, &pb.GetTaskRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.GetTask(), nil
}

func (c *TaskClient) ListTasks(ctx context.Context, userID, status string, page, pageSize int32) ([]*pb.Task, int32, error) {
	resp, err := c.client.ListTasks(ctx, &pb.ListTasksRequest{
		UserId:   userID,
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	return resp.GetTasks(), resp.GetTotal(), nil
}

func (c *TaskClient) UpdateTask(ctx context.Context, id int32, text, status string) (*pb.Task, error) {
	resp, err := c.client.UpdateTask(ctx, &pb.UpdateTaskRequest{
		Id:     id,
		Text:   text,
		Status: status,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetTask(), nil
}

func (c *TaskClient) DeleteTask(ctx context.Context, id int32) (bool, error) {
	resp, err := c.client.DeleteTask(ctx, &pb.DeleteTaskRequest{Id: id})
	if err != nil {
		return false, err
	}
	return resp.GetSuccess(), nil
}

func (c *TaskClient) SearchTasks(ctx context.Context, query, userID string, page, pageSize int32) ([]*pb.Task, int32, error) {
    resp, err := c.client.SearchTasks(ctx, &pb.SearchTasksRequest{
        Query:    query,
        UserId:   userID,
        Page:     page,
        PageSize: pageSize,
    })
    if err != nil {
        return nil, 0, err
    }
    return resp.GetTasks(), resp.GetTotal(), nil
}