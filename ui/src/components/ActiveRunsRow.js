import React from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import { get } from "lodash"
import moment from "moment"
import Button from "./Button"
import EnhancedRunStatus from "./EnhancedRunStatus"

const ActiveRunsRow = ({ data, onStopButtonClick }) => (
  <Link className="pl-tr unstyled-link hoverable" to={`/runs/${data.run_id}`}>
    <div className="pl-td" style={{ flex: 1 }}>
      <Button onClick={onStopButtonClick}>Stop</Button>
    </div>
    <div className="pl-td" style={{ flex: 1 }}>
      <EnhancedRunStatus
        status={get(data, "status")}
        exitCode={get(data, "exit_code")}
        iconOnly={window.innerWidth < 550}
      />
    </div>
    <div className="pl-td" style={{ flex: 1.5 }}>
      {moment(data.started_at).fromNow()}
    </div>
    <div className="pl-td" style={{ flex: 4 }}>
      {get(data, "alias", data.definition_id)}
    </div>
    <div className="pl-td pl-hide-small" style={{ flex: 1.5 }}>
      {data.cluster}
    </div>
  </Link>
)

ActiveRunsRow.propTypes = {
  data: PropTypes.shape({
    definition_id: PropTypes.string,
    run_id: PropTypes.string,
    status: PropTypes.string,
    started_at: PropTypes.string,
    cluster: PropTypes.string,
  }),
  onStopButtonClick: PropTypes.func,
}

export default ActiveRunsRow
