import React, { Component } from "react"
import { Link } from "react-router-dom"
import JsonView from "react-json-view"
import { Card, FormGroup, Button, Tag, colors } from "aa-ui-components"
import { has, get } from "lodash"
import moment from "moment"
import { reactJsonViewProps } from "../constants/"
import EnhancedRunStatus from "./EnhancedRunStatus"
import KeyValueContainer from "./KeyValueContainer"

export default function RunInfo({ data }) {
  return (
    <div className="flot-detail-view-sidebar with-vertical-child-margin">
      <KeyValueContainer header="Run Info">
        {({ json, collapsed }) => {
          if (json) {
            return <JsonView {...reactJsonViewProps} src={data} />
          }

          return (
            <div className="flot-detail-view-sidebar-card-content">
              <FormGroup isStatic label="Status">
                <EnhancedRunStatus
                  status={get(data, "status", "")}
                  exitCode={get(data, "exit_code", "")}
                />
              </FormGroup>
              <FormGroup isStatic label="Exit Code">
                {get(data, "exit_code", "-")}
              </FormGroup>
              <FormGroup isStatic label="Started At">
                {has(data, "started_at") ? (
                  <div className="flex ff-rn j-fs a-bl with-horizontal-child-margin">
                    <div>{data.started_at}</div>
                    <div className="text-small">
                      {moment(data.started_at).fromNow()}
                    </div>
                  </div>
                ) : (
                  "-"
                )}
              </FormGroup>
              <FormGroup isStatic label="Finished At">
                {has(data, "finished_at") ? (
                  <div className="flex ff-rn j-fs a-bl with-horizontal-child-margin">
                    <div>{data.finished_at}</div>
                    <div className="text-small">
                      {moment(data.finished_at).fromNow()}
                    </div>
                  </div>
                ) : (
                  "-"
                )}
              </FormGroup>
              <FormGroup isStatic label="Run ID">
                {get(data, "run_id", "-")}
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
                  horizontal
                  isStatic
                  label={<Tag>{env.name}</Tag>}
                  key={`env-${i}`}
                >
                  <Tag>{env.value}</Tag>
                </FormGroup>
              ))}
            </div>
          )
        }}
      </KeyValueContainer>
    </div>
  )
}
