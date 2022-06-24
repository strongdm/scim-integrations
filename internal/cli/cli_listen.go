package cli

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"scim-integrations/internal/repository"
	"scim-integrations/internal/repository/query"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const intervalSeconds = 60

var (
	exposeMetricsCommand = &command{
		Name: "expose-metrics",
		Exec: exposeMetrics,
	}
	metricsPort       = 2112
	createdUsersGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_last_execution_created_users_count",
		Help: "Count of Users prepared to be created in SDM",
	})
	updatedUsersGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_last_execution_updated_users_count",
		Help: "Count of Users prepared to be updated in SDM",
	})
	deletedUsersGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_last_execution_deleted_users_count",
		Help: "Count of Users prepared to be deleted in SDM",
	})
	createdGroupsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_last_execution_created_groups_count",
		Help: "Count of Groups prepared to be created in SDM",
	})
	updatedGroupsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_last_execution_updated_groups_count",
		Help: "Count of Groups prepared to be updated in SDM",
	})
	deletedGroupsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_last_execution_deleted_groups_count",
		Help: "Count of Groups prepared to be deleted in SDM",
	})
	lastRunSucceededGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_last_execution_succeeded",
		Help: "Last execution succeeded status",
	})
	consecutiveErrorsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_total_consecutive_errors_count",
		Help: "Count of errors occurred during the synchronizing process",
	})
)

func exposeMetrics() error {
	http.Handle("/metrics", promhttp.Handler())
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", metricsPort))
	if err != nil {
		return fmt.Errorf("An error occurred running prometheus metrics: %v", err)
	}
	fmt.Printf("Prometheus metrics endpoint is listening in http://localhost:%v\n", metricsPort)
	iterateMetrics()
	if err = http.Serve(listener, nil); err != nil {
		return fmt.Errorf("An error occurred running prometheus metrics: %v", err)
	}
	return nil
}

func iterateMetrics() {
	go func() {
		for {
			loadReport()
			time.Sleep(time.Second * intervalSeconds)
		}
	}()
}

func loadReport() {
	reports, err := repository.NewReportRepository().Select(&query.SelectFilter{OrderBy: "id desc"})
	if err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred when collecting report metrics:", err.Error())
		return
	}
	if reports == nil || len(reports) == 0 {
		return
	}
	report := reports[0]
	createdUsersGauge.Set(float64(report.CreatedUsersCount))
	updatedUsersGauge.Set(float64(report.UpdatedUsersCount))
	deletedUsersGauge.Set(float64(report.DeletedUsersCount))
	createdGroupsGauge.Set(float64(report.CreatedGroupsCount))
	updatedGroupsGauge.Set(float64(report.UpdatedGroupsCount))
	deletedGroupsGauge.Set(float64(report.DeletedGroupsCount))
	lastRunSucceededGauge.Set(float64(report.Succeed))
	consecutiveErrorsGauge.Set(float64(getConsecutiveErrorsCount(reports)))
}

func getConsecutiveErrorsCount(reports []*repository.ReportsRow) int {
	var count int
	for _, report := range reports {
		if report.Succeed == 0 {
			break
		}
		count++
	}
	return count
}
