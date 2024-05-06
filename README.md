# sensu-rri-write Asset

Creating releases for sensu is handled by GitHub Actions.

- run `./publish_release.sh v1.x.y`
- run the sensu pipeline (or apply `sensu/asset.yaml` via sensuctl for all relevant namespaces)
