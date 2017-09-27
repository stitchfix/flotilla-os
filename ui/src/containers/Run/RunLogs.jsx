import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'
import Ansi from 'ansi-to-react'
import { ChevronsUp, ChevronsDown } from 'react-feather'
import { Loader } from '../../components/'

class RunLogs extends Component {
  static propTypes = {
    mini: PropTypes.bool
  }
  static defaultProps = {
    mini: false
  }
  constructor(props) {
    super(props)
    this.toggleAutoscroll = this.toggleAutoscroll.bind(this)
    this.scrollToTop = this.scrollToTop.bind(this)
    this.scrollToBottom = this.scrollToBottom.bind(this)
  }
  state = {
    isAutoscrollEnabled: true
  }
  componentDidUpdate(prevProps) {
    if (this.state.isAutoscrollEnabled &&
        !!this.props.logs &&
        prevProps.logs.length !== this.props.logs.length) {
      this.scrollToBottom()
    }
  }
  toggleAutoscroll() {
    this.setState({ isAutoscrollEnabled: !this.state.isAutoscrollEnabled })
  }
  scrollToTop() {
    const runLogsContainer = document.querySelector('#runLogsContainer')
    runLogsContainer.scrollTop = 0
  }
  scrollToBottom() {
    const runLogsContainer = document.querySelector('#runLogsContainer')
    runLogsContainer.scrollTop = runLogsContainer.scrollHeight
  }
  render() {
    const { active, logs, mini } = this.props
    const loaderProps = mini ?
      { mini: true, style: { marginTop: 4, marginBottom: 4 } } :
      { style: { marginTop: 20, marginBottom: 20 } }
    return (
      <div className="section-container run-logs-container">
        {
          !mini && (
            <div className="section-header">
              <div className="section-header-text">Logs</div>
              <div className="flex">
                <div
                  className="flex ff-rn j-sb a-c"
                  style={{
                    paddingLeft: 12,
                    paddingRight: 12,
                  }}
                >
                  <input
                    type="checkbox"
                    style={{ marginRight: 4 }}
                    checked={this.state.isAutoscrollEnabled}
                    onChange={this.toggleAutoscroll}
                  />
                  <span>Autoscroll</span>
                </div>
                <button className="button" onClick={this.scrollToTop}>
                  <ChevronsUp size={14} />
                </button>
                <button className="button" onClick={this.scrollToBottom}>
                  <ChevronsDown size={14} />
                </button>
              </div>
            </div>
          )
        }
        <div className="run-logs code" id="runLogsContainer">
          {
            !!logs ? logs.map((logSet, i) => {
              return (
                <pre key={`logSet-${i}`}>
                  {logSet.logs.map((line, j) => (<Ansi key={`logSet-${i}-line-${j}`}>{`${line}\n`}</Ansi>))}
                </pre>
              )
            }) : <Loader />
          }
          {!!active && <Loader {...loaderProps} />}
        </div>
      </div>
    )
  }
}

function mapStateToProps(state) {
  return ({
    logs: state.run.logs,
    active: state.run._logsAreFetching || (!!state.run.info && state.run.info.status !== 'STOPPED')
  })
}

export default connect(mapStateToProps)(RunLogs)
