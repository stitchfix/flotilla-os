const FLOTILLA_API = process.env.FLOTILLA_API || "FILL_ME_IN"
const CLUSTERS_API = process.env.CLUSTERS_API || "config/clusters.json"
const DOCKER_REPOSITORY_HOST = process.env.DOCKER_REPOSITORY_HOST || ""
const DEFAULT_CLUSTER = process.env.DEFAULT_CLUSTER || "default"

export default {
  FLOTILLA_API,
  CLUSTERS_API,
  DOCKER_REPOSITORY_HOST,
  DEFAULT_CLUSTER,
}
