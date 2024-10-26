package website

import (
	"context"
	"fmt"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/website"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	client website.RoomServiceClient
	conn   *grpc.ClientConn
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gRPC connection")
	}

	return &Client{
		client: website.NewRoomServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) CreateRoom(ctx context.Context, token string, name, ownerID string) (*entities.Room, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := c.client.CreateRoom(ctx, &website.CreateRoomRequest{
		Name:    name,
		OwnerId: ownerID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create room")
	}

	return protoToRoom(resp.Room), nil
}

func (c *Client) GetRoom(ctx context.Context, token string, roomID string) (*entities.Room, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := c.client.GetRoom(ctx, &website.GetRoomRequest{
		RoomId: roomID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get room")
	}

	return protoToRoom(resp), nil
}

func (c *Client) GetOwnerRooms(ctx context.Context, token string, ownerID string) ([]*entities.Room, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.client.GetOwnerRooms(ctx, &website.GetOwnerRoomsRequest{
		OwnerId: ownerID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get owner rooms")
	}

	rooms := make([]*entities.Room, len(resp.Rooms))
	for i, r := range resp.Rooms {
		rooms[i] = protoToRoom(r)
	}
	return rooms, nil
}

func (c *Client) SearchRooms(ctx context.Context, token string, query string, limit, offset int32) ([]*entities.Room, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := c.client.SearchRooms(ctx, &website.SearchRoomsRequest{
		Name:   query,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to search rooms")
	}

	rooms := make([]*entities.Room, len(resp.Rooms))
	for i, r := range resp.Rooms {
		rooms[i] = protoToRoom(r)
	}
	return rooms, nil
}

func (c *Client) DeleteRoom(ctx context.Context, token string, roomID, ownerID string) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))
	_, err := c.client.DeleteRoom(ctx, &website.DeleteRoomRequest{
		RoomId:  roomID,
		OwnerId: ownerID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete room")
	}
	return nil
}

func protoToRoom(r *website.Room) *entities.Room {
	roomID, err := uuid.Parse(r.Id)
	if err != nil {
		return nil
	}

	return &entities.Room{
		ID:        roomID,
		Name:      r.Name,
		OwnerID:   r.OwnerId,
		CreatedAt: r.CreatedAt.AsTime(),
		UpdatedAt: r.UpdatedAt.AsTime(),
	}
}
