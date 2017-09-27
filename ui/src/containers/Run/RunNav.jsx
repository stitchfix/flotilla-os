import React from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import { isEmpty } from 'lodash'
import { RefreshCw, X } from 'react-feather'
import { allowedLocations, invalidEnv } from '../../constants/'
import { AppHeader } from '../../components/'

function RunNav({ taskID, onStop, runInfo, isStopped, hasError }) {
  const retryQuery = !isEmpty(runInfo) ? {
    cluster: runInfo.cluster,
    env: !!runInfo.env ? runInfo.env
      .filter(e => !invalidEnv.includes(e.name))
      .map(e => `${e.name}|${e.value}`) : null
  } : null
  const stopButton = (
    <button className="button button-error" onClick={onStop}>
      <div className="flex">
        <X size={14} />
        <span className="button-icon-text">Stop</span>
      </div>
    </button>
  )
  const retryButton = (
    <Link
      to={{
        pathname: `/tasks/${taskID}/run`,
        query: retryQuery
      }}
      className="button"
    >
      <RefreshCw size={14} />&nbsp;Retry
    </Link>
  )
  let buttons

  if (hasError) {
    buttons = null
  } else {
    buttons = !isStopped ? [stopButton, retryButton] : [retryButton]
  }

  return (
    <AppHeader
      currentLocation={allowedLocations.run}
      buttons={buttons}
    />
  )
}

RunNav.propTypes = {
  onStop: PropTypes.func.isRequired,
}

function mapStateToProps(state) {
  return ({
    taskID: state.run.info.definition_id,
    taskAlias: state.run.info.alias,
    runID: state.run.info.run_id,
    runInfo: state.run.info
  })
}

export default connect(mapStateToProps)(RunNav)
