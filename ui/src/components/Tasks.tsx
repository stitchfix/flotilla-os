import * as React from "react"
import { Link } from "react-router-dom"
import { get, omit } from "lodash"
import { DebounceInput } from "react-debounce-input"
import { FormGroup, Classes, Spinner } from "@blueprintjs/core"
import ListRequest, { ChildProps as ListRequestChildProps } from "./ListRequest"
import api from "../api"
import { ListTaskParams, ListTaskResponse, SortOrder, Task } from "../types"
import pageToOffsetLimit from "../helpers/pageToOffsetLimit"
import Table from "./Table"
import Pagination from "./Pagination"
import GroupNameSelect from "./GroupNameSelect"
import ViewHeader from "./ViewHeader"
import ListFiltersDropdown from "./ListFiltersDropdown"
import { pageSize } from "../constants"
import { RequestStatus } from "./Request"
import ErrorCallout from "./ErrorCallout"

export const initialQuery = {
  page: 1,
  sort_by: "alias",
  order: SortOrder.ASC,
}

export type Props = ListRequestChildProps<
  ListTaskResponse,
  { params: ListTaskParams }
>

export const Tasks: React.FunctionComponent<Props> = props => {
  const {
    query,
    data,
    updateFilter,
    updatePage,
    updateSort,
    currentPage,
    currentSortKey,
    currentSortOrder,
    isLoading,
    requestStatus,
    error,
  } = props

  let content: React.ReactNode

  switch (requestStatus) {
    case RequestStatus.ERROR:
      content = <ErrorCallout error={error} />
      break
    case RequestStatus.READY:
      content = (
        <Table<Task>
          items={get(data, "definitions", [])}
          getItemKey={(task: Task) => task.definition_id}
          updateSort={updateSort}
          currentSortKey={currentSortKey}
          currentSortOrder={currentSortOrder}
          columns={{
            alias: {
              displayName: "Alias",
              render: (item: Task) => (
                <Link to={`/tasks/${item.definition_id}`}>{item.alias}</Link>
              ),
              isSortable: true,
            },
            group_name: {
              displayName: "Group Name",
              render: (item: Task) => item.group_name,
              isSortable: true,
            },
            image: {
              displayName: "Image",
              render: (item: Task) => item.image,
              isSortable: true,
            },
            memory: {
              displayName: "Memory (MB)",
              render: (item: Task) => item.memory,
              isSortable: true,
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
        breadcrumbs={[{ text: "Tasks", href: "/tasks" }]}
        buttons={
          <Link
            className={`${Classes.BUTTON} ${Classes.INTENT_PRIMARY}`}
            to={`/tasks/create`}
          >
            Create Task
          </Link>
        }
      />
      <div className="flotilla-list-utils-container">
        <FormGroup label="Alias" helperText="Search by task alias.">
          <DebounceInput
            id="tasksAliasFilter"
            style={{ flex: 1 }}
            className="bp3-input flotilla-list-utils-searchbar"
            debounceTimeout={500}
            value={get(query, "alias", "")}
            onChange={(evt: React.ChangeEvent<HTMLInputElement>) => {
              updateFilter("alias", evt.target.value)
            }}
            placeholder="Search by task alias..."
          />
        </FormGroup>
        <ListFiltersDropdown>
          <FormGroup label="Group Name" helperText="Search by group name.">
            <GroupNameSelect
              value={get(query, "group_name", "")}
              onChange={value => {
                updateFilter("group_name", value)
              }}
            />
          </FormGroup>
          <FormGroup label="Image" helperText="Search by Docker image.">
            <DebounceInput
              id="tasksImageFilter"
              className="bp3-input"
              debounceTimeout={500}
              value={get(query, "image", "")}
              onChange={(evt: React.ChangeEvent<HTMLInputElement>) => {
                updateFilter("image", evt.target.value)
              }}
            />
          </FormGroup>
        </ListFiltersDropdown>
        <Pagination
          updatePage={updatePage}
          currentPage={currentPage}
          isLoading={isLoading}
          pageSize={pageSize}
          numItems={data ? data.total : 0}
        />
      </div>
      {content}
    </>
  )
}

const ConnectedTasks: React.FunctionComponent = () => (
  <ListRequest<ListTaskResponse, { params: ListTaskParams }>
    requestFn={api.listTasks}
    initialQuery={initialQuery}
    getRequestArgs={params => ({
      params: {
        ...omit(params, "page"),
        ...pageToOffsetLimit({ page: get(params, "page", 1), limit: pageSize }),
      },
    })}
  >
    {props => <Tasks {...props} />}
  </ListRequest>
)

export default ConnectedTasks
