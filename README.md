# Grafana plugins schema SDK

> [!WARNING]  
> This is an experimental project and still under active development

# Releasing a new dsconfig SDK version

Go to https://github.com/grafana/dsconfig/tags to verify the latest version that has been published.

From the root folder run, with the above version bumped:

`git tag dsconfig/v<x.x.x> && git tag schema/v<x.x.x>`

Then push the tags:

`git push --tags`
