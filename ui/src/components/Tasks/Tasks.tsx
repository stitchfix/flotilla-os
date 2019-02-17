import * as React from "react"
import { Link } from "react-router-dom"
import Helmet from "react-helmet"
import { get } from "lodash"
import AsyncDataTable from "../AsyncDataTable/AsyncDataTable"
import Navigation from "../Navigation/Navigation"
import ButtonLink from "../styled/ButtonLink"
import View from "../styled/View"
import api from "../../api"
import config from "../../config"
import { flotillaUIIntents, flotillaUIAsyncDataTableFilters } from "../../types"

class Tasks extends React.PureComponent {
  render() {
    return (
      <View>
        <Helmet>
          <title>Tasks</title>
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
        <AsyncDataTable
          isView
          getRequestArgs={query => ({ query })}
          limit={50}
          shouldContinuouslyFetch={false}
          getItemKey={(item, i) => get(item, "definition_id", i)}
          requestFn={api.getTasks}
          columns={{
            run_task: {
              allowSort: false,
              displayName: "Run",
              render: item => (
                <ButtonLink to={`/tasks/${item.definition_id}/run`}>
                  Run
                </ButtonLink>
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
          }}
          getItems={data => data.definitions}
          getTotal={data => data.total}
          filters={{
            alias: {
              name: "alias",
              displayName: "Alias",
              type: flotillaUIAsyncDataTableFilters.INPUT,
              description: "Search tasks by alias.",
            },
            group_name: {
              name: "group_name",
              displayName: "Group Name",
              type: flotillaUIAsyncDataTableFilters.SELECT,
              description: "Search tasks by existing group names.",
              filterProps: {
                shouldRequestOptions: true,
                requestOptionsFn: api.getGroups,
                isRequired: false,
                isMulti: true,
              },
            },
            image: {
              name: "image",
              displayName: "Image",
              type: flotillaUIAsyncDataTableFilters.INPUT,
              description: "Search tasks by Docker image.",
            },
          }}
          initialQuery={{
            page: 1,
            sort_by: "alias",
            order: "asc",
          }}
          emptyTableTitle="No tasks were found. Create one?"
          emptyTableBody={
            <ButtonLink to="/tasks/create">Create New Task</ButtonLink>
          }
        />
      </View>
    )
  }
}

export default Tasks
