import moment from "moment"
import api from "../api"
import { stringToSelectOpt } from "./reactSelectHelpers"
import {
  flotillaUIAsyncDataTableFilters,
  IFlotillaUIAsyncDataTableFilterProps,
  flotillaRunStatuses,
} from "../.."

const isValidISOString = (value: string): boolean =>
  value !== "" ? moment.utc(value).isValid() : true

const historyTableFilters: {
  [key: string]: IFlotillaUIAsyncDataTableFilterProps
} = {
  alias: {
    displayName: "Alias",
    type: flotillaUIAsyncDataTableFilters.SELECT,
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
    type: flotillaUIAsyncDataTableFilters.SELECT,
    name: "status",
    filterProps: {
      options: Object.values(flotillaRunStatuses)
        .filter(
          v =>
            v !== flotillaRunStatuses.FAILED &&
            v !== flotillaRunStatuses.SUCCESS &&
            v !== flotillaRunStatuses.STOPPED &&
            v !== flotillaRunStatuses.NEEDS_RETRY
        )
        .map(stringToSelectOpt),
      isMulti: true,
    },
    description: "Search by run status.",
  },
  cluster_name: {
    displayName: "Cluster Name",
    type: flotillaUIAsyncDataTableFilters.SELECT,
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
    type: flotillaUIAsyncDataTableFilters.KV,
    description: "Search environemnt variables",
    filterProps: {
      keyField: "name",
      valueField: "value",
    },
  },
  started_at_since: {
    name: "started_at_since",
    displayName: "Started At Since",
    type: flotillaUIAsyncDataTableFilters.INPUT,
    description: "Filter by runs that started since a certain time (ISO8601)",
    filterProps: { validate: isValidISOString },
  },
  started_at_until: {
    name: "started_at_until",
    displayName: "Started At Until",
    type: flotillaUIAsyncDataTableFilters.INPUT,
    description: "Filter by runs that started before a certain time (ISO8601)",
    filterProps: { validate: isValidISOString },
  },
  finished_at_since: {
    name: "finished_at_since",
    displayName: "Finished At Since",
    type: flotillaUIAsyncDataTableFilters.INPUT,
    description: "Filter by runs that ended after a certain time (ISO8601)",
    filterProps: { validate: isValidISOString },
  },
  finished_at_until: {
    name: "finished_at_until",
    displayName: "Finished At Until",
    type: flotillaUIAsyncDataTableFilters.INPUT,
    description: "Filter by runs that ended before a certain time (ISO8601)",
    filterProps: { validate: isValidISOString },
  },
}

export default historyTableFilters
