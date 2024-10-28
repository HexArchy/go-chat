package controllers

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/website"
	"github.com/HexArch/go-chat/internal/services/website/internal/controllers/cache"
	"github.com/HexArch/go-chat/internal/services/website/internal/controllers/middleware"
	"github.com/HexArch/go-chat/internal/services/website/internal/metrics"
	createRoomUC "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/create-room"
	deleteRoomUC "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/delete-room"
	getallrooms "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/get-all-rooms"
	getOwnerRoomsUC "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/get-owner-rooms"
	getRoomUC "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/get-room"
	searchRoomsUC "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/search-rooms"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WebsiteServiceServer struct {
	logger    *zap.Logger
	metrics   *metrics.WebsiteMetrics
	roomCache *cache.RoomCache
	website.UnimplementedRoomServiceServer
	createRoomUC    *createRoomUC.UseCase
	deleteRoomUC    *deleteRoomUC.UseCase
	getRoomUC       *getRoomUC.UseCase
	getOwnerRoomsUC *getOwnerRoomsUC.UseCase
	searchRoomsUC   *searchRoomsUC.UseCase
	getAllRoomsUC   *getallrooms.UseCase
}

func NewWebsiteServiceServer(
	logger *zap.Logger,
	metrics *metrics.WebsiteMetrics,
	roomCache *cache.RoomCache,
	createRoomUC *createRoomUC.UseCase,
	deleteRoomUC *deleteRoomUC.UseCase,
	getRoomUC *getRoomUC.UseCase,
	getOwnerRoomsUC *getOwnerRoomsUC.UseCase,
	searchRoomsUC *searchRoomsUC.UseCase,
	getAllRoomsUC *getallrooms.UseCase,
) *WebsiteServiceServer {
	return &WebsiteServiceServer{
		logger:          logger,
		metrics:         metrics,
		roomCache:       roomCache,
		createRoomUC:    createRoomUC,
		deleteRoomUC:    deleteRoomUC,
		getRoomUC:       getRoomUC,
		getOwnerRoomsUC: getOwnerRoomsUC,
		searchRoomsUC:   searchRoomsUC,
		getAllRoomsUC:   getAllRoomsUC,
	}
}

func (s *WebsiteServiceServer) CreateRoom(ctx context.Context, req *website.CreateRoomRequest) (*website.CreateRoomResponse, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordRequestDuration("CreateRoom", "success", time.Since(start).Seconds())
	}()

	// Validate requester.
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		s.metrics.RecordError("unauthorized")
		return nil, status.Error(codes.Unauthenticated, "unauthorized access")
	}

	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		s.metrics.RecordError("invalid_owner_id")
		return nil, status.Error(codes.InvalidArgument, "invalid owner ID format")
	}

	// Verify that requester is the owner.
	if userID != req.OwnerId {
		s.metrics.RecordError("permission_denied")
		return nil, status.Error(codes.PermissionDenied, "can only create rooms for yourself")
	}

	// Validate room name.
	if len(req.Name) < 3 || len(req.Name) > 50 {
		s.metrics.RecordError("invalid_room_name")
		return nil, status.Error(codes.InvalidArgument, "room name must be between 3 and 50 characters")
	}

	room, err := s.createRoomUC.Execute(ctx, req.Name, ownerID)
	if err != nil {
		s.logger.Error("Failed to create room",
			zap.Error(err),
			zap.String("name", req.Name),
			zap.String("owner_id", req.OwnerId))
		s.metrics.RecordError("create_room_failed")
		return nil, status.Error(codes.Internal, "failed to create room")
	}

	s.metrics.RecordRoomCreation()
	s.roomCache.Set(room.ID, room)

	s.logger.Info("Room created successfully",
		zap.String("room_id", room.ID.String()),
		zap.String("name", room.Name),
		zap.String("owner_id", room.OwnerID.String()))

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

func (s *WebsiteServiceServer) GetAllRooms(ctx context.Context, req *website.GetAllRoomsRequest) (*website.RoomsResponse, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordRequestDuration("GetAllRooms", "success", time.Since(start).Seconds())
	}()

	limit := int(req.Limit)
	offset := int(req.Offset)

	// Validate pagination parameters.
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rooms, err := s.getAllRoomsUC.Execute(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get rooms",
			zap.Error(err),
			zap.Int("limit", limit),
			zap.Int("offset", offset))
		s.metrics.RecordError("get_rooms_failed")
		return nil, status.Error(codes.Internal, "failed to fetch rooms")
	}

	response := &website.RoomsResponse{
		Rooms: make([]*website.Room, len(rooms)),
	}

	for i, room := range rooms {
		response.Rooms[i] = &website.Room{
			Id:        room.ID.String(),
			Name:      room.Name,
			OwnerId:   room.OwnerID.String(),
			CreatedAt: timestamppb.New(room.CreatedAt),
			UpdatedAt: timestamppb.New(room.UpdatedAt),
		}
	}

	return response, nil
}

