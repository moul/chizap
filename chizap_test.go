package chizap_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"moul.io/chizap"
)

func TestBasic(t *testing.T) {
	r := chi.NewRouter()
	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	middleware := chizap.New(logger, &chizap.Opts{})
	r.Use(middleware)
	r.Get("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("foobar"))
	})
	ts := httptest.NewServer(r)
	defer ts.Close()

	{
		req, err := http.NewRequest("GET", ts.URL+"/foo", nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		respBody, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, string(respBody), "foobar")
		allLogs := logs.All()
		require.Len(t, allLogs, 1)
		log := allLogs[0]
		require.Equal(t, log.Message, "Served")
		require.Equal(t, log.Level, zap.InfoLevel)
		require.NotEmpty(t, log.Time)
		require.Empty(t, log.LoggerName)
		require.False(t, log.Caller.Defined)
		fields := log.Context
		require.Len(t, fields, 6)
		require.Equal(t, fieldByKey(fields, "proto"), &zapcore.Field{Key: "proto", String: "HTTP/1.1", Type: zapcore.StringType})
		require.Equal(t, fieldByKey(fields, "path"), &zapcore.Field{Key: "path", String: "/foo", Type: zapcore.StringType})
		require.Equal(t, fieldByKey(fields, "reqId"), &zapcore.Field{Key: "reqId", String: "", Type: zapcore.StringType})
		require.Equal(t, fieldByKey(fields, "status"), &zapcore.Field{Key: "status", Integer: 200, Type: zapcore.Int64Type})
		require.Equal(t, fieldByKey(fields, "size"), &zapcore.Field{Key: "size", Integer: 6, Type: zapcore.Int64Type})
		duration := fieldByKey(fields, "lat")
		require.NotZero(t, duration.Integer)
		duration.Integer = 0
		require.Equal(t, duration, &zapcore.Field{Key: "lat", Type: zapcore.DurationType})
	}
}

func TestWithFeatures(t *testing.T) {
	r := chi.NewRouter()
	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	middleware := chizap.New(logger, &chizap.Opts{
		WithReferer:   true,
		WithUserAgent: true,
	})
	r.Use(middleware)
	r.Get("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Referer", "https://manfred.life/")
		w.Header().Set("User-Agent", "TestMan")
		w.Write([]byte("foobar"))
	})
	ts := httptest.NewServer(r)
	defer ts.Close()

	{
		req, err := http.NewRequest("GET", ts.URL+"/foo", nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		respBody, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, string(respBody), "foobar")
		allLogs := logs.All()
		require.Len(t, allLogs, 1)
		log := allLogs[0]
		require.Equal(t, log.Message, "Served")
		require.Equal(t, log.Level, zap.InfoLevel)
		require.NotEmpty(t, log.Time)
		require.Empty(t, log.LoggerName)
		require.False(t, log.Caller.Defined)
		fields := log.Context
		require.Len(t, fields, 8)
		require.Equal(t, fieldByKey(fields, "proto"), &zapcore.Field{Key: "proto", String: "HTTP/1.1", Type: zapcore.StringType})
		require.Equal(t, fieldByKey(fields, "path"), &zapcore.Field{Key: "path", String: "/foo", Type: zapcore.StringType})
		require.Equal(t, fieldByKey(fields, "reqId"), &zapcore.Field{Key: "reqId", String: "", Type: zapcore.StringType})
		require.Equal(t, fieldByKey(fields, "status"), &zapcore.Field{Key: "status", Integer: 200, Type: zapcore.Int64Type})
		require.Equal(t, fieldByKey(fields, "size"), &zapcore.Field{Key: "size", Integer: 6, Type: zapcore.Int64Type})
		require.Equal(t, fieldByKey(fields, "ref"), &zapcore.Field{Key: "ref", String: "https://manfred.life/", Type: zapcore.StringType})
		require.Equal(t, fieldByKey(fields, "ua"), &zapcore.Field{Key: "ua", String: "TestMan", Type: zapcore.StringType})
		duration := fieldByKey(fields, "lat")
		require.NotZero(t, duration.Integer)
		duration.Integer = 0
		require.Equal(t, duration, &zapcore.Field{Key: "lat", Type: zapcore.DurationType})
	}
}

func fieldByKey(fields []zapcore.Field, key string) *zapcore.Field {
	for _, field := range fields {
		if field.Key == key {
			return &field
		}
	}
	return nil
}
