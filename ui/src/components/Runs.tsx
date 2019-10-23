import * as React from "react"
import { Link } from "react-router-dom"
import { get, omit } from "lodash"
import { DebounceInput } from "react-debounce-input"
import ListRequest, { ChildProps as ListRequestChildProps } from "./ListRequest"
import api from "../api"
import {
  ListRunParams,
  ListRunResponse,
  SortOrder,
  Run,
  RunStatus,
} from "../types"
import pageToOffsetLimit from "../helpers/pageToOffsetLimit"
import Table from "./Table"
import ViewHeader from "./ViewHeader"
import ListFiltersDropdown from "./ListFiltersDropdown"
import Pagination from "./Pagination"
import GenericMultiSelect from "./GenericMultiSelect"
import RunStatusSelect from "./RunStatusSelect"
import { FormGroup, Classes, Spinner } from "@blueprintjs/core"
import { PAGE_SIZE } from "../constants"
import { RequestStatus } from "./Request"
import ErrorCallout from "./ErrorCallout"

export const initialQuery = {
  page: 1,
  sort_by: "started_at",
  order: SortOrder.DESC,
  status: [RunStatus.PENDING, RunStatus.QUEUED, RunStatus.RUNNING],
}
export type Props = ListRequestChildProps<
  ListRunResponse,
  { params: ListRunParams }
>

export const Runs: React.FunctionComponent<Props> = ({
  data,
  updateSort,
  currentSortKey,
  currentSortOrder,
  updatePage,
  currentPage,
  query,
  updateFilter,
  isLoading,
  requestStatus,
  error,
}) => {
  let content: React.ReactNode

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
            status: {
              displayName: "Status",
              render: (r: Run) => r.status,
              isSortable: true,
            },
            started_at: {
              displayName: "Started At",
              render: (r: Run) => r.started_at,
              isSortable: true,
            },
            run_id: {
              displayName: "Run ID",
              render: (r: Run) => (
                <Link to={`/runs/${r.run_id}`}>{r.run_id}</Link>
              ),
              isSortable: true,
            },
            alias: {
              displayName: "Alias",
              render: (r: Run) => r.alias,
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
      <ViewHeader
        breadcrumbs={[
          { text: "Runs", href: "/runs?page=1&sort_by=started_at&order=desc" },
        ]}
      />
      <div className="flotilla-list-utils-container">
        <FormGroup label="Alias" helperText="Search by task alias.">
          <GenericMultiSelect
            value={get(query, "alias", [])}
            onChange={(value: string[]) => {
              updateFilter("alias", value)
            }}
          />
        </FormGroup>
        <FormGroup label="Run Status" helperText="Search by run status.">
          <RunStatusSelect
            value={get(query, "status", [])}
            onChange={(value: string[]) => {
              updateFilter("status", value)
            }}
          />
        </FormGroup>
        <ListFiltersDropdown>
          <FormGroup label="Cluster" helperText="Search by ECS cluster.">
            <GenericMultiSelect
              value={get(query, "cluster", [])}
              onChange={(value: string[]) => {
                updateFilter("cluster", value)
              }}
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

const ConnectedRuns: React.FunctionComponent<{}> = () => (
  <ListRequest<ListRunResponse, { params: ListRunParams }>
    requestFn={api.listRun}
    initialQuery={initialQuery}
    getRequestArgs={params => ({
      params: {
        ...omit(params, "page"),
        ...pageToOffsetLimit({
          page: get(params, "page", 1),
          limit: PAGE_SIZE,
        }),
      },
    })}
  >
    {props => <Runs {...props} />}
  </ListRequest>
)

export default ConnectedRuns
