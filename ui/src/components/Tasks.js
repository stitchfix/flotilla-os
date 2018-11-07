import React, { Component } from "react"
import { connect } from "react-redux"
import Helmet from "react-helmet"
import { Link } from "react-router-dom"
import Select from "react-select"
import qs from "qs"
import DebounceInput from "react-debounce-input"
import { has, get, pickBy, identity } from "lodash"
import Card from "./Card"
import EmptyTable from "./EmptyTable"
import FormGroup from "./FormGroup"
import PaginationButtons from "./PaginationButtons"
import SortHeader from "./SortHeader"
import TasksRow from "./TasksRow"
import View from "./View"
import ViewHeader from "./ViewHeader"
import withServerList from "./withServerList"
import getHelmetTitle from "../utils/getHelmetTitle"
import queryUpdateTypes from "../utils/queryUpdateTypes"
import config from "../config"

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
    const { isLoading, data, error, history, query, updateQuery } = this.props

    let content = <EmptyTable isLoading />

    if (isLoading) {
      content = <EmptyTable isLoading />
    } else if (error) {
      const errorDisplay = error.toString() || "An error occured."
      content = <EmptyTable title={errorDisplay} error />
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
        content = (
          <EmptyTable
            title="No tasks were found. Create one?"
            actions={
              <Link className="pl-button pl-intent-primary" to="/tasks/create">
                <span style={{ marginLeft: 4 }}>Create New Task</span>
              </Link>
            }
          />
        )
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
                className="pl-hide-small"
              />
              <div className="pl-th pl-hide-small" style={{ flex: 1 }}>
                Image
              </div>
              <SortHeader
                currentSortKey={query.sort_by}
                currentOrder={query.order}
                display="Memory"
                sortKey="memory"
                updateQuery={updateQuery}
                style={{ flex: 1 }}
                className="pl-hide-small"
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
