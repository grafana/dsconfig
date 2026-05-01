export type alertmanagerConfig = {
  jsonData: {
    /** OPTIONAL. Alertmanager implementation type. */
    implementation?: AlertManagerImplementation;
    /** OPTIONAL. Whether to handle Grafana-managed alerts. */
    handleGrafanaManagedAlerts?: boolean;
  };
  secureJsonData: {};
};

export type AlertManagerImplementation = "cortex" | "mimir" | "prometheus";
