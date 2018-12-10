import * as React from "react"
import { withRouter, RouteComponentProps } from "react-router-dom"
import ConfirmModal from "./ConfirmModal"
import api from "../../api"

interface IDeleteTaskModalProps extends RouteComponentProps {
  definitionID: string
}

const DeleteTaskModal: React.SFC<IDeleteTaskModalProps> = props => (
  <ConfirmModal
    body="Are you sure you want to delete this task?"
    getRequestArgs={() => ({ definitionID: props.definitionID })}
    requestFn={api.deleteTask}
    onSuccess={() => {
      props.history.push("/tasks")
    }}
  />
)

DeleteTaskModal.displayName = "DeleteTaskModal"
DeleteTaskModal.defaultProps = {}
export default withRouter(DeleteTaskModal)
