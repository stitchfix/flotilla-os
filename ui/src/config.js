export default {
  /** The default ECS cluster tasks will be executed on. */
  DEFAULT_CLUSTER: process.env.DEFAULT_CLUSTER || "default",

  /** The location of the Flotilla API, e.g. `flotilla.mycompany.com/api/v1` */
  FLOTILLA_API: process.env.FLOTILLA_API || "FILL_ME_IN",

  /** Prefix for Docker images, e.g. `my-docker-repository:4567` */
  IMAGE_PREFIX: process.env.IMAGE_PREFIX || "",

  /** The rate at which run data will be requested in the RunView */
  RUN_REQUEST_INTERVAL_MS: 10000,

  /** The rate at which run logs will be requested in the RunLogs component */
  RUN_LOGS_REQUEST_INTERVAL_MS: 5000,
}
