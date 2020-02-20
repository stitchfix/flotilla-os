import * as React from "react"
import { get } from "lodash"
import { Card, Button, Colors } from "@blueprintjs/core"
import JsonView from "react-json-view"
import { ExecutionEngine, Run, ExecutableType } from "../types"
import EnvList from "./EnvList"
import RunAttributes from "./RunAttributes"
import RunDebugAttributes from "./RunDebugAttributes"

const RunSidebar: React.FC<{ data: Run }> = ({ data }) => {
  const templatePayload = get(
    data,
    ["execution_request_custom", "template_payload"],
    {}
  )

  return (
    <div className="flotilla-sidebar-view-sidebar">
      {data && data.executable_type === ExecutableType.ExecutableTypeTemplate && (
        <Card style={{ marginBottom: 12 }}>
          <div className="flotilla-card-header-container">
            <div className="flotilla-card-header">Template Payload</div>
          </div>
          <JsonView
            src={templatePayload}
            name={false}
            collapsed={2}
            enableClipboard={false}
            displayDataTypes={false}
            displayObjectSize={false}
            theme="ocean"
            style={{
              background: Colors.DARK_GRAY1,
              fontFamily: "Roboto Mono",
              fontSize: "0.8rem",
            }}
          />
        </Card>
      )}
      <RunAttributes data={data} />
      <Card>
        <div className="flotilla-card-header-container">
          <div className="flotilla-card-header">Environment Variables</div>
        </div>
        <EnvList env={data.env} />
      </Card>
      {data && data.engine === ExecutionEngine.EKS && (
        <RunDebugAttributes data={data} />
      )}
    </div>
  )
}

export default RunSidebar
