package grpcserver

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/Julia-ivv/shortener-url.git/internal/authorizer"
	"github.com/Julia-ivv/shortener-url.git/internal/config"
	pb "github.com/Julia-ivv/shortener-url.git/internal/proto"
	"github.com/Julia-ivv/shortener-url.git/internal/storage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var cfg config.Flags

const testUserID = 123

type testURL struct {
	shortURL    string
	originURL   string
	deletedFlag bool
	userID      int
}

type testURLs struct {
	originalURLs []testURL
}

func Init() {
	cfg = *config.NewConfig()
}

func (urls *testURLs) DeleteUserURLs(ctx context.Context, delURLs []string, userID int) (err error) {
	for _, delURL := range delURLs {
		for k, curURL := range urls.originalURLs {
			if (delURL == curURL.shortURL) && (userID == curURL.userID) {
				urls.originalURLs[k].deletedFlag = true
				break
			}
		}
	}
	return nil
}

func (urls *testURLs) GetURL(ctx context.Context, shortURL string) (originURL string, isDel bool, ok bool) {
	for _, v := range urls.originalURLs {
		if v.shortURL == shortURL {
			return v.originURL, v.deletedFlag, true
		}
	}
	return "", false, false
}

func (urls *testURLs) AddURL(ctx context.Context, shortURL string, originURL string, userID int) (findURL string, err error) {
	urls.originalURLs = append(urls.originalURLs, testURL{
		userID:    userID,
		shortURL:  shortURL,
		originURL: originURL,
	})
	return "", nil
}

func (urls *testURLs) AddBatch(ctx context.Context, shortURLBatch []storage.ResponseBatch, originURLBatch []storage.RequestBatch, userID int) (err error) {
	allUrls := make([]testURL, len(originURLBatch))
	for k, v := range shortURLBatch {
		allUrls = append(allUrls, testURL{
			userID:    userID,
			shortURL:  v.ShortURL,
			originURL: originURLBatch[k].OriginalURL,
		})
	}

	urls.originalURLs = append(urls.originalURLs, allUrls...)
	return nil
}

func (urls *testURLs) GetAllUserURLs(ctx context.Context, baseURL string, userID int) (userURLs []storage.UserURL, err error) {

	for _, v := range urls.originalURLs {
		if v.userID == userID {
			userURLs = append(userURLs, storage.UserURL{
				ShortURL:    baseURL + v.shortURL,
				OriginalURL: v.originURL,
			})
		}
	}

	return userURLs, nil
}

func (urls *testURLs) PingStor(ctx context.Context) (err error) {
	if urls == nil {
		return errors.New("storage storage does not exist")
	}
	return nil
}

func (urls *testURLs) Close() (err error) {
	return nil
}

func (urls *testURLs) GetStats(ctx context.Context) (stats storage.ServiceStats, err error) {
	stats.URLs = len(urls.originalURLs)
	stats.Users = 0

	tmp := make([]int, len(urls.originalURLs))
	for _, v := range urls.originalURLs {
		if !slices.Contains(tmp, v.userID) {
			tmp = append(tmp, v.userID)
			stats.Users++
		}
	}

	return stats, nil
}

func createTestRepo() *testURLs {
	testR := make([]testURL, 0)
	testR = append(testR, testURL{
		userID:      testUserID,
		shortURL:    "EwH",
		deletedFlag: false,
		originURL:   "https://practicum.yandex.ru/",
	})
	testR = append(testR, testURL{
		userID:      testUserID,
		shortURL:    "Eorp",
		deletedFlag: true,
		originURL:   "https://yandex.ru/",
	})
	testR = append(testR, testURL{
		userID:      456,
		shortURL:    "etrygh",
		deletedFlag: false,
		originURL:   "https://mai.ru/",
	})
	return &testURLs{originalURLs: testR}
}
func TestNewShortenerServer(t *testing.T) {
	t.Run("create new service", func(t *testing.T) {
		res := NewShortenerServer(createTestRepo(), cfg, &sync.WaitGroup{})
		assert.NotEmpty(t, res)
	})
}

func TestGetUrl(t *testing.T) {
	testRepo := createTestRepo()
	testServ := NewShortenerServer(testRepo, cfg, &sync.WaitGroup{})
	tests := []struct {
		name      string
		in        *pb.GetUrlRequest
		res       *pb.GetUrlResponse
		wantError bool
		wantCode  codes.Code
	}{
		{
			name:      "ok test",
			in:        &pb.GetUrlRequest{ShortUrl: "EwH"},
			res:       &pb.GetUrlResponse{OriginalUrl: "https://practicum.yandex.ru/"},
			wantError: false,
			wantCode:  codes.OK,
		},
		{
			name:      "not found",
			in:        &pb.GetUrlRequest{ShortUrl: "errr"},
			res:       nil,
			wantError: true,
			wantCode:  codes.NotFound,
		},
		{
			name:      "deleted",
			in:        &pb.GetUrlRequest{ShortUrl: "Eorp"},
			res:       nil,
			wantError: true,
			wantCode:  codes.NotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := testServ.GetUrl(context.Background(), test.in)
			if test.wantError {
				assert.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, test.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.res, r)
			}
		})
	}
}

