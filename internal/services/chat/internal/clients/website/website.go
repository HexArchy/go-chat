package website

import (
	"context"
	"sync"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/website"
	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	logger       *zap.Logger
	conn         *grpc.ClientConn
	client       website.RoomServiceClient
	serviceToken string
	mutex        sync.RWMutex
}

func NewClient(logger *zap.Logger, address string, serviceToken string) (*Client, error) {
	if serviceToken == "" {
		return nil, errors.New("service token is required")
	}

	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to website service")
	}

	return &Client{
		logger:       logger,
		conn:         conn,
		client:       website.NewRoomServiceClient(conn),
		serviceToken: serviceToken,
		mutex:        sync.RWMutex{},
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) createServiceContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"authorization": "Bearer " + c.serviceToken,
	}))
}

func (c *Client) GetRoom(ctx context.Context, roomID uuid.UUID) (*entities.Room, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ctx = c.createServiceContext(ctx)

	resp, err := c.client.GetRoom(ctx, &website.GetRoomRequest{
		RoomId: roomID.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get room")
	}

	return c.protoToRoom(resp)
}

func (c *Client) IsRoomOwner(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	room, err := c.GetRoom(ctx, roomID)
	if err != nil {
		return false, errors.Wrap(err, "failed to get room")
	}

	return room.OwnerID == userID, nil
}

func (c *Client) RoomExists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	_, err := c.GetRoom(ctx, roomID)
	if err != nil {
		return false, nil // Room not found or other error.
	}
	return true, nil
}

func (c *Client) GetOwnerRooms(ctx context.Context, ownerID uuid.UUID) ([]*entities.Room, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ctx = c.createServiceContext(ctx)

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
			return nil, errors.Wrap(err, "failed to convert room")
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (c *Client) SearchRooms(ctx context.Context, name string, limit, offset int) ([]*entities.Room, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ctx = c.createServiceContext(ctx)

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
			return nil, errors.Wrap(err, "failed to convert room")
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (c *Client) protoToRoom(protoRoom *website.Room) (*entities.Room, error) {
	roomID, err := uuid.Parse(protoRoom.Id)
	if err != nil {
		return nil, errors.Wrap(err, "invalid room ID format")
	}

	ownerID, err := uuid.Parse(protoRoom.OwnerId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid owner ID format")
	}

	return &entities.Room{
		ID:        roomID,
		Name:      protoRoom.Name,
		OwnerID:   ownerID,
		CreatedAt: protoRoom.CreatedAt.AsTime(),
		UpdatedAt: protoRoom.UpdatedAt.AsTime(),
	}, nil
}
