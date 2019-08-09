import * as React from "react"
import { Link } from "react-router-dom"
import { get, omit } from "lodash"
import ListRequest, { ChildProps as ListRequestChildProps } from "./ListRequest"
import api from "../api"
import {
  ListTaskRunsParams,
  ListTaskRunsResponse,
  SortOrder,
  Run,
} from "../types"
import pageToOffsetLimit from "../helpers/pageToOffsetLimit"
import Table from "./Table"
import { FormGroup, Classes } from "@blueprintjs/core"
import GenericMultiSelect from "./GenericMultiSelect"
import RunStatusSelect from "./RunStatusSelect"
import ListFiltersDropdown from "./ListFiltersDropdown"
import { DebounceInput } from "react-debounce-input"
import Pagination from "./Pagination"

export const TaskRuns: React.FunctionComponent<
  ListRequestChildProps<ListTaskRunsResponse, { params: ListTaskRunsParams }>
> = ({
  data,
  updateSort,
  currentSortKey,
  currentSortOrder,
  query,
  updateFilter,
  updatePage,
  currentPage,
}) => (
  <>
    <div className="flotilla-list-utils-container">
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
        totalPages={data ? data.total : 1}
      />
    </div>
    <Table<Run>
      items={get(data, "history", [])}
      getItemKey={(r: Run) => r.run_id}
      updateSort={updateSort}
      currentSortKey={currentSortKey}
      currentSortOrder={currentSortOrder}
      columns={{
        run_id: {
          displayName: "Run ID",
          render: (r: Run) => <Link to={`/runs/${r.run_id}`}>{r.run_id}</Link>,
          isSortable: true,
        },
        status: {
          displayName: "Status",
          render: (r: Run) => r.status,
          isSortable: true,
        },
        started_at: {
          displayName: "Started At",
          render: (r: Run) => r.started_at || "-",
          isSortable: true,
        },
        finished_at: {
          displayName: "Finished At",
          render: (r: Run) => r.finished_at || "-",
          isSortable: true,
        },
        cluster: {
          displayName: "Cluster",
          render: (r: Run) => r.cluster,
          isSortable: false,
        },
      }}
    />
  </>
)

const ConnectedTaskRuns: React.FunctionComponent<{ definitionID: string }> = ({
  definitionID,
}) => (
  <ListRequest<
    ListTaskRunsResponse,
    { definitionID: string; params: ListTaskRunsParams }
  >
    requestFn={api.listTaskRuns}
    initialQuery={{
      page: 1,
      sort_by: "started_at",
      order: SortOrder.DESC,
    }}
    getRequestArgs={params => ({
      definitionID,
      params: {
        ...omit(params, "page"),
        ...pageToOffsetLimit({ page: get(params, "page", 1), limit: 20 }),
      },
    })}
  >
    {props => <TaskRuns {...props} />}
  </ListRequest>
)

export default ConnectedTaskRuns