func TestPostBatch(t *testing.T) {
	testRepo := createTestRepo()
	testServ := NewShortenerServer(testRepo, cfg, &sync.WaitGroup{})
	tests := []struct {
		name      string
		in        *pb.PostBatchRequest
		res       *pb.PostBatchResponse
		ctx       context.Context
		wantError bool
		wantCode  codes.Code
	}{
		{
			name: "ok test",
			in: &pb.PostBatchRequest{
				RequestBatchs: []*pb.PostBatchRequest_RequestBatch{
					{
						CorrelationId: "1",
						OriginalUrl:   "http://qwert.ru",
					},
					{
						CorrelationId: "2",
						OriginalUrl:   "http://yuio.ru",
					},
				},
			},
			res: &pb.PostBatchResponse{
				ResponseBatchs: []*pb.PostBatchResponse_ResponseBatch{
					{
						CorrelationId: "1",
						ShortUrl:      "some url",
					},
					{
						CorrelationId: "2",
						ShortUrl:      "other url",
					},
				},
			},
			ctx:       context.WithValue(context.Background(), authorizer.UserContextKey, testUserID),
			wantError: false,
			wantCode:  codes.OK,
		},
		{
			name: "missing user id",
			in: &pb.PostBatchRequest{
				RequestBatchs: []*pb.PostBatchRequest_RequestBatch{
					{
						CorrelationId: "1",
						OriginalUrl:   "http://qwert.ru",
					},
				},
			},
			res:       nil,
			ctx:       context.Background(),
			wantError: true,
			wantCode:  codes.Unauthenticated,
		},
		{
			name:      "empty request",
			in:        &pb.PostBatchRequest{},
			res:       nil,
			ctx:       context.WithValue(context.Background(), authorizer.UserContextKey, testUserID),
			wantError: true,
			wantCode:  codes.DataLoss,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := testServ.PostBatch(test.ctx, test.in)
			if test.wantError {
				assert.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, test.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(test.res.ResponseBatchs), len(r.ResponseBatchs))
			}
		})
	}
}

func TestPostUrl(t *testing.T) {
	testRepo := createTestRepo()
	testServ := NewShortenerServer(testRepo, cfg, &sync.WaitGroup{})
	tests := []struct {
		name      string
		in        *pb.PostUrlRequest
		res       *pb.PostUrlResponse
		ctx       context.Context
		wantError bool
		wantCode  codes.Code
	}{
		{
			name:      "ok test",
			in:        &pb.PostUrlRequest{OriginalUrl: "https://pract.ru/"},
			res:       &pb.PostUrlResponse{ShortUrl: "some url"},
			ctx:       context.WithValue(context.Background(), authorizer.UserContextKey, testUserID),
			wantError: false,
			wantCode:  codes.OK,
		},
		{
			name:      "missing id",
			in:        &pb.PostUrlRequest{OriginalUrl: "errr"},
			res:       nil,
			ctx:       context.Background(),
			wantError: true,
			wantCode:  codes.Unauthenticated,
		},
		{
			name:      "empty request",
			in:        &pb.PostUrlRequest{},
			res:       nil,
			ctx:       context.WithValue(context.Background(), authorizer.UserContextKey, testUserID),
			wantError: true,
			wantCode:  codes.DataLoss,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := testServ.PostUrl(test.ctx, test.in)
			if test.wantError {
				assert.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, test.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, r.ShortUrl)
			}
		})
	}
}

