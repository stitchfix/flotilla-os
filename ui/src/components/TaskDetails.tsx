import * as React from "react"
import { Link } from "react-router-dom"
import { Card, ButtonGroup, Pre, Classes } from "@blueprintjs/core"
import { TaskContext } from "./Task"
import Attribute from "./Attribute"
import TaskRuns from "./TaskRuns"
import ViewHeader from "./ViewHeader"
import EnvList from "./EnvList"
import DeleteTaskButton from "./DeleteTaskButton"
import { RequestStatus } from "./Request"

const TaskDetails: React.FunctionComponent = () => (
  <TaskContext.Consumer>
    {ctx => {
      if (ctx.requestStatus === RequestStatus.READY && ctx.data) {
        return (
          <>
            <ViewHeader
              breadcrumbs={[
                { text: "Tasks", href: "/tasks" },
                {
                  text: ctx.data.alias || ctx.definitionID,
                  href: `/tasks/${ctx.definitionID}`,
                },
              ]}
              buttons={
                <ButtonGroup>
                  <DeleteTaskButton definitionID={ctx.definitionID} />
                  <Link
                    className={Classes.BUTTON}
                    to={`/tasks/${ctx.definitionID}/copy`}
                  >
                    Copy
                  </Link>
                  <Link
                    className={Classes.BUTTON}
                    to={`/tasks/${ctx.definitionID}/update`}
                  >
                    Update
                  </Link>
                  <Link
                    className={Classes.BUTTON}
                    to={`/tasks/${ctx.definitionID}/execute`}
                  >
                    Run
                  </Link>
                </ButtonGroup>
              }
            />
            <div className="flotilla-sidebar-view-container">
              <div className="flotilla-sidebar-view-sidebar">
                <Card style={{ marginBottom: 12 }}>
                  <div className="flotilla-card-header">Attributes</div>
                  <div className="flotilla-attributes-container">
                    <Attribute name="Alias" value={ctx.data.alias} />
                    <Attribute
                      name="Definition ID"
                      value={ctx.data.definition_id}
                    />
                    <Attribute
                      name="Container Name"
                      value={ctx.data.container_name}
                    />
                    <Attribute name="Group Name" value={ctx.data.group_name} />
                    <Attribute name="Image" value={ctx.data.image} />
                    <Attribute
                      name="Command"
                      value={
                        <Pre className="flotilla-pre">{ctx.data.command}</Pre>
                      }
                    />
                    <Attribute name="Memory" value={ctx.data.memory} />
                    <Attribute name="CPU" value={ctx.data.cpu} />
                    <Attribute name="Arn" value={ctx.data.arn} />
                    <Attribute
                      name="Privileged"
                      value={ctx.data.privileged === true ? "Yes" : "No"}
                    />
                    <Attribute name="Tags" value={ctx.data.tags} />
                  </div>
                </Card>
                {ctx.data.env && ctx.data.env.length > 0 && (
                  <Card>
                    <div className="flotilla-card-header">
                      Environment Variables
                    </div>
                    <EnvList env={ctx.data.env} />
                  </Card>
                )}
              </div>
              <div className="flotilla-sidebar-view-content">
                <TaskRuns definitionID={ctx.definitionID} />
              </div>
            </div>
          </>
        )
      }
    }}
  </TaskContext.Consumer>
)

export default TaskDetails
