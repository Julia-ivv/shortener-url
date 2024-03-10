// Package grpcserver contains gRPC methods.
package grpcserver

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/Julia-ivv/shortener-url.git/internal/authorizer"
	"github.com/Julia-ivv/shortener-url.git/internal/config"
	pb "github.com/Julia-ivv/shortener-url.git/internal/proto"
	"github.com/Julia-ivv/shortener-url.git/internal/storage"
	"github.com/Julia-ivv/shortener-url.git/pkg/randomizer"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ShortenerServer stores the repository and settings of this application.
type ShortenerServer struct {
	pb.UnimplementedShortUrlServer
	stor storage.Repositories
	cfg  config.Flags
	wg   *sync.WaitGroup
}

// NewShortenerServer creates an instance with storage and settings for grpc methods.
func NewShortenerServer(stor storage.Repositories, cfg config.Flags, wg *sync.WaitGroup) *ShortenerServer {
	h := &ShortenerServer{}
	h.stor = stor
	h.cfg = cfg
	h.wg = wg
	return h
}

// GetURL gets a long URL from the storage using shortURL.
func (h *ShortenerServer) GetURL(ctx context.Context, in *pb.GetUrlRequest) (*pb.GetUrlResponse, error) {
	shortURL := in.ShortUrl

	originURL, isDel, ok := h.stor.GetURL(ctx, shortURL)
	if !ok {
		return nil, status.Error(codes.NotFound, "short URL not found")
	}
	if isDel {
		return nil, status.Error(codes.NotFound, "short URL has been removed")
	}

	return &pb.GetUrlResponse{
		OriginalUrl: originURL,
	}, nil
}

// PostBatch gets a slice of the original URLs from the request body.
// Adds it to storage, returns a slice of the short URLs in the response body.
func (h *ShortenerServer) PostBatch(ctx context.Context, in *pb.PostBatchRequest) (*pb.PostBatchResponse, error) {
	v := ctx.Value(authorizer.UserContextKey)
	if v == nil {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	id := v.(int)

	if len(in.RequestBatchs) == 0 {
		return nil, status.Error(codes.DataLoss, "empty request")
	}

	reqBatch := make([]storage.RequestBatch, len(in.RequestBatchs))
	for k, v := range in.RequestBatchs {
		reqBatch[k].CorrelationID = v.CorrelationId
		reqBatch[k].OriginalURL = v.OriginalUrl
	}

	resBatch := make([]storage.ResponseBatch, len(in.RequestBatchs))
	for k, v := range reqBatch {
		shortURL, err := randomizer.GenerateRandomString(randomizer.LengthShortURL)
		if err != nil {
			return nil, status.Error(codes.Internal, "randomizer error")
		}
		resBatch[k].CorrelationID = v.CorrelationID
		resBatch[k].ShortURLFull = h.cfg.URL + "/" + shortURL
		resBatch[k].ShortURL = shortURL
	}

	err := h.stor.AddBatch(ctx, resBatch, reqBatch, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := make([]*pb.PostBatchResponse_ResponseBatch, 0, len(resBatch))
	for _, v := range resBatch {
		res = append(res, &pb.PostBatchResponse_ResponseBatch{
			CorrelationId: v.CorrelationID,
			ShortUrl:      v.ShortURLFull,
		})
	}

	return &pb.PostBatchResponse{ResponseBatchs: res}, nil
}

// PostURL gets a long URL from the request body.
// Adds it to storage, returns a short URL in the response body.
func (h *ShortenerServer) PostURL(ctx context.Context, in *pb.PostUrlRequest) (*pb.PostUrlResponse, error) {
	v := ctx.Value(authorizer.UserContextKey)
	if v == nil {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	id := v.(int)

	if len(in.OriginalUrl) == 0 {
		return nil, status.Error(codes.DataLoss, "empty request")
	}

	shortURL, err := randomizer.GenerateRandomString(randomizer.LengthShortURL)
	if err != nil {
		return nil, status.Error(codes.Internal, "randomizer error")
	}
	findURL, err := h.stor.AddURL(ctx, shortURL, in.OriginalUrl, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return &pb.PostUrlResponse{ShortUrl: h.cfg.URL + "/" + findURL},
				status.Error(codes.AlreadyExists, "this URL already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.PostUrlResponse{ShortUrl: h.cfg.URL + "/" + shortURL}, nil
}

// GetUserURLs gets all the user's short urls from the repository.
func (h *ShortenerServer) GetUserUrls(ctx context.Context, in *pb.GetUserUrlsRequest) (*pb.GetUserUrlsResponse, error) {
	v := ctx.Value(authorizer.UserContextKey)
	if v == nil {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	id := v.(int)

	allURLs, err := h.stor.GetAllUserURLs(ctx, h.cfg.URL+"/", id)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}
	if len(allURLs) == 0 {
		return nil, status.Error(codes.NotFound, "no content")
	}

	res := make([]*pb.GetUserUrlsResponse_UserUrl, 0, len(allURLs))
	for _, v := range allURLs {
		res = append(res, &pb.GetUserUrlsResponse_UserUrl{
			ShortUrl:    v.ShortURL,
			OriginalUrl: v.OriginalURL,
		})
	}

	return &pb.GetUserUrlsResponse{UserUrls: res}, nil
}

// DeleteUserURLs adds a removal flag for URLs from the request body.
func (h *ShortenerServer) DeleteUserUrls(ctx context.Context, in *pb.DeleteUserUrlsRequest) (*pb.DeleteUserUrlsResponse, error) {
	v := ctx.Value(authorizer.UserContextKey)
	if v == nil {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	id := v.(int)

	if len(in.DelUrls) == 0 {
		return nil, status.Error(codes.DataLoss, "empty request")
	}

	h.wg.Add(1)
	go func() {
		h.stor.DeleteUserURLs(ctx, in.DelUrls, id)
		h.wg.Done()
	}()

	return nil, nil
}

// GetStats gets the amount of all users and URLs in the service.
// Available only for IP addresses from a trusted subnet.
func (h *ShortenerServer) GetStats(ctx context.Context, in *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	if h.cfg.TrustedSubnet == "" {
		return nil, status.Error(codes.PermissionDenied, "empty trusted subnet")
	}

	var ipStr string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("X-Real-IP")
		if len(values) > 0 {
			ipStr = values[0]
		}
	}
	if len(ipStr) == 0 {
		return nil, status.Error(codes.Internal, "missing IP")
	}

	ip := net.ParseIP(ipStr)
	_, ipNet, err := net.ParseCIDR(h.cfg.TrustedSubnet)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !ipNet.Contains(ip) {
		return nil, status.Error(codes.PermissionDenied, "not trusted IP")
	}

	stats, err := h.stor.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetStatsResponse{
		Urls:  int32(stats.URLs),
		Users: int32(stats.Users),
	}, nil
}

// GetPingDB checks storage access.
func (h *ShortenerServer) GetPing(ctx context.Context, in *pb.GetPingRequest) (*pb.GetPingResponse, error) {
	if err := h.stor.PingStor(ctx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, nil
}
