import * as React from "react"
import { Link } from "react-router-dom"
import {
  Collapse,
  Card,
  ButtonGroup,
  Pre,
  Classes,
  Button,
  Spinner,
  Icon,
} from "@blueprintjs/core"
import { TaskContext } from "./Task"
import Attribute from "./Attribute"
import TaskRuns from "./TaskRuns"
import ViewHeader from "./ViewHeader"
import EnvList from "./EnvList"
import DeleteTaskButton from "./DeleteTaskButton"
import Toggler from "./Toggler"
import { RequestStatus } from "./Request"
import ErrorCallout from "./ErrorCallout"

const TaskDetails: React.FC<{}> = () => (
  <TaskContext.Consumer>
    {({ requestStatus, data, error, definitionID }) => {
      switch (requestStatus) {
        case RequestStatus.ERROR:
          return <ErrorCallout error={error} />
        case RequestStatus.READY:
          if (data) {
            return (
              <>
                <ViewHeader
                  breadcrumbs={[
                    { text: "Tasks", href: "/tasks" },
                    {
                      text: data.alias || definitionID,
                      href: `/tasks/${definitionID}`,
                    },
                  ]}
                  buttons={
                    <ButtonGroup>
                      <DeleteTaskButton definitionID={definitionID} />
                      <Link
                        className={Classes.BUTTON}
                        to={`/tasks/${definitionID}/copy`}
                      >
                        <div className="bp3-button-text">Copy</div>
                        <Icon icon="duplicate" />
                      </Link>
                      <Link
                        className={Classes.BUTTON}
                        to={`/tasks/${definitionID}/update`}
                      >
                        <div className="bp3-button-text">Update</div>
                        <Icon icon="edit" />
                      </Link>
                      <Link
                        className={Classes.BUTTON}
                        to={`/tasks/${definitionID}/execute`}
                      >
                        Run
                      </Link>
                    </ButtonGroup>
                  }
                />
                <div className="flotilla-sidebar-view-container">
                  <div className="flotilla-sidebar-view-sidebar">
                    <Toggler>
                      {({ isVisible, toggleVisibility }) => (
                        <Card style={{ marginBottom: 12 }}>
                          <div className="flotilla-card-header-container">
                            <div className="flotilla-card-header">
                              Attributes
                            </div>
                            <ButtonGroup>
                              <Button
                                small
                                onClick={toggleVisibility}
                                rightIcon={isVisible ? "minimize" : "maximize"}
                              >
                                {isVisible ? "Hide" : "Show"}
                              </Button>
                            </ButtonGroup>
                          </div>
                          <Collapse isOpen={isVisible}>
                            <div className="flotilla-attributes-container">
                              <Attribute name="Alias" value={data.alias} />
                              <Attribute
                                name="Definition ID"
                                value={data.definition_id}
                              />
                              <Attribute
                                name="Container Name"
                                value={data.container_name}
                              />
                              <Attribute
                                name="Group Name"
                                value={data.group_name}
                              />
                              <Attribute name="Image" value={data.image} />
                              <Attribute
                                name="Command"
                                value={
                                  <Pre className="flotilla-pre">
                                    {data.command}
                                  </Pre>
                                }
                              />
                              <Attribute name="Memory" value={data.memory} />
                              <Attribute name="CPU" value={data.cpu} />
                              <Attribute name="Arn" value={data.arn} />
                              <Attribute
                                name="Privileged"
                                value={data.privileged === true ? "Yes" : "No"}
                              />
                              <Attribute name="Tags" value={data.tags} />
                            </div>
                          </Collapse>
                        </Card>
                      )}
                    </Toggler>
                    {data.env && (
                      <Toggler>
                        {({ isVisible, toggleVisibility }) => (
                          <Card>
                            <div className="flotilla-card-header-container">
                              <div className="flotilla-card-header">
                                Environment Variables
                              </div>
                              <ButtonGroup>
                                <Button
                                  small
                                  onClick={toggleVisibility}
                                  rightIcon={
                                    isVisible ? "minimize" : "maximize"
                                  }
                                >
                                  {isVisible ? "Hide" : "Show"}
                                </Button>
                              </ButtonGroup>
                            </div>
                            <Collapse isOpen={isVisible}>
                              <EnvList env={data.env} />
                            </Collapse>
                          </Card>
                        )}
                      </Toggler>
                    )}
                  </div>
                  <div className="flotilla-sidebar-view-content">
                    <TaskRuns definitionID={definitionID} />
                  </div>
                </div>
              </>
            )
          }
          return null
        case RequestStatus.NOT_READY:
        default:
          return <Spinner />
      }
    }}
  </TaskContext.Consumer>
)
export default TaskDetails
