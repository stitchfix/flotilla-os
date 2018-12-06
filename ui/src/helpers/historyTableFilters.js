import { asyncDataTableFilterTypes } from "../components/AsyncDataTable/AsyncDataTableFilter"
import runStatusTypes from "./runStatusTypes"
import api from "../api"
import { stringToSelectOpt } from "./reactSelectHelpers"

const historyTableFilters = {
  alias: {
    displayName: "Alias",
    type: asyncDataTableFilterTypes.SELECT,
    options: [],
    description: "Search by task alias.",
    isMulti: true,
    isCreatable: true,
  },
  status: {
    displayName: "Run Status",
    type: asyncDataTableFilterTypes.SELECT,
    options: Object.values(runStatusTypes)
      .filter(
        v =>
          v !== runStatusTypes.failed &&
          v !== runStatusTypes.success &&
          v !== runStatusTypes.stopped &&
          v !== runStatusTypes.needs_retry
      )
      .map(stringToSelectOpt),
    description: "Search by run status.",
    isMulti: true,
  },
  cluster_name: {
    displayName: "Cluster Name",
    type: asyncDataTableFilterTypes.SELECT,
    description: "Search runs running on a specific cluster.",
    shouldRequestOptions: true,
    requestOptionsFn: api.getClusters,
  },
  env: {
    displayName: "Environment Variables",
    type: asyncDataTableFilterTypes.KV,
    description: "Search environemnt variables",
  },
  started_at_since: {
    displayName: "Started At Since",
    type: asyncDataTableFilterTypes.INPUT,
    description: "Filter by runs that started since a certain time (ISO8601)",
  },
  started_at_until: {
    displayName: "Started At Until",
    type: asyncDataTableFilterTypes.INPUT,
    description: "Filter by runs that started before a certain time (ISO8601)",
  },
  finished_at_since: {
    displayName: "Finished At Since",
    type: asyncDataTableFilterTypes.INPUT,
    description: "Filter by runs that ended after a certain time (ISO8601)",
  },
  finished_at_until: {
    displayName: "Finished At Until",
    type: asyncDataTableFilterTypes.INPUT,
    description: "Filter by runs that ended before a certain time (ISO8601)",
  },
}

export default historyTableFilters
