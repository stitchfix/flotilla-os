import React from "react"
import { Link } from "react-router-dom"
import { has, get } from "lodash"
import moment from "moment"
import KeyValues from "../styled/KeyValueContainer"
import colors from "../../constants/colors"
import Pre from "../styled/Pre"
import RunContext from "./RunContext"
import * as requestStateTypes from "../../constants/requestStateTypes"

const RunSidebar = props => {
  return (
    <RunContext.Consumer>
      {({ data, requestState }) => {
        if (requestState === requestStateTypes.READY) {
          return (
            <KeyValues
              items={{
                Cluster: get(data, "cluster", "-"),
                "Exit Code": get(data, "exit_code", "-"),
                "Started At": has(data, "started_at") ? (
                  <div className="flex ff-rn j-fs a-bl with-horizontal-child-margin">
                    <div>{moment(data.started_at).fromNow()}</div>
                    <div className="text-small">{data.started_at}</div>
                  </div>
                ) : (
                  "-"
                ),
                "Finished At": has(data, "finished_at") ? (
                  <div className="flex ff-rn j-fs a-bl with-horizontal-child-margin">
                    <div>{moment(data.finished_at).fromNow()}</div>
                    <div className="text-small">{data.finished_at}</div>
                  </div>
                ) : (
                  "-"
                ),
                "Run ID": has(data, "run_id") ? (
                  <Link
                    to={`/runs/${data.run_id}`}
                    style={{
                      textDecoration: "underline",
                      color: colors.gray[3],
                    }}
                  >
                    {data.run_id}
                  </Link>
                ) : (
                  "-"
                ),
                "Task Definition ID": has(data, "definition_id") ? (
                  <Link
                    to={`/tasks/${data.definition_id}`}
                    style={{
                      textDecoration: "underline",
                      color: colors.gray[3],
                    }}
                  >
                    {data.definition_id}
                  </Link>
                ) : (
                  "-"
                ),
                Image: get(data, "image", "-"),
                "Task Arn": get(data, "task_arn", "-"),
                "Instance ID": get(data, "instance.instance_id", "-"),
                "Instance DNS Name": get(data, "instance.dns_name", "-"),
                "Environment Vars": (
                  <KeyValues
                    items={get(data, "env", []).reduce((acc, env) => {
                      acc[env.name] = <Pre>{env.value}</Pre>
                      return acc
                    }, {})}
                  />
                ),
              }}
            />
          )
        }
      }}
    </RunContext.Consumer>
  )
}

export default RunSidebar
