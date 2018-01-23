import React, { Component } from 'react'
import { connect } from 'react-redux'
import { Link, withRouter } from 'react-router'
import { Logo } from './'
import { allowedLocations } from '../constants/'

class Breadcrumbs extends Component {
  generateBreadcrumbs() {
    const { currentLocation } = this.props
    let breadcrumbs = [{
      path: `/`,
      displayName: (
        <div className="flex a-c">
          <Logo />
          <div style={{ marginLeft: 6 }}>Flotilla</div>
        </div>
      )
    }]

    switch (currentLocation) {
      case allowedLocations.createTask:
        breadcrumbs = [
          ...breadcrumbs,
          {
            path: `/create-task`,
            displayName: 'Create Task'
          }
        ]
        break
      case allowedLocations.run:
        breadcrumbs = [
          ...breadcrumbs,
          {
            path: `/tasks/${this.props.run.info.definition_id}`,
            displayName: this.props.run.info.definition_id
          },
          {
            path: `/runs/${this.props.run.info.run_id}`,
            displayName: this.props.run.info.run_id
          },
        ]
        break
      case allowedLocations.task:
        breadcrumbs = [
          ...breadcrumbs,
          {
            path: `/tasks/${this.props.task.task.definition_id}`,
            displayName: this.props.task.task.alias || this.props.task.task.definition_id
          }
        ]
        break
      case allowedLocations.editTask:
        breadcrumbs = [
          ...breadcrumbs,
          {
            path: `/tasks/${this.props.task.task.definition_id}`,
            displayName: this.props.task.task.alias || this.props.task.task.definition_id
          },
          {
            path: `/tasks/${this.props.task.task.definition_id}/edit`,
            displayName: 'Edit Task'
          }
        ]
        break
      case allowedLocations.copyTask:
        breadcrumbs = [
          ...breadcrumbs,
          {
            path: `/tasks/${this.props.task.task.definition_id}`,
            displayName: this.props.task.task.alias || this.props.task.task.definition_id
          },
          {
            path: `/tasks/${this.props.task.task.definition_id}/copy`,
            displayName: 'Copy Task'
          }
        ]
        break
      case allowedLocations.runTask:
        breadcrumbs = [
          ...breadcrumbs,
          {
            path: `/tasks/${this.props.task.task.definition_id}`,
            displayName: this.props.task.task.alias || this.props.task.task.definition_id
          },
          {
            path: `/tasks/${this.props.task.task.definition_id}/run`,
            displayName: 'Run'
          },
        ]
        break
      case allowedLocations.tasks:
        breadcrumbs = [
          ...breadcrumbs,
          {
            path: `/tasks`,
            displayName: 'Tasks'
          }
        ]
        break
      case allowedLocations.runs:
        breadcrumbs = [
          ...breadcrumbs,
          {
            path: `/runs`,
            displayName: 'Runs'
          }
        ]
        break
      default:
        break
    }
    return breadcrumbs.filter(c => !!c.displayName)
  }
  render() {
    const breadcrumbs = this.generateBreadcrumbs()
    return (
      <div className="breadcrumbs-container">
        <div className="flex ff-rn j-fs a-c">
          {breadcrumbs && breadcrumbs.map((crumb, i) => (
            <h3 className="flex" key={`crumb-${i}`}>
              <Link className="breadcrumb" to={crumb.path}>{crumb.displayName}</Link>
              {i !== breadcrumbs.length - 1 && `>`}
            </h3>
          ))}
        </div>
      </div>
    )
  }
}

const mapStateToProps = state => state

export default withRouter(connect(mapStateToProps)(Breadcrumbs))
