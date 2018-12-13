import api from "../api"
import { stringToSelectOpt } from "./reactSelectHelpers"
import {
  asyncDataTableFilters,
  IAsyncDataTableFilterProps,
  ecsRunStatuses,
} from "../.."

const historyTableFilters: { [key: string]: IAsyncDataTableFilterProps } = {
  alias: {
    displayName: "Alias",
    type: asyncDataTableFilters.SELECT,
    description: "Search by task alias.",
    name: "alias",
    filterProps: {
      options: [],
      isMulti: true,
      isCreatable: true,
    },
  },
  status: {
    displayName: "Run Status",
    type: asyncDataTableFilters.SELECT,
    name: "status",
    filterProps: {
      options: Object.values(ecsRunStatuses)
        .filter(
          v =>
            v !== ecsRunStatuses.FAILED &&
            v !== ecsRunStatuses.SUCCESS &&
            v !== ecsRunStatuses.STOPPED &&
            v !== ecsRunStatuses.NEEDS_RETRY
        )
        .map(stringToSelectOpt),
      isMulti: true,
    },
    description: "Search by run status.",
  },
  cluster_name: {
    displayName: "Cluster Name",
    type: asyncDataTableFilters.SELECT,
    description: "Search runs running on a specific cluster.",
    name: "cluster_name",
    filterProps: {
      shouldRequestOptions: true,
      requestOptionsFn: api.getClusters,
    },
  },
  env: {
    name: "env",
    displayName: "Environment Variables",
    type: asyncDataTableFilters.KV,
    description: "Search environemnt variables",
    filterProps: {
      keyField: "name",
      valueField: "value",
    },
  },
  started_at_since: {
    name: "started_at_since",
    displayName: "Started At Since",
    type: asyncDataTableFilters.INPUT,
    description: "Filter by runs that started since a certain time (ISO8601)",
  },
  started_at_until: {
    name: "started_at_until",
    displayName: "Started At Until",
    type: asyncDataTableFilters.INPUT,
    description: "Filter by runs that started before a certain time (ISO8601)",
  },
  finished_at_since: {
    name: "finished_at_since",
    displayName: "Finished At Since",
    type: asyncDataTableFilters.INPUT,
    description: "Filter by runs that ended after a certain time (ISO8601)",
  },
  finished_at_until: {
    name: "finished_at_until",
    displayName: "Finished At Until",
    type: asyncDataTableFilters.INPUT,
    description: "Filter by runs that ended before a certain time (ISO8601)",
  },
}

export default historyTableFilters
