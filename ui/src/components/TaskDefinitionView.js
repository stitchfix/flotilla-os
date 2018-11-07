import React from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import Helmet from "react-helmet"
import { has } from "lodash"
import Button from "./Button"
import DeleteTaskModal from "./DeleteTaskModal"
import TaskHistory from "./TaskHistory"
import TaskInfo from "./TaskInfo"
import View from "./View"
import ViewHeader from "./ViewHeader"
import modalActions from "../actions/modalActions"
import intentTypes from "../constants/intentTypes"
import getHelmetTitle from "../utils/getHelmetTitle"

const TaskDefinitionView = props => {
  let title = ""

  if (has(props.data, "alias")) {
    title = props.data.alias
  } else if (has(props.data, "definition_id")) {
    title = props.data.definition_id
  }

  return (
    <View>
      <Helmet>
        <title>{getHelmetTitle(title)}</title>
      </Helmet>
      <ViewHeader
        title={<div className="overflow-ellipsis">{title}</div>}
        actions={
          <div className="flex ff-rn j-fs a-c with-horizontal-child-margin">
            <Button
              intent={intentTypes.error}
              onClick={() => {
                props.dispatch(
                  modalActions.renderModal(
                    <DeleteTaskModal definitionId={props.definitionId} />
                  )
                )
              }}
            >
              Delete
            </Button>
            <Link
              to={`/tasks/${props.definitionId}/copy`}
              className="pl-button"
            >
              Copy
            </Link>
            <Link
              to={`/tasks/${props.definitionId}/edit`}
              className="pl-button"
            >
              Edit
            </Link>
            <Link
              to={`/tasks/${props.definitionId}/run`}
              className="pl-button pl-intent-primary"
            >
              Run
            </Link>
          </div>
        }
      />
      <div className="flot-detail-view">
        <TaskInfo {...props} />
        <TaskHistory definitionId={props.definitionId} />
      </div>
    </View>
  )
}

TaskDefinitionView.propTypes = {
  data: PropTypes.object,
  definitionId: PropTypes.string,
  dispatch: PropTypes.func,
}

export default TaskDefinitionView
