import React from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import { connect } from "react-redux"
import Helmet from "react-helmet"
import { get } from "lodash"
import AsyncDataTable from "../AsyncDataTable/AsyncDataTable"
import { asyncDataTableFilterTypes } from "../AsyncDataTable/AsyncDataTableFilter"
import View from "../styled/View"
import ViewHeader from "../styled/ViewHeader"
import api from "../../api"
import config from "../../config"

const Tasks = props => (
  <View>
    <Helmet>
      <title>Tasks</title>
    </Helmet>
    <ViewHeader
      title="Tasks"
      actions={
        <Link className="pl-button pl-intent-primary" to="/tasks/create">
          Create New Task
        </Link>
      }
    />
    <AsyncDataTable
      requestFn={api.getTasks}
      columns={{
        run_task: {
          allowSort: false,
          displayName: "Run",
          render: item => (
            <Link className="pl-button" to={`/tasks/${item.definition_id}/run`}>
              Run
            </Link>
          ),
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
        },
        image: {
          allowSort: true,
          displayName: "Image",
          render: item => item.image.substr(config.IMAGE_PREFIX.length),
        },
        memory: {
          allowSort: true,
          displayName: "Memory",
          render: item => item.memory,
        },
      }}
      getItems={data => data.definitions}
      getTotal={data => data.total}
      filters={{
        alias: {
          displayName: "Alias",
          type: asyncDataTableFilterTypes.INPUT,
        },
        group_name: {
          displayName: "Group Name",
          type: asyncDataTableFilterTypes.SELECT,
          options: props.groupOptions,
        },
        image: {
          displayName: "Image",
          type: asyncDataTableFilterTypes.INPUT,
        },
      }}
      initialQuery={{
        page: 1,
        sort_by: "alias",
        order: "asc",
      }}
      emptyTableTitle="No tasks were found. Create one?"
      emptyTableBody={
        <Link className="pl-button pl-intent-primary" to="/tasks/create">
          Create New Task
        </Link>
      }
    />
  </View>
)

Tasks.propTypes = {
  groupOptions: PropTypes.arrayOf(
    PropTypes.shape({ label: PropTypes.string, value: PropTypes.string })
  ),
}

Tasks.defaultProps = {
  groupOptions: [],
}

const mapStateToProps = state => ({
  groupOptions: get(state, "selectOpts.group", []),
})

export default connect(mapStateToProps)(Tasks)
