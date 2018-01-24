import React, { Component } from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import Helmet from "react-helmet"
import { Link } from "react-router-dom"
import Select from "react-select"
import {
  View,
  ViewHeader,
  Loader,
  Button,
  Card,
  FormGroup,
  queryUpdateTypes,
} from "aa-ui-components"
import qs from "query-string"
import DebounceInput from "react-debounce-input"
import { has, isEmpty, isEqual, get, pickBy, identity } from "lodash"
import config from "../config"
import { getHelmetTitle } from "../utils/"
import PaginationButtons from "./PaginationButtons"
import SortHeader from "./SortHeader"
import withServerList from "./withServerList"
import TasksRow from "./TasksRow"

const limit = 20
const defaultQuery = {
  page: 1,
  sort_by: "alias",
  order: "asc",
}

export class Tasks extends Component {
  constructor(props) {
    super(props)
    this.handleRunButtonClick = this.handleRunButtonClick.bind(this)
  }
  componentDidMount() {
    this.aliasInput.focus()
  }
  handleRunButtonClick(definitionId) {
    this.props.history.push(`/tasks/${definitionId}/run`)
  }
  render() {
    const { isLoading, error, data, history, query, updateQuery } = this.props
    const loaderContainerStyle = { height: 960 }

    let content = <Loader containerStyle={loaderContainerStyle} />

    if (isLoading) {
      content = <Loader containerStyle={loaderContainerStyle} />
    } else if (error) {
      content = (
        <div className="table-error-container">
          {get(error, "response.data.error", error.toString())}
        </div>
      )
    } else if (has(data, "definitions")) {
      if (data.definitions.length > 0) {
        content = data.definitions.map(d => (
          <TasksRow
            key={d.definition_id}
            data={d}
            onRunButtonClick={evt => {
              evt.preventDefault()
              evt.stopPropagation()
              this.handleRunButtonClick(d.definition_id)
            }}
          />
        ))
      } else {
        content = "No tasks were found."
      }
    }
    return (
      <View>
        <Helmet>
          <title>{getHelmetTitle("Tasks")}</title>
        </Helmet>
        <ViewHeader
          title="Tasks"
          actions={
            <Link className="pl-button pl-intent-primary" to="/tasks/create">
              <span style={{ marginLeft: 4 }}>Create New Task</span>
            </Link>
          }
        />
        <div className="flot-list-view">
          <Card
            className="flot-list-view-filters-container"
            contentStyle={{ padding: 0 }}
          >
            <div className="flot-list-view-filters">
              <FormGroup
                label="Alias"
                input={
                  <DebounceInput
                    minLength={1}
                    debounceTimeout={250}
                    inputRef={aliasInput => {
                      this.aliasInput = aliasInput
                    }}
                    type="text"
                    className="pl-input"
                    value={get(query, "alias", "")}
                    onChange={evt => {
                      updateQuery([
                        {
                          key: "alias",
                          value: evt.target.value,
                          updateType: queryUpdateTypes.SHALLOW,
                        },
                        {
                          key: "page",
                          value: 1,
                          updateType: queryUpdateTypes.SHALLOW,
                        },
                      ])
                    }}
                  />
                }
              />
              <FormGroup
                label="Group Name"
                input={
                  <Select
                    value={get(query, "group_name", "")}
                    options={this.props.groupOptions}
                    onChange={selection => {
                      updateQuery([
                        {
                          key: "group_name",
                          value: selection === null ? null : selection.value,
                          updateType: queryUpdateTypes.SHALLOW,
                        },
                        {
                          key: "page",
                          value: 1,
                          updateType: queryUpdateTypes.SHALLOW,
                        },
                      ])
                    }}
                  />
                }
              />
              <FormGroup
                label="Image"
                input={
                  <DebounceInput
                    minLength={1}
                    debounceTimeout={250}
                    type="text"
                    className="pl-input"
                    value={get(query, "image", "")}
                    onChange={evt => {
                      updateQuery([
                        {
                          key: "image",
                          value: evt.target.value,
                          updateType: queryUpdateTypes.SHALLOW,
                        },
                        {
                          key: "page",
                          value: 1,
                          updateType: queryUpdateTypes.SHALLOW,
                        },
                      ])
                    }}
                  />
                }
              />
            </div>
          </Card>
          <div className="pl-table pl-bordered">
            <div className="pl-tr">
              <div className="pl-th" style={{ flex: 1 }}>
                Actions
              </div>
              <SortHeader
                currentSortKey={query.sort_by}
                currentOrder={query.order}
                display="Alias"
                sortKey="alias"
                updateQuery={updateQuery}
                style={{ flex: 4 }}
              />
              <SortHeader
                currentSortKey={query.sort_by}
                currentOrder={query.order}
                display="Group Name"
                sortKey="group_name"
                updateQuery={updateQuery}
                style={{ flex: 1 }}
              />
              <div className="pl-th" style={{ flex: 1 }}>
                Image
              </div>
              <SortHeader
                currentSortKey={query.sort_by}
                currentOrder={query.order}
                display="Memory"
                sortKey="memory"
                updateQuery={updateQuery}
                style={{ flex: 1 }}
              />
            </div>
            {content}
            <PaginationButtons
              total={get(data, "total", 20)}
              buttonCount={5}
              offset={query.offset}
              limit={query.limit}
              updateQuery={updateQuery}
              activeButtonClassName="pl-intent-primary"
              wrapperEl={
                <div className="table-footer flex ff-rn j-c a-c with-horizontal-child-margin" />
              }
            />
          </div>
        </div>
      </View>
    )
  }
}

const mapStateToProps = state => ({
  groupOptions: get(state, "selectOpts.group", []),
})

export default withServerList({
  getUrl: (props, query) => {
    // Strip falsy values and stringify query.
    const q = qs.stringify(pickBy(query, identity))
    return `${config.FLOTILLA_API}/task?${q}`
  },
  defaultQuery,
  limit,
})(connect(mapStateToProps)(Tasks)).withHOCStack
