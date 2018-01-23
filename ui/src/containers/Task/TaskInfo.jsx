import React, { Component } from 'react'
import { connect } from 'react-redux'
import ReactJson from 'react-json-view'
import { has } from 'lodash'
import { reactJsonViewProps } from '../../constants/'
import { FormGroup } from '../../components/'

class TaskInfo extends Component {
  constructor(props) {
    super(props)
    this.toggleJsonView = this.toggleJsonView.bind(this)
  }
  state = {
    jsonMode: false
  }
  toggleJsonView() {
    this.setState({ jsonMode: !this.state.jsonMode })
  }
  render() {
    const { task } = this.props
    const { jsonMode } = this.state
    return (
      <div>
        <div className="section-container task-definition">
          <div className="section-header">
            <div className="section-header-text">Task Definition</div>
            <button className="button" onClick={this.toggleJsonView}>
              {jsonMode ? 'View Standard' : 'View JSON'}
            </button>
          </div>
          {
            jsonMode ?
              <ReactJson src={task} {...reactJsonViewProps} /> :
              <div>
                <FormGroup isStatic label="Alias">
                  {task.alias}
                </FormGroup>
                <FormGroup isStatic label="Definition ID">
                  {task.definition_id}
                </FormGroup>
                <FormGroup isStatic label="Group Name">
                  {task.group_name}
                </FormGroup>
                <FormGroup isStatic label="Image">
                  {task.image}
                </FormGroup>
                <FormGroup isStatic label="Command">
                  <pre>{task.command}</pre>
                </FormGroup>
                <FormGroup isStatic label="Ports">
                  {task.ports}
                </FormGroup>
                <FormGroup isStatic label="Memory">
                  {task.memory}
                </FormGroup>
                <FormGroup isStatic label="Arn">
                  {task.arn}
                </FormGroup>
                <FormGroup isStatic label="Tags">
                  <div className="flex ff-rw j-fs a-fs">
                    {task.tags && task.tags.map(t => <div key={t} className="tag code">{t}</div>)}
                  </div>
                </FormGroup>
              </div>
          }
        </div>
        <div className="section-container task-env-vars">
          <div className="section-header">
            <div className="section-header-text">Environment Variables</div>
          </div>
          <div>
            {
              has(task, 'env') &&
              task.env.map((env, i) => (
                <FormGroup
                  key={i}
                  label={<div className="tag code">{env.name}</div>}
                  isStatic
                  horizontal
                >
                  <div className="tag code">{env.value}</div>
                </FormGroup>
              ))
            }
          </div>
        </div>
      </div>
    )
  }
}

function mapStateToProps(state) {
  return ({ task: state.task.task })
}

export default connect(mapStateToProps)(TaskInfo)
