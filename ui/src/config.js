const FLOTILLA_API = process.env.FLOTILLA_API || "FILL_ME_IN"
const CLUSTERS_API = process.env.CLUSTERS_API || "config/clusters.json"
const DEFAULT_CLUSTER = process.env.DEFAULT_CLUSTER || "default"

export default {
  FLOTILLA_API,
  CLUSTERS_API,
  DEFAULT_CLUSTER,
}
