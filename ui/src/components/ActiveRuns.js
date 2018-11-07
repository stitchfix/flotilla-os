import React from "react"
import { connect } from "react-redux"
import Helmet from "react-helmet"
import Select from "react-select"
import qs from "qs"
import { has, get, pickBy, identity } from "lodash"
import modalActions from "../actions/modalActions"
import queryUpdateTypes from "../utils/queryUpdateTypes"
import getHelmetTitle from "../utils/getHelmetTitle"
import ActiveRunsRow from "./ActiveRunsRow"
import Card from "./Card"
import EmptyTable from "./EmptyTable"
import FormGroup from "./FormGroup"
import PaginationButtons from "./PaginationButtons"
import SortHeader from "./SortHeader"
import StopRunModal from "./StopRunModal"
import View from "./View"
import ViewHeader from "./ViewHeader"
import withServerList from "./withServerList"

import config from "../config"

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
  let content = <EmptyTable isLoading />

  if (isLoading) {
    content = <EmptyTable isLoading />
  } else if (error) {
    const errorDisplay = error.toString() || "An error occurred."
    content = <EmptyTable title={errorDisplay} error />
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
      content = <EmptyTable title="No tasks are currently running." />
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
