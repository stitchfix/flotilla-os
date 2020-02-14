import * as React from "react"
import { Card } from "@blueprintjs/core"
import { ExecutionEngine, Run } from "../types"
import EnvList from "./EnvList"
import RunAttributes from "./RunAttributes"
import RunDebugAttributes from "./RunDebugAttributes"

const RunSidebar: React.FC<{ data: Run }> = ({ data }) => (
  <div className="flotilla-sidebar-view-sidebar">
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
