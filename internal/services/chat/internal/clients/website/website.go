package website

import (
	"context"
	"fmt"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/website"
	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn         *grpc.ClientConn
	client       website.RoomServiceClient
	serviceToken string
}

func NewClient(address string, serviceToken string) (*Client, error) {
	if serviceToken == "" {
		return nil, errors.New("service token is required")
	}

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gRPC connection")
	}

	return &Client{
		conn:         conn,
		client:       website.NewRoomServiceClient(conn),
		serviceToken: serviceToken,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) createAuthContext(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", c.serviceToken),
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *Client) GetRoom(ctx context.Context, roomID uuid.UUID) (*entities.Room, error) {
	ctx = c.createAuthContext(ctx)
	resp, err := c.client.GetRoom(ctx, &website.GetRoomRequest{
		RoomId: roomID.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get room")
	}

	room, err := c.protoToRoom(resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert proto room")
	}

	return room, nil
}

func (c *Client) IsRoomOwner(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	ctx = c.createAuthContext(ctx)
	room, err := c.GetRoom(ctx, roomID)
	if err != nil {
		return false, errors.Wrap(err, "failed to get room")
	}

	return room.OwnerID == userID, nil
}

func (c *Client) RoomExists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	ctx = c.createAuthContext(ctx)
	resp, err := c.client.GetRoom(ctx, &website.GetRoomRequest{
		RoomId: roomID.String(),
	})

	if err != nil {
		return false, errors.Wrap(err, "failed to check room existence")
	}

	return resp != nil, nil
}

func (c *Client) GetOwnerRooms(ctx context.Context, ownerID uuid.UUID) ([]*entities.Room, error) {
	ctx = c.createAuthContext(ctx)
	resp, err := c.client.GetOwnerRooms(ctx, &website.GetOwnerRoomsRequest{
		OwnerId: ownerID.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get owner rooms")
	}

	rooms := make([]*entities.Room, 0, len(resp.Rooms))
	for _, protoRoom := range resp.Rooms {
		room, err := c.protoToRoom(protoRoom)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert proto room")
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (c *Client) SearchRooms(ctx context.Context, name string, limit, offset int) ([]*entities.Room, error) {
	ctx = c.createAuthContext(ctx)
	resp, err := c.client.SearchRooms(ctx, &website.SearchRoomsRequest{
		Name:   name,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to search rooms")
	}

	rooms := make([]*entities.Room, 0, len(resp.Rooms))
	for _, protoRoom := range resp.Rooms {
		room, err := c.protoToRoom(protoRoom)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert proto room")
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (c *Client) protoToRoom(protoRoom *website.Room) (*entities.Room, error) {
	roomID, err := uuid.Parse(protoRoom.Id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse room ID")
	}

	ownerID, err := uuid.Parse(protoRoom.OwnerId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse owner ID")
	}

	return &entities.Room{
		ID:        roomID,
		Name:      protoRoom.Name,
		OwnerID:   ownerID,
		CreatedAt: protoRoom.CreatedAt.AsTime(),
		UpdatedAt: protoRoom.UpdatedAt.AsTime(),
	}, nil
}
