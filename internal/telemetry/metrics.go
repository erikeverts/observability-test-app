package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type AppMetrics struct {
	OrdersCreated    metric.Int64Counter
	ProductViews     metric.Int64Counter
	ChaosErrorsTotal metric.Int64Counter
	ChaosLatencyTotal metric.Int64Counter
}

func NewAppMetrics() (*AppMetrics, error) {
	meter := otel.Meter("observability-test-app")

	ordersCreated, err := meter.Int64Counter("app.orders.created",
		metric.WithDescription("Total orders created"),
	)
	if err != nil {
		return nil, err
	}

	productViews, err := meter.Int64Counter("app.products.views",
		metric.WithDescription("Total product views"),
	)
	if err != nil {
		return nil, err
	}

	chaosErrors, err := meter.Int64Counter("app.chaos.errors_injected",
		metric.WithDescription("Total chaos errors injected"),
	)
	if err != nil {
		return nil, err
	}

	chaosLatency, err := meter.Int64Counter("app.chaos.latency_injected",
		metric.WithDescription("Total chaos latency injections"),
	)
	if err != nil {
		return nil, err
	}

	return &AppMetrics{
		OrdersCreated:     ordersCreated,
		ProductViews:      productViews,
		ChaosErrorsTotal:  chaosErrors,
		ChaosLatencyTotal: chaosLatency,
	}, nil
}
