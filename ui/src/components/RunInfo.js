import React, { Component } from "react"
import { Link } from "react-router-dom"
import JsonView from "react-json-view"
import { Card, FormGroup, Button, Tag, colors } from "aa-ui-components"
import { has, get } from "lodash"
import moment from "moment"
import { reactJsonViewProps } from "../constants/"
import EnhancedRunStatus from "./EnhancedRunStatus"
import KeyValueContainer from "./KeyValueContainer"
import RunStatusBar from "./RunStatusBar"

export default function RunInfo({ data }) {
  return (
    <div className="flot-detail-view-sidebar with-vertical-child-margin">
      <RunStatusBar
        startedAt={get(data, "started_at")}
        finishedAt={get(data, "finished_at")}
        status={get(data, "status", "")}
        exitCode={get(data, "exit_code", "")}
      />
      <KeyValueContainer header="Run Info">
        {({ json, collapsed }) => {
          if (json) {
            return <JsonView {...reactJsonViewProps} src={data} />
          }

          return (
            <div className="flot-detail-view-sidebar-card-content">
              <FormGroup isStatic label="Cluster">
                {get(data, "cluster", "-")}
              </FormGroup>
              <FormGroup isStatic label="Exit Code">
                {get(data, "exit_code", "-")}
              </FormGroup>
              <FormGroup isStatic label="Started At">
                {has(data, "started_at") ? (
                  <div className="flex ff-rn j-fs a-bl with-horizontal-child-margin">
                    <div>{moment(data.started_at).fromNow()}</div>
                    <div className="text-small">{data.started_at}</div>
                  </div>
                ) : (
                  "-"
                )}
              </FormGroup>
              <FormGroup isStatic label="Finished At">
                {has(data, "finished_at") ? (
                  <div className="flex ff-rn j-fs a-bl with-horizontal-child-margin">
                    <div>{moment(data.finished_at).fromNow()}</div>
                    <div className="text-small">{data.finished_at}</div>
                  </div>
                ) : (
                  "-"
                )}
              </FormGroup>
              <FormGroup isStatic label="Run ID">
                {has(data, "run_id") ? (
                  <Link
                    to={`/runs/${data.run_id}`}
                    style={{
                      textDecoration: "underline",
                      color: colors.gray.gray_3,
                    }}
                  >
                    {data.run_id}
                  </Link>
                ) : (
                  "-"
                )}
              </FormGroup>
              <FormGroup isStatic label="Task Definition ID">
                {has(data, "definition_id") ? (
                  <Link
                    to={`/tasks/${data.definition_id}`}
                    style={{
                      textDecoration: "underline",
                      color: colors.gray.gray_3,
                    }}
                  >
                    {data.definition_id}
                  </Link>
                ) : (
                  "-"
                )}
              </FormGroup>
              <FormGroup isStatic label="Image">
                {get(data, "image", "-")}
              </FormGroup>
              <FormGroup isStatic label="Task Arn">
                {get(data, "task_arn", "-")}
              </FormGroup>
              <FormGroup isStatic label="Instance ID">
                {get(data, "instance.instance_id", "-")}
              </FormGroup>
              <FormGroup isStatic label="Instance DNS Name">
                {get(data, "instance.dns_name", "-")}
              </FormGroup>
            </div>
          )
        }}
      </KeyValueContainer>
      <KeyValueContainer header="Environment Variables">
        {({ json, collapsed }) => {
          if (json) {
            return (
              <JsonView
                {...reactJsonViewProps}
                src={get(data, "env", []).reduce((acc, val) => {
                  acc[val.name] = val.value
                  return acc
                }, {})}
              />
            )
          }

          return (
            <div className="flot-detail-view-sidebar-card-content">
              {get(data, "env", []).map((env, i) => (
                <FormGroup
                  isStatic
                  label={
                    <span className="code" style={{ color: "white" }}>
                      {env.name}
                    </span>
                  }
                  key={`env-${i}`}
                >
                  <span className="code" style={{ wordBreak: "break-all" }}>
                    {env.value}
                  </span>
                </FormGroup>
              ))}
            </div>
          )
        }}
      </KeyValueContainer>
    </div>
  )
}