func TestGetUserUrls(t *testing.T) {
	testRepo := createTestRepo()
	testServ := NewShortenerServer(testRepo, cfg, &sync.WaitGroup{})
	tests := []struct {
		name      string
		in        *pb.GetUserUrlsRequest
		res       *pb.GetUserUrlsResponse
		ctx       context.Context
		wantError bool
		wantCode  codes.Code
	}{
		{
			name: "ok test",
			in:   &pb.GetUserUrlsRequest{},
			res: &pb.GetUserUrlsResponse{
				UserUrls: []*pb.GetUserUrlsResponse_UserUrl{
					{
						ShortUrl:    "/EwH",
						OriginalUrl: "https://practicum.yandex.ru/",
					},
					{
						ShortUrl:    "/Eorp",
						OriginalUrl: "https://yandex.ru/",
					},
				},
			},
			ctx:       context.WithValue(context.Background(), authorizer.UserContextKey, testUserID),
			wantError: false,
			wantCode:  codes.OK,
		},
		{
			name:      "missing id",
			in:        &pb.GetUserUrlsRequest{},
			res:       nil,
			ctx:       context.Background(),
			wantError: true,
			wantCode:  codes.Unauthenticated,
		},
		{
			name:      "no content",
			in:        &pb.GetUserUrlsRequest{},
			res:       nil,
			ctx:       context.WithValue(context.Background(), authorizer.UserContextKey, 6),
			wantError: true,
			wantCode:  codes.NotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := testServ.GetUserUrls(test.ctx, test.in)
			if test.wantError {
				assert.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, test.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.res.UserUrls, r.UserUrls)
			}
		})
	}
}

func TestDeleteUserUrls(t *testing.T) {
	testRepo := createTestRepo()
	testServ := NewShortenerServer(testRepo, cfg, &sync.WaitGroup{})
	tests := []struct {
		name      string
		in        *pb.DeleteUserUrlsRequest
		res       *pb.DeleteUserUrlsResponse
		ctx       context.Context
		wantError bool
		wantCode  codes.Code
	}{
		{
			name: "ok test",
			in: &pb.DeleteUserUrlsRequest{
				DelUrls: []string{"EwH"},
			},
			res:       &pb.DeleteUserUrlsResponse{},
			ctx:       context.WithValue(context.Background(), authorizer.UserContextKey, testUserID),
			wantError: false,
			wantCode:  codes.OK,
		},
		{
			name:      "missing id",
			in:        &pb.DeleteUserUrlsRequest{},
			res:       nil,
			ctx:       context.Background(),
			wantError: true,
			wantCode:  codes.Unauthenticated,
		},
		{
			name:      "empty requesrt",
			in:        &pb.DeleteUserUrlsRequest{},
			res:       nil,
			ctx:       context.WithValue(context.Background(), authorizer.UserContextKey, testUserID),
			wantError: true,
			wantCode:  codes.DataLoss,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := testServ.DeleteUserUrls(test.ctx, test.in)
			if test.wantError {
				assert.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, test.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	testRepo := createTestRepo()
	testServ := NewShortenerServer(testRepo, cfg, &sync.WaitGroup{})
	ctxWithMd := metadata.NewIncomingContext(context.Background(),
		metadata.New(map[string]string{"X-Real-IP": "192.168.0.1"}))
	trSubn := "192.168.0.0/24"
	tests := []struct {
		name          string
		in            *pb.GetStatsRequest
		res           *pb.GetStatsResponse
		ctx           context.Context
		trustedSubnet string
		wantError     bool
		wantCode      codes.Code
	}{
		{
			name: "ok test",
			in:   &pb.GetStatsRequest{},
			res: &pb.GetStatsResponse{
				Urls:  3,
				Users: 2,
			},
			ctx:           ctxWithMd,
			trustedSubnet: trSubn,
			wantError:     false,
			wantCode:      codes.OK,
		},
		{
			name:          "empty trusted subnet",
			in:            &pb.GetStatsRequest{},
			res:           &pb.GetStatsResponse{},
			ctx:           ctxWithMd,
			trustedSubnet: "",
			wantError:     true,
			wantCode:      codes.PermissionDenied,
		},
		{
			name:          "missing ip",
			in:            &pb.GetStatsRequest{},
			res:           &pb.GetStatsResponse{},
			ctx:           context.Background(),
			trustedSubnet: trSubn,
			wantError:     true,
			wantCode:      codes.Internal,
		},
		{
			name:          "not trusted IP",
			in:            &pb.GetStatsRequest{},
			res:           &pb.GetStatsResponse{},
			ctx:           ctxWithMd,
			trustedSubnet: "192.168.1.1/24",
			wantError:     true,
			wantCode:      codes.PermissionDenied,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testServ.cfg.TrustedSubnet = test.trustedSubnet
			r, err := testServ.GetStats(test.ctx, test.in)
			if test.wantError {
				assert.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, test.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.res.Urls, r.Urls)
			}
		})
	}
}

func TestGetPing(t *testing.T) {
	testRepo := createTestRepo()
	testServ := NewShortenerServer(testRepo, cfg, &sync.WaitGroup{})

	t.Run("ok ping", func(t *testing.T) {
		_, err := testServ.GetPing(context.Background(), nil)
		assert.NoError(t, err)
	})

	testRepo = nil
	testServ = NewShortenerServer(testRepo, cfg, &sync.WaitGroup{})
	t.Run("error ping", func(t *testing.T) {
		_, err := testServ.GetPing(context.Background(), nil)
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}
