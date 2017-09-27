import React, { Component } from 'react'
import { connect } from 'react-redux'
import { runTask, saveRunConfig, fetchRunConfig } from '../../actions/'
import { RunForm, RunFormNav } from '../'

class RunFormContainer extends Component {
  componentDidMount() {
    this.props.dispatch(fetchRunConfig({ id: this.props.params.taskID }))
  }
  runTask(values) {
    const { router, dispatch, params } = this.props

    dispatch(runTask({
      cluster: values.cluster,
      env: values.env,
      taskID: params.taskID,
      routerPush: router.push
    }))

    if (values.saveConfig) {
      dispatch(saveRunConfig({
        cluster: values.cluster,
        env: values.env,
        taskID: params.taskID,
      }))
    }
  }
  render() {
    return (
      <div className="view-container">
        <RunFormNav />
        <div className="view">
          <div className="layout-standard flex ff-cn j-fs a-c">
            <RunForm onSubmit={(values) => { this.runTask(values) }} />
          </div>
        </div>
      </div>
    )
  }
}

function mapStateToProps(state) {
  return ({
    taskID: state.task.task.definition_id,
    clusterOptions: state.dropdownOpts.cluster,
    formValues: state.form.run ? state.form.run.values : {},
    localRunConfig: state.task.runConfig
  })
}

export default connect(mapStateToProps)(RunFormContainer)