func (s *WebsiteServiceServer) GetRoom(ctx context.Context, req *website.GetRoomRequest) (*website.Room, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordRequestDuration("GetRoom", "success", time.Since(start).Seconds())
	}()

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		s.metrics.RecordError("invalid_room_id")
		return nil, status.Error(codes.InvalidArgument, "invalid room ID format")
	}

	// Check cache first.
	if room, found := s.roomCache.Get(roomID); found {
		s.metrics.RecordCacheHit("room")
		return &website.Room{
			Id:        room.ID.String(),
			Name:      room.Name,
			OwnerId:   room.OwnerID.String(),
			CreatedAt: timestamppb.New(room.CreatedAt),
			UpdatedAt: timestamppb.New(room.UpdatedAt),
		}, nil
	}

	room, err := s.getRoomUC.Execute(ctx, roomID)
	if err != nil {
		s.logger.Error("Failed to get room",
			zap.Error(err),
			zap.String("room_id", req.RoomId))
		s.metrics.RecordError("get_room_failed")
		return nil, status.Error(codes.Internal, "failed to fetch room")
	}

	// Cache the room for future requests.
	s.roomCache.Set(room.ID, room)

	return &website.Room{
		Id:        room.ID.String(),
		Name:      room.Name,
		OwnerId:   room.OwnerID.String(),
		CreatedAt: timestamppb.New(room.CreatedAt),
		UpdatedAt: timestamppb.New(room.UpdatedAt),
	}, nil
}

func (s *WebsiteServiceServer) GetOwnerRooms(ctx context.Context, req *website.GetOwnerRoomsRequest) (*website.RoomsResponse, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordRequestDuration("GetOwnerRooms", "success", time.Since(start).Seconds())
	}()

	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		s.metrics.RecordError("invalid_owner_id")
		return nil, status.Error(codes.InvalidArgument, "invalid owner ID format")
	}

	rooms, err := s.getOwnerRoomsUC.Execute(ctx, ownerID)
	if err != nil {
		s.logger.Error("Failed to get owner rooms",
			zap.Error(err),
			zap.String("owner_id", req.OwnerId))
		s.metrics.RecordError("get_owner_rooms_failed")
		return nil, status.Error(codes.Internal, "failed to fetch owner rooms")
	}

	response := &website.RoomsResponse{
		Rooms: make([]*website.Room, len(rooms)),
	}

	for i, room := range rooms {
		response.Rooms[i] = &website.Room{
			Id:        room.ID.String(),
			Name:      room.Name,
			OwnerId:   room.OwnerID.String(),
			CreatedAt: timestamppb.New(room.CreatedAt),
			UpdatedAt: timestamppb.New(room.UpdatedAt),
		}
	}

	return response, nil
}

func (s *WebsiteServiceServer) SearchRooms(ctx context.Context, req *website.SearchRoomsRequest) (*website.RoomsResponse, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordRequestDuration("SearchRooms", "success", time.Since(start).Seconds())
	}()

	// Validate search parameters.
	if len(req.Name) < 2 {
		s.metrics.RecordError("invalid_search_query")
		return nil, status.Error(codes.InvalidArgument, "search query must be at least 2 characters")
	}

	limit := int(req.Limit)
	offset := int(req.Offset)

	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	s.metrics.RecordSearchQuery()

	rooms, err := s.searchRoomsUC.Execute(ctx, req.Name, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search rooms",
			zap.Error(err),
			zap.String("query", req.Name),
			zap.Int("limit", limit),
			zap.Int("offset", offset))
		s.metrics.RecordError("search_rooms_failed")
		return nil, status.Error(codes.Internal, "failed to search rooms")
	}

	response := &website.RoomsResponse{
		Rooms: make([]*website.Room, len(rooms)),
	}

	for i, room := range rooms {
		response.Rooms[i] = &website.Room{
			Id:        room.ID.String(),
			Name:      room.Name,
			OwnerId:   room.OwnerID.String(),
			CreatedAt: timestamppb.New(room.CreatedAt),
			UpdatedAt: timestamppb.New(room.UpdatedAt),
		}
	}

	return response, nil
}

func (s *WebsiteServiceServer) DeleteRoom(ctx context.Context, req *website.DeleteRoomRequest) (*emptypb.Empty, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordRequestDuration("DeleteRoom", "success", time.Since(start).Seconds())
	}()

	// Validate requester.
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		s.metrics.RecordError("unauthorized")
		return nil, status.Error(codes.Unauthenticated, "unauthorized access")
	}

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		s.metrics.RecordError("invalid_room_id")
		return nil, status.Error(codes.InvalidArgument, "invalid room ID format")
	}

	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		s.metrics.RecordError("invalid_owner_id")
		return nil, status.Error(codes.InvalidArgument, "invalid owner ID format")
	}

	// Verify that requester is the owner.
	if userID != req.OwnerId {
		s.metrics.RecordError("permission_denied")
		return nil, status.Error(codes.PermissionDenied, "only room owner can delete the room")
	}

	err = s.deleteRoomUC.Execute(ctx, roomID, ownerID)
	if err != nil {
		s.logger.Error("Failed to delete room",
			zap.Error(err),
			zap.String("room_id", req.RoomId),
			zap.String("owner_id", req.OwnerId))
		s.metrics.RecordError("delete_room_failed")
		return nil, status.Error(codes.Internal, "failed to delete room")
	}

	// Remove from cache.
	s.roomCache.Delete(roomID)
	s.metrics.RecordRoomDeletion()

	s.logger.Info("Room deleted successfully",
		zap.String("room_id", req.RoomId),
		zap.String("owner_id", req.OwnerId))

	return &emptypb.Empty{}, nil
}
