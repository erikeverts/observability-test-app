package telemetry

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"strings"
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
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	LoggerProvider *sdklog.LoggerProvider
}

// exportOpts holds resolved exporter configuration.
type exportOpts struct {
	endpoint            string
	headers             map[string]string
	insecure            bool
	metricsTemporality  string
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

	// Clear OTEL_EXPORTER_OTLP_* env vars so the SDK doesn't auto-apply them
	// on top of our explicit WithEndpoint/WithHeaders/WithInsecure calls.
	os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	os.Unsetenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	os.Unsetenv("OTEL_EXPORTER_OTLP_INSECURE")
	os.Unsetenv("OTEL_EXPORTER_OTLP_HEADERS")

	opts := resolveExportOpts(cfg)
	protocol := cfg.OTLPProtocol
	if cfg.GrafanaOTLPEndpoint != "" {
		protocol = "http/protobuf"
	}

	var (
		tp *sdktrace.TracerProvider
		mp *sdkmetric.MeterProvider
		lp *sdklog.LoggerProvider
	)

	switch protocol {
	case "http/protobuf", "http":
		isGrafana := cfg.GrafanaOTLPEndpoint != ""
		tp, mp, lp, err = initHTTPExporters(ctx, opts, res, isGrafana)
	default:
		tp, mp, lp, err = initGRPCExporters(ctx, opts, res)
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

// normalizeEndpoint strips http:// or https:// from the endpoint and infers
// the insecure flag from the scheme. The OTEL SDK WithEndpoint expects
// host:port without a scheme.
func normalizeEndpoint(ep string, fallbackInsecure bool) (string, bool) {
	if strings.HasPrefix(ep, "https://") {
		return strings.TrimPrefix(ep, "https://"), false
	}
	if strings.HasPrefix(ep, "http://") {
		return strings.TrimPrefix(ep, "http://"), true
	}
	return ep, fallbackInsecure
}

// resolveExportOpts merges generic OTLP config with Grafana overlay.
func resolveExportOpts(cfg *config.Config) exportOpts {
	endpoint, insecure := normalizeEndpoint(cfg.OTLPEndpoint, cfg.OTLPInsecure)
	opts := exportOpts{
		endpoint:           endpoint,
		headers:            copyHeaders(cfg.OTLPHeaders),
		insecure:           insecure,
		metricsTemporality: cfg.OTLPMetricsTemporality,
	}

	if cfg.GrafanaOTLPEndpoint != "" {
		opts.endpoint, opts.insecure = normalizeEndpoint(cfg.GrafanaOTLPEndpoint, false)
		if cfg.GrafanaAPIToken != "" {
			encoded := base64.StdEncoding.EncodeToString(
				[]byte(cfg.GrafanaAPIToken),
			)
			opts.headers["Authorization"] = "Basic " + encoded
		}
	}

	return opts
}

func copyHeaders(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func initGRPCExporters(ctx context.Context, opts exportOpts, res *resource.Resource) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider, *sdklog.LoggerProvider, error) {
	dialOpts := []grpc.DialOption{}
	if opts.insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}
	if len(opts.headers) > 0 {
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(headerInterceptor(opts.headers)))
	}

	traceOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(opts.endpoint),
		otlptracegrpc.WithDialOption(dialOpts...),
	}
	metricOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(opts.endpoint),
		otlpmetricgrpc.WithDialOption(dialOpts...),
	}
	if sel := temporalitySelector(opts.metricsTemporality); sel != nil {
		metricOpts = append(metricOpts, otlpmetricgrpc.WithTemporalitySelector(sel))
	}
	logOpts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(opts.endpoint),
		otlploggrpc.WithDialOption(dialOpts...),
	}

	traceExp, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating gRPC trace exporter: %w", err)
	}

	metricExp, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating gRPC metric exporter: %w", err)
	}

	logExp, err := otlploggrpc.New(ctx, logOpts...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating gRPC log exporter: %w", err)
	}

	tp, mp, lp := buildProviders(res, traceExp, metricExp, logExp)
	return tp, mp, lp, nil
}

// headerInterceptor injects headers into every gRPC call as metadata.
func headerInterceptor(headers map[string]string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md := metadata.New(headers)
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func initHTTPExporters(ctx context.Context, opts exportOpts, res *resource.Resource, isGrafana bool) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider, *sdklog.LoggerProvider, error) {
	traceOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(opts.endpoint),
	}
	metricOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(opts.endpoint),
	}
	if sel := temporalitySelector(opts.metricsTemporality); sel != nil {
		metricOpts = append(metricOpts, otlpmetrichttp.WithTemporalitySelector(sel))
	}
	logOpts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(opts.endpoint),
	}

	if len(opts.headers) > 0 {
		traceOpts = append(traceOpts, otlptracehttp.WithHeaders(opts.headers))
		metricOpts = append(metricOpts, otlpmetrichttp.WithHeaders(opts.headers))
		logOpts = append(logOpts, otlploghttp.WithHeaders(opts.headers))
	}

	if opts.insecure {
		traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
		metricOpts = append(metricOpts, otlpmetrichttp.WithInsecure())
		logOpts = append(logOpts, otlploghttp.WithInsecure())
	} else {
		traceOpts = append(traceOpts, otlptracehttp.WithTLSClientConfig(&tls.Config{}))
		metricOpts = append(metricOpts, otlpmetrichttp.WithTLSClientConfig(&tls.Config{}))
		logOpts = append(logOpts, otlploghttp.WithTLSClientConfig(&tls.Config{}))
	}

	if isGrafana {
		traceOpts = append(traceOpts, otlptracehttp.WithURLPath("/otlp/v1/traces"))
		metricOpts = append(metricOpts, otlpmetrichttp.WithURLPath("/otlp/v1/metrics"))
		logOpts = append(logOpts, otlploghttp.WithURLPath("/otlp/v1/logs"))
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

func temporalitySelector(preference string) func(sdkmetric.InstrumentKind) metricdata.Temporality {
	if strings.EqualFold(preference, "cumulative") {
		return func(sdkmetric.InstrumentKind) metricdata.Temporality {
			return metricdata.CumulativeTemporality
		}
	}
	return nil
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
