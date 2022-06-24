# CONFIGURE PROMETHEUS

When using Docker, a prometheus metrics endpoint is enabled listening in the port `2112`. You can use the following gauges to monitor the last synchronized report:

- `scim_integrations_last_execution_created_users_count` - count of the planned users to be created
- `scim_integrations_last_execution_updated_users_count` - count of the planned users to be updated
- `scim_integrations_last_execution_deleted_users_count` - count of the planned users to be deleted
- `scim_integrations_last_execution_created_groups_count` - count of the planned groups to be created
- `scim_integrations_last_execution_updated_groups_count` - count of the planned groups to be updated
- `scim_integrations_last_execution_deleted_groups_count` - count of the planned groups to be deleted
- `scim_integrations_last_execution_succeeded` - status of the last execution. Zero (0) indicates that the application was ran successfully, and one (1) indicates that the application was failed
- `scim_integrations_total_consecutive_errors_count` - count of the consecutive application failures

If you're using the [docker-compose-prometheus.yml](../docker-compose-prometheus.yml) file, you can go to the `Grafana` and browse to the `SCIM Integration Stats` dashboard an see an usage example of this gauge properties.

To run the example environment with Prometheus and Grafana, just run `docker-compose -f docker-compose-prometheus.yml up`
