import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'
import {
  fetchRun,
  resetRun,
  clearRunInterval
} from '../../actions/'

class RunContainer extends Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.shape({
      runID: PropTypes.string,
    })
  }
  componentDidMount() {
    const { dispatch, params } = this.props
    dispatch(fetchRun({ runID: params.runID }))
  }
  componentDidUpdate(prevProps) {
    const { params, dispatch } = this.props
    if ((prevProps.params.runID !== params.runID)) {
      this.reset()
      dispatch(fetchRun({ runID: params.runID }))
    }
  }
  componentWillUnmount() {
    this.reset()
  }
  reset() {
    this.props.dispatch(resetRun())
    clearRunInterval()
  }
  render() {
    return (<div>{this.props.children}</div>)
  }
}

export default connect()(RunContainer)
