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
  modalActions,
} from "aa-ui-components"
import qs from "query-string"
import { has, get, pickBy, identity } from "lodash"
import config from "../config"
import { getHelmetTitle } from "../utils/"
import withServerList from "./withServerList"
import SortHeader from "./SortHeader"
import StopRunModal from "./StopRunModal"
import PaginationButtons from "./PaginationButtons"
import ActiveRunsRow from "./ActiveRunsRow"

const limit = 20
const defaultQuery = {
  page: 1,
  sort_by: "started_at",
  order: "desc",
}

export const ActiveRuns = ({
  isLoading,
  error,
  data,
  updateQuery,
  query,
  clusterOptions,
  dispatch,
}) => {
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
  } else if (has(data, "history")) {
    if (Array.isArray(data.history) && data.history.length > 0) {
      content = data.history.map(d => (
        <ActiveRunsRow
          data={d}
          onStopButtonClick={evt => {
            // Prevent from propagating click to <Link> component.
            evt.preventDefault()
            evt.stopPropagation()
            dispatch(
              modalActions.renderModal(
                <StopRunModal runId={d.run_id} definitionId={d.definition_id} />
              )
            )
          }}
          key={d.run_id}
        />
      ))
    } else {
      content = <div>No data was found.</div>
    }
  }
  return (
    <View>
      <Helmet>
        <title>{getHelmetTitle("Active Runs")}</title>
      </Helmet>
      <ViewHeader title="Active Runs" />
      <div className="flot-list-view">
        <Card
          className="flot-list-view-filters-container"
          contentStyle={{ padding: 0 }}
          // header="Filters"
        >
          <div className="flot-list-view-filters">
            <FormGroup
              label="Cluster"
              input={
                <Select
                  value={get(query, "cluster_name", "")}
                  options={clusterOptions}
                  onChange={selection => {
                    const value = selection === null ? null : selection.value
                    updateQuery([
                      {
                        key: "cluster_name",
                        value,
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
              style={{ flex: 1 }}
              currentSortKey={query.sort_by}
              currentOrder={query.order}
              display="Status"
              sortKey="status"
              updateQuery={updateQuery}
            />
            <SortHeader
              style={{ flex: 1.5 }}
              currentSortKey={query.sort_by}
              currentOrder={query.order}
              display="Started At"
              sortKey="started_at"
              updateQuery={updateQuery}
            />
            <div className="pl-th" style={{ flex: 4 }}>
              Alias
            </div>
            <SortHeader
              style={{ flex: 1.5 }}
              currentSortKey={query.sort_by}
              currentOrder={query.order}
              display="Cluster"
              sortKey="cluster_name"
              updateQuery={updateQuery}
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

const mapStateToProps = state => ({
  clusterOptions: get(state, "selectOpts.cluster", []),
})

export default withServerList({
  getUrl: (props, query) => {
    // Strip falsy values and stringify query.
    const q = qs.stringify(pickBy(query, identity))
    return `${
      config.FLOTILLA_API
    }/history?status=RUNNING&status=PENDING&status=QUEUED&${q}`
  },
  defaultQuery,
  limit,
})(connect(mapStateToProps)(ActiveRuns)).withHOCStack
