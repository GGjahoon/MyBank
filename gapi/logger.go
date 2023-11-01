package gapi

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

func GrpcLogger(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	//get the time started
	startTime := time.Now()
	//get the result and err from handler
	result, err := handler(ctx, req)
	//compute the duration
	duration := time.Since(startTime)
	//initialize the statusCode is unknown
	statusCode := codes.Unknown
	//get the status from err
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}
	logger := log.Info()
	//if err is not nil,convert the level of log to error
	if err != nil {
		logger = log.Error().Err(err)
	}
	logger.Str("protocol", "gRPC").
		Str("method", info.FullMethod).
		Int("status_code", int(statusCode)).
		Str("status_text", statusCode.String()).
		Dur("duration", duration).
		Msg("receive a gRPC request")
	return result, err
}

type responseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (rec *responseRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}
func (rec *responseRecorder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

func HttpLogger(handler http.Handler) http.Handler {
	//http.Handler need implement method : ServeHTTP()

	//type HandlerFunc func(ResponseWriter, *Request)
	// HandlerFunc已经实现了ServeHTTP的方法
	//func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	//	f(w, r)
	//}
	//类似于 string()的类型转换，此处将闭包函数func转换为http.HandlerFunc
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		rec := &responseRecorder{
			ResponseWriter: res,
			StatusCode:     http.StatusOK,
			Body:           nil,
		}
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)

		logger := log.Info()
		if rec.StatusCode != http.StatusOK {
			logger = log.Error().Bytes("body", rec.Body)
		}

		logger.Str("protocol", "HTTP").
			Str("method", req.Method).
			Str("path", req.RequestURI).
			Int("status_code", rec.StatusCode).
			Str("status_text", http.StatusText(rec.StatusCode)).
			Dur("duration", duration).
			Msg("receive a http request")
	})
}
