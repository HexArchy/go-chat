package website

import (
	"context"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/website"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/shared"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Client encapsulates the gRPC client for the RoomService.
type Client struct {
	logger    *zap.Logger
	client    website.RoomServiceClient
	conn      *grpc.ClientConn
	retryConf *shared.RetryConfig
}

// NewClient initializes a new Website Client.
// It establishes a gRPC connection to the website service with the provided interceptors.
func NewClient(logger *zap.Logger, address string, authInterceptor *shared.AuthInterceptor, retryConf *shared.RetryConfig) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInterceptor.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial website service")
	}

	return &Client{
		logger:    logger,
		client:    website.NewRoomServiceClient(conn),
		conn:      conn,
		retryConf: retryConf,
	}, nil
}

// Close gracefully closes the gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// CreateRoom creates a new chat room.
// It uses retry logic to handle transient failures.
func (c *Client) CreateRoom(ctx context.Context, name string, ownerID uuid.UUID) (*entities.Room, error) {
	var room *entities.Room
	err := shared.RetryWithBackoff(ctx, c.logger, c.retryConf, func() error {
		resp, err := c.client.CreateRoom(ctx, &website.CreateRoomRequest{
			Name:    name,
			OwnerId: ownerID.String(),
		})
		if err != nil {
			return errors.Wrap(err, "CreateRoom RPC failed")
		}

		room, err = protoToRoom(resp.Room)
		if err != nil {
			return errors.Wrap(err, "failed to convert proto Room to entities.Room")
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return room, nil
}

// GetRoom retrieves a chat room by its ID.
// It uses retry logic to handle transient failures.
func (c *Client) GetRoom(ctx context.Context, roomID uuid.UUID) (*entities.Room, error) {
	var room *entities.Room
	err := shared.RetryWithBackoff(ctx, c.logger, c.retryConf, func() error {
		resp, err := c.client.GetRoom(ctx, &website.GetRoomRequest{
			RoomId: roomID.String(),
		})
		if err != nil {
			return errors.Wrap(err, "GetRoom RPC failed")
		}

		room, err = protoToRoom(resp)
		if err != nil {
			return errors.Wrap(err, "failed to convert proto Room to entities.Room")
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return room, nil
}

// GetOwnerRooms retrieves all chat rooms owned by a specific user.
// It uses retry logic to handle transient failures.
func (c *Client) GetOwnerRooms(ctx context.Context, ownerID uuid.UUID) ([]*entities.Room, error) {
	var rooms []*entities.Room
	err := shared.RetryWithBackoff(ctx, c.logger, c.retryConf, func() error {
		resp, err := c.client.GetOwnerRooms(ctx, &website.GetOwnerRoomsRequest{
			OwnerId: ownerID.String(),
		})
		if err != nil {
			return errors.Wrap(err, "GetOwnerRooms RPC failed")
		}

		for _, protoRoom := range resp.Rooms {
			room, err := protoToRoom(protoRoom)
			if err != nil {
				return errors.Wrap(err, "failed to convert proto Room to entities.Room")
			}
			rooms = append(rooms, room)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return rooms, nil
}

// GetAllRooms retrieves all chat rooms with pagination (limit and offset).
// It uses retry logic to handle transient failures.
func (c *Client) GetAllRooms(ctx context.Context, limit, offset int32) ([]*entities.Room, error) {
	var rooms []*entities.Room
	err := shared.RetryWithBackoff(ctx, c.logger, c.retryConf, func() error {
		resp, err := c.client.GetAllRooms(ctx, &website.GetAllRoomsRequest{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return errors.Wrap(err, "GetAllRooms RPC failed")
		}

		for _, protoRoom := range resp.Rooms {
			room, err := protoToRoom(protoRoom)
			if err != nil {
				return errors.Wrap(err, "failed to convert proto Room to entities.Room")
			}
			rooms = append(rooms, room)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return rooms, nil
}

// SearchRooms searches for chat rooms by name with pagination.
// It uses retry logic to handle transient failures.
func (c *Client) SearchRooms(ctx context.Context, name string, limit, offset int32) ([]*entities.Room, error) {
	var rooms []*entities.Room
	err := shared.RetryWithBackoff(ctx, c.logger, c.retryConf, func() error {
		resp, err := c.client.SearchRooms(ctx, &website.SearchRoomsRequest{
			Name:   name,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return errors.Wrap(err, "SearchRooms RPC failed")
		}

		for _, protoRoom := range resp.Rooms {
			room, err := protoToRoom(protoRoom)
			if err != nil {
				return errors.Wrap(err, "failed to convert proto Room to entities.Room")
			}
			rooms = append(rooms, room)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return rooms, nil
}

// DeleteRoom deletes a chat room by its ID.
// It uses retry logic to handle transient failures.
func (c *Client) DeleteRoom(ctx context.Context, roomID, ownerID uuid.UUID) error {
	return shared.RetryWithBackoff(ctx, c.logger, c.retryConf, func() error {
		_, err := c.client.DeleteRoom(ctx, &website.DeleteRoomRequest{
			RoomId:  roomID.String(),
			OwnerId: ownerID.String(),
		})
		if err != nil {
			return errors.Wrap(err, "DeleteRoom RPC failed")
		}
		return nil
	})
}

// protoToRoom converts a proto.Room to an entities.Room.
func protoToRoom(protoRoom *website.Room) (*entities.Room, error) {
	roomID, err := uuid.Parse(protoRoom.Id)
	if err != nil {
		return nil, errors.Wrap(err, "invalid room ID")
	}

	ownerID, err := uuid.Parse(protoRoom.OwnerId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid owner ID")
	}

	return &entities.Room{
		ID:        roomID,
		Name:      protoRoom.Name,
		OwnerID:   ownerID,
		CreatedAt: protoRoom.CreatedAt.AsTime(),
		UpdatedAt: protoRoom.UpdatedAt.AsTime(),
	}, nil
}

// roomToProto converts an entities.Room to a proto.Room.
func roomToProto(room *entities.Room) *website.Room {
	return &website.Room{
		Id:        room.ID.String(),
		Name:      room.Name,
		OwnerId:   room.OwnerID.String(),
		CreatedAt: timestamppb.New(room.CreatedAt),
		UpdatedAt: timestamppb.New(room.UpdatedAt),
	}
}
