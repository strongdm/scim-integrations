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
	listenCommand = &command{
		Name: "listen",
		Exec: listen,
	}
	metricsPort        = 2112
	usersToCreateGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_users_to_create_count",
		Help: "Count of Users prepared to be created in SDM",
	})
	usersToUpdateGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_users_to_update_count",
		Help: "Count of Users prepared to be updated in SDM",
	})
	usersToDeleteGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_users_to_delete_count",
		Help: "Count of Users prepared to be deleted in SDM",
	})
	groupsToCreateGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_groups_to_create_count",
		Help: "Count of Groups prepared to be created in SDM",
	})
	groupsToUpdateGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_groups_to_update_count",
		Help: "Count of Groups prepared to be updated in SDM",
	})
	groupsToDeleteGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_groups_to_delete_count",
		Help: "Count of Groups prepared to be deleted in SDM",
	})
	errorsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scim_integrations_errors_count",
		Help: "Count of errors occurred during the synchronizing process",
	})
)

func listen() error {
	http.Handle("/metrics", promhttp.Handler())
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", metricsPort))
	if err != nil {
		return fmt.Errorf("An error was occurred running prometheus metrics: %v", err)
	}
	fmt.Printf("Prometheus metrics endpoint is listening in http://localhost:%v\n", metricsPort)
	iterateMetrics()
	if err = http.Serve(listener, nil); err != nil {
		return fmt.Errorf("An error was occurred running prometheus metrics: %v", err)
	}
	return nil
}

func iterateMetrics() {
	go func() {
		for {
			loadReport()
			loadErr()
			time.Sleep(time.Second * intervalSeconds)
		}
	}()
}

func loadReport() {
	reports, err := repository.NewReportRepository().Select(&query.SelectFilter{Limit: 1, OrderBy: "id desc"})
	if err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred when collecting report metrics:", err.Error())
		return
	}
	if len(reports) == 0 {
		return
	}
	report := reports[0]
	usersToCreateGauge.Set(float64(report.UsersToCreateCount))
	usersToUpdateGauge.Set(float64(report.UsersToUpdateCount))
	usersToDeleteGauge.Set(float64(report.UsersToDeleteCount))
	groupsToCreateGauge.Set(float64(report.GroupsToCreateCount))
	groupsToUpdateGauge.Set(float64(report.GroupsToUpdateCount))
	groupsToDeleteGauge.Set(float64(report.GroupsToDeleteCount))
}

func loadErr() {
	errors, err := repository.NewErrorsRepository().Select(nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred when collecting error metrics:", err.Error())
		return
	}
	errorsGauge.Set(float64(len(errors)))
}
