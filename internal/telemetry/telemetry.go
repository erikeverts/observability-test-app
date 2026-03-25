package telemetry

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"time"

	"github.com/erikeverts/observability-test-app/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	LoggerProvider *sdklog.LoggerProvider
}

func InitTelemetry(ctx context.Context, cfg *config.Config) (*Telemetry, func(), error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("creating resource: %w", err)
	}

	endpoint := cfg.OTLPEndpoint
	protocol := cfg.OTLPProtocol
	isGrafana := cfg.GrafanaOTLPEndpoint != ""

	if isGrafana {
		endpoint = cfg.GrafanaOTLPEndpoint
		protocol = "http/protobuf"
	}

	var (
		tp *sdktrace.TracerProvider
		mp *sdkmetric.MeterProvider
		lp *sdklog.LoggerProvider
	)

	switch protocol {
	case "http/protobuf", "http":
		tp, mp, lp, err = initHTTPExporters(ctx, endpoint, res, cfg.GrafanaAPIToken, isGrafana)
	default:
		tp, mp, lp, err = initGRPCExporters(ctx, endpoint, res)
	}
	if err != nil {
		return nil, nil, err
	}

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	cleanup := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutting down tracer provider", "error", err)
		}
		if err := mp.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutting down meter provider", "error", err)
		}
		if err := lp.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutting down logger provider", "error", err)
		}
	}

	return &Telemetry{
		TracerProvider: tp,
		MeterProvider:  mp,
		LoggerProvider: lp,
	}, cleanup, nil
}

func initGRPCExporters(ctx context.Context, endpoint string, res *resource.Resource) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider, *sdklog.LoggerProvider, error) {
	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating gRPC trace exporter: %w", err)
	}

	metricExp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating gRPC metric exporter: %w", err)
	}

	logExp, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating gRPC log exporter: %w", err)
	}

	tp, mp, lp := buildProviders(res, traceExp, metricExp, logExp)
	return tp, mp, lp, nil
}

func initHTTPExporters(ctx context.Context, endpoint string, res *resource.Resource, apiToken string, isGrafana bool) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider, *sdklog.LoggerProvider, error) {
	traceOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
	}
	metricOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(endpoint),
	}
	logOpts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint),
	}

	if isGrafana && apiToken != "" {
		headers := map[string]string{
			"Authorization": "Basic " + apiToken,
		}
		traceOpts = append(traceOpts,
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithTLSClientConfig(&tls.Config{}),
			otlptracehttp.WithURLPath("/otlp/v1/traces"),
		)
		metricOpts = append(metricOpts,
			otlpmetrichttp.WithHeaders(headers),
			otlpmetrichttp.WithTLSClientConfig(&tls.Config{}),
			otlpmetrichttp.WithURLPath("/otlp/v1/metrics"),
		)
		logOpts = append(logOpts,
			otlploghttp.WithHeaders(headers),
			otlploghttp.WithTLSClientConfig(&tls.Config{}),
			otlploghttp.WithURLPath("/otlp/v1/logs"),
		)
	} else {
		traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
		metricOpts = append(metricOpts, otlpmetrichttp.WithInsecure())
		logOpts = append(logOpts, otlploghttp.WithInsecure())
	}

	traceExp, err := otlptracehttp.New(ctx, traceOpts...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating HTTP trace exporter: %w", err)
	}

	metricExp, err := otlpmetrichttp.New(ctx, metricOpts...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating HTTP metric exporter: %w", err)
	}

	logExp, err := otlploghttp.New(ctx, logOpts...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating HTTP log exporter: %w", err)
	}

	tp, mp, lp := buildProviders(res, traceExp, metricExp, logExp)
	return tp, mp, lp, nil
}

func buildProviders(res *resource.Resource, traceExp sdktrace.SpanExporter, metricExp sdkmetric.Exporter, logExp sdklog.Exporter) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider, *sdklog.LoggerProvider) {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(15*time.Second))),
		sdkmetric.WithResource(res),
	)

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)

	return tp, mp, lp
}
