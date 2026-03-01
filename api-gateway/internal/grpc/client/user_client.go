package client

import (
    "context"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    
    taskv1 "api-gateway/internal/grpc/task/pb"  // используем существующий pb
)

type UserClient struct {
    client taskv1.TaskServiceClient
    conn   *grpc.ClientConn
}

func NewUserClient(addr string) (*UserClient, error) {
    conn, err := grpc.Dial(addr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock())
    if err != nil {
        return nil, err
    }

    client := taskv1.NewTaskServiceClient(conn)
    return &UserClient{
        client: client,
        conn:   conn,
    }, nil
}

func (c *UserClient) GetUserByUsername(ctx context.Context, username string) (*taskv1.User, error) {
    // TODO: добавить метод GetUserByUsername в task.proto
    resp, err := c.client.GetUserByUsername(ctx, &taskv1.GetUserByUsernameRequest{
        Username: username,
    })
    if err != nil {
        return nil, err
    }
    return resp.GetUser(), nil
}

func (c *UserClient) CreateUser(ctx context.Context, username, password string) (*taskv1.User, error) {
    // TODO: добавить метод CreateUser в task.proto
    resp, err := c.client.CreateUser(ctx, &taskv1.CreateUserRequest{
        Username: username,
        Password: password,
    })
    if err != nil {
        return nil, err
    }
    return resp.GetUser(), nil
}

func (c *UserClient) Close() error {
    return c.conn.Close()
}