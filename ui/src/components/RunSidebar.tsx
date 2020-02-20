import * as React from "react"
import { get } from "lodash"
import { Card } from "@blueprintjs/core"
import { ExecutionEngine, Run, ExecutableType } from "../types"
import EnvList from "./EnvList"
import RunAttributes from "./RunAttributes"
import RunDebugAttributes from "./RunDebugAttributes"
import RecursiveKeyValueRenderer from "./RecursiveKeyValueRenderer"

const RunSidebar: React.FC<{ data: Run }> = ({ data }) => (
  <div className="flotilla-sidebar-view-sidebar">
    {data && data.executable_type === ExecutableType.ExecutableTypeTemplate && (
      <Card style={{ marginBottom: 12 }}>
        <RecursiveKeyValueRenderer
          data={get(data, ["execution_request_custom", "template_payload"], {})}
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

export default RunSidebar
