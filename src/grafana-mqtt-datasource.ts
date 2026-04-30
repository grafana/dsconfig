export type grafanaMqttDatasourceConfig = {
  jsonData: {
    /** REQUIRED. MQTT broker URI (e.g., tcp://localhost:1883). */
    uri: string;
    /** OPTIONAL. MQTT username. */
    username?: string;
    /** OPTIONAL. MQTT client ID. */
    clientID?: string;
    /** OPTIONAL. Enable TLS client auth. */
    tlsAuth: boolean;
    /** OPTIONAL. Enable TLS CA cert verification. */
    tlsAuthWithCACert: boolean;
    /** OPTIONAL. Skip TLS certificate verification. */
    tlsSkipVerify: boolean;
  };
  secureJsonData: {
    /** OPTIONAL. MQTT password. */
    password?: string;
    /** OPTIONAL. TLS CA certificate. */
    tlsCACert?: string;
    /** OPTIONAL. TLS client key. */
    tlsClientKey?: string;
    /** OPTIONAL. TLS client certificate. */
    tlsClientCert?: string;
  };
};
