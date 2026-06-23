#### Change 1 : Add `pkg/schema/dsconfig.json`

add the `dsconfig.json` file to the specified location. ( either manually authored / copy the already baselined ones from plugin-ui.

```json
{
  "$schema": "https://raw.githubusercontent.com/grafana/dsconfig/refs/heads/main/dsconfig/schema.json",
    "schemaVersion": "v1",
    "pluginType": "grafana-athena-datasource",
    "pluginName": "Amazon Athena",
      ....
}
```

### Change 2 : Add `pkg/schema/dsconfig_test.go`

```go
package schema_test

import (
	_ "embed"
	"testing"

	"github.com/grafana/athena-datasource/pkg/athena/models"
	"github.com/grafana/dsconfig/schema"
)

//go:embed dsconfig.json
var configSchemaJSON []byte

//go:generate go test -run TestPlugin -generateArtifacts
func TestPlugin(t *testing.T) {
	schema.RunPluginTests(t, schema.PluginUnderTest{
		ID:                "grafana-athena-datasource",
		ConfigSchemaJSON:  configSchemaJSON,
		SettingsJSONModel: models.AthenaDataSourceSettings{},
		SecureKeys:        []string{"accessKey", "secretKey", "sessionToken", "proxyPassword"},
	})
}
```

### Change 3 : Add / Update `webpack.config.ts` in the root folder

Add/Update the `webpack.config.ts` in the root folder

```ts
import { type Configuration } from "webpack";
import { merge } from "webpack-merge";
import CopyWebpackPlugin from "copy-webpack-plugin";
import grafanaConfig, { type Env } from "./.config/webpack/webpack.config";

const config = async (env: Env): Promise<Configuration> => {
  const baseConfig = await grafanaConfig(env);
  return merge(baseConfig, {
    plugins: [
      new CopyWebpackPlugin({
        patterns: [
          {
            from: "../pkg/schema/dsconfig.json",
            to: "./schema/dsconfig.json",
            noErrorOnMissing: true,
          },
          {
            from: "../pkg/schema/schema.gen.json",
            to: "./schema/v0alpha1.json",
            noErrorOnMissing: true,
          },
          {
            from: "../pkg/schema/settings.gen.json",
            to: "./schema/v0alpha1/settings.json",
            noErrorOnMissing: true,
          },
          {
            from: "../pkg/schema/settings.examples.gen.json",
            to: "./schema/v0alpha1/settings.examples.json",
            noErrorOnMissing: true,
          },
        ],
      }),
    ],
  });
};

export default config;
```

### Change 4 : Update the package.json script

```diff
-    "build": "webpack -c ./.config/webpack/webpack.config.ts --env production",
+    "build": "webpack -c ./webpack.config.ts --env production",
-    "dev": "webpack -w -c ./.config/webpack/webpack.config.ts --env development",
+    "dev": "webpack -w -c ./webpack.config.ts --env development",
```
