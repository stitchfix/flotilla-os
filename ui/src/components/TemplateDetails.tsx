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
import { TemplateContext } from "./Template"
import Attribute from "./Attribute"
// import TemplateRuns from "./TemplateRuns"
import ViewHeader from "./ViewHeader"
import EnvList from "./EnvList"
import Toggler from "./Toggler"
import { RequestStatus } from "./Request"
import ErrorCallout from "./ErrorCallout"
import ARASwitch from "./ARASwitch"

const TemplateDetails: React.FC<{}> = () => (
  <TemplateContext.Consumer>
    {({ requestStatus, data, error, templateID, request }) => {
      switch (requestStatus) {
        case RequestStatus.ERROR:
          return <ErrorCallout error={error} />
        case RequestStatus.READY:
          if (data) {
            return (
              <>
                <ViewHeader
                  breadcrumbs={[
                    { text: "Templates", href: "/Templates" },
                    {
                      text:
                        `${data.template_name} v${data.version}` || templateID,
                      href: `/templates/${templateID}`,
                    },
                  ]}
                  buttons={
                    <ButtonGroup>
                      <Link
                        className={[
                          Classes.BUTTON,
                          Classes.INTENT_PRIMARY,
                        ].join(" ")}
                        to={`/templates/${templateID}/execute`}
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
                            <div className="flotilla-attributes-container flotilla-attributes-container-vertical">
                              <Attribute
                                name="Container Name"
                                value={data.container_name}
                              />
                              <Attribute name="Image" value={data.image} />
                              <Attribute name="CPU (Units)" value={data.cpu} />
                              <Attribute
                                name="Memory (MB)"
                                value={data.memory}
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
                    do me later
                    {/* <TaskRuns definitionID={definitionID} /> */}
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
  </TemplateContext.Consumer>
)
export default TemplateDetails
