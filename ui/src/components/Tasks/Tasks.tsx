import * as React from "react"
import { Link } from "react-router-dom"
import Helmet from "react-helmet"
import { DebounceInput } from "react-debounce-input"
import { get } from "lodash"
import Navigation from "../Navigation/Navigation"
import ButtonLink from "../styled/ButtonLink"
import StyledField from "../styled/Field"
import { Input } from "../styled/Inputs"
import View from "../styled/View"
import ListRequest, {
  IChildProps as IListRequestChildProps,
} from "../ListRequest/ListRequest"
import DataTable, { IDataTableColumn } from "../DataTable/DataTable"
import api from "../../api"
import config from "../../config"
import { flotillaUIIntents } from "../../types"
import ReactSelectWrapper from "../ReactSelectWrapper/ReactSelectWrapper"
import formConfiguration from "../../helpers/formConfiguration"

const TASKS_TABLE_COLUMNS: { [key: string]: IDataTableColumn } = {
  run_task: {
    allowSort: false,
    displayName: "Run",
    render: item => (
      <ButtonLink to={`/tasks/${item.definition_id}/run`}>Run</ButtonLink>
    ),
    width: 0.4,
  },
  alias: {
    allowSort: true,
    displayName: "Alias",
    render: item => (
      <Link to={`/tasks/${item.definition_id}`}>{item.alias}</Link>
    ),
    width: 3,
  },
  group_name: {
    allowSort: true,
    displayName: "Group Name",
    render: item => item.group_name,
    width: 1,
  },
  image: {
    allowSort: true,
    displayName: "Image",
    render: item => item.image.substr(config.IMAGE_PREFIX.length),
    width: 1,
  },
  memory: {
    allowSort: true,
    displayName: "Memory",
    render: item => item.memory,
    width: 0.5,
  },
}

export const Tasks: React.FunctionComponent<IListRequestChildProps> = props => {
  return (
    <View>
      <Helmet>
        <title>Flotilla | Tasks</title>
      </Helmet>
      <Navigation
        actions={[
          {
            isLink: true,
            href: "/tasks/create",
            text: "Create New Task",
            buttonProps: {
              intent: flotillaUIIntents.PRIMARY,
            },
          },
        ]}
      />
      <div
        style={{
          display: "flex",
          flexFlow: "row nowrap",
          justifyContent: "flex-start",
          alignItems: "flex-start",
          width: "100%",
        }}
      >
        <div style={{}}>
          <StyledField label="Alias" description="Search tasks by alias.">
            <DebounceInput
              value={get(props, ["queryParams", "alias"], "")}
              onChange={(evt: React.SyntheticEvent) => {
                const target = evt.target as HTMLInputElement
                props.updateSearch(formConfiguration.alias.key, target.value)
              }}
              element={Input}
              debounceTimeout={500}
              minLength={1}
              type="text"
            />
          </StyledField>
          <StyledField
            label="Group Name"
            description="Search tasks by existing group names."
          >
            <ReactSelectWrapper
              isCreatable
              isMulti={false}
              name="group_name"
              onChange={(value: string | string[]) => {
                props.updateSearch("group_name", value as string)
              }}
              requestOptionsFn={api.getGroups}
              shouldRequestOptions
              value={get(props, ["queryParams", "group_name"], "")}
              // onRequestError
              getOptionsFromResponse={r => r || []}
            />
          </StyledField>
          <StyledField
            label="Image"
            description="Search tasks by Docker image."
          >
            <DebounceInput
              value={get(props, ["queryParams", "image"], "")}
              onChange={(evt: React.SyntheticEvent) => {
                const target = evt.target as HTMLInputElement
                props.updateSearch("image", target.value)
              }}
              element={Input}
              debounceTimeout={500}
              minLength={1}
              type="text"
            />
          </StyledField>
        </div>
        <DataTable
          items={get(props, ["data", "definitions"], [])}
          columns={TASKS_TABLE_COLUMNS}
          onSortableHeaderClick={props.updateSort}
          getItemKey={(_, index) => index}
          currentSortKey={props.currentSortKey}
          currentSortOrder={props.currentSortOrder}
          currentPage={props.currentPage}
        />
      </div>
    </View>
  )
}

const ConnectedTasks: React.FunctionComponent = () => (
  <ListRequest
    getRequestArgs={query => ({ query })}
    initialQuery={{
      page: 1,
      sort_by: "alias",
      order: "asc",
    }}
    limit={50}
    requestFn={api.getTasks}
    shouldContinuouslyFetch={false}
  >
    {(props: IListRequestChildProps) => <Tasks {...props} />}
  </ListRequest>
)

export default ConnectedTasks
