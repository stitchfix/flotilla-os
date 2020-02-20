import * as React from "react"
import { Link } from "react-router-dom"
import { get, omit, isArray, isString } from "lodash"
import ListRequest, { ChildProps as ListRequestChildProps } from "./ListRequest"
import api from "../api"
import {
  ListTemplateHistoryParams,
  ListTemplateHistoryResponse,
  SortOrder,
  Run,
  RunStatus,
  ExecutionEngine,
} from "../types"
import pageToOffsetLimit from "../helpers/pageToOffsetLimit"
import Table from "./Table"
import { FormGroup, Classes, Spinner, Tag } from "@blueprintjs/core"
import GenericMultiSelect from "./GenericMultiSelect"
import RunStatusSelect from "./RunStatusSelect"
import ListFiltersDropdown from "./ListFiltersDropdown"
import { DebounceInput } from "react-debounce-input"
import Pagination from "./Pagination"
import { PAGE_SIZE } from "../constants"
import { RequestStatus } from "./Request"
import ErrorCallout from "./ErrorCallout"
import RunTag from "./RunTag"
import ISO8601AttributeValue from "./ISO8601AttributeValue"
import EnvQueryFilter from "./EnvQueryFilter"
import Duration from "./Duration"

export const initialQuery = {
  page: 1,
  sort_by: "started_at",
  order: SortOrder.DESC,
}

export type Props = ListRequestChildProps<
  ListTemplateHistoryResponse,
  { params: ListTemplateHistoryParams }
>

export const TemplateHistoryTable: React.FunctionComponent<Props> = ({
  data,
  updateSort,
  currentSortKey,
  currentSortOrder,
  query,
  updateFilter,
  updatePage,
  currentPage,
  isLoading,
  requestStatus,
  error,
}) => {
  let content: React.ReactNode

  // Preprocess `env` query to ensure that it's an array.
  let env: string | string[] = get(query, "env", [])
  if (!isArray(env) && isString(env)) env = [env]

  switch (requestStatus) {
    case RequestStatus.ERROR:
      content = <ErrorCallout error={error} />
      break
    case RequestStatus.READY:
      content = (
        <Table<Run>
          items={get(data, "history", [])}
          getItemKey={(r: Run) => r.run_id}
          updateSort={updateSort}
          currentSortKey={currentSortKey}
          currentSortOrder={currentSortOrder}
          columns={{
            run_id: {
              displayName: "Run ID",
              render: (r: Run) => (
                <Link to={`/runs/${r.run_id}`}>{r.run_id}</Link>
              ),
              isSortable: true,
            },
            status: {
              displayName: "Status",
              render: (r: Run) => <RunTag {...r}></RunTag>,
              isSortable: true,
            },
            engine: {
              displayName: "Engine",
              render: (r: Run) => <Tag>{r.engine}</Tag>,
              isSortable: false,
            },
            duration: {
              displayName: "Duration",
              render: (r: Run) =>
                r.started_at ? (
                  <Duration
                    start={r.started_at}
                    end={r.finished_at}
                    isActive={r.status !== RunStatus.STOPPED}
                  />
                ) : (
                  "-"
                ),
              isSortable: false,
            },
            started_at: {
              displayName: "Started At",
              render: (r: Run) => (
                <ISO8601AttributeValue
                  time={r.started_at}
                ></ISO8601AttributeValue>
              ),
              isSortable: true,
            },
            finished_at: {
              displayName: "Finished At",
              render: (r: Run) => (
                <ISO8601AttributeValue
                  time={r.finished_at}
                ></ISO8601AttributeValue>
              ),
              isSortable: true,
            },
            cluster: {
              displayName: "Cluster",
              render: (r: Run) =>
                r.engine === ExecutionEngine.EKS ? "-" : r.cluster,
              isSortable: false,
            },
          }}
        />
      )
      break
    case RequestStatus.NOT_READY:
    default:
      content = <Spinner />
      break
  }

  return (
    <>
      <div className="flotilla-list-utils-container">
        <FormGroup label="Run Status" helperText="Search by run status.">
          <RunStatusSelect
            value={get(query, "status", [])}
            onChange={(value: string[]) => {
              updateFilter("status", value)
            }}
            isDisabled={false}
          />
        </FormGroup>
        <ListFiltersDropdown>
          <EnvQueryFilter
            value={env}
            onChange={value => {
              updateFilter("env", value)
            }}
          />
          <FormGroup label="Cluster" helperText="Search by ECS cluster.">
            <GenericMultiSelect
              value={get(query, "cluster_name", [])}
              onChange={(value: string[]) => {
                updateFilter("cluster_name", value)
              }}
              isDisabled={false}
            />
          </FormGroup>
          <FormGroup
            label="Started At Since"
            helperText="Enter a valid ISO8601 string."
          >
            <DebounceInput
              style={{ flex: 1 }}
              className={Classes.INPUT}
              debounceTimeout={500}
              value={get(query, "started_at_since", "")}
              onChange={(evt: React.ChangeEvent<HTMLInputElement>) => {
                updateFilter("started_at_since", evt.target.value)
              }}
            />
          </FormGroup>
          <FormGroup
            label="Started At Until"
            helperText="Enter a valid ISO8601 string."
          >
            <DebounceInput
              style={{ flex: 1 }}
              className={Classes.INPUT}
              debounceTimeout={500}
              value={get(query, "started_at_until", "")}
              onChange={(evt: React.ChangeEvent<HTMLInputElement>) => {
                updateFilter("started_at_until", evt.target.value)
              }}
            />
          </FormGroup>
          <FormGroup
            label="Finished At Since"
            helperText="Enter a valid ISO8601 string."
          >
            <DebounceInput
              style={{ flex: 1 }}
              className={Classes.INPUT}
              debounceTimeout={500}
              value={get(query, "finished_at_since", "")}
              onChange={(evt: React.ChangeEvent<HTMLInputElement>) => {
                updateFilter("finished_at_since", evt.target.value)
              }}
            />
          </FormGroup>
          <FormGroup
            label="Finished At Until"
            helperText="Enter a valid ISO8601 string."
          >
            <DebounceInput
              style={{ flex: 1 }}
              className={Classes.INPUT}
              debounceTimeout={500}
              value={get(query, "finished_at_until", "")}
              onChange={(evt: React.ChangeEvent<HTMLInputElement>) => {
                updateFilter("finished_at_until", evt.target.value)
              }}
            />
          </FormGroup>
        </ListFiltersDropdown>
        <Pagination
          updatePage={updatePage}
          currentPage={currentPage}
          isLoading={isLoading}
          pageSize={PAGE_SIZE}
          numItems={data ? data.total : 0}
        />
      </div>
      {content}
    </>
  )
}

const ConnectedTaskRuns: React.FunctionComponent<{ templateID: string }> = ({
  templateID,
}) => (
  <ListRequest<
    ListTemplateHistoryResponse,
    { templateID: string; params: ListTemplateHistoryParams }
  >
    requestFn={api.listTemplateHistory}
    initialQuery={initialQuery}
    // @TODO: this function should be extracted and tested.
    getRequestArgs={params => ({
      templateID,
      params: {
        ...omit(params, "page"),
        ...pageToOffsetLimit({
          page: get(params, "page", 1),
          limit: PAGE_SIZE,
        }),
      },
    })}
  >
    {props => <TemplateHistoryTable {...props} />}
  </ListRequest>
)

export default ConnectedTaskRuns
