import React, { Fragment } from "react"
import { Link } from "react-router-dom"
import { has, get, omit } from "lodash"
import moment from "moment"
import KeyValues from "../styled/KeyValues"
import { Pre } from "../styled/Monospace"
import SecondaryText from "../styled/SecondaryText"
import RunContext from "./RunContext"
import { flotillaUIRequestStates, IFlotillaEnv } from "../../.."
import getRunDuration from "../../helpers/getRunDuration"
import RunDurationCounter from "./RunDurationCounter"

const RunSidebar = () => {
  return (
    <RunContext.Consumer>
      {({ data, requestState }) => {
        if (requestState === flotillaUIRequestStates.READY && data) {
          return (
            <Fragment>
              <KeyValues
                raw={omit(data, ["env"])}
                label="Run Info"
                items={{
                  Duration: (
                    <RunDurationCounter
                      started_at={data.started_at}
                      finished_at={data.finished_at}
                    />
                  ),
                  Cluster: get(data, "cluster", "-"),
                  "Exit Code": get(data, "exit_code", "-"),
                  "Started At": has(data, "started_at") ? (
                    <Fragment>
                      <div style={{ marginBottom: 4 }}>
                        {moment(data.started_at).fromNow()}
                      </div>
                      <SecondaryText>{data.started_at}</SecondaryText>
                    </Fragment>
                  ) : (
                    "-"
                  ),
                  "Finished At": has(data, "finished_at") ? (
                    <Fragment>
                      <div style={{ marginBottom: 4 }}>
                        {moment(data.finished_at).fromNow()}
                      </div>
                      <SecondaryText>{data.finished_at}</SecondaryText>
                    </Fragment>
                  ) : (
                    "-"
                  ),
                  "Run ID": has(data, "run_id") ? (
                    <Link to={`/runs/${data.run_id}`}>{data.run_id}</Link>
                  ) : (
                    "-"
                  ),
                  "Task Definition ID": has(data, "definition_id") ? (
                    <Link to={`/tasks/${data.definition_id}`}>
                      {data.definition_id}
                    </Link>
                  ) : (
                    "-"
                  ),
                  Image: get(data, "image", "-"),
                  "Task Arn": get(data, "task_arn", "-"),
                  "Instance ID": get(data, "instance.instance_id", "-"),
                  "Instance DNS Name": get(data, "instance.dns_name", "-"),
                }}
              />
              <KeyValues
                raw={get(data, "env", [])}
                label="Environment Variables"
                items={
                  data &&
                  data.env &&
                  data.env.reduce((acc: any, env: IFlotillaEnv): any => {
                    acc[env.name] = <Pre>{env.value}</Pre>
                    return acc
                  }, {})
                }
              />
            </Fragment>
          )
        }
      }}
    </RunContext.Consumer>
  )
}

export default RunSidebar
