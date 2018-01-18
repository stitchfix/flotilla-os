import React from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import { Button } from "platforma"
import { getImageDisplayName } from "../utils/"

const TasksRow = ({ data, onRunButtonClick }) => {
  return (
    <Link
      className="pl-tr unstyled-link hoverable"
      to={`/tasks/${data.definition_id}`}
      key={data.definition_id}
    >
      <div className="pl-td" style={{ flex: 1 }}>
        <Button onClick={onRunButtonClick}>Run</Button>
      </div>
      <div className="pl-td" style={{ flex: 4 }}>
        {data.alias}
      </div>
      <div className="pl-td" style={{ flex: 1 }}>
        {data.group_name}
      </div>
      <div className="pl-td overflow-ellipsis" style={{ flex: 1 }}>
        {getImageDisplayName(data.image)}
      </div>
      <div className="pl-td" style={{ flex: 1 }}>
        {data.memory}
      </div>
    </Link>
  )
}

TasksRow.propTypes = {
  data: PropTypes.object,
  onRunButtonClick: PropTypes.func,
}

export default TasksRow
