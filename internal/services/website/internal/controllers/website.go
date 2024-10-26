package controllers

import (
	"context"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/website"
	createRoomUC "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/create-room"
	deleteRoomUC "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/delete-room"
	getallrooms "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/get-all-rooms"
	getOwnerRoomsUC "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/get-owner-rooms"
	getRoomUC "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/get-room"
	searchRoomsUC "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/search-rooms"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WebsiteServiceServer struct {
	website.UnimplementedRoomServiceServer
	createRoomUC    *createRoomUC.UseCase
	deleteRoomUC    *deleteRoomUC.UseCase
	getRoomUC       *getRoomUC.UseCase
	getOwnerRoomsUC *getOwnerRoomsUC.UseCase
	searchRoomsUC   *searchRoomsUC.UseCase
	getAllRoomsUC   *getallrooms.UseCase
}

func NewWebsiteServiceServer(
	createRoomUC *createRoomUC.UseCase,
	deleteRoomUC *deleteRoomUC.UseCase,
	getRoomUC *getRoomUC.UseCase,
	getOwnerRoomsUC *getOwnerRoomsUC.UseCase,
	searchRoomsUC *searchRoomsUC.UseCase,
	getAllRoomsUC *getallrooms.UseCase,
) *WebsiteServiceServer {
	return &WebsiteServiceServer{
		createRoomUC:    createRoomUC,
		deleteRoomUC:    deleteRoomUC,
		getRoomUC:       getRoomUC,
		getOwnerRoomsUC: getOwnerRoomsUC,
		searchRoomsUC:   searchRoomsUC,
		getAllRoomsUC:   getAllRoomsUC,
	}
}

// CreateRoom handles the creation of a new room.
func (s *WebsiteServiceServer) CreateRoom(ctx context.Context, req *website.CreateRoomRequest) (*website.CreateRoomResponse, error) {
	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid owner ID")
	}

	room, err := s.createRoomUC.Execute(ctx, req.Name, ownerID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create room")
	}

	return &website.CreateRoomResponse{
		Room: &website.Room{
			Id:        room.ID.String(),
			Name:      room.Name,
			OwnerId:   room.OwnerID.String(),
			CreatedAt: timestamppb.New(room.CreatedAt),
			UpdatedAt: timestamppb.New(room.UpdatedAt),
		},
	}, nil
}

// GetAllRooms fetching all rooms with pagination support.
func (s *WebsiteServiceServer) GetAllRooms(ctx context.Context, req *website.GetAllRoomsRequest) (*website.RoomsResponse, error) {
	rooms, err := s.getAllRoomsUC.Execute(ctx, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve all rooms")
	}

	var roomProtos []*website.Room
	for _, room := range rooms {
		roomProtos = append(roomProtos, &website.Room{
			Id:        room.ID.String(),
			Name:      room.Name,
			OwnerId:   room.OwnerID.String(),
			CreatedAt: timestamppb.New(room.CreatedAt),
			UpdatedAt: timestamppb.New(room.UpdatedAt),
		})
	}

	return &website.RoomsResponse{Rooms: roomProtos}, nil
}

// GetRoom handles fetching a room by its ID.
func (s *WebsiteServiceServer) GetRoom(ctx context.Context, req *website.GetRoomRequest) (*website.Room, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid room ID")
	}

	room, err := s.getRoomUC.Execute(ctx, roomID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get room")
	}

	return &website.Room{
		Id:        room.ID.String(),
		Name:      room.Name,
		OwnerId:   room.OwnerID.String(),
		CreatedAt: timestamppb.New(room.CreatedAt),
		UpdatedAt: timestamppb.New(room.UpdatedAt),
	}, nil
}

// GetOwnerRooms handles fetching rooms by owner ID.
func (s *WebsiteServiceServer) GetOwnerRooms(ctx context.Context, req *website.GetOwnerRoomsRequest) (*website.RoomsResponse, error) {
	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid owner ID")
	}

	rooms, err := s.getOwnerRoomsUC.Execute(ctx, ownerID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get owner rooms")
	}

	var roomList []*website.Room
	for _, room := range rooms {
		roomList = append(roomList, &website.Room{
			Id:        room.ID.String(),
			Name:      room.Name,
			OwnerId:   room.OwnerID.String(),
			CreatedAt: timestamppb.New(room.CreatedAt),
			UpdatedAt: timestamppb.New(room.UpdatedAt),
		})
	}

	return &website.RoomsResponse{Rooms: roomList}, nil
}

// SearchRooms handles searching rooms by name.
func (s *WebsiteServiceServer) SearchRooms(ctx context.Context, req *website.SearchRoomsRequest) (*website.RoomsResponse, error) {
	rooms, err := s.searchRoomsUC.Execute(ctx, req.Name, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, errors.Wrap(err, "failed to search rooms")
	}

	var roomList []*website.Room
	for _, room := range rooms {
		roomList = append(roomList, &website.Room{
			Id:        room.ID.String(),
			Name:      room.Name,
			OwnerId:   room.OwnerID.String(),
			CreatedAt: timestamppb.New(room.CreatedAt),
			UpdatedAt: timestamppb.New(room.UpdatedAt),
		})
	}

	return &website.RoomsResponse{Rooms: roomList}, nil
}

// DeleteRoom handles the deletion of a room.
func (s *WebsiteServiceServer) DeleteRoom(ctx context.Context, req *website.DeleteRoomRequest) (*emptypb.Empty, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid room ID")
	}

	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid owner ID")
	}

	err = s.deleteRoomUC.Execute(ctx, roomID, ownerID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete room")
	}

	return &emptypb.Empty{}, nil
}
