export type alertmanagerConfig = {
  jsonData: {
    
    implementation?: AlertManagerImplementation;
    
    handleGrafanaManagedAlerts?: boolean;
  };
  secureJsonData: {};
};

export type AlertManagerImplementation = "cortex" | "mimir" | "prometheus";
