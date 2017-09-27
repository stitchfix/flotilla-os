import React, { Component } from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import moment from 'moment'
import { has, isEqual } from 'lodash'
import ReactJson from 'react-json-view'
import { ExternalLink } from 'react-feather'
import { reactJsonViewProps } from '../../constants/'
import { FormGroup, RunStatusText } from '../../components/'
import { getRunStatus } from '../../utils/'

class RunInfo extends Component {
  constructor(props) {
    super(props)
    this.toggleJsonView = this.toggleJsonView.bind(this)
    this.openRunMiniView = this.openRunMiniView.bind(this)
  }
  state = {
    jsonMode: false
  }
  shouldComponentUpdate(nextProps, nextState) {
    return !isEqual(this.state, nextState) || !isEqual(this.props.info, nextProps.info)
  }
  toggleJsonView() {
    this.setState({ jsonMode: !this.state.jsonMode })
  }
  openRunMiniView() {
    const { info } = this.props
    const url = `${window.location.origin}/#/runs/${info.run_id}/mini`
    window.open(
      url,
      'Run Mini',
      `menubar=no,location=no,toolbar=no,resizable=0,width=360,height=130,toolbar=no`
    )
  }
  render() {
    const { info } = this.props
    const { jsonMode } = this.state
    const enhancedStatus = has(info, 'status') ? getRunStatus({
      status: info.status,
      exitCode: has(info, 'exit_code') ? info.exit_code : null
    }) : '-'
    return (
      <div>
        <div className="section-container run-info">
          <div className="section-header">
            <div className="section-header-text">Run Info</div>
            <div className="flex">
              <button className="button" onClick={this.openRunMiniView}>
                <div className="flex">
                  <ExternalLink size={14} />
                  <span className="button-icon-text">Mini View</span>
                </div>
              </button>
              <button className="button" onClick={this.toggleJsonView}>
                {jsonMode ? 'Default' : 'JSON'}
              </button>
            </div>
          </div>
          {
            jsonMode ?
              <ReactJson src={info} {...reactJsonViewProps} /> :
              <div>
                <FormGroup
                  label="Status"
                  isStatic
                >
                  <RunStatusText
                    enhancedStatus={enhancedStatus}
                    status={info.status}
                    exitCode={info.exit_code}
                  />
                </FormGroup>
                <FormGroup
                  label="Exit Code"
                  isStatic
                >
                  {has(info, 'exit_code') ? info.exit_code : '-'}
                </FormGroup>
                <FormGroup
                  label="Started At"
                  isStatic
                >
                  {
                    has(info, 'started_at') ?
                      <div className="flex ff-rn j-fs a-bl">
                        {info.started_at}
                        <span className="text-secondary" style={{ marginLeft: 6 }}>
                          {moment(info.started_at).fromNow()}
                        </span>
                      </div> : '-'
                  }
                </FormGroup>
                <FormGroup
                  label="Finished At"
                  isStatic
                >
                  {
                    has(info, 'finished_at') ?
                      <div className="flex ff-rn j-fs a-bl">
                        {info.finished_at}
                        <span className="text-secondary" style={{ marginLeft: 6 }}>
                          {moment(info.finished_at).fromNow()}
                        </span>
                      </div> : '-'
                  }
                </FormGroup>
                <FormGroup
                  label="Run ID"
                  isStatic
                >
                  {has(info, 'run_id') ? info.run_id : '-'}
                </FormGroup>
                <FormGroup
                  label="Task Definition ID"
                  isStatic
                >
                  <Link to={`/tasks/${info.definition_id}`}>
                    {has(info, 'definition_id') ? info.definition_id : '-'}
                  </Link>
                </FormGroup>
                <FormGroup
                  label="Instance ID"
                  isStatic
                >
                  {has(info, 'instance.instance_id') ? info.instance.instance_id : '-'}
                </FormGroup>
                <FormGroup
                  label="DNS Name"
                  isStatic
                  style={{ borderBottom: 0 }}
                >
                  {has(info, 'instance.dns_name') ? info.instance.dns_name : '-'}
                </FormGroup>
              </div>
          }
        </div>
        <div className="section-container run-env-vars">
          <div className="section-header">
            <div className="section-header-text">Environment Variables</div>
          </div>
          <div>
            {
              has(info, 'env') &&
              info.env.map((env, i) => (
                <FormGroup
                  key={i}
                  label={<div className="tag code">{env.name}</div>}
                  isStatic
                  horizontal
                  style={i === info.env.length - 1 ? { borderBottom: 0 } : {}}
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
  return ({ info: state.run.info })
}

export default connect(mapStateToProps)(RunInfo)
